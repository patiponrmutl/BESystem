package handlers

import "github.com/labstack/echo/v4"

// HealthCheck godoc
// @Summary      Health check
// @Tags         health
// @Success      200 {object} map[string]string
// @Router       /health [get]
func HealthCheck(c echo.Context) error {
	return c.JSON(200, map[string]string{"status": "ok"})
}
