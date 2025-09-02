package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

// -----------------------------
// Handler & ctor
// -----------------------------

type TeacherAccountHandler struct{}

func NewTeacherAccountHandler() *TeacherAccountHandler { return &TeacherAccountHandler{} }

// -----------------------------
// Request/Response payloads
// -----------------------------

type createTeacherAccountReq struct {
	TeacherID uint   `json:"teacher_id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type patchTeacherAccountReq struct {
	// โปรเจกต์ตอนนี้ยังไม่มีคอลัมน์รองรับจริง ๆ
	// เก็บไว้ให้ FE เรียกได้ โดยจะ NO-OP หากไม่มีฟิลด์ใน DB
	Enabled             *bool `json:"enabled,omitempty"`
	ForcePasswordChange *bool `json:"force_password_change,omitempty"`
}

type resetPasswordReq struct {
	Length int `json:"length"`
}

type teacherAccountDTO struct {
	ID                  uint      `json:"id"`
	TeacherID           uint      `json:"teacher_id"`
	Username            string    `json:"username"`
	Enabled             bool      `json:"enabled"`               // ไม่มีใน DB ตอนนี้ -> ใส่ค่า default
	ForcePasswordChange bool      `json:"force_password_change"` // ไม่มีใน DB ตอนนี้ -> ใส่ค่า default
	UpdatedAt           time.Time `json:"updated_at"`
}

type updateFlagsReq struct {
	Enabled             *bool `json:"enabled"`
	ForcePasswordChange *bool `json:"force_password_change"`
}

// -----------------------------
// Helpers
// -----------------------------

func toDTO(u models.User) teacherAccountDTO {
	// TeacherID ในโมเดลเป็น *uint
	var tid uint
	if u.TeacherID != nil {
		tid = *u.TeacherID
	}
	return teacherAccountDTO{
		ID:                  u.ID,
		TeacherID:           tid,
		Username:            u.Username,
		Enabled:             true,  // ค่า default ชั่วคราว (ยังไม่มีคอลัมน์)
		ForcePasswordChange: false, // ค่า default ชั่วคราว (ยังไม่มีคอลัมน์)
		UpdatedAt:           u.UpdatedAt,
	}
}

func (h *TeacherAccountHandler) findUserByID(id uint) (*models.User, error) {
	var u models.User
	err := database.DB.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func randomPassword(n int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz0123456789"
	if n < 8 {
		n = 8
	}
	out := make([]byte, n)
	// ใช้เวลาเป็น seed แบบง่าย ๆ
	for i := 0; i < n; i++ {
		out[i] = alphabet[int(time.Now().UnixNano())%len(alphabet)]
	}
	return string(out)
}

// -----------------------------
// List accounts
// GET /admin/teacher-accounts
// -----------------------------

func (h *TeacherAccountHandler) List(c echo.Context) error {
	// ดึงเฉพาะ user role=teacher ที่มี username
	var users []models.User
	q := database.DB.Where("role = ?", "teacher").Where("username <> ''")
	if err := q.Order("updated_at desc").Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_ERROR"})
	}

	out := make([]teacherAccountDTO, 0, len(users))
	for _, u := range users {
		out = append(out, toDTO(u))
	}
	return c.JSON(http.StatusOK, out)
}

// -----------------------------
// Create account
// POST /admin/teacher-accounts
// body: { teacher_id, username, password }
// -----------------------------

func (h *TeacherAccountHandler) Create(c echo.Context) error {
	var req createTeacherAccountReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.TeacherID == 0 || req.Username == "" || len(req.Password) < 8 {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{
			"error": "VALIDATION_ERROR",
			"fields": map[string]string{
				"teacher_id": "required",
				"username":   "required",
				"password":   "min_length_8",
			},
		})
	}

	// ตรวจว่ามี teacher จริงไหม
	var t models.Teacher
	if err := database.DB.First(&t, req.TeacherID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]any{"error": "TEACHER_NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_ERROR"})
	}

	// username ซ้ำหรือยัง
	var cnt int64
	if err := database.DB.Model(&models.User{}).
		Where("username = ?", req.Username).Count(&cnt).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_ERROR"})
	}
	if cnt > 0 {
		return c.JSON(http.StatusConflict, map[string]any{"error": "USERNAME_TAKEN"})
	}

	// teacher คนนี้มีบัญชีแล้วหรือยัง (อิงจาก TeacherID pointer)
	if err := database.DB.Model(&models.User{}).
		Where("teacher_id = ?", req.TeacherID).Count(&cnt).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_ERROR"})
	}
	if cnt > 0 {
		return c.JSON(http.StatusConflict, map[string]any{"error": "TEACHER_ALREADY_HAS_ACCOUNT"})
	}

	// สร้างรหัสผ่าน
	hashed, err := hashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "HASH_ERROR"})
	}

	// TeacherID ต้องเป็น *uint
	tid := req.TeacherID
	u := models.User{
		Username:     req.Username,
		PasswordHash: hashed,
		Role:         "teacher",
		TeacherID:    &tid,
	}
	if err := database.DB.Create(&u).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_SAVE_ERROR"})
	}

	return c.JSON(http.StatusCreated, toDTO(u))
}

// -----------------------------
// Reset password (one-time)
// POST /admin/teacher-accounts/:id/reset
// body: { length }
// resp: { one_time_password }
// -----------------------------

func (h *TeacherAccountHandler) ResetPassword(c echo.Context) error {
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 64)
	if id64 == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_ID"})
	}
	var req resetPasswordReq
	if err := c.Bind(&req); err != nil {
		req.Length = 12
	}
	if req.Length < 8 {
		req.Length = 8
	}

	u, err := h.findUserByID(uint(id64))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]any{"error": "ACCOUNT_NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_ERROR"})
	}
	if u.Role != "teacher" {
		return c.JSON(http.StatusForbidden, map[string]any{"error": "NOT_TEACHER_ACCOUNT"})
	}

	newPW := randomPassword(req.Length)
	hash, err := hashPassword(newPW)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "HASH_ERROR"})
	}
	u.PasswordHash = hash

	if err := database.DB.Save(u).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_SAVE_ERROR"})
	}
	return c.JSON(http.StatusOK, map[string]any{"one_time_password": newPW})
}

// -----------------------------
// Patch fields (NO-OP ถ้าโปรเจกต์ยังไม่มีคอลัมน์)
// PATCH /admin/teacher-accounts/:id
// body: { enabled?, force_password_change? }
// -----------------------------

func (h *TeacherAccountHandler) Patch(c echo.Context) error {
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 64)
	if id64 == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_ID"})
	}

	var req patchTeacherAccountReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}

	u, err := h.findUserByID(uint(id64))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]any{"error": "ACCOUNT_NOT_FOUND"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "DB_ERROR"})
	}
	if u.Role != "teacher" {
		return c.JSON(http.StatusForbidden, map[string]any{"error": "NOT_TEACHER_ACCOUNT"})
	}

	// ตอนนี้โปรเจกต์ยังไม่มีคอลัมน์ enabled/force_password_change ในตาราง users
	// ถ้าส่งมาแต่ไม่มีอะไรอัปเดตจริง ให้แจ้งกลับว่าไม่มีฟิลด์ให้ปรับ
	if req.Enabled == nil && req.ForcePasswordChange == nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "NO_FIELDS_TO_UPDATE"})
	}

	// ทำ NO-OP แล้วคืนค่าปัจจุบันกลับไป (ค่า default)
	// หากอนาคตเพิ่มคอลัมน์ใน models.User เมื่อไร มาปรับตรงนี้ให้เซฟจริงได้ทันที
	return c.JSON(http.StatusOK, toDTO(*u))
}

func (h *TeacherAccountHandler) UpdateFlags(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_ID"})
	}

	var u models.User
	if err := database.DB.First(&u, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, map[string]any{"error": "NOT_FOUND"})
		}
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	var req updateFlagsReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}
	updates := map[string]any{}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.ForcePasswordChange != nil {
		updates["force_password_change"] = *req.ForcePasswordChange
	}
	if len(updates) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "NO_FIELDS"})
	}
	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&u).Updates(updates).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// reload to return latest
	if err := database.DB.First(&u, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":                    u.ID,
		"teacher_id":            u.TeacherID,
		"username":              u.Username,
		"enabled":               u.Enabled,
		"force_password_change": u.ForcePasswordChange,
		"updated_at":            u.UpdatedAt,
	})
}
