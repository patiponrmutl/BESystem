package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

type CalendarHandler struct{}

func NewCalendarHandler() *CalendarHandler { return &CalendarHandler{} }

// ─── Validators ────────────────────────────────────────────────────────────────
var reHHMM = regexp.MustCompile(`^\d{2}:\d{2}$`)

func isDateYYYYMMDD(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func mustID(c echo.Context) (uint64, error) {
	idStr := c.Param("id")
	return strconv.ParseUint(idStr, 10, 64)
}

// ─── LIST (GET) ────────────────────────────────────────────────────────────────

// GET /calendar/normals
func (h *CalendarHandler) ListNormals(c echo.Context) error {
	var items []models.CalendarItem
	if err := database.DB.Where("type = ?", "normal").Order("id DESC").Find(&items).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, items)
}

// GET /calendar/holidays
func (h *CalendarHandler) ListHolidays(c echo.Context) error {
	var items []models.CalendarItem
	if err := database.DB.Where("type = ?", "holiday").Order("id DESC").Find(&items).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, items)
}

// GET /calendar/events
func (h *CalendarHandler) ListEvents(c echo.Context) error {
	var items []models.CalendarItem
	if err := database.DB.Where("type = ?", "event").Order("id DESC").Find(&items).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, items)
}

// GET /calendar/:id  (อ่านตัวเดียวด้วย id ตัวเลขเท่านั้น)
func (h *CalendarHandler) GetByID(c echo.Context) error {
	if _, err := mustID(c); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	var it models.CalendarItem
	if err := database.DB.First(&it, "id = ?", c.Param("id")).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.JSON(http.StatusOK, it)
}

// ─── CREATE (POST) ─────────────────────────────────────────────────────────────

// POST /calendar/normals
func (h *CalendarHandler) CreateNormal(c echo.Context) error {
	var v models.CalendarItem
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	v.Type = "normal"

	fields := map[string]string{}
	if strings.TrimSpace(v.Semester) == "" {
		fields["semester"] = "กรุณาเลือกภาคการศึกษา"
	}
	if strings.TrimSpace(v.AcademicYear) == "" {
		fields["academic_year"] = "กรุณาเลือกปีการศึกษา"
	}
	if !isDateYYYYMMDD(v.OpenDate) {
		fields["open_date"] = "ต้องเป็น YYYY-MM-DD"
	}
	if !isDateYYYYMMDD(v.CloseDate) {
		fields["close_date"] = "ต้องเป็น YYYY-MM-DD"
	}
	if v.OpenDate != "" && v.CloseDate != "" && v.CloseDate < v.OpenDate {
		fields["close_date"] = "ต้องไม่ก่อนวันเปิดภาคเรียน"
	}
	if !reHHMM.MatchString(v.TimeIn) {
		fields["time_in"] = "รูปแบบเวลา HH:MM"
	}
	if !reHHMM.MatchString(v.TimeOut) || (v.TimeIn != "" && v.TimeOut <= v.TimeIn) {
		fields["time_out"] = "ต้องมากกว่าเวลาเข้าเรียน (HH:MM)"
	}
	if len(fields) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": fields})
	}

	if err := database.DB.Create(&v).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": v.ID})
}

// POST /calendar/holidays
func (h *CalendarHandler) CreateHoliday(c echo.Context) error {
	var v models.CalendarItem
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	v.Type = "holiday"

	fields := map[string]string{}
	if strings.TrimSpace(v.Name) == "" {
		fields["name"] = "กรุณากรอกชื่อวันหยุด"
	}
	if !isDateYYYYMMDD(v.StartDate) {
		fields["start_date"] = "ต้องเป็น YYYY-MM-DD"
	}
	if v.EndDate != "" && (!isDateYYYYMMDD(v.EndDate) || v.EndDate < v.StartDate) {
		fields["end_date"] = "ต้องไม่ก่อนวันที่เริ่มต้น"
	}
	if len(fields) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": fields})
	}

	if err := database.DB.Create(&v).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": v.ID})
}

// POST /calendar/events
func (h *CalendarHandler) CreateEvent(c echo.Context) error {
	var v models.CalendarItem
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	v.Type = "event"

	fields := map[string]string{}
	if strings.TrimSpace(v.Title) == "" {
		fields["title"] = "กรุณากรอกชื่อกิจกรรม"
	}
	if !isDateYYYYMMDD(v.Date) {
		fields["date"] = "ต้องเป็น YYYY-MM-DD"
	}
	if v.StartTime != "" && !reHHMM.MatchString(v.StartTime) {
		fields["start_time"] = "รูปแบบเวลา HH:MM"
	}
	if v.EndTime != "" && (!reHHMM.MatchString(v.EndTime) || (v.StartTime != "" && v.EndTime <= v.StartTime)) {
		fields["end_time"] = "ต้องมากกว่าเวลาเริ่ม (HH:MM)"
	}
	if len(fields) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": fields})
	}

	if err := database.DB.Create(&v).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": v.ID})
}

// ─── UPDATE (PUT) ──────────────────────────────────────────────────────────────

