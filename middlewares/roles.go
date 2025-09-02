package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// RequireAnyRole("admin","teacher") → ผ่านถ้า role ของผู้ใช้ตรงอย่างน้อย 1 ค่า
func RequireAnyRole(roles ...string) echo.MiddlewareFunc {
	need := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		need[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get("auth.role").(string) // set ไว้โดย JWT middleware/handler
			if _, ok := need[strings.ToLower(role)]; !ok {
				return echo.NewHTTPError(http.StatusForbidden, map[string]any{"error": "FORBIDDEN"})
			}
			return next(c)
		}
	}
}
