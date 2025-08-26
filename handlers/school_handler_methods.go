package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

// NOTE: ไม่ประกาศ type SchoolHandler ใหม่ (มีอยู่แล้วในโปรเจกต์ของคุณ)
// แค่เติมเมธอดที่ routes.go เรียกให้ครบ

// GET /school
func (h *SchoolHandler) GetSchool(c echo.Context) error {
	var s models.School
	if err := database.DB.Order("id ASC").First(&s).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{"error": "NOT_FOUND"})
	}
	return c.JSON(http.StatusOK, s)
}

// POST /school (create or update แถวแรก)
func (h *SchoolHandler) CreateOrUpdate(c echo.Context) error {
	var in models.School
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}

	var s models.School
	tx := database.DB
	if err := tx.Order("id ASC").First(&s).Error; err != nil {
		// ไม่มีข้อมูล → create
		if err := tx.Create(&in).Error; err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, in)
	}

	// มีอยู่แล้ว → update เฉพาะฟิลด์ที่ส่งมา
	if err := tx.Model(&s).Updates(in).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, s)
}

// DELETE /school (ลบแถวแรก)
func (h *SchoolHandler) DeleteSchool(c echo.Context) error {
	var s models.School
	if err := database.DB.Order("id ASC").First(&s).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{"error": "NOT_FOUND"})
	}
	if err := database.DB.Delete(&s).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// (ทางเลือก) สำหรับ routes ที่เรียก GetSchoolAdmin
func (h *SchoolHandler) GetSchoolAdmin(c echo.Context) error {
	// reuse จาก GetSchool ไปก่อน
	return h.GetSchool(c)
}
