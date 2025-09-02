package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

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
		secret = "dev-secret"
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

// แปลง interface{} เป็น string แบบปลอดภัย
func asString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

/* ====================== DTOs ====================== */

type StaffLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

/* ====================== Seed Admin ====================== */

func EnsureDefaultAdmin() error {
	var cnt int64
	if err := database.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}

	username := os.Getenv("ADMIN_SEED_USERNAME")
	if username == "" {
		username = "Admin"
	}
	password := os.Getenv("ADMIN_SEED_PASSWORD")
	if password == "" {
		password = "1234"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := models.User{Username: username, PasswordHash: string(hash), Role: "admin"}
	if err := database.DB.Create(&u).Error; err != nil {
		return err
	}
	log.Printf("[bootstrap] default admin created: %s/%s", username, password)
	return nil
}

/* ====================== Handlers ====================== */

// POST /auth/login (staff/admin/teacher ใช้เส้นเดียวกัน)
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
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_CREDENTIALS"})
	}

	token, err := h.signJWT(uint(u.ID), u.Role, u.Username, 8*time.Hour)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]any{"error": "TOKEN_GEN_FAILED"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"user": map[string]any{
			"id":       u.ID,
			"username": u.Username,
			"role":     u.Role,
		},
	})
}

// GET /auth/me
func (h *AuthHandler) Me(c echo.Context) error {
	claims, ok := c.Get("auth.claims").(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_TOKEN"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":       claims["sub"],
		"username": claims["name"],
		"role":     claims["role"],
	})
}

/* ====================== Middleware ====================== */

// RequireAuth: parse JWT แล้วใส่ claims ลง context
func (h *AuthHandler) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	secret := h.JWTSecret
	return func(c echo.Context) error {
		ah := c.Request().Header.Get("Authorization")
		if ah == "" || !strings.HasPrefix(ah, "Bearer ") {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "MISSING_AUTH_HEADER"})
		}
		tokenStr := strings.TrimPrefix(ah, "Bearer ")
		tk, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("invalid sign method")
			}
			return []byte(secret), nil
		})
		if err != nil || !tk.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_TOKEN"})
		}
		claims, ok := tk.Claims.(jwt.MapClaims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_TOKEN"})
		}
		c.Set("auth.claims", claims)
		return next(c)
	}
}

// RequireRoles: ใช้ต่อท้ายจาก RequireAuth เพื่อบังคับสิทธิ์ตาม role
// ตัวอย่าง: group.GET("/teachers", h.List, auth.RequireRoles("admin"))
func (h *AuthHandler) RequireRoles(roles ...string) echo.MiddlewareFunc {
	// แปลง roles slice เป็น set
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get("auth.claims").(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_TOKEN"})
			}
			role := strings.ToLower(asString(claims["role"]))
			if _, ok := allowed[role]; !ok {
				return echo.NewHTTPError(http.StatusForbidden, map[string]any{"error": "FORBIDDEN"})
			}
			return next(c)
		}
	}
}
