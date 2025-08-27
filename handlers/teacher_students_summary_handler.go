package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/patiponrmutl/BESystem/database"
)

type TeacherStudentsSummaryHandler struct{}

func NewTeacherStudentsSummaryHandler() *TeacherStudentsSummaryHandler {
	return &TeacherStudentsSummaryHandler{}
}

// GET /teacher/students-summary?q=&grade=&room=&limit=
// คืนฟิลด์แบบย่อ: id, code, full_name, grade, room
func (h *TeacherStudentsSummaryHandler) List(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	grade := strings.TrimSpace(c.QueryParam("grade"))
	room := strings.TrimSpace(c.QueryParam("room"))
	limit := atoiOr(c.QueryParam("limit"), 200)
	if limit <= 0 || limit > 1000 {
		limit = 200
	}

	type row struct {
		ID        uint   `json:"id"`
		Code      string `json:"code"`
		Prefix    string `json:"-"`
		FirstName string `json:"-"`
		LastName  string `json:"-"`
		FullName  string `json:"full_name"`
		Grade     string `json:"grade"`
		Room      string `json:"room"`
	}

	tx := database.DB.Table("students").
		Select("id, student_id AS code, prefix, first_name, last_name, grade, room")

	if grade != "" {
		tx = tx.Where("grade = ?", grade)
	}
	if room != "" {
		tx = tx.Where("room = ?", room)
	}
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		tx = tx.Where("LOWER(student_id) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?", like, like, like)
	}

	var rows []row
	if err := tx.Order("grade, room, student_id").Limit(limit).Scan(&rows).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusOK, []row{})
		}
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	for i := range rows {
		fn := strings.TrimSpace(rows[i].FirstName)
		ln := strings.TrimSpace(rows[i].LastName)
		pf := strings.TrimSpace(rows[i].Prefix)
		name := strings.TrimSpace(strings.Join([]string{pf, fn, ln}, " "))
		rows[i].FullName = strings.Join(strings.Fields(name), " ")
	}

	return c.JSON(http.StatusOK, rows)
}
