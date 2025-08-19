package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
	"gorm.io/gorm"
)

// ====== ค่าคงที่ให้ตรงกับฟอร์ม FE ======
var (
	hmReYear = regexp.MustCompile(`^[0-9]{4}$`)   // พ.ศ. 4 หลัก (เช่น 2568)
	hmReRoom = regexp.MustCompile(`^[0-9]{1,3}$`) // ห้องเป็นตัวเลข ≤ 3 หลัก

	Stages = map[string]bool{
		"อนุบาลศึกษา": true, "ประถมศึกษา": true, "มัธยมศึกษา": true,
	}
	Positions = map[string]bool{
		"ครูประจำชั้นหลัก": true, "ครูประจำชั้นรอง": true,
	}
	StatusOptions = map[string]bool{
		"ปฏิบัติงาน": true, "เลิกจ้าง": true, "เปลี่ยนครูประจำชั้นหลัก": true, "เปลี่ยนครูประจำชั้นรอง": true,
	}
)

type HomeroomHandler struct{}

func NewHomeroomHandler() *HomeroomHandler { return &HomeroomHandler{} }

type homeroomPayload struct {
	AcademicYear   string `json:"academic_year"`   // พ.ศ. (string)
	EducationStage string `json:"education_stage"` // อนุบาล/ประถม/มัธยม
	Grade          string `json:"grade"`           // เช่น "ประถม 6"
	Room           string `json:"room"`            // ตัวเลขล้วน
	Position       string `json:"position"`        // หลัก/รอง
	TeacherID      any    `json:"teacher_id"`      // FE อาจส่งเป็น number หรือ string
	Status         string `json:"status"`
	Note           string `json:"note"`
}

func (p *homeroomPayload) norm() {
	trim := func(s string) string { return strings.TrimSpace(s) }
	p.AcademicYear = trim(p.AcademicYear)
	p.EducationStage = trim(p.EducationStage)
	p.Grade = trim(p.Grade)
	p.Room = trim(p.Room)
	p.Position = trim(p.Position)
	p.Status = trim(p.Status)
	p.Note = trim(p.Note)
}

