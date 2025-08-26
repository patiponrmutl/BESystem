package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Claims ที่เราคาดหวัง (ตามที่เซ็นใน auth_handler.go)
type Claims struct {
	Sub  uint   `json:"sub"`
	Role string `json:"role"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

// ดึง token จาก Authorization header
func extractBearer(c echo.Context) (string, error) {
	h := c.Request().Header.Get("Authorization")
	if h == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "MISSING_AUTH_HEADER"})
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_AUTH_HEADER"})
	}
	return parts[1], nil
}

// ตรวจ JWT (HS256) และแนบ claims ไว้ใน context
func RequireAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tok, err := extractBearer(c)
			if err != nil {
				return err
			}
			token, err := jwt.ParseWithClaims(tok, &Claims{}, func(t *jwt.Token) (any, error) {
				// ป้องกัน alg โดนสลับ
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_TOKEN_METHOD"})
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_TOKEN"})
			}
			claims, ok := token.Claims.(*Claims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "INVALID_CLAIMS"})
			}
			// ตรวจ expiry (กัน lib ถูก config ปิด)
			if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{"error": "TOKEN_EXPIRED"})
			}
			// แนบไว้ใน context
			c.Set("user_id", claims.Sub)
			c.Set("role", claims.Role)
			c.Set("name", claims.Name)
			return next(c)
		}
	}
}

// จำกัดบทบาทที่อนุญาต เช่น RequireRole("admin") หรือ RequireRole("teacher","admin")
func RequireRole(roles ...string) echo.MiddlewareFunc {
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleAny := c.Get("role")
			role, _ := roleAny.(string)
			if _, ok := allowed[strings.ToLower(role)]; !ok {
				return echo.NewHTTPError(http.StatusForbidden, map[string]any{"error": "FORBIDDEN"})
			}
			return next(c)
		}
	}
}
