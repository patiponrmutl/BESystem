package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/labstack/echo/v4"

	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

/*
Context values ที่ JWT middleware ของคุณตั้งไว้:
- user_id (uint)
- role    (string)
- name    (string)
*/

// --------------------------------------------------------------------
// DTOs
// --------------------------------------------------------------------

type accountInfo struct {
	Username           string     `json:"username"`
	LastLogin          *time.Time `json:"last_login"`
	LastPasswordChange *time.Time `json:"last_password_change"`
}

type meResponse struct {
	Teacher  any `json:"teacher"`
	Homeroom any `json:"homeroom"`
	Account  any `json:"account"`
}

type profileGetResponse struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
}

type profileUpdateRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
}

type changePasswordRequest struct {
	Current string `json:"current"`
	Next    string `json:"next"`
}

// --------------------------------------------------------------------
// Utilities
// --------------------------------------------------------------------

func currentUser(c echo.Context) (uid uint, role string) {
	roleAny := c.Get("role")
	role, _ = roleAny.(string)
	idAny := c.Get("user_id")
	switch v := idAny.(type) {
	case uint:
		uid = v
	case int:
		uid = uint(v)
	default:
		uid = 0
	}
	return
}

// map user → teacher
// 1) โดยปกติ: map จาก users.username = teachers.teacher_code
// 2) fallback: หา teacher ด้วย phone/email ถ้าต้องการ
func findTeacherForUser(u *models.User) (*models.Teacher, error) {
	var t models.Teacher

	// พยายาม map ด้วย teacher_code = username
	if u.Username != "" {
		if err := database.DB.Where("teacher_code = ?", u.Username).First(&t).Error; err == nil {
			return &t, nil
		}
	}

	// fallback เบอร์/อีเมล (ถ้าอยากผูกแบบนี้)
	if u.Phone != "" {
		if err := database.DB.Where("phone = ?", u.Phone).First(&t).Error; err == nil {
			return &t, nil
		}
	}
	// ถ้าตาราง teacher มี email ให้ใช้ด้วย (ตัวอย่างนี้ model ไม่มี email)
	// if u.Email != "" { ... }

	return nil, errors.New("teacher not found for user")
}

// ดึง homeroom ล่าสุดของครูคนนั้น (ถ้าไม่มีคืน nil)
func findLatestHomeroom(teacherID uint) *models.Homeroom {
	if teacherID == 0 {
		return nil
	}
	var hr models.Homeroom
	// เลือกตัวที่สถานะปฏิบัติงานก่อน ถ้าไม่มีค่อยเอาตัวล่าสุด
	if err := database.DB.
		Where("teacher_id = ?", teacherID).
		Where("status = ?", "ปฏิบัติงาน").
		Order("id DESC").
		First(&hr).Error; err == nil {
		return &hr
	}
	if err := database.DB.
		Where("teacher_id = ?", teacherID).
		Order("id DESC").
		First(&hr).Error; err == nil {
		return &hr
	}
	return nil
}

// --------------------------------------------------------------------
// Handlers
// --------------------------------------------------------------------

// GET /teacher/me
func TeacherMe(c echo.Context) error {
	uid, role := currentUser(c)
	if uid == 0 || (role != "teacher" && role != "admin") {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "UNAUTHORIZED"})
	}

	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "USER_NOT_FOUND"})
	}

	var t *models.Teacher
	if role == "teacher" || role == "admin" {
		if tt, err := findTeacherForUser(&u); err == nil {
			t = tt
		}
	}

	var hr *models.Homeroom
	if t != nil {
		hr = findLatestHomeroom(t.ID)
	}

	resp := meResponse{
		Teacher: t, // ถ้า nil FE ควรรองรับได้
		Homeroom: func() any {
			if hr == nil {
				return nil
			}
			// ปั้น code ห้อง เช่น "ป.3/2"
			roomCode := strings.TrimSpace(hr.Grade)
			if hr.Room != "" {
				roomCode = roomCode + "/" + hr.Room
			}
			return map[string]any{
				"academic_year":   hr.AcademicYear,
				"education_stage": hr.EducationStage,
				"grade":           hr.Grade,
				"room":            hr.Room,
				"code":            roomCode,
				"position":        hr.Position,
				"status":          hr.Status,
			}
		}(),
		Account: accountInfo{
			Username:           u.Username,
			LastLogin:          u.LastLogin,
			LastPasswordChange: u.LastPasswordChange,
		},
	}
	return c.JSON(http.StatusOK, resp)
}

// GET /teacher/profile
func TeacherGetProfile(c echo.Context) error {
	uid, role := currentUser(c)
	if uid == 0 || (role != "teacher" && role != "admin") {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "UNAUTHORIZED"})
	}
	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "USER_NOT_FOUND"})
	}

	resp := profileGetResponse{
		Email:    u.Email,
		Phone:    u.Phone,
		Timezone: u.Timezone,
		Locale:   u.Locale,
	}
	return c.JSON(http.StatusOK, resp)
}

// PUT /teacher/profile
func TeacherUpdateProfile(c echo.Context) error {
	uid, role := currentUser(c)
	if uid == 0 || (role != "teacher" && role != "admin") {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "UNAUTHORIZED"})
	}

	var req profileUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}

	// sanitize
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.Timezone = strings.TrimSpace(req.Timezone)
	req.Locale = strings.TrimSpace(req.Locale)

	// validate เบื้องต้น
	if req.Email == "" && req.Phone == "" && req.Timezone == "" && req.Locale == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "EMPTY"})
	}
	if req.Email != "" && !strings.Contains(req.Email, "@") {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_EMAIL"})
	}
	// (เพิ่ม validation timezone/locale ที่ต้องการได้)

	// update user
	if err := database.DB.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]any{
		"email":    req.Email,
		"phone":    req.Phone,
		"timezone": req.Timezone,
		"locale":   req.Locale,
	}).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// sync phone ไปตารางครู (ถ้ามี)
	if req.Phone != "" {
		var u models.User
		if err := database.DB.First(&u, uid).Error; err == nil {
			if t, err := findTeacherForUser(&u); err == nil && t != nil {
				_ = database.DB.Model(&models.Teacher{}).Where("id = ?", t.ID).Update("phone", req.Phone).Error
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}

// POST /teacher/password/change
func TeacherChangePassword(c echo.Context) error {
	uid, role := currentUser(c)
	if uid == 0 || (role != "teacher" && role != "admin") {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "UNAUTHORIZED"})
	}

	var req changePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}
	req.Current = strings.TrimSpace(req.Current)
	req.Next = strings.TrimSpace(req.Next)

	if len(req.Next) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "WEAK_PASSWORD"})
	}

	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "USER_NOT_FOUND"})
	}

	// verify current
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Current)) != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_CURRENT_PASSWORD"})
	}

	// update to new hash
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Next), bcrypt.DefaultCost)
	now := time.Now()
	if err := database.DB.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]any{
		"password":             string(hash),
		"last_password_change": &now,
	}).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}
