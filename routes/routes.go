package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/patiponrmutl/BESystem/handlers"
)

func RegisterRoutes(e *echo.Echo) {
	// health
	e.GET("/health", handlers.Health)

	// auth
	auth := handlers.NewAuthHandler()
	e.POST("/auth/login", auth.StaffLogin)
	e.GET("/auth/me", auth.Me, auth.RequireAuth)

	// ===== Protected root group (ต้องมี token) =====
	secured := e.Group("", auth.RequireAuth)

	/* ===== Admin-only endpoints ===== */
	adminOnly := secured.Group("", auth.RequireRoles("admin"))

	// School (อ่าน/แก้ไข)
	school := handlers.NewSchoolHandler()
	adminOnly.GET("/school", school.GetSchool)
	adminOnly.POST("/school", school.CreateOrUpdate)
	adminOnly.PUT("/school", school.CreateOrUpdate)
	adminOnly.DELETE("/school", school.DeleteSchool)

	// Teachers / Students (รายการ)
	teacher := handlers.NewTeacherHandler()
	adminOnly.GET("/teachers", teacher.List)

	student := handlers.NewStudentHandler()
	adminOnly.GET("/students", student.List)

	// Teacher accounts (สร้าง/จัดการบัญชีครู)
	acc := handlers.NewTeacherAccountHandler()
	adminOnly.GET("/teacher-accounts", acc.List)
	adminOnly.POST("/teacher-accounts", acc.Create)
	adminOnly.POST("/teacher-accounts/:id/reset", acc.ResetPassword)
	adminOnly.PATCH("/teacher-accounts/:id", acc.UpdateFlags)

	// ย้ายนักเรียน (move)
	mv := handlers.NewStudentMoveHandler()
	adminOnly.GET("/moves", mv.List)
	adminOnly.POST("/moves", mv.Create)
	adminOnly.PUT("/moves/:id", mv.Update)
	adminOnly.DELETE("/moves/:id", mv.Delete)

	// Calendar (สร้าง/แก้/ลบ)
	cal := handlers.NewCalendarHandler()
	adminOnly.POST("/calendar/:kind", cal.Create)
	adminOnly.PUT("/calendar/:kind/:id", cal.Update)
	adminOnly.DELETE("/calendar/:kind/:id", cal.Delete)

	/* ===== Admin + Teacher (อ่านได้) ===== */
	adminOrTeacher := secured.Group("", auth.RequireRoles("admin", "teacher"))

	// homerooms, calendar (read)
	homeroom := handlers.NewHomeroomHandler()
	adminOrTeacher.GET("/homerooms", homeroom.List)

	adminOrTeacher.GET("/calendar/:kind", cal.List)

	// leave requests (อ่านจากแอพผู้ปกครอง)
	leave := handlers.NewLeaveRequestHandler()
	adminOrTeacher.GET("/leave-requests", leave.List)
	adminOrTeacher.GET("/leave-requests/:id", leave.Get)

	// dashboard/summary (อ่าน)
	dash := handlers.NewDashboardHandler()
	adminOrTeacher.GET("/dashboard/summary", dash.Summary)

	// teacher-accounts
	e.PATCH("/teacher-accounts/:id", acc.UpdateFlags)

	// leave requests
	e.GET("/leave-requests", leave.Get)

	// dashboard
	e.GET("/dashboard/summary", dash.Summary)

}
