package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// สตับสำหรับ /teacher/attendance/mark
func MarkAttendance(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]any{
		"error": "NOT_IMPLEMENTED",
		"hint":  "implement attendance marking logic here",
	})
}

// สตับสำหรับ /parent/children
func ParentChildren(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]any{
		"error": "NOT_IMPLEMENTED",
		"hint":  "return children list for the parent (by JWT subject)",
	})
}
