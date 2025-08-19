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

	// === School ===
	school := handlers.NewSchoolHandler()
	e.GET("/school", school.Get)
	e.POST("/school", school.Create)
	e.PUT("/school", school.Update)
	e.DELETE("/school", school.Delete) // ✅ เพิ่ม

	// === Students ===
	st := handlers.NewStudentHandler()
	e.GET("/students", st.List)
	e.GET("/students/:id", st.Get)
	e.POST("/students", st.Create)
	e.PUT("/students/:id", st.Update)
	e.DELETE("/students/:id", st.Delete)
	e.POST("/students/import", st.Import)

	tch := handlers.NewTeacherHandler()
	e.GET("/teachers", tch.List) // ตารางรายชื่อ + ค้นหาแบบ AND ใน UI :contentReference[oaicite:12]{index=12}
	e.GET("/teachers/:id", tch.Get)
	e.POST("/teachers", tch.Create)
	e.PUT("/teachers/:id", tch.Update)
	e.DELETE("/teachers/:id", tch.Delete)
	e.POST("/teachers/import", tch.Import) // รองรับไฟล์ .xlsx/.csv ตามหัวคอลัมน์ใน UI :contentReference[oaicite:13]{index=13}

	// ✅ homerooms
	hm := handlers.NewHomeroomHandler()
	e.GET("/homerooms", hm.List)
	e.GET("/homerooms/:id", hm.Get)
	e.POST("/homerooms", hm.Create)
	e.PUT("/homerooms/:id", hm.Update)
	e.DELETE("/homerooms/:id", hm.Delete)

	// ...
	mv := handlers.NewStudentMoveHandler()

	// List & Detail
	e.GET("/moves", mv.List)
	e.GET("/moves/:id", mv.GetByID)

	// Create (alias) + แบบแยก single/bulk
	e.POST("/moves", mv.MoveAuto)
	e.POST("/students/move/single", mv.MoveSingle)
	e.POST("/students/move/bulk", mv.MoveBulk)

	// Update / Delete
	e.PUT("/moves/:id", mv.Update)
	e.DELETE("/moves/:id", mv.Delete)

}
