package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
	"gorm.io/gorm"
)

type StudentHandler struct{}

func NewStudentHandler() *StudentHandler { return &StudentHandler{} }

// ===== Validation rules (ให้ตรง AddStudentForm) =====
var (
	stuReNatID  = regexp.MustCompile(`^[0-9]{1,13}$`)         // ≤13
	stuReStuID  = regexp.MustCompile(`^[A-Za-z0-9\-]{1,20}$`) // เดี๋ยวจำกัดความยาวด้วยค่าจากโรงเรียน
	stuRePrefix = regexp.MustCompile(`^[ก-๙A-Za-z\.]{1,20}$`)
	stuReName   = regexp.MustCompile(`^[ก-๙A-Za-z\s]{1,50}$`)
	stuReRoom   = regexp.MustCompile(`^[0-9]{1,5}$`)
	stuRePhone  = regexp.MustCompile(`^[0-9\- ]{1,15}$`)
)

type studentPayload struct {
	NationalID     string `json:"national_id"`
	StudentID      string `json:"student_id"`
	Prefix         string `json:"prefix"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	BirthDate      string `json:"birth_date"`      // YYYY-MM-DD หรือว่าง
	EducationStage string `json:"education_stage"` // อนุบาล/ประถม/มัธยม
	Grade          string `json:"grade"`
	Room           string `json:"room"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	Status         string `json:"status"`
}

func (p *studentPayload) normalize() {
	trim := func(s string) string { return strings.TrimSpace(s) }
	p.NationalID = trim(p.NationalID)
	p.StudentID = trim(p.StudentID)
	p.Prefix = trim(p.Prefix)
	p.FirstName = strings.Join(strings.Fields(p.FirstName), " ")
	p.LastName = strings.Join(strings.Fields(p.LastName), " ")
	p.BirthDate = trim(p.BirthDate)
	p.EducationStage = trim(p.EducationStage)
	p.Grade = trim(p.Grade)
	p.Room = trim(p.Room)
	p.Address = trim(p.Address)
	p.Phone = trim(p.Phone)
	p.Status = trim(p.Status)
}

func stuDigitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// อ่านจำนวนหลักรหัสนักเรียนจาก school (ไม่มีข้อมูล → 0 = ไม่บังคับ)
func getStudentCodeLimit() int {
	type tmp struct {
		StudentCodeDigits int
	}
	var t tmp
	if err := database.DB.Table("schools").Select("student_code_digits").First(&t).Error; err != nil {
		return 0
	}
	if t.StudentCodeDigits < 0 {
		return 0
	}
	return t.StudentCodeDigits
}

func validateStudent(p *studentPayload) map[string]string {
	errs := map[string]string{}

	if !stuReNatID.MatchString(p.NationalID) {
		errs["national_id"] = "เลขบัตรประชาชนต้องเป็นตัวเลขไม่เกิน 13 หลัก"
	}
	if p.StudentID == "" || !stuReStuID.MatchString(p.StudentID) {
		errs["student_id"] = "รหัสนักเรียนไม่ถูกต้อง"
	} else {
		if lim := getStudentCodeLimit(); lim > 0 && len(p.StudentID) > lim {
			errs["student_id"] = fmt.Sprintf("รหัสนักเรียนต้องไม่เกิน %d ตัว", lim)
		}
	}
	if p.Prefix == "" || !stuRePrefix.MatchString(p.Prefix) {
		errs["prefix"] = "คำนำหน้าไม่ถูกต้อง"
	}
	if p.FirstName == "" || !stuReName.MatchString(p.FirstName) {
		errs["first_name"] = "ชื่อต้องเป็นตัวอักษร (ไทย/อังกฤษ)"
	}
	if p.LastName == "" || !stuReName.MatchString(p.LastName) {
		errs["last_name"] = "นามสกุลต้องเป็นตัวอักษร (ไทย/อังกฤษ)"
	}
	if p.BirthDate != "" {
		if _, err := time.Parse("2006-01-02", p.BirthDate); err != nil {
			errs["birth_date"] = "วันเกิดต้องเป็น YYYY-MM-DD หรือเว้นว่าง"
		}
	}
	validStages := map[string]bool{"อนุบาลศึกษา": true, "ประถมศึกษา": true, "มัธยมศึกษา": true}
	if !validStages[p.EducationStage] {
		errs["education_stage"] = "กรุณาเลือกระดับการศึกษา"
	}
	if strings.TrimSpace(p.Grade) == "" {
		errs["grade"] = "กรุณาเลือกชั้นเรียน"
	}
	if !stuReRoom.MatchString(p.Room) {
		errs["room"] = "ห้องต้องเป็นตัวเลข"
	}
	if strings.TrimSpace(p.Address) == "" {
		errs["address"] = "กรุณากรอกที่อยู่"
	}
	if !stuRePhone.MatchString(p.Phone) {
		errs["phone"] = "รูปแบบเบอร์โทรไม่ถูกต้อง"
	} else {
		d := stuDigitsOnly(p.Phone)
		if len(d) < 9 || len(d) > 10 {
			errs["phone"] = "เบอร์โทรต้องมี 9–10 หลัก"
		}
	}
	if strings.TrimSpace(p.Status) == "" {
		errs["status"] = "กรุณาเลือกสถานะ"
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ===== Handlers =====

func (h *StudentHandler) List(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	page := 1
	if v, err := strconv.Atoi(c.QueryParam("page")); err == nil && v > 0 {
		page = v
	}
	size := 20
	if v, err := strconv.Atoi(c.QueryParam("size")); err == nil {
		if v < 1 {
			size = 1
		} else if v > 100 {
			size = 100
		} else {
			size = v
		}
	}

	var items []models.Student
	tx := database.DB.Model(&models.Student{})

	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("student_id ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", like, like, like)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_COUNT_FAILED"})
	}
	if err := tx.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&items).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data":  items,
		"page":  page,
		"size":  size,
		"total": total,
	})
}

