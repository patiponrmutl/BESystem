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

/*** Validation rules ***/
var validPrefixes = map[string]bool{
	"นาย": true, "นาง": true, "นางสาว": true, "ว่าที่ รต.": true, "ดร.": true,
}

// ✅ ใช้ชื่อไม่ชนไฟล์อื่น
var (
	tchReCode  = regexp.MustCompile(`^[A-Za-z0-9]{1,20}$`)
	tchReName  = regexp.MustCompile(`^[A-Za-zก-๙\- ]{1,50}$`)
	tchReEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	tchRePhone = regexp.MustCompile(`^[0-9\- ]{1,15}$`)
)

type TeacherHandler struct{}

func NewTeacherHandler() *TeacherHandler { return &TeacherHandler{} }

type teacherPayload struct {
	TeacherCode string `json:"teacher_code"`
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Position    string `json:"position"`
}

func (p *teacherPayload) norm() {
	trim := func(s string) string { return strings.TrimSpace(s) }
	p.TeacherCode = trim(p.TeacherCode)
	p.Prefix = trim(p.Prefix)
	p.FirstName = strings.Join(strings.Fields(p.FirstName), " ")
	p.LastName = strings.Join(strings.Fields(p.LastName), " ")
	p.Phone = trim(p.Phone)
	p.Email = strings.ToLower(trim(p.Email))
	p.Position = strings.Join(strings.Fields(p.Position), " ")
}

func tchOnlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func validateTeacher(p *teacherPayload) map[string]string {
	errs := map[string]string{}
	if p.TeacherCode == "" || !tchReCode.MatchString(p.TeacherCode) {
		errs["teacher_code"] = "รหัสครูต้องเป็น A–Z/a–z/0–9 ไม่เกิน 20 ตัว"
	}
	if p.Prefix == "" || !validPrefixes[p.Prefix] {
		errs["prefix"] = "คำนำหน้าไม่ถูกต้อง"
	}
	if p.FirstName == "" || !tchReName.MatchString(p.FirstName) {
		errs["first_name"] = "ชื่อรองรับไทย/อังกฤษ เว้นวรรค/ขีด (≤50)"
	}
	if p.LastName == "" || !tchReName.MatchString(p.LastName) {
		errs["last_name"] = "นามสกุลองรับไทย/อังกฤษ เว้นวรรค/ขีด (≤50)"
	}
	d := tchOnlyDigits(p.Phone)
	if d == "" || len(d) < 9 || len(d) > 10 {
		errs["phone"] = "เบอร์โทรต้องมี 9–10 หลัก (ใส่ขีด/ช่องว่างได้)"
	}
	if p.Email == "" || len(p.Email) > 50 || !tchReEmail.MatchString(p.Email) {
		errs["email"] = "รูปแบบอีเมลไม่ถูกต้อง (≤50 ตัวอักษร)"
	}
	if p.Position == "" || len(p.Position) > 50 {
		errs["position"] = "ตำแหน่งต้องไม่เกิน 50 ตัวอักษร"
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

/*** CRUD ***/

// GET /teachers?q=&page=&size=
func (h *TeacherHandler) List(c echo.Context) error {
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

	var items []models.Teacher
	tx := database.DB.Model(&models.Teacher{})
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where(`teacher_code ILIKE ? OR prefix ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR position ILIKE ? OR email ILIKE ?`,
			like, like, like, like, like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_COUNT_FAILED"})
	}
	if err := tx.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&items).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items, "page": page, "size": size, "total": total})
}

func (h *TeacherHandler) Get(c echo.Context) error {
	id := c.Param("id")
	var t models.Teacher
	if err := database.DB.First(&t, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, t)
}

