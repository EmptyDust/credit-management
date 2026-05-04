package router

import (
	"log/slog"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	activityhttp "credit-management/backend/internal/modules/activities/http"
	apphttp "credit-management/backend/internal/modules/applications/http"
	authhttp "credit-management/backend/internal/modules/auth/http"
	dashboardhttp "credit-management/backend/internal/modules/dashboard/http"
	devtoolshttp "credit-management/backend/internal/modules/devtools/http"
	filehttp "credit-management/backend/internal/modules/files/http"
	importhttp "credit-management/backend/internal/modules/imports/http"
	optionshttp "credit-management/backend/internal/modules/options/http"
	participanthttp "credit-management/backend/internal/modules/participants/http"
	usershttp "credit-management/backend/internal/modules/users/http"
	"credit-management/backend/internal/platform/authz"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Logger             *slog.Logger
	AuthHandler        *authhttp.Handler
	ActivityHandler    *activityhttp.Handler
	ApplicationHandler *apphttp.Handler
	AdminHandler       *usershttp.AdminHandler
	DashboardHandler   *dashboardhttp.Handler
	DevtoolsHandler    *devtoolshttp.Handler
	FileHandler        *filehttp.Handler
	ImportHandler      *importhttp.Handler
	MeHandler          *usershttp.MeHandler
	OptionsHandler     *optionshttp.Handler
	ParticipantHandler *participanthttp.Handler
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.RequestID(), middleware.Recovery(deps.Logger))

	healthHandler := func(c *gin.Context) {
		requestID, _ := c.Get(middleware.RequestIDKey)
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"service":    "backend",
			"request_id": requestID,
		})
	}

	engine.GET("/healthz", healthHandler)

	api := engine.Group("/api/v1")
	api.GET("/healthz", healthHandler)

	if deps.AuthHandler != nil {
		api.POST("/auth/login", deps.AuthHandler.Login)
		api.POST("/auth/register/student", deps.AuthHandler.RegisterStudent)
		api.POST("/auth/refresh", deps.AuthHandler.Refresh)
		api.POST("/auth/logout", middleware.RequireAuth(), deps.AuthHandler.Logout)
	}

	if deps.OptionsHandler != nil {
		api.GET("/options/organization/tree", deps.OptionsHandler.GetOrganizationTree)
		api.GET("/options/grades", deps.OptionsHandler.GetGrades)
	}

	protected := api.Group("")
	protected.Use(middleware.RequireAuth())

	teacherOrAdmin := protected.Group("")
	teacherOrAdmin.Use(middleware.RequireRoles(authz.RoleTeacher, authz.RoleAdmin))

	adminOnly := protected.Group("")
	adminOnly.Use(middleware.RequireRoles(authz.RoleAdmin))

	if deps.MeHandler != nil {
		protected.GET("/me/profile", deps.MeHandler.GetProfile)
		protected.PATCH("/me/profile", deps.MeHandler.UpdateProfile)
		protected.POST("/me/change-password", deps.MeHandler.ChangePassword)
		protected.POST("/me/avatar", deps.MeHandler.UploadAvatar)
		protected.DELETE("/me/avatar", deps.MeHandler.DeleteAvatar)
	}

	if deps.OptionsHandler != nil {
		protected.GET("/options/academic-years", deps.OptionsHandler.GetAcademicYears)
		protected.GET("/options/semesters", deps.OptionsHandler.GetSemesters)
		protected.GET("/options/teacher-titles", deps.OptionsHandler.GetTeacherTitles)
	}

	if deps.ActivityHandler != nil {
		protected.GET("/activity-categories", deps.ActivityHandler.ListCategories)
		protected.GET("/activity-categories/:category/schema", deps.ActivityHandler.GetCategorySchema)
		protected.GET("/activities", deps.ActivityHandler.List)
		protected.GET("/activities/stats", deps.ActivityHandler.Stats)
		protected.GET("/activities/export", deps.ActivityHandler.Export)
		protected.POST("/activities", deps.ActivityHandler.Create)
		protected.GET("/activities/:id", deps.ActivityHandler.Get)
		protected.PATCH("/activities/:id", deps.ActivityHandler.Update)
		protected.DELETE("/activities/:id", deps.ActivityHandler.Delete)
		protected.POST("/activities/:id/submit", deps.ActivityHandler.Submit)
		protected.POST("/activities/:id/withdraw", deps.ActivityHandler.Withdraw)
		protected.POST("/activities/:id/review", deps.ActivityHandler.Review)
		protected.POST("/activities/:id/copy", deps.ActivityHandler.Copy)
		protected.POST("/activities/:id/transfer-owner", deps.ActivityHandler.TransferOwner)
	}

	if deps.ParticipantHandler != nil {
		protected.GET("/activities/:id/participants", deps.ParticipantHandler.List)
		protected.POST("/activities/:id/participants", deps.ParticipantHandler.Add)
		protected.POST("/activities/:id/participants/leave", deps.ParticipantHandler.Leave)
	}

	if deps.FileHandler != nil {
		protected.GET("/activities/:id/files", deps.FileHandler.List)
		protected.POST("/activities/:id/files", deps.FileHandler.Upload)
		protected.DELETE("/activities/:id/files/:file_id", deps.FileHandler.Delete)
		protected.GET("/files/:file_id/download", deps.FileHandler.Download)
	}

	if deps.ApplicationHandler != nil {
		protected.GET("/applications", deps.ApplicationHandler.List)
		protected.GET("/applications/stats", deps.ApplicationHandler.Stats)
		protected.GET("/applications/export", deps.ApplicationHandler.Export)
		protected.GET("/applications/:id", deps.ApplicationHandler.Get)
	}

	if deps.DashboardHandler != nil {
		protected.GET("/dashboard/overview", deps.DashboardHandler.Overview)
	}

	if deps.AdminHandler != nil {
		teacherOrAdmin.GET("/students", deps.AdminHandler.ListStudents)
		teacherOrAdmin.GET("/students/export", deps.AdminHandler.ExportStudents)
		teacherOrAdmin.GET("/students/:id", deps.AdminHandler.GetStudent)

		adminOnly.POST("/students", deps.AdminHandler.CreateStudent)
		adminOnly.PATCH("/students/:id", deps.AdminHandler.UpdateStudent)
		adminOnly.DELETE("/students/:id", deps.AdminHandler.DeleteStudent)

		adminOnly.GET("/teachers", deps.AdminHandler.ListTeachers)
		adminOnly.GET("/teachers/export", deps.AdminHandler.ExportTeachers)
		adminOnly.GET("/teachers/:id", deps.AdminHandler.GetTeacher)
		adminOnly.POST("/teachers", deps.AdminHandler.CreateTeacher)
		adminOnly.PATCH("/teachers/:id", deps.AdminHandler.UpdateTeacher)
		adminOnly.DELETE("/teachers/:id", deps.AdminHandler.DeleteTeacher)

		adminOnly.PATCH("/users/:id/status", deps.AdminHandler.UpdateStatus)
		adminOnly.POST("/users/:id/reset-password", deps.AdminHandler.ResetPassword)
	}

	if deps.ImportHandler != nil {
		protected.POST("/imports/activities/precheck", deps.ImportHandler.ActivityPrecheck)
		protected.POST("/imports/activities/confirm", deps.ImportHandler.ActivityConfirm)

		adminOnly.GET("/imports/templates/students", deps.ImportHandler.StudentTemplate)
		adminOnly.GET("/imports/templates/teachers", deps.ImportHandler.TeacherTemplate)
		adminOnly.POST("/imports/students/precheck", deps.ImportHandler.PrecheckStudents)
		adminOnly.POST("/imports/students/confirm", deps.ImportHandler.ConfirmStudents)
		adminOnly.POST("/imports/teachers/precheck", deps.ImportHandler.PrecheckTeachers)
		adminOnly.POST("/imports/teachers/confirm", deps.ImportHandler.ConfirmTeachers)
	}

	if deps.DevtoolsHandler != nil {
		adminOnly.GET("/devtools/health", deps.DevtoolsHandler.DetailedHealth)
		adminOnly.GET("/devtools/logs", deps.DevtoolsHandler.Logs)
		adminOnly.GET("/devtools/audit-logs", deps.DevtoolsHandler.AuditLogs)
		adminOnly.GET("/devtools/audit-logs/export", deps.DevtoolsHandler.ExportAuditLogs)
	}

	return engine
}
