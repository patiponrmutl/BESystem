package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler { return &DashboardHandler{} }

// GET /teacher/dashboard/daily?date=YYYY-MM-DD&classroom=1/1
// คืนรูปแบบที่ FE ใช้ในหน้า Dashboard:
// { holiday: { isHoliday: bool, name: string }, rows: [{ id, student_id, status, time, note, operator, retro }] }
func (h *DashboardHandler) Daily(c echo.Context) error {
	date := strings.TrimSpace(c.QueryParam("date"))
	classroom := strings.TrimSpace(c.QueryParam("classroom")) // "ชั้น/ห้อง" เช่น "1/1" (อาจว่าง)

	if date == "" {
		// default: วันนี้ (เขตเวลาของเครื่องรัน)
		date = time.Now().Format("2006-01-02")
	}

	// 1) ตรวจวันหยุด (optional ถ้าคุณมีตารางวันหยุด)
	holiday := map[string]any{"isHoliday": false, "name": ""}
	type holidayRow struct {
		Name     string
		DateFrom string // YYYY-MM-DD
		DateTo   string // YYYY-MM-DD
	}
	var hRows []holidayRow
	// ตัวอย่าง: ถ้าคุณเก็บวันหยุดในตาราง calendar_holidays (แก้ชื่อตาราง/ฟิลด์ให้ตรงโปรเจกต์จริง)
	_ = database.DB.
		Table("calendar_holidays").
		Select("name, date_from, date_to").
		Where("? BETWEEN date_from AND date_to", date).
		Scan(&hRows)
	if len(hRows) > 0 {
		holiday["isHoliday"] = true
		holiday["name"] = strings.TrimSpace(hRows[0].Name)
	}

	// 2) โหลด attendance ของวันนั้น (อาจกรองตาม classroom)
	// ถ้าส่ง classroom มา → join students เพื่อกรอง grade/room
	type row struct {
		ID        any    `json:"id"`
		StudentID uint   `json:"student_id"`
		Status    string `json:"status"`
		Time      string `json:"time"`
		Note      string `json:"note"`
		Operator  string `json:"operator"`
		Retro     bool   `json:"retro"`
		StudentNm string `json:"student_name"` // FE เผื่อใช้
	}

	tx := database.DB.Table("attendances AS a").
		Select("a.id, a.student_id, a.status, COALESCE(a.time,'—') AS time, COALESCE(a.note,'') AS note, '' AS operator, false AS retro")

	if classroom != "" {
		// classroom = "<grade>/<room>" → ดึงเฉพาะห้องนี้
		var grade, room string
		if m := parseClassroom(classroom); m != nil {
			grade, room = m[0], m[1]
			tx = tx.Joins("JOIN students s ON s.id = a.student_id").
				Where("s.grade = ? AND s.room = ?", grade, room)
		}
	}

	tx = tx.Where("a.date = ?", date)

	var rows []row
	if err := tx.Order("a.student_id ASC, a.time ASC, a.id ASC").Scan(&rows).Error; err != nil && err != gorm.ErrRecordNotFound {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// 3) เติมใบลาที่ "อนุมัติแล้ว" ให้กลายเป็นสถานะ "ลา" ของวันนี้
	var leaves []struct {
		ID        uint
		StudentID uint
		Type      string
		DateFrom  string
		DateTo    string
		Status    string
	}
	_ = database.DB.Table("leave_requests").
		Select("id, student_id, type, date_from, date_to, status").
		Where("? BETWEEN date_from AND date_to", date).
		Where("status = ?", "อนุมัติ").
		Scan(&leaves)

	// แปลง leave → แถว "ลา"
	leaveMap := map[uint]row{}
	for _, lv := range leaves {
		note := ""
		switch strings.TrimSpace(lv.Type) {
		case "ลาป่วย", "ป่วย":
			note = "ป่วย"
		case "ลากิจ", "ธุระส่วนตัว":
			note = "ธุระส่วนตัว"
		default:
			note = "" // อื่นๆ
		}
		leaveMap[lv.StudentID] = row{
			ID:        "leave-" + date + "-" + itoa(lv.ID),
			StudentID: lv.StudentID,
			Status:    "ลา",
			Time:      "—",
			Note:      note,
			Operator:  "",
			Retro:     false,
		}
	}

	// รวม attendance + leave (เอา leave ทับสถานะอื่นของวันนั้น)
	latestByStu := map[uint]row{}
	for _, r := range rows {
		latestByStu[r.StudentID] = r
	}
	for sid, lr := range leaveMap {
		latestByStu[sid] = lr
	}

	out := make([]row, 0, len(latestByStu))
	for _, r := range latestByStu {
		out = append(out, r)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"holiday": holiday,
		"rows":    out,
	})
}

// parse "1/1" → ["1", "1"], คืน nil ถ้า parse ไม่ได้
func parseClassroom(code string) []string {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	// รูปแบบเลข/เลข เท่านั้น
	var g, r string
	n, _ := fmtSscanf(code, "%d/%d", &g, &r) // ใช้ shim เดิม
	if n == 2 && g != "" && r != "" {
		return []string{g, r}
	}
	// สำรองแบบมีเว้นวรรค
	parts := strings.Split(code, "/")
	if len(parts) != 2 {
		return nil
	}
	return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
}

// แปลง uint → string แบบง่าย
func itoa(u uint) string {
	return fmtUint(u)
}

// GET /dashboard/summary
// คืนค่าจำนวนคร่าว ๆ สำหรับหน้าแดชบอร์ด
func (h *DashboardHandler) Summary(c echo.Context) error {
	var (
		cntStudents int64
		cntTeachers int64
		cntRooms    int64
		cntLeaves   int64
	)

	database.DB.Model(&models.Student{}).Count(&cntStudents)
	database.DB.Model(&models.Teacher{}).Count(&cntTeachers)
	database.DB.Model(&models.Homeroom{}).Count(&cntRooms)
	database.DB.Model(&models.LeaveRequest{}).Where("status = ?", "pending").Count(&cntLeaves)

	return c.JSON(http.StatusOK, map[string]any{
		"students":       cntStudents,
		"teachers":       cntTeachers,
		"homerooms":      cntRooms,
		"pending_leaves": cntLeaves,
	})
}