func (h *TeacherHandler) Create(c echo.Context) error {
	var p teacherPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	p.norm()
	if errs := validateTeacher(&p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}

	// กันซ้ำ (รหัส/อีเมล)
	var cnt int64
	database.DB.Model(&models.Teacher{}).
		Where("teacher_code = ? OR LOWER(email) = ?", p.TeacherCode, p.Email).
		Count(&cnt)
	if cnt > 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "DUP_CODE_OR_EMAIL"})
	}

	t := models.Teacher{
		TeacherCode: p.TeacherCode, Prefix: p.Prefix,
		FirstName: p.FirstName, LastName: p.LastName,
		Phone: p.Phone, Email: p.Email, Position: p.Position,
	}
	if err := database.DB.Create(&t).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, t)
}

func (h *TeacherHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var cur models.Teacher
	if err := database.DB.First(&cur, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	var p teacherPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	p.norm()
	if errs := validateTeacher(&p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}

	// กันซ้ำรหัส/อีเมลกับคนอื่น
	var cnt int64
	database.DB.Model(&models.Teacher{}).
		Where("(teacher_code = ? OR LOWER(email) = ?) AND id <> ?", p.TeacherCode, p.Email, cur.ID).
		Count(&cnt)
	if cnt > 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "DUP_CODE_OR_EMAIL"})
	}

	cur.TeacherCode = p.TeacherCode
	cur.Prefix = p.Prefix
	cur.FirstName = p.FirstName
	cur.LastName = p.LastName
	cur.Phone = p.Phone
	cur.Email = p.Email
	cur.Position = p.Position

	if err := database.DB.Save(&cur).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, cur)
}

func (h *TeacherHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Teacher{}, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// Import หลายรายการ
func (h *TeacherHandler) Import(c echo.Context) error {
	var rows []teacherPayload
	if err := c.Bind(&rows); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	errs := []map[string]any{}
	insert := make([]models.Teacher, 0, len(rows))

	// ซ้ำกับ DB เดิม
	var existed []models.Teacher
	database.DB.Find(&existed)
	dupCode := map[string]bool{}
	dupEmail := map[string]bool{}
	for _, t := range existed {
		dupCode[strings.TrimSpace(t.TeacherCode)] = true
		dupEmail[strings.ToLower(strings.TrimSpace(t.Email))] = true
	}

	// นับซ้ำภายในไฟล์
	countCode := map[string]int{}
	countEmail := map[string]int{}
	for _, r := range rows {
		code := strings.TrimSpace(r.TeacherCode)
		mail := strings.ToLower(strings.TrimSpace(r.Email))
		if code != "" {
			countCode[code]++
		}
		if mail != "" {
			countEmail[mail]++
		}
	}

	for i, r := range rows {
		r.norm()
		e := validateTeacher(&r)
		if r.TeacherCode != "" && countCode[r.TeacherCode] > 1 {
			if e == nil {
				e = map[string]string{}
			}
			e["teacher_code"] = "รหัสครูซ้ำในไฟล์"
		}
		if r.Email != "" && countEmail[r.Email] > 1 {
			if e == nil {
				e = map[string]string{}
			}
			e["email"] = "อีเมลซ้ำในไฟล์"
		}
		if dupCode[r.TeacherCode] {
			if e == nil {
				e = map[string]string{}
			}
			e["teacher_code"] = "รหัสครูซ้ำกับข้อมูลเดิม"
		}
		if dupEmail[r.Email] {
			if e == nil {
				e = map[string]string{}
			}
			e["email"] = "อีเมลซ้ำกับข้อมูลเดิม"
		}

		if e != nil {
			errs = append(errs, map[string]any{"index": i, "fields": e})
			continue
		}

		insert = append(insert, models.Teacher{
			TeacherCode: r.TeacherCode, Prefix: r.Prefix,
			FirstName: r.FirstName, LastName: r.LastName,
			Phone: r.Phone, Email: r.Email, Position: r.Position,
		})
	}
	if len(errs) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":  "BULK_VALIDATION_ERROR",
			"issues": errs,
		})
	}
	if err := database.DB.Create(&insert).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"inserted": len(insert)})
}
