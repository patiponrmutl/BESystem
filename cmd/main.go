package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/patiponrmutl/BESystem/config"
	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/routes"
)

// @title           YourApp API
// @version         1.0
// @description     Echo + PostgreSQL + Swagger (Starter)
// @BasePath        /
func main() {
	cfg := config.Load()

	// เชื่อมต่อฐานข้อมูล (ถ้า DB ยังไม่ขึ้น โปรแกรมจะ error ทันที — เหมาะสำหรับ early fail)
	database.Connect(cfg)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	// ===== Pretty Logger with ANSI colors =====
	const (
		cReset  = "\033[0m"
		cCyan   = "\033[36m"
		cGreen  = "\033[32m"
		cYellow = "\033[33m"
	)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		// ตัวอย่างฟอร์แมต: เวลา (ฟ้า) METHOD URI -> STATUS (เขียว) (latency) from IP
		Format: cCyan + "[${time_rfc3339}]" + cReset +
			" ${method} ${uri} -> " + cGreen + "${status}" + cReset +
			" (${latency_human}) from ${remote_ip}\n",
	}))

	routes.Register(e)

	addr := ":" + cfg.AppPort
	log.Printf("server listening at %s", addr)
	if err := e.Start(addr); err != nil {
		log.Fatal(err)
	}
}
