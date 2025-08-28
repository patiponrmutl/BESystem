package handlers

import (
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

// แปลง string -> int; ถ้าแปลงไม่ได้ให้คืนค่าเริ่มต้น
func atoiOr(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// ดึง user id จาก JWT ของ Echo JWT middleware (c.Get("user") เป็น *jwt.Token)
func getUserIDFromContext(c echo.Context) (uint, bool) {
	v := c.Get("user")
	if v == nil {
		return 0, false
	}
	tok, ok := v.(*jwt.Token)
	if !ok || tok == nil {
		return 0, false
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}
	// เราเซ็ต "sub" เป็น user id ตอนออก token
	subVal, ok := claims["sub"]
	if !ok {
		return 0, false
	}
	switch id := subVal.(type) {
	case float64:
		return uint(id), true
	case int:
		return uint(id), true
	case int64:
		return uint(id), true
	case string:
		if n, err := strconv.ParseUint(id, 10, 64); err == nil {
			return uint(n), true
		}
	}
	return 0, false
}

func fmtUint(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
