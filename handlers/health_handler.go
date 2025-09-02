package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Health ใช้สำหรับ /health
func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
