package handlers

import (
	"encoding/json"
	"io"
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

type StudentMoveHandler struct{}

func NewStudentMoveHandler() *StudentMoveHandler { return &StudentMoveHandler{} }

var (
	mvReRoom = regexp.MustCompile(`^[0-9]{1,3}$`) // ห้องเลข ≤3
)

/* -------------------- Payload structs -------------------- */

type studentLite struct {
	ID        any    `json:"id"`
	StudentID string `json:"student_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Grade     string `json:"grade"`
	Room      string `json:"room"`
}

type moveSinglePayload struct {
	Type        string        `json:"type"` // "single"
	MoveDate    string        `json:"moveDate"`
	FromYear    string        `json:"fromYear"`
	FromGrade   string        `json:"fromGrade"`
	FromRoom    string        `json:"fromRoom"`
	ToYear      string        `json:"toYear"`
	ToGrade     string        `json:"toGrade"`
	ToRoom      string        `json:"toRoom"`
	FromDisplay string        `json:"fromDisplay"`
	ToDisplay   string        `json:"toDisplay"`
	Students    []studentLite `json:"students"` // ต้องมี 1 คน
	Count       int           `json:"count"`
	Note        string        `json:"note"`
}

type moveBulkPayload struct {
	Type        string        `json:"type"` // "bulk"
	MoveDate    string        `json:"moveDate"`
	FromYear    string        `json:"fromYear"`
	FromGrade   string        `json:"fromGrade"`
	FromRoom    string        `json:"fromRoom"`
	ToYear      string        `json:"toYear"`
	ToGrade     string        `json:"toGrade"`
	ToRoom      string        `json:"toRoom"`
	FromDisplay string        `json:"fromDisplay"`
	ToDisplay   string        `json:"toDisplay"`
	Students    []studentLite `json:"students"` // >= 1 คน
	Count       int           `json:"count"`
	Note        string        `json:"note"`
}

/* -------------------- Helpers -------------------- */

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func parseStudentDBID(v any) (uint, bool) {
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
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return uint(n), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func validateMoveSingle(p *moveSinglePayload) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(p.Type) != "single" {
		errs["type"] = "ชนิดการย้ายไม่ถูกต้อง"
	}
	if strings.TrimSpace(p.MoveDate) == "" {
		errs["moveDate"] = "กรุณาเลือกวันที่ย้าย"
	} else if _, err := time.Parse("2006-01-02", p.MoveDate); err != nil {
		errs["moveDate"] = "รูปแบบวันที่ต้องเป็น YYYY-MM-DD"
	}
	if strings.TrimSpace(p.ToGrade) == "" {
		errs["toGrade"] = "กรุณาเลือกชั้นใหม่"
	}
	if strings.TrimSpace(p.ToYear) == "" {
		errs["toYear"] = "กรุณาระบุปีการศึกษาใหม่"
	}
	if !mvReRoom.MatchString(strings.TrimSpace(p.ToRoom)) {
		errs["toRoom"] = "ห้องใหม่ต้องเป็นตัวเลขไม่เกิน 3 หลัก"
	}
	if len(p.Students) != 1 {
		errs["students"] = "ต้องเลือกนักเรียน 1 คน"
	} else {
		if _, ok := parseStudentDBID(p.Students[0].ID); !ok {
			errs["students"] = "รหัสนักเรียนไม่ถูกต้อง"
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func validateMoveBulk(p *moveBulkPayload) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(p.Type) != "bulk" {
		errs["type"] = "ชนิดการย้ายไม่ถูกต้อง"
	}
	if strings.TrimSpace(p.MoveDate) == "" {
		errs["moveDate"] = "กรุณาเลือกวันที่ย้าย"
	} else if _, err := time.Parse("2006-01-02", p.MoveDate); err != nil {
		errs["moveDate"] = "รูปแบบวันที่ต้องเป็น YYYY-MM-DD"
	}
	if strings.TrimSpace(p.ToGrade) == "" {
		errs["toGrade"] = "กรุณาเลือกชั้นใหม่"
	}
	if strings.TrimSpace(p.ToYear) == "" {
		errs["toYear"] = "กรุณาระบุปีการศึกษาใหม่"
	}
	if !mvReRoom.MatchString(strings.TrimSpace(p.ToRoom)) {
		errs["toRoom"] = "ห้องใหม่ต้องเป็นตัวเลขไม่เกิน 3 หลัก"
	}
	if len(p.Students) < 1 {
		errs["students"] = "ต้องเลือกนักเรียนอย่างน้อย 1 คน"
	} else {
		okCnt := 0
		for _, s := range p.Students {
			if _, ok := parseStudentDBID(s.ID); ok {
				okCnt++
			}
		}
		if okCnt == 0 {
			errs["students"] = "รหัสนักเรียนไม่ถูกต้อง"
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

/* -------------------- Processors -------------------- */

func (h *StudentMoveHandler) processMoveSingle(c echo.Context, p *moveSinglePayload) error {
	p.ToRoom = onlyDigits(p.ToRoom)
	if errs := validateMoveSingle(p); errs != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}

	stuID, _ := parseStudentDBID(p.Students[0].ID)
	var stu models.Student
	if err := database.DB.First(&stu, "id = ?", stuID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "STUDENT_NOT_FOUND"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}

	mvDate, _ := time.Parse("2006-01-02", p.MoveDate)
	rec := models.StudentMove{
		StudentID: stu.ID,
		FromYear:  strings.TrimSpace(p.FromYear),
		FromGrade: strings.TrimSpace(p.FromGrade),
		FromRoom:  strings.TrimSpace(p.FromRoom),
		ToYear:    strings.TrimSpace(p.ToYear),
		ToGrade:   strings.TrimSpace(p.ToGrade),
		ToRoom:    strings.TrimSpace(p.ToRoom),
		MoveDate:  mvDate,
		Note:      strings.TrimSpace(p.Note),
	}

	tx := database.DB.Begin()
	if err := tx.Create(&rec).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	stu.Grade = p.ToGrade
	stu.Room = p.ToRoom
	if err := tx.Save(&stu).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 201 Created ตามสเปก
	return c.JSON(http.StatusCreated, map[string]any{
		"id":      rec.ID,
		"moved":   1,
		"student": stu,
		"record":  rec,
	})
}

func (h *StudentMoveHandler) processMoveBulk(c echo.Context, p *moveBulkPayload) error {
	p.ToRoom = onlyDigits(p.ToRoom)
	if errs := validateMoveBulk(p); errs != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": errs})
	}

	mvDate, _ := time.Parse("2006-01-02", p.MoveDate)

	ids := make([]uint, 0, len(p.Students))
	for _, s := range p.Students {
		if id, ok := parseStudentDBID(s.ID); ok {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "NO_VALID_STUDENT_IDS"})
	}

	var stus []models.Student
	if err := database.DB.Where("id IN ?", ids).Find(&stus).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	if len(stus) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "STUDENTS_NOT_FOUND"})
	}

	tx := database.DB.Begin()

	created := 0
	notFound := []uint{}
	indexByID := map[uint]*models.Student{}
	for i := range stus {
		indexByID[stus[i].ID] = &stus[i]
	}

	for _, id := range ids {
		stu, ok := indexByID[id]
		if !ok {
			notFound = append(notFound, id)
			continue
		}
		rec := models.StudentMove{
			StudentID: stu.ID,
			FromYear:  strings.TrimSpace(p.FromYear),
			FromGrade: strings.TrimSpace(stu.Grade),
			FromRoom:  strings.TrimSpace(stu.Room),
			ToYear:    strings.TrimSpace(p.ToYear),
			ToGrade:   strings.TrimSpace(p.ToGrade),
			ToRoom:    strings.TrimSpace(p.ToRoom),
			MoveDate:  mvDate,
			Note:      strings.TrimSpace(p.Note),
		}
		if err := tx.Create(&rec).Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		stu.Grade = p.ToGrade
		stu.Room = p.ToRoom
		if err := tx.Save(stu).Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		created++
	}

	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 201 Created ตามสเปก
	return c.JSON(http.StatusCreated, map[string]any{
		"moved":         created,
		"requested":     len(ids),
		"not_found_ids": notFound,
	})
}

/* -------------------- Handlers: CRUD -------------------- */

// GET /moves  — ตามสเปก: ?fromYear&fromGrade&fromRoom&toYear&toGrade&toRoom&type=single|bulk&q=&limit=&offset=
func (h *StudentMoveHandler) List(c echo.Context) error {
	limit := 20
	offset := 0
	if v, err := strconv.Atoi(c.QueryParam("limit")); err == nil {
		if v < 1 {
			limit = 1
		} else if v > 100 {
			limit = 100
		} else {
			limit = v
		}
	}
	if v, err := strconv.Atoi(c.QueryParam("offset")); err == nil && v >= 0 {
		offset = v
	}

	tx := database.DB.Model(&models.StudentMove{})

	// ฟิลเตอร์ตามสเปก
	if s := strings.TrimSpace(c.QueryParam("fromYear")); s != "" {
		tx = tx.Where("from_year = ?", s)
	}
	if s := strings.TrimSpace(c.QueryParam("fromGrade")); s != "" {
		tx = tx.Where("from_grade = ?", s)
	}
	if s := strings.TrimSpace(c.QueryParam("fromRoom")); s != "" {
		tx = tx.Where("from_room = ?", s)
	}
	if s := strings.TrimSpace(c.QueryParam("toYear")); s != "" {
		tx = tx.Where("to_year = ?", s)
	}
	if s := strings.TrimSpace(c.QueryParam("toGrade")); s != "" {
		tx = tx.Where("to_grade = ?", s)
	}
	if s := strings.TrimSpace(c.QueryParam("toRoom")); s != "" {
		tx = tx.Where("to_room = ?", s)
	}
	if s := strings.TrimSpace(c.QueryParam("type")); s != "" {
		s = strings.ToLower(s)
		if s == "single" || s == "bulk" {
			tx = tx.Where("type = ?", s) // <-- ลบทิ้ง เพราะไม่มีคอลัมน์ type
		}
	}

	if q := strings.TrimSpace(c.QueryParam("q")); q != "" {
		like := "%" + q + "%"
		tx = tx.Where(`
			from_year ILIKE ? OR to_year ILIKE ? OR
			from_grade ILIKE ? OR to_grade ILIKE ? OR
			from_room ILIKE ? OR to_room ILIKE ? OR
			note ILIKE ?
		`, like, like, like, like, like, like, like)
	}

	var items []models.StudentMove
	if err := tx.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}
	return c.JSON(http.StatusOK, items)
}

// GET /moves/:id — รายละเอียด + นักเรียนแบบสรุป
func (h *StudentMoveHandler) GetByID(c echo.Context) error {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}

	var rec models.StudentMove
	if err := database.DB.First(&rec, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}

	// โหลดนักเรียน (brief fields)
	var stu models.Student
	if err := database.DB.First(&stu, "id = ?", rec.StudentID).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_STUDENT_FAILED"})
	}
	brief := map[string]any{
		"id":         stu.ID,
		"student_id": stu.StudentID,
		"prefix":     stu.Prefix,
		"first_name": stu.FirstName,
		"last_name":  stu.LastName,
		"grade":      stu.Grade,
		"room":       stu.Room,
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id": rec.ID,
		// ไม่มี rec.Type ในโมเดล จึงไม่ส่ง "type"
		"moveDate":   rec.MoveDate.Format("2006-01-02"),
		"fromYear":   rec.FromYear,
		"fromGrade":  rec.FromGrade,
		"fromRoom":   rec.FromRoom,
		"toYear":     rec.ToYear,
		"toGrade":    rec.ToGrade,
		"toRoom":     rec.ToRoom,
		"note":       rec.Note,
		"created_at": rec.CreatedAt,
		"updated_at": rec.UpdatedAt,
		"students":   []any{brief}, // รายชื่อนักเรียนแบบสรุป
	})

}

// POST /students/move/single
func (h *StudentMoveHandler) MoveSingle(c echo.Context) error {
	var p moveSinglePayload
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	return h.processMoveSingle(c, &p)
}

// POST /students/move/bulk
func (h *StudentMoveHandler) MoveBulk(c echo.Context) error {
	var p moveBulkPayload
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}
	return h.processMoveBulk(c, &p)
}

// POST /moves — alias สำหรับ FE (รองรับ json/form + type หลายรูปแบบ)
func (h *StudentMoveHandler) MoveAuto(c echo.Context) error {
	body, _ := io.ReadAll(c.Request().Body)
	defer func() { c.Request().Body = io.NopCloser(strings.NewReader(string(body))) }()

	getStr := func(m map[string]any, keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok && v != nil {
				switch t := v.(type) {
				case string:
					return strings.TrimSpace(t)
				case float64:
					return strings.TrimSpace(strconv.FormatFloat(t, 'f', -1, 64))
				case int:
					return strings.TrimSpace(strconv.Itoa(t))
				}
			}
		}
		return ""
	}

	ctype := strings.ToLower(c.Request().Header.Get("Content-Type"))
	var typ string
	var root map[string]any

	if len(body) > 0 && strings.Contains(ctype, "application/json") {
		if err := json.Unmarshal(body, &root); err == nil && root != nil {
			typ = strings.ToLower(getStr(root, "type", "Type", "move_type"))
		}
	}
	if typ == "" && (strings.Contains(ctype, "application/x-www-form-urlencoded") || strings.Contains(ctype, "multipart/form-data")) {
		if err := c.Request().ParseForm(); err == nil {
			typ = strings.ToLower(strings.TrimSpace(c.Request().Form.Get("type")))
			if typ == "" {
				typ = strings.ToLower(strings.TrimSpace(c.Request().Form.Get("move_type")))
			}
		}
	}
	if typ == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error":  "VALIDATION_ERROR",
			"fields": map[string]string{"type": "required (single หรือ bulk)"},
		})
	}

	switch typ {
	case "single":
		var p moveSinglePayload
		if len(body) > 0 && strings.Contains(ctype, "application/json") {
			if err := json.Unmarshal(body, &p); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
			}
		} else {
			p.Type = "single"
			p.MoveDate = c.FormValue("moveDate")
			p.FromYear = c.FormValue("fromYear")
			p.FromGrade = c.FormValue("fromGrade")
			p.FromRoom = c.FormValue("fromRoom")
			p.ToYear = c.FormValue("toYear")
			p.ToGrade = c.FormValue("toGrade")
			p.ToRoom = c.FormValue("toRoom")
			p.Note = c.FormValue("note")
			if arr := c.Request().Form["students[]"]; len(arr) > 0 {
				if id, err := strconv.Atoi(strings.TrimSpace(arr[0])); err == nil {
					p.Students = []studentLite{{ID: id}}
					p.Count = 1
				}
			}
		}
		return h.processMoveSingle(c, &p)

	case "bulk":
		var p moveBulkPayload
		if len(body) > 0 && strings.Contains(ctype, "application/json") {
			if err := json.Unmarshal(body, &p); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
			}
		} else {
			p.Type = "bulk"
			p.MoveDate = c.FormValue("moveDate")
			p.FromYear = c.FormValue("fromYear")
			p.FromGrade = c.FormValue("fromGrade")
			p.FromRoom = c.FormValue("fromRoom")
			p.ToYear = c.FormValue("toYear")
			p.ToGrade = c.FormValue("toGrade")
			p.ToRoom = c.FormValue("toRoom")
			p.Note = c.FormValue("note")
			raw := c.Request().Form["students[]"]
			for _, s := range raw {
				if id, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && id > 0 {
					p.Students = append(p.Students, studentLite{ID: id})
				}
			}
			p.Count = len(p.Students)
		}
		return h.processMoveBulk(c, &p)

	default:
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
			"error":  "VALIDATION_ERROR",
			"fields": map[string]string{"type": "must be 'single' or 'bulk'"},
		})
	}
}

// PUT /moves/:id — แก้ไขปลายทาง/วันที่/หมายเหตุ
func (h *StudentMoveHandler) Update(c echo.Context) error {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}

	var req map[string]any
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_PAYLOAD"})
	}

	var rec models.StudentMove
	if err := database.DB.First(&rec, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_FAILED"})
	}

	newToYear := rec.ToYear
	newToGrade := rec.ToGrade
	newToRoom := rec.ToRoom
	newMoveDate := rec.MoveDate
	newNote := rec.Note

	getStr := func(k string) string {
		if v, ok := req[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				return strings.TrimSpace(t)
			case float64:
				return strings.TrimSpace(strconv.FormatFloat(t, 'f', -1, 64))
			case int:
				return strings.TrimSpace(strconv.Itoa(t))
			}
		}
		return ""
	}

	if v := getStr("toYear"); v != "" {
		newToYear = v
	}
	if v := getStr("toGrade"); v != "" {
		newToGrade = v
	}
	if v := getStr("toRoom"); v != "" {
		newToRoom = onlyDigits(v)
	}
	if v := getStr("note"); v != "" || (req["note"] != nil && v == "") {
		newNote = v
	}
	if v := getStr("moveDate"); v != "" {
		tm, e := time.Parse("2006-01-02", v)
		if e != nil {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]any{
				"error":  "VALIDATION_ERROR",
				"fields": map[string]string{"moveDate": "รูปแบบวันที่ต้องเป็น YYYY-MM-DD"},
			})
		}
		newMoveDate = tm
	}

	fields := map[string]string{}
	if strings.TrimSpace(newToYear) == "" {
		fields["toYear"] = "กรุณาระบุปีการศึกษา"
	}
	if strings.TrimSpace(newToGrade) == "" {
		fields["toGrade"] = "กรุณาเลือกชั้นปลายทาง"
	}
	if !mvReRoom.MatchString(strings.TrimSpace(newToRoom)) {
		fields["toRoom"] = "ห้องปลายทางต้องเป็นเลขไม่เกิน 3 หลัก"
	}
	if len(fields) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "VALIDATION_ERROR", "fields": fields})
	}

	var stu models.Student
	if err := database.DB.First(&stu, "id = ?", rec.StudentID).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_QUERY_STUDENT_FAILED"})
	}

	tx := database.DB.Begin()

	rec.ToYear = newToYear
	rec.ToGrade = newToGrade
	rec.ToRoom = newToRoom
	rec.MoveDate = newMoveDate
	rec.Note = newNote

	if err := tx.Save(&rec).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if stu.Grade != newToGrade || stu.Room != newToRoom {
		stu.Grade = newToGrade
		stu.Room = newToRoom
		if err := tx.Save(&stu).Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
	}
	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{"record": rec, "student": stu})
}

// DELETE /moves/:id
func (h *StudentMoveHandler) Delete(c echo.Context) error {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "INVALID_ID"})
	}

	tx := database.DB.Delete(&models.StudentMove{}, "id = ?", id)
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{"error": "DB_DELETE_FAILED"})
	}
	if tx.RowsAffected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "NOT_FOUND"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Create คือ alias ของ MoveAuto เพื่อให้ routes.go เรียกได้
func (h *StudentMoveHandler) Create(c echo.Context) error {
	return h.MoveAuto(c)
}
