package routes

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/patiponrmutl/BESystem/handlers"
	"github.com/patiponrmutl/BESystem/middlewares"
)

func RegisterRoutes(e *echo.Echo) {
	// ====== Auth ======
	auth := handlers.NewAuthHandler()
	e.POST("/auth/parents/register", auth.ParentRegister)
	e.GET("/auth/check-email", auth.CheckEmail)
	e.POST("/auth/parent/login", auth.ParentLogin)
	e.POST("/auth/staff/login", auth.StaffLogin)

	// ====== Public / Common ======
	school := handlers.NewSchoolHandler()
	e.GET("/school", school.GetSchool)
	e.POST("/school", school.CreateOrUpdate)
	e.DELETE("/school", school.DeleteSchool)

	std := handlers.NewStudentHandler()
	e.GET("/students", std.List)
	e.POST("/students", std.Create)
	e.PUT("/students/:id", std.Update)
	e.DELETE("/students/:id", std.Delete)

	tch := handlers.NewTeacherHandler()
	e.GET("/teachers", tch.List)
	e.POST("/teachers", tch.Create)
	e.PUT("/teachers/:id", tch.Update)
	e.DELETE("/teachers/:id", tch.Delete)

	hr := handlers.NewHomeroomHandler()
	e.GET("/homerooms", hr.List)
	e.POST("/homerooms", hr.Create)
	e.PUT("/homerooms/:id", hr.Update)
	e.DELETE("/homerooms/:id", hr.Delete)

	mv := handlers.NewStudentMoveHandler()
	e.GET("/moves", mv.List)
	e.POST("/moves", mv.Create)
	e.PUT("/moves/:id", mv.Update)
	e.DELETE("/moves/:id", mv.Delete)

	cal := handlers.NewCalendarHandler()
	e.GET("/calendar/:id", cal.GetByID)
	e.GET("/calendar/normals", cal.ListNormals)
	e.POST("/calendar/normals", cal.CreateNormal)
	e.PUT("/calendar/normals/:id", cal.UpdateNormal)
	e.DELETE("/calendar/normals/:id", cal.DeleteNormal)
	e.GET("/calendar/holidays", cal.ListHolidays)
	e.POST("/calendar/holidays", cal.CreateHoliday)
	e.PUT("/calendar/holidays/:id", cal.UpdateHoliday)
	e.DELETE("/calendar/holidays/:id", cal.DeleteHoliday)
	e.GET("/calendar/events", cal.ListEvents)
	e.POST("/calendar/events", cal.CreateEvent)
	e.PUT("/calendar/events/:id", cal.UpdateEvent)
	e.DELETE("/calendar/events/:id", cal.DeleteEvent)

	// ====== Protected Groups ======
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	authMW := middlewares.RequireAuth(secret)

	// Admin routes
	admin := e.Group("/admin", authMW, middlewares.RequireRole("admin"))
	admin.GET("/schools", school.GetSchoolAdmin)
	admin.POST("/schools", school.CreateOrUpdate)
	admin.GET("/teachers", tch.List)
	admin.POST("/teachers", tch.Create)
	admin.PUT("/teachers/:id", tch.Update)
	admin.DELETE("/teachers/:id", tch.Delete)
	admin.GET("/students", std.List)
	admin.POST("/students", std.Create)
	admin.PUT("/students/:id", std.Update)
	admin.DELETE("/students/:id", std.Delete)
	admin.GET("/homerooms", hr.List)
	admin.POST("/homerooms", hr.Create)
	admin.PUT("/homerooms/:id", hr.Update)
	admin.DELETE("/homerooms/:id", hr.Delete)
	admin.GET("/moves", mv.List)
	admin.POST("/moves", mv.Create)
	admin.PUT("/moves/:id", mv.Update)
	admin.DELETE("/moves/:id", mv.Delete)
	admin.GET("/calendar/normals", cal.ListNormals)
	admin.POST("/calendar/normals", cal.CreateNormal)
	admin.PUT("/calendar/normals/:id", cal.UpdateNormal)
	admin.DELETE("/calendar/normals/:id", cal.DeleteNormal)
	admin.GET("/calendar/holidays", cal.ListHolidays)
	admin.POST("/calendar/holidays", cal.CreateHoliday)
	admin.PUT("/calendar/holidays/:id", cal.UpdateHoliday)
	admin.DELETE("/calendar/holidays/:id", cal.DeleteHoliday)
	admin.GET("/calendar/events", cal.ListEvents)
	admin.POST("/calendar/events", cal.CreateEvent)
	admin.PUT("/calendar/events/:id", cal.UpdateEvent)
	admin.DELETE("/calendar/events/:id", cal.DeleteEvent)

	// --- เพิ่มบรรทัดนี้ตรงส่วนประกาศ handler อื่น ๆ ---
	lv := handlers.NewLeaveRequestHandler()

	// Teacher routes
	teacher := e.Group("/teacher", authMW, middlewares.RequireRole("teacher", "admin"))
	{
		// ที่มีอยู่แล้ว
		teacher.GET("/students", std.List)
		teacher.GET("/homerooms", hr.List)
		teacher.POST("/attendance/mark", handlers.MarkAttendance)

		// ===== Leave Requests (ใหม่) =====
		teacher.GET("/leave-requests", lv.List)
		teacher.GET("/leave-requests/pending-count", lv.PendingCount)
		teacher.POST("/leave-requests/:id/approve", lv.Approve)
		teacher.POST("/leave-requests/:id/reject", lv.Reject)
	}

	// Parent routes
	parent := e.Group("/parent", authMW, middlewares.RequireRole("parent"))
	parent.GET("/children", handlers.ParentChildren)
	parent.GET("/calendar/events", cal.ListEvents)
}
