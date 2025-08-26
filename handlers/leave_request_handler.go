package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

type LeaveRequestHandler struct{}

func NewLeaveRequestHandler() *LeaveRequestHandler { return &LeaveRequestHandler{} }

// GET /teacher/leave-requests?status=&type=&studentId=&from=&to=&q=&page=&size=
func (h *LeaveRequestHandler) List(c echo.Context) error {
	var rows []models.LeaveRequest

	// filters
	status := strings.TrimSpace(c.QueryParam("status"))       // รออนุมัติ/อนุมัติ/ปฏิเสธ
	typ := strings.TrimSpace(c.QueryParam("type"))            // ป่วย/ธุระส่วนตัว/อื่นๆ
	studentID := strings.TrimSpace(c.QueryParam("studentId")) // id
	from := strings.TrimSpace(c.QueryParam("from"))           // YYYY-MM-DD
	to := strings.TrimSpace(c.QueryParam("to"))               // YYYY-MM-DD
	q := strings.TrimSpace(c.QueryParam("q"))                 // คีย์เวิร์ดใน reason

	page := atoiOr(c.QueryParam("page"), 1)
	size := atoiOr(c.QueryParam("size"), 10)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	tx := database.DB.Model(&models.LeaveRequest{})

	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if typ != "" {
		tx = tx.Where("type = ?", typ)
	}
	if studentID != "" {
		tx = tx.Where("student_id = ?", studentID)
	}
	if from != "" && to != "" {
		// ทับซ้อนช่วง (overlap): (DateFrom <= to) AND (DateTo >= from)
		tx = tx.Where("date_from <= ? AND date_to >= ?", to, from)
	}
	if q != "" {
		tx = tx.Where("reason ILIKE ?", "%"+q+"%")
	}

	// เรียงล่าสุดก่อน
	offset := (page - 1) * size
	if err := tx.Order("submitted_at DESC, id DESC").Offset(offset).Limit(size).Find(&rows).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// รูปแบบฟิลด์ให้ตรง FE (LeaveRequestsPage.jsx) — ฟิลด์เราตรงอยู่แล้ว
	return c.JSON(http.StatusOK, rows)
}

// GET /teacher/leave-requests/pending-count
func (h *LeaveRequestHandler) PendingCount(c echo.Context) error {
	var n int64
	if err := database.DB.Model(&models.LeaveRequest{}).
		Where("status = ?", "รออนุมัติ").Count(&n).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"count": n})
}

type updateReq struct {
	Status       string `json:"status"`       // "อนุมัติ"|"ปฏิเสธ"
	RejectReason string `json:"rejectReason"` // ถ้า "ปฏิเสธ" ต้องมี
}

// POST /teacher/leave-requests/:id/approve
func (h *LeaveRequestHandler) Approve(c echo.Context) error {
	id := c.Param("id")
	return h.updateStatus(c, id, updateReq{Status: "อนุมัติ"})
}

// POST /teacher/leave-requests/:id/reject
func (h *LeaveRequestHandler) Reject(c echo.Context) error {
	id := c.Param("id")
	var body updateReq
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}
	body.Status = "ปฏิเสธ"
	return h.updateStatus(c, id, body)
}

func (h *LeaveRequestHandler) updateStatus(c echo.Context, id string, body updateReq) error {
	var row models.LeaveRequest
	if err := database.DB.First(&row, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]any{"error": "NOT_FOUND"})
		}
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// ตรวจความถูกต้อง
	if body.Status == "ปฏิเสธ" && strings.TrimSpace(body.RejectReason) == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "REJECT_REASON_REQUIRED"})
	}

	now := time.Now()
	updates := map[string]any{
		"status":     body.Status,
		"decided_at": &now,
	}
	// เก็บ user_id คนอนุมัติ/ปฏิเสธ ถ้ามีใน context JWT
	if uid, _ := getUserID(c); uid > 0 {
		updates["decided_by"] = uid
	}
	if body.Status == "ปฏิเสธ" {
		updates["reject_reason"] = strings.TrimSpace(body.RejectReason)
	} else {
		updates["reject_reason"] = ""
	}

	if err := database.DB.Model(&models.LeaveRequest{}).Where("id = ?", row.ID).Updates(updates).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}

// helper สำหรับอ่าน user_id จาก JWT middleware ที่คุณมีอยู่
func getUserID(c echo.Context) (uint, bool) {
	switch v := c.Get("user_id").(type) {
	case uint:
		return v, true
	case int:
		return uint(v), true
	default:
		return 0, false
	}
}
func atoiOr(s string, def int) int {
	var n int
	_, err := fmtSscanf(s, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

// ใช้ fmt.Sscanf โดยเลี่ยง import fmt ตรง ๆ
func fmtSscanf(str string, format string, a ...any) (int, error) {
	return fmtSscanfImpl(str, format, a...)
}
