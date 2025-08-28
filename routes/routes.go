package routes

import (
	"os"

	"github.com/labstack/echo/v4"

	"github.com/patiponrmutl/BESystem/handlers"
	"github.com/patiponrmutl/BESystem/middlewares"
)

// RegisterRoutes wires all HTTP routes.
func RegisterRoutes(e *echo.Echo) {
	// ===== Handlers (shared singletons) =====
	auth := handlers.NewAuthHandler()
	school := handlers.NewSchoolHandler()
	std := handlers.NewStudentHandler()
	tch := handlers.NewTeacherHandler()
	hr := handlers.NewHomeroomHandler()
	mv := handlers.NewStudentMoveHandler()
	cal := handlers.NewCalendarHandler()
	att := handlers.NewAttendanceHandler()
	tss := handlers.NewTeacherStudentsSummaryHandler()
	lv := handlers.NewLeaveRequestHandler()
	dash := handlers.NewDashboardHandler()
	tp := handlers.NewTeacherProfileHandler()

	// ===== Public Auth =====
	e.POST("/admin/login", auth.AdminLogin)
	e.POST("/auth/parents/register", auth.ParentRegister)
	e.GET("/auth/check-email", auth.CheckEmail)
	e.POST("/auth/parent/login", auth.ParentLogin)
	e.POST("/auth/staff/login", auth.StaffLogin)

	// ===== Public resources (minimal) =====
	e.GET("/school", school.GetSchool)
	e.POST("/school", school.CreateOrUpdate) // ใช้ตอน initial setup
	e.DELETE("/school", school.DeleteSchool)

	e.GET("/students", std.List)
	e.POST("/students", std.Create)
	e.PUT("/students/:id", std.Update)
	e.DELETE("/students/:id", std.Delete)

	e.GET("/teachers", tch.List)
	e.POST("/teachers", tch.Create)
	e.PUT("/teachers/:id", tch.Update)
	e.DELETE("/teachers/:id", tch.Delete)

	e.GET("/homerooms", hr.List)
	e.POST("/homerooms", hr.Create)
	e.PUT("/homerooms/:id", hr.Update)
	e.DELETE("/homerooms/:id", hr.Delete)

	e.GET("/moves", mv.List)
	e.POST("/moves", mv.Create)
	e.PUT("/moves/:id", mv.Update)
	e.DELETE("/moves/:id", mv.Delete)

	// ปฏิทินการศึกษา (อ่าน/สร้าง/แก้/ลบ)
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

	// ===== Protected Groups =====
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	authMW := middlewares.RequireAuth(secret)

	// ===== Admin routes =====
	admin := e.Group("/admin", authMW, middlewares.RequireRole("admin"))

	// School/People Mgmt (admin scope)
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

	// ย้ายห้อง/ย้ายนักเรียน (admin scope)
	admin.GET("/moves", mv.List)
	admin.POST("/moves", mv.Create)
	admin.PUT("/moves/:id", mv.Update)
	admin.DELETE("/moves/:id", mv.Delete)

	// ปฏิทินการศึกษา (admin scope)
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

	// ===== Admin: Teacher Accounts (รองรับหน้า TeacherAccountsPage.jsx) =====
	ta := handlers.NewTeacherAccountHandler()
	// FE ใช้เส้นทางเหล่านี้:
	// GET    /admin/teacher-accounts             → รายชื่อบัญชีครูทั้งหมด
	// POST   /admin/teacher-accounts             → { teacher_id, username, password }
	// POST   /admin/teacher-accounts/:id/reset   → รีเซ็ตรหัสผ่าน (คืน one_time_password)
	// PATCH  /admin/teacher-accounts/:id         → อัปเดต enabled / force_password_change
	admin.GET("/teacher-accounts", ta.List)
	admin.POST("/teacher-accounts", ta.Create)
	admin.POST("/teacher-accounts/:id/reset", ta.ResetPassword)
	admin.PATCH("/teacher-accounts/:id", ta.Patch)

	// ===== Teacher routes =====
	teacher := e.Group("/teacher", authMW, middlewares.RequireRole("teacher", "admin"))

	// Profile
	teacher.PUT("/profile/password", tp.ChangePassword)

	// Dashboard / Summary / Attendance
	teacher.GET("/dashboard/daily", dash.Daily)
	teacher.GET("/students-summary", tss.List)
	teacher.GET("/attendance", att.List)
	teacher.POST("/attendance/mark", att.Mark) // (ลบตัวซ้ำที่เคยเรียก handlers.MarkAttendance โดยตรง)

	// Basic data reads used in teacher UI
	teacher.GET("/students", std.List)
	teacher.GET("/homerooms", hr.List)

	// Leave Requests (สำหรับครูตรวจ/อนุมัติ)
	teacher.GET("/leave-requests", lv.List)
	teacher.GET("/leave-requests/pending-count", lv.PendingCount)
	teacher.POST("/leave-requests/:id/approve", lv.Approve)
	teacher.POST("/leave-requests/:id/reject", lv.Reject)

	// ===== Parent routes =====
	parent := e.Group("/parent", authMW, middlewares.RequireRole("parent"))
	parent.GET("/children", handlers.ParentChildren)
	parent.GET("/calendar/events", cal.ListEvents)
}