// รับ teacher_id เป็น id หรือ teacher_code ก็ได้
func resolveTeacherID(v any) (uint, bool) {
	switch t := v.(type) {
	case float64:
		if t <= 0 {
			return 0, false
		}
		return uint(t), true
	case int:
		if t <= 0 {
			return 0, false
		}
		return uint(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		// ลอง parse เป็นตัวเลขก่อน
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return uint(n), true
		}
		// ไม่ใช่ตัวเลข → ลองหาโดย teacher_code
		var teacher models.Teacher
		if err := database.DB.First(&teacher, "teacher_code = ?", s).Error; err == nil && teacher.ID > 0 {
			return teacher.ID, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func validateHomeroom(p *homeroomPayload) map[string]string {
	errs := map[string]string{}
	if !hmReYear.MatchString(p.AcademicYear) {
		errs["academic_year"] = "ปีการศึกษาต้องเป็น พ.ศ. 4 หลัก"
	}
	if !Stages[p.EducationStage] {
		errs["education_stage"] = "เลือกระดับการศึกษาไม่ถูกต้อง"
	}
	if p.Grade == "" {
		errs["grade"] = "กรุณาเลือกชั้น"
	}
	if !hmReRoom.MatchString(p.Room) {
		errs["room"] = "ห้องต้องเป็นตัวเลขไม่เกิน 3 หลัก"
	}
	if !Positions[p.Position] {
		errs["position"] = "กรุณาเลือกตำแหน่งให้ถูกต้อง"
	}
	if !StatusOptions[p.Status] {
		errs["status"] = "กรุณาเลือกสถานะให้ถูกต้อง"
	}
	// note: optional (จำกัด 255 ใน Model)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ========== List ==========
func (h *HomeroomHandler) List(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	page, size := 1, 20
	if v, err := strconv.Atoi(c.QueryParam("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.QueryParam("size")); err == nil {
		if v < 1 {
			size = 1
		} else if v > 100 {
			size = 100
		} else {
			size = v
		}
	}

	var items []models.Homeroom
	tx := database.DB.Model(&models.Homeroom{})
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where(`
			academic_year ILIKE ? OR education_stage ILIKE ? OR grade ILIKE ? OR room ILIKE ? OR position ILIKE ? OR status ILIKE ?
		`, like, like, like, like, like, like)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_COUNT_FAILED"})
	}
	if err := tx.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&items).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}

	// เติม teacher_name (optional)
	if len(items) > 0 {
		var ts []models.Teacher
		if err := database.DB.Find(&ts).Error; err == nil {
			name := map[uint]string{}
			for _, t := range ts {
				full := strings.TrimSpace((t.Prefix + " " + t.FirstName + " " + t.LastName))
				name[t.ID] = full
			}
			type out struct {
				models.Homeroom
				TeacherName string `json:"teacher_name"`
			}
			resp := make([]out, 0, len(items))
			for _, r := range items {
				resp = append(resp, out{Homeroom: r, TeacherName: name[r.TeacherID]})
			}
			return c.JSON(http.StatusOK, map[string]any{"data": resp, "page": page, "size": size, "total": total})
		}
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items, "page": page, "size": size, "total": total})
}

// ========== Get ==========
func (h *HomeroomHandler) Get(c echo.Context) error {
	id := c.Param("id")
	var r models.Homeroom
	if err := database.DB.First(&r, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, r)
}

// ========== Create ==========
func (h *HomeroomHandler) Create(c echo.Context) error {
	var p homeroomPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	p.norm()
	if errs := validateHomeroom(&p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}
	tid, ok := resolveTeacherID(p.TeacherID)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": map[string]string{"teacher_id": "ไม่ถูกต้อง"}})
	}

	// Rule 1: กันซ้ำ ปี/ระดับ/ชั้น/ห้อง/ตำแหน่ง
	var cnt int64
	database.DB.Model(&models.Homeroom{}).
		Where("academic_year = ? AND education_stage = ? AND grade = ? AND room = ? AND position = ?",
			p.AcademicYear, p.EducationStage, p.Grade, p.Room, p.Position).
		Count(&cnt)
	if cnt > 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "DUP_CLASS_POSITION"})
	}

	// Rule 2: ถ้าเป็น "หลัก" ครูคนนี้ห้ามเป็นหลักที่อื่น
	if p.Position == "ครูประจำชั้นหลัก" {
		var used int64
		database.DB.Model(&models.Homeroom{}).
			Where("position = ? AND teacher_id = ?", "ครูประจำชั้นหลัก", tid).
			Count(&used)
		if used > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "TEACHER_ALREADY_MAIN"})
		}
	}

	r := models.Homeroom{
		AcademicYear:   p.AcademicYear,
		EducationStage: p.EducationStage,
		Grade:          p.Grade,
		Room:           p.Room,
		Position:       p.Position,
		TeacherID:      tid,
		Status:         p.Status,
		Note:           p.Note,
	}
	if err := database.DB.Create(&r).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, r)
}

// ========== Update ==========
func (h *HomeroomHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var cur models.Homeroom
	if err := database.DB.First(&cur, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}

	var p homeroomPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	p.norm()
	if errs := validateHomeroom(&p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}
	tid, ok := resolveTeacherID(p.TeacherID)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": map[string]string{"teacher_id": "ไม่ถูกต้อง"}})
	}

	// Rule 1: กันซ้ำ ปี/ระดับ/ชั้น/ห้อง/ตำแหน่ง (ยกเว้นตัวเอง)
	var cnt int64
	database.DB.Model(&models.Homeroom{}).
		Where("academic_year = ? AND education_stage = ? AND grade = ? AND room = ? AND position = ? AND id <> ?",
			p.AcademicYear, p.EducationStage, p.Grade, p.Room, p.Position, cur.ID).
		Count(&cnt)
	if cnt > 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "DUP_CLASS_POSITION"})
	}

	// Rule 2: ถ้าเป็น "หลัก" ครูคนนี้ห้ามเป็นหลักที่อื่น (ยกเว้นตัวเอง)
	if p.Position == "ครูประจำชั้นหลัก" {
		var used int64
		database.DB.Model(&models.Homeroom{}).
			Where("position = ? AND teacher_id = ? AND id <> ?", "ครูประจำชั้นหลัก", tid, cur.ID).
			Count(&used)
		if used > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "TEACHER_ALREADY_MAIN"})
		}
	}

	cur.AcademicYear = p.AcademicYear
	cur.EducationStage = p.EducationStage
	cur.Grade = p.Grade
	cur.Room = p.Room
	cur.Position = p.Position
	cur.TeacherID = tid
	cur.Status = p.Status
	cur.Note = p.Note

	if err := database.DB.Save(&cur).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, cur)
}

// ========== Delete ==========
func (h *HomeroomHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Homeroom{}, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
