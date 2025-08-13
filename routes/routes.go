package routes

import (
	"github.com/labstack/echo/v4"
	_ "github.com/patiponrmutl/BESystem/docs" // swagger docs
	"github.com/patiponrmutl/BESystem/handlers"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Register(e *echo.Echo) {
	// Health
	e.GET("/health", handlers.HealthCheck)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}
