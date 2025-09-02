package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

// NOTE: ไม่ประกาศ type SchoolHandler ใหม่ (มีอยู่แล้วในโปรเจกต์ของคุณ)
// แค่เติมเมธอดที่ routes.go เรียกให้ครบ

// GET /school
func (h *SchoolHandler) GetSchool(c echo.Context) error {
	var s models.School
	// ใช้ตัวแรก (มี 1 row)
	if err := database.DB.Order("id asc").First(&s).Error; err != nil {
		// ไม่มีข้อมูล → 404 แบบอ่านง่าย
		return c.JSON(http.StatusNotFound, map[string]any{
			"error": "NOT_FOUND",
		})
	}

	resp := SchoolResp{
		ID:         s.ID,
		SchoolCode: s.SchoolCode,
		SchoolName: s.SchoolName,
		Address:    s.Address,
		Phone:      s.Phone,
		Education:  s.EducationLevel,
		CodeLengths: CodeLengthsPayload{
			TeacherCodeDigits: s.TeacherCodeDigits, // <— มาจากคอลัมน์เดิม
			StudentCodeDigits: s.StudentCodeDigits, // <— มาจากคอลัมน์เดิม
		},
	}
	return c.JSON(http.StatusOK, resp)
}

// POST /school (create or update แถวแรก)
func (h *SchoolHandler) CreateOrUpdate(c echo.Context) error {
	var in SchoolUpsertReq
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}

	// ตรวจฟิลด์หลักแบบสั้น ๆ (คุณมีชุดตรวจของเดิมอยู่แล้วก็ใช้ต่อได้)
	if strings.TrimSpace(in.SchoolCode) == "" || strings.TrimSpace(in.SchoolName) == "" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{
			"error": "VALIDATION_ERROR",
			"fields": map[string]string{
				"school_code": "required",
				"school_name": "required",
			},
		})
	}

	// หา row แรก ถ้าไม่มีจะสร้างใหม่
	var s models.School
	err := database.DB.Order("id asc").First(&s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_ERROR"})
	}

	// อัปเดตฟิลด์หลัก
	s.SchoolCode = in.SchoolCode
	s.SchoolName = in.SchoolName
	s.Address = in.Address
	s.Phone = in.Phone
	s.EducationLevel = in.Education

	// >>> รับค่าความยาวรหัส (รองรับทั้งซ้อนและ flat)
	if in.CodeLengths != nil {
		s.TeacherCodeDigits = max0(in.CodeLengths.TeacherCodeDigits)
		s.StudentCodeDigits = max0(in.CodeLengths.StudentCodeDigits)
	} else {
		// เผื่อส่งแบบ flat มา
		if in.TeacherCodeDigits != nil {
			s.TeacherCodeDigits = max0(*in.TeacherCodeDigits)
		}
		if in.StudentCodeDigits != nil {
			s.StudentCodeDigits = max0(*in.StudentCodeDigits)
		}
	}

	// Save / Upsert
	if err == gorm.ErrRecordNotFound {
		if err := database.DB.Create(&s).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_SAVE_ERROR"})
		}
	} else {
		if err := database.DB.Save(&s).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_SAVE_ERROR"})
		}
	}

	// ตอบกลับรูปแบบเดียวกับ GetSchool (รวม code_lengths)
	resp := SchoolResp{
		ID:         s.ID,
		SchoolCode: s.SchoolCode,
		SchoolName: s.SchoolName,
		Address:    s.Address,
		Phone:      s.Phone,
		Education:  s.EducationLevel,
		CodeLengths: CodeLengthsPayload{
			TeacherCodeDigits: s.TeacherCodeDigits,
			StudentCodeDigits: s.StudentCodeDigits,
		},
	}
	return c.JSON(http.StatusOK, resp)
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
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
