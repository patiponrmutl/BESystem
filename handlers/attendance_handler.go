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

type AttendanceHandler struct{}

func NewAttendanceHandler() *AttendanceHandler { return &AttendanceHandler{} }

// GET /teacher/attendance?start=YYYY-MM-DD&end=YYYY-MM-DD&studentId=&statuses=เข้า,ออก,มาสาย,ขาด,ลา
// optional: grade, room, q
func (h *AttendanceHandler) List(c echo.Context) error {
	start := strings.TrimSpace(c.QueryParam("start"))
	end := strings.TrimSpace(c.QueryParam("end"))
	studentID := strings.TrimSpace(c.QueryParam("studentId"))
	statuses := strings.TrimSpace(c.QueryParam("statuses"))
	grade := strings.TrimSpace(c.QueryParam("grade"))
	room := strings.TrimSpace(c.QueryParam("room"))
	q := strings.TrimSpace(c.QueryParam("q"))

	tx := database.DB.Model(&models.Attendance{})

	if start != "" && end != "" {
		tx = tx.Where(`date >= ? AND date <= ?`, start, end)
	}
	if studentID != "" {
		tx = tx.Where("student_id = ?", studentID)
	}
	if statuses != "" {
		parts := splitCSV(statuses)
		if len(parts) > 0 {
			tx = tx.Where("status IN ?", parts)
		}
	}

	// join students เพื่อ filter grade/room หรือค้นชื่อ
	if grade != "" || room != "" || q != "" {
		tx = tx.Joins("JOIN students s ON s.id = attendances.student_id")
		if grade != "" {
			tx = tx.Where("s.grade = ?", grade)
		}
		if room != "" {
			tx = tx.Where("s.room = ?", room)
		}
		if q != "" {
			like := "%" + strings.ToLower(q) + "%"
			tx = tx.Where("LOWER(s.student_id) LIKE ? OR LOWER(s.first_name) LIKE ? OR LOWER(s.last_name) LIKE ?",
				like, like, like)
		}
	}

	var rows []models.Attendance
	if err := tx.Order("date ASC, time ASC, id ASC").Find(&rows).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusOK, []models.Attendance{})
		}
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, rows)
}

func (h *AttendanceHandler) Mark(c echo.Context) error {
	var req struct {
		StudentID uint   `json:"student_id"`
		Date      string `json:"date"`
		Status    string `json:"status"`
		Note      string `json:"note"`
		Operator  string `json:"operator"` // เก็บลง Note เพิ่ม/หรือช่องแยก ถ้ามีคอลัมน์
		Retro     bool   `json:"retro"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}
	if req.StudentID == 0 || req.Date == "" || req.Status == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "MISSING_FIELDS"})
	}

	// map "ลากิจ/ลาป่วย" → เก็บเป็น "ลา" + note แยก (ให้เข้ากับ FE ที่รวมเป็น 'ลา')
	status := strings.TrimSpace(req.Status)
	note := strings.TrimSpace(req.Note)
	if status == "ลากิจ" {
		status = "ลา"
		if note == "" {
			note = "ธุระส่วนตัว"
		}
	} else if status == "ลาป่วย" {
		status = "ลา"
		if note == "" {
			note = "ป่วย"
		}
	}

	// ออกแบบ: 1 วัน/นักเรียน อนุญาตหลายแถว (เข้า/ออก) → ที่ Dashboard เราจะดึง “ล่าสุด” อยู่แล้ว
	rec := models.Attendance{
		StudentID: req.StudentID,
		Date:      req.Date,
		Time:      time.Now().Format("15:04"),
		Status:    status,
		Note:      note,
	}
	if status == "ขาด" || status == "ลา" || status == "ยังไม่เข้าโรงเรียน" {
		rec.Time = "—"
	}

	if err := database.DB.Create(&rec).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// ตอบกลับรูปแบบที่หน้า Dashboard ใช้
	out := map[string]any{
		"id":         rec.ID,
		"student_id": rec.StudentID,
		"status":     rec.Status,
		"time":       rec.Time,
		"note":       rec.Note,
		"operator":   strings.TrimSpace(req.Operator),
		"retro":      req.Retro,
	}
	return c.JSON(http.StatusOK, out)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
