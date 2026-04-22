package router

import (
	"log/slog"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	authhttp "credit-management/backend/internal/modules/auth/http"
	importhttp "credit-management/backend/internal/modules/imports/http"
	optionshttp "credit-management/backend/internal/modules/options/http"
	usershttp "credit-management/backend/internal/modules/users/http"
	"credit-management/backend/internal/platform/authz"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Logger         *slog.Logger
	AuthHandler    *authhttp.Handler
	AdminHandler   *usershttp.AdminHandler
	ImportHandler  *importhttp.UserImportHandler
	MeHandler      *usershttp.MeHandler
	OptionsHandler *optionshttp.Handler
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
		adminOnly.GET("/imports/templates/students", deps.ImportHandler.StudentTemplate)
		adminOnly.GET("/imports/templates/teachers", deps.ImportHandler.TeacherTemplate)
		adminOnly.POST("/imports/students/precheck", deps.ImportHandler.PrecheckStudents)
		adminOnly.POST("/imports/students/confirm", deps.ImportHandler.ConfirmStudents)
		adminOnly.POST("/imports/teachers/precheck", deps.ImportHandler.PrecheckTeachers)
		adminOnly.POST("/imports/teachers/confirm", deps.ImportHandler.ConfirmTeachers)
	}

	return engine
}
