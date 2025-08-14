package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

type SchoolHandler struct{}

func NewSchoolHandler() *SchoolHandler { return &SchoolHandler{} }

type schoolPayload struct {
	SchoolCode     string `json:"school_code"`
	SchoolName     string `json:"school_name"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	EducationLevel string `json:"education_level"`
}

var (
	reCode      = regexp.MustCompile(`^[ก-๙A-Za-z0-9]{1,20}$`)
	reName      = regexp.MustCompile(`^[ก-๙0-9\s]{1,50}$`)
	reAddr      = regexp.MustCompile(`^[ก-๙0-9\s.,/]{1,100}$`)
	rePhone     = regexp.MustCompile(`^[0-9\- ]{1,10}$`)
	validLevels = map[string]bool{
		"อนุบาลศึกษา":    true,
		"ประถมศึกษา":     true,
		"มัธยมศึกษา":     true,
		"ทุกระดับการสอน": true,
	}
)

func validate(p schoolPayload) map[string]string {
	errs := map[string]string{}
	if !reCode.MatchString(strings.TrimSpace(p.SchoolCode)) {
		errs["school_code"] = "รูปแบบรหัสโรงเรียนไม่ถูกต้อง"
	}
	if !reName.MatchString(strings.TrimSpace(p.SchoolName)) {
		errs["school_name"] = "รูปแบบชื่อโรงเรียนไม่ถูกต้อง"
	}
	if !reAddr.MatchString(strings.TrimSpace(p.Address)) {
		errs["address"] = "รูปแบบที่อยู่ไม่ถูกต้อง"
	}
	if !rePhone.MatchString(strings.TrimSpace(p.Phone)) {
		errs["phone"] = "รูปแบบเบอร์โทรไม่ถูกต้อง"
	}
	if !validLevels[strings.TrimSpace(p.EducationLevel)] {
		errs["education_level"] = "กรุณาเลือกระดับการสอนให้ถูกต้อง"
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// GetSchool godoc
// @Summary      Get school (single)
// @Tags         school
// @Success      200 {object} models.School
// @Failure      404 {object} map[string]string
// @Router       /school [get]
func (h *SchoolHandler) Get(c echo.Context) error {
	var s models.School
	if err := database.DB.First(&s).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.JSON(http.StatusOK, s)
}

// CreateSchool godoc
// @Summary      Create school (only once)
// @Tags         school
// @Accept       json
// @Produce      json
// @Param        payload body schoolPayload true "school payload"
// @Success      201 {object} models.School
// @Failure      400 {object} map[string]any
// @Failure      409 {object} map[string]string
// @Router       /school [post]
func (h *SchoolHandler) Create(c echo.Context) error {
	// ถ้ามีอยู่แล้ว ไม่ให้เพิ่ม (ให้เหมือนฝั่ง UI ตอนนี้)
	var count int64
	database.DB.Model(&models.School{}).Count(&count)
	if count > 0 {
		return c.JSON(http.StatusConflict, map[string]string{"error": "EXISTS"})
	}

	var p schoolPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	if errs := validate(p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}

	s := models.School{
		SchoolCode:     strings.TrimSpace(p.SchoolCode),
		SchoolName:     strings.TrimSpace(p.SchoolName),
		Address:        strings.TrimSpace(p.Address),
		Phone:          strings.TrimSpace(p.Phone),
		EducationLevel: strings.TrimSpace(p.EducationLevel),
	}
	if err := database.DB.Create(&s).Error; err != nil {
		// อาจชน unique school_code
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, s)
}

// UpdateSchool godoc
// @Summary      Update school (the only record)
// @Tags         school
// @Accept       json
// @Produce      json
// @Param        payload body schoolPayload true "school payload"
// @Success      200 {object} models.School
// @Failure      404 {object} map[string]string
// @Failure      400 {object} map[string]any
// @Router       /school [put]
func (h *SchoolHandler) Update(c echo.Context) error {
	var current models.School
	if err := database.DB.First(&current).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}

	var p schoolPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	if errs := validate(p); errs != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}

	current.SchoolCode = strings.TrimSpace(p.SchoolCode)
	current.SchoolName = strings.TrimSpace(p.SchoolName)
	current.Address = strings.TrimSpace(p.Address)
	current.Phone = strings.TrimSpace(p.Phone)
	current.EducationLevel = strings.TrimSpace(p.EducationLevel)

	if err := database.DB.Save(&current).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, current)
}
