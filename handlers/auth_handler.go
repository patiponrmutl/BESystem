package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

/* ====================== Config & Helpers ====================== */

type AuthHandler struct {
	JWTSecret string
}

func NewAuthHandler() *AuthHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret" // กันล่มในเครื่อง dev (โปรดตั้งใน .env จริง)
	}
	return &AuthHandler{JWTSecret: secret}
}

func (h *AuthHandler) signJWT(sub uint, role, name string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  sub,
		"role": role,
		"name": name,
		"exp":  time.Now().Add(ttl).Unix(),
		"iat":  time.Now().Unix(),
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tk.SignedString([]byte(h.JWTSecret))
}

/* ====================== DTOs ====================== */

type ParentRegisterReq struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
	AgreePDPA bool   `json:"agree_pdpa"`
}

type ParentLoginReq struct {
	Identity string `json:"identity"` // email หรือ phone
	Password string `json:"password"`
}

type StaffLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

/* ====================== Handlers ====================== */

// POST /auth/parents/register
func (h *AuthHandler) ParentRegister(c echo.Context) error {
	var req ParentRegisterReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	phone := strings.TrimSpace(req.Phone)
	pass := strings.TrimSpace(req.Password)
	if email == "" || phone == "" || pass == "" || !req.AgreePDPA {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "MISSING_FIELDS"})
	}

	// ตรวจซ้ำ email
	var dup models.Parent
	if err := database.DB.Where("email = ?", email).First(&dup).Error; err == nil {
		return echo.NewHTTPError(http.StatusConflict, map[string]any{"error": "EMAIL_EXISTS", "code": "EMAIL_EXISTS"})
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	rec := models.Parent{
		Email:    email,
		Phone:    phone,
		Password: string(hash),
		PdpaOK:   true,
	}
	if err := database.DB.Create(&rec).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"id": rec.ID})
}

// GET /auth/check-email?email=...
func (h *AuthHandler) CheckEmail(c echo.Context) error {
	email := strings.TrimSpace(strings.ToLower(c.QueryParam("email")))
	if email == "" {
		return c.JSON(http.StatusOK, map[string]bool{"exists": false})
	}
	var p models.Parent
	err := database.DB.Where("email = ?", email).First(&p).Error
	return c.JSON(http.StatusOK, map[string]bool{"exists": err == nil})
}

// POST /auth/parent/login
func (h *AuthHandler) ParentLogin(c echo.Context) error {
	var req ParentLoginReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}

	id := strings.TrimSpace(strings.ToLower(req.Identity))
	if id == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "MISSING_FIELDS"})
	}

	var p models.Parent
	q := database.DB
	if strings.Contains(id, "@") {
		q = q.Where("email = ?", id)
	} else {
		q = q.Where("phone = ?", id)
	}
	if err := q.First(&p).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
	}
	if bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(req.Password)) != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
	}

	token, err := h.signJWT(uint(p.ID), "parent", p.Email, 7*24*time.Hour)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{"error": "TOKEN_GEN_FAILED"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": p.ID, "role": "parent", "emailOrPhone": id, "name": "ผู้ปกครอง"},
	})
}

// POST /auth/staff/login
func (h *AuthHandler) StaffLogin(c echo.Context) error {
	var req StaffLoginReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]any{"error": "MISSING_FIELDS"})
	}

	var u models.User
	if err := database.DB.Where("username = ?", username).First(&u).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
	}

	token, err := h.signJWT(uint(u.ID), u.Role, u.Username, 7*24*time.Hour)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{"error": "TOKEN_GEN_FAILED"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": u.ID, "role": u.Role, "username": u.Username, "name": u.Username},
	})
}

// AdminLogin: POST /admin/login
func (h *AuthHandler) AdminLogin(c echo.Context) error {
	var req loginReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "INVALID_PAYLOAD"})
	}
	fields := map[string]string{}
	if req.Username == "" {
		fields["username"] = "required"
	}
	if req.Password == "" {
		fields["password"] = "required"
	}
	if len(fields) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{"error": "VALIDATION_ERROR", "fields": fields})
	}

	// หา user
	var u models.User
	if err := database.DB.Where("username = ?", req.Username).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
		}
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	// role ต้องเป็น admin
	if u.Role != "admin" {
		return c.JSON(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
	}

	// ตรวจรหัสผ่าน
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
	}

	// ออก JWT
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}

	ttlHours := 8
	if v := os.Getenv("JWT_TTL_HOURS"); v != "" {
		if n := atoiOr(v, 8); n > 0 && n <= 24*7 {
			ttlHours = n
		}
	}
	exp := time.Now().Add(time.Duration(ttlHours) * time.Hour)

	claims := jwt.MapClaims{
		"sub":        u.ID,
		"username":   u.Username,
		"role":       u.Role,
		"teacher_id": u.TeacherID, // อาจเป็น nil
		"iat":        time.Now().Unix(),
		"exp":        exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": "TOKEN_SIGN_ERROR"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"access_token": signed,
		"token_type":   "Bearer",
		"expires_in":   int(time.Until(exp).Seconds()),
		"user": map[string]any{
			"id":       u.ID,
			"username": u.Username,
			"role":     u.Role,
		},
	})
}
