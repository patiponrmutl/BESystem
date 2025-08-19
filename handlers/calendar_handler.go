package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

type CalendarHandler struct{}

func NewCalendarHandler() *CalendarHandler { return &CalendarHandler{} }

var (
	reTimeHHMM = regexp.MustCompile(`^\d{2}:\d{2}$`)
	reRoom     = regexp.MustCompile(`^\d{1,3}$`) // ไว้เผื่ออนาคต
)

func parseDateYYYYMMDD(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func (h *CalendarHandler) List(c echo.Context) error {
	q := database.DB.Model(&models.CalendarItem{})
	if t := strings.TrimSpace(c.QueryParam("type")); t != "" {
		q = q.Where("type = ?", t)
	}
	var items []models.CalendarItem
	if err := q.Order("id DESC").Find(&items).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *CalendarHandler) GetByID(c echo.Context) error {
	var it models.CalendarItem
	if err := database.DB.First(&it, "id = ?", c.Param("id")).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.JSON(http.StatusOK, it)
}

func (h *CalendarHandler) Create(c echo.Context) error {
	var v models.CalendarItem
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	v.Type = strings.ToLower(strings.TrimSpace(v.Type))

	// ----- Validate ตามชนิด -----
	fields := map[string]string{}
	switch v.Type {
	case "normal":
		if v.Semester == "" {
			fields["semester"] = "กรุณาเลือกภาคการศึกษา"
		}
		if v.AcademicYear == "" {
			fields["academic_year"] = "กรุณาเลือกปีการศึกษา"
		}
		if !parseDateYYYYMMDD(v.OpenDate) {
			fields["open_date"] = "รูปแบบวันที่ต้องเป็น YYYY-MM-DD"
		}
		if !parseDateYYYYMMDD(v.CloseDate) {
			fields["close_date"] = "รูปแบบวันที่ต้องเป็น YYYY-MM-DD"
		}
		if v.OpenDate != "" && v.CloseDate != "" && v.CloseDate < v.OpenDate {
			fields["close_date"] = "วันปิดภาคเรียนต้องไม่ก่อนวันเปิดภาคเรียน"
		}
		if !reTimeHHMM.MatchString(v.TimeIn) {
			fields["time_in"] = "รูปแบบเวลา HH:MM"
		}
		if !reTimeHHMM.MatchString(v.TimeOut) || (v.TimeIn != "" && v.TimeOut <= v.TimeIn) {
			fields["time_out"] = "เวลาเลิกเรียนต้องมากกว่าเวลาเข้าเรียน (HH:MM)"
		}

	case "holiday":
		if v.Name == "" {
			fields["name"] = "กรุณากรอกชื่อวันหยุด"
		}
		if !parseDateYYYYMMDD(v.StartDate) {
			fields["start_date"] = "รูปแบบวันที่ต้องเป็น YYYY-MM-DD"
		}
		if v.EndDate != "" && (!parseDateYYYYMMDD(v.EndDate) || v.EndDate < v.StartDate) {
			fields["end_date"] = "วันที่สิ้นสุดต้องไม่ก่อนวันที่เริ่มต้น"
		}

	case "event":
		if v.Title == "" {
			fields["title"] = "กรุณากรอกชื่อกิจกรรม"
		}
		if !parseDateYYYYMMDD(v.Date) {
			fields["date"] = "รูปแบบวันที่ต้องเป็น YYYY-MM-DD"
		}
		if v.StartTime != "" && !reTimeHHMM.MatchString(v.StartTime) {
			fields["start_time"] = "รูปแบบเวลา HH:MM"
		}
		if v.EndTime != "" && (!reTimeHHMM.MatchString(v.EndTime) || (v.StartTime != "" && v.EndTime <= v.StartTime)) {
			fields["end_time"] = "เวลาเลิกกิจกรรมต้องมากกว่าเวลาเริ่ม (HH:MM)"
		}

	default:
		fields["type"] = "ต้องเป็น normal | holiday | event"
	}
	if len(fields) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error":  "VALIDATION_ERROR",
			"fields": fields,
		})
	}

	if err := database.DB.Create(&v).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": v.ID})
}

func (h *CalendarHandler) Update(c echo.Context) error {
	var it models.CalendarItem
	if err := database.DB.First(&it, "id = ?", c.Param("id")).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	var p models.CalendarItem
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}

	// อนุญาตแก้เฉพาะฟิลด์ของชนิดนั้น ๆ
	switch it.Type {
	case "normal":
		if p.Semester != "" {
			it.Semester = p.Semester
		}
		if p.AcademicYear != "" {
			it.AcademicYear = p.AcademicYear
		}
		if p.OpenDate != "" {
			if !parseDateYYYYMMDD(p.OpenDate) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "open_date invalid"})
			}
			it.OpenDate = p.OpenDate
		}
		if p.CloseDate != "" {
			if !parseDateYYYYMMDD(p.CloseDate) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "close_date invalid"})
			}
			if it.OpenDate != "" && p.CloseDate < it.OpenDate {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "close_date before open_date"})
			}
			it.CloseDate = p.CloseDate
		}
		if p.TimeIn != "" {
			if !reTimeHHMM.MatchString(p.TimeIn) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "time_in invalid"})
			}
			it.TimeIn = p.TimeIn
		}
		if p.TimeOut != "" {
			if !reTimeHHMM.MatchString(p.TimeOut) || (it.TimeIn != "" && p.TimeOut <= it.TimeIn) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "time_out invalid"})
			}
			it.TimeOut = p.TimeOut
		}

	case "holiday":
		if p.Name != "" {
			it.Name = p.Name
		}
		if p.StartDate != "" {
			if !parseDateYYYYMMDD(p.StartDate) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "start_date invalid"})
			}
			it.StartDate = p.StartDate
		}
		if p.EndDate != "" {
			if !parseDateYYYYMMDD(p.EndDate) || (it.StartDate != "" && p.EndDate < it.StartDate) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "end_date invalid"})
			}
			it.EndDate = p.EndDate
		}

	case "event":
		if p.Title != "" {
			it.Title = p.Title
		}
		if p.Date != "" {
			if !parseDateYYYYMMDD(p.Date) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "date invalid"})
			}
			it.Date = p.Date
		}
		if p.StartTime != "" {
			if !reTimeHHMM.MatchString(p.StartTime) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "start_time invalid"})
			}
			it.StartTime = p.StartTime
		}
		if p.EndTime != "" {
			if !reTimeHHMM.MatchString(p.EndTime) || (it.StartTime != "" && p.EndTime <= it.StartTime) {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "end_time invalid"})
			}
			it.EndTime = p.EndTime
		}
	}

	// note ใช้ได้ทุกชนิด
	if p.Note != "" || (p.Note == "" && c.Request().Body != nil) {
		it.Note = p.Note
	}

	if err := database.DB.Save(&it).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, it)
}

func (h *CalendarHandler) Delete(c echo.Context) error {
	tx := database.DB.Delete(&models.CalendarItem{}, "id = ?", c.Param("id"))
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_DELETE_FAILED"})
	}
	if tx.RowsAffected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.NoContent(http.StatusNoContent)
}