func (h *StudentHandler) Get(c echo.Context) error {
	id := c.Param("id")
	var s models.Student
	if err := database.DB.First(&s, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, s)
}

func (h *StudentHandler) Create(c echo.Context) error {
	var p studentPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	p.normalize()
	if errs := validateStudent(&p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}

	var birth *time.Time
	if p.BirthDate != "" {
		if b, err := time.Parse("2006-01-02", p.BirthDate); err == nil {
			birth = &b
		}
	}
	s := models.Student{
		NationalID: p.NationalID, StudentID: p.StudentID, Prefix: p.Prefix,
		FirstName: p.FirstName, LastName: p.LastName, BirthDate: birth,
		Education: p.EducationStage, Grade: p.Grade, Room: p.Room,
		Address: p.Address, Phone: p.Phone, Status: p.Status,
	}
	if err := database.DB.Create(&s).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, s)
}

func (h *StudentHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var existing models.Student
	if err := database.DB.First(&existing, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	var p studentPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	p.normalize()
	if errs := validateStudent(&p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}
	if p.BirthDate != "" {
		if b, err := time.Parse("2006-01-02", p.BirthDate); err == nil {
			existing.BirthDate = &b
		}
	}
	existing.NationalID = p.NationalID
	existing.StudentID = p.StudentID
	existing.Prefix = p.Prefix
	existing.FirstName = p.FirstName
	existing.LastName = p.LastName
	existing.Education = p.EducationStage
	existing.Grade = p.Grade
	existing.Room = p.Room
	existing.Address = p.Address
	existing.Phone = p.Phone
	existing.Status = p.Status

	if err := database.DB.Save(&existing).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, existing)
}

func (h *StudentHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Student{}, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *StudentHandler) Import(c echo.Context) error {
	var arr []studentPayload
	if err := c.Bind(&arr); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	var inserted []models.Student
	errFields := []map[string]any{}

	for i, p := range arr {
		p.normalize()
		if errs := validateStudent(&p); errs != nil {
			errFields = append(errFields, map[string]any{"index": i, "fields": errs})
			continue
		}
		var birth *time.Time
		if p.BirthDate != "" {
			if b, err := time.Parse("2006-01-02", p.BirthDate); err == nil {
				birth = &b
			}
		}
		inserted = append(inserted, models.Student{
			NationalID: p.NationalID, StudentID: p.StudentID, Prefix: p.Prefix,
			FirstName: p.FirstName, LastName: p.LastName, BirthDate: birth,
			Education: p.EducationStage, Grade: p.Grade, Room: p.Room,
			Address: p.Address, Phone: p.Phone, Status: p.Status,
		})
	}
	if len(errFields) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":  "BULK_VALIDATION_ERROR",
			"issues": errFields,
		})
	}
	if err := database.DB.Create(&inserted).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"inserted": len(inserted)})
}