// PUT /calendar/normals/:id
func (h *CalendarHandler) UpdateNormal(c echo.Context) error {
	id, err := mustID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}
	var it models.CalendarItem
	if err := database.DB.First(&it, "id = ? AND type = ?", id, "normal").Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	var p models.CalendarItem
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}

	if p.Semester != "" {
		it.Semester = p.Semester
	}
	if p.AcademicYear != "" {
		it.AcademicYear = p.AcademicYear
	}
	if p.OpenDate != "" {
		if !isDateYYYYMMDD(p.OpenDate) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "open_date invalid"})
		}
		it.OpenDate = p.OpenDate
	}
	if p.CloseDate != "" {
		if !isDateYYYYMMDD(p.CloseDate) || (it.OpenDate != "" && p.CloseDate < it.OpenDate) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "close_date invalid"})
		}
		it.CloseDate = p.CloseDate
	}
	if p.TimeIn != "" {
		if !reHHMM.MatchString(p.TimeIn) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "time_in invalid"})
		}
		it.TimeIn = p.TimeIn
	}
	if p.TimeOut != "" {
		if !reHHMM.MatchString(p.TimeOut) || (it.TimeIn != "" && p.TimeOut <= it.TimeIn) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "time_out invalid"})
		}
		it.TimeOut = p.TimeOut
	}
	// อนุญาตให้ล้าง note ด้วยการส่ง "" มา
	if p.Note != "" || (p.Note == "" && c.Request().Body != nil) {
		it.Note = p.Note
	}

	if err := database.DB.Save(&it).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, it)
}

// PUT /calendar/holidays/:id
func (h *CalendarHandler) UpdateHoliday(c echo.Context) error {
	id, err := mustID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}
	var it models.CalendarItem
	if err := database.DB.First(&it, "id = ? AND type = ?", id, "holiday").Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	var p models.CalendarItem
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}

	if p.Name != "" {
		it.Name = p.Name
	}
	if p.StartDate != "" {
		if !isDateYYYYMMDD(p.StartDate) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "start_date invalid"})
		}
		it.StartDate = p.StartDate
	}
	if p.EndDate != "" {
		if !isDateYYYYMMDD(p.EndDate) || (it.StartDate != "" && p.EndDate < it.StartDate) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "end_date invalid"})
		}
		it.EndDate = p.EndDate
	}
	// อนุญาตให้ล้าง note
	if p.Note != "" || (p.Note == "" && c.Request().Body != nil) {
		it.Note = p.Note
	}

	if err := database.DB.Save(&it).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, it)
}

// PUT /calendar/events/:id
func (h *CalendarHandler) UpdateEvent(c echo.Context) error {
	id, err := mustID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}
	var it models.CalendarItem
	if err := database.DB.First(&it, "id = ? AND type = ?", id, "event").Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	var p models.CalendarItem
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}

	if p.Title != "" {
		it.Title = p.Title
	}
	if p.Date != "" {
		if !isDateYYYYMMDD(p.Date) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "date invalid"})
		}
		it.Date = p.Date
	}
	if p.StartTime != "" {
		if !reHHMM.MatchString(p.StartTime) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "start_time invalid"})
		}
		it.StartTime = p.StartTime
	}
	if p.EndTime != "" {
		if !reHHMM.MatchString(p.EndTime) || (it.StartTime != "" && p.EndTime <= it.StartTime) {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "end_time invalid"})
		}
		it.EndTime = p.EndTime
	}
	// อนุญาตให้ล้าง note
	if p.Note != "" || (p.Note == "" && c.Request().Body != nil) {
		it.Note = p.Note
	}

	if err := database.DB.Save(&it).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, it)
}

// ─── DELETE (DELETE) ───────────────────────────────────────────────────────────

// DELETE /calendar/normals/:id
func (h *CalendarHandler) DeleteNormal(c echo.Context) error {
	id, err := mustID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}
	tx := database.DB.Delete(&models.CalendarItem{}, "id = ? AND type = ?", id, "normal")
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_DELETE_FAILED"})
	}
	if tx.RowsAffected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.NoContent(http.StatusNoContent)
}

// DELETE /calendar/holidays/:id
func (h *CalendarHandler) DeleteHoliday(c echo.Context) error {
	id, err := mustID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}
	tx := database.DB.Delete(&models.CalendarItem{}, "id = ? AND type = ?", id, "holiday")
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_DELETE_FAILED"})
	}
	if tx.RowsAffected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.NoContent(http.StatusNoContent)
}

// DELETE /calendar/events/:id
func (h *CalendarHandler) DeleteEvent(c echo.Context) error {
	id, err := mustID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}
	tx := database.DB.Delete(&models.CalendarItem{}, "id = ? AND type = ?", id, "event")
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_DELETE_FAILED"})
	}
	if tx.RowsAffected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.NoContent(http.StatusNoContent)
}

/* ====================== เมธอดรวม สำหรับ routes แบบ /calendar/:kind ====================== */

// GET /calendar/:kind
func (h *CalendarHandler) List(c echo.Context) error {
	switch c.Param("kind") {
	case "normals":
		return h.ListNormals(c)
	case "holidays":
		return h.ListHolidays(c)
	case "events":
		return h.ListEvents(c)
	default:
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
}

// POST /calendar/:kind
func (h *CalendarHandler) Create(c echo.Context) error {
	switch c.Param("kind") {
	case "normals":
		return h.CreateNormal(c)
	case "holidays":
		return h.CreateHoliday(c)
	case "events":
		return h.CreateEvent(c)
	default:
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
}

// PUT /calendar/:kind/:id
func (h *CalendarHandler) Update(c echo.Context) error {
	switch c.Param("kind") {
	case "normals":
		return h.UpdateNormal(c)
	case "holidays":
		return h.UpdateHoliday(c)
	case "events":
		return h.UpdateEvent(c)
	default:
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
}

// DELETE /calendar/:kind/:id
func (h *CalendarHandler) Delete(c echo.Context) error {
	switch c.Param("kind") {
	case "normals":
		return h.DeleteNormal(c)
	case "holidays":
		return h.DeleteHoliday(c)
	case "events":
		return h.DeleteEvent(c)
	default:
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
}
