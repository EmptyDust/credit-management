package router

import (
	"log/slog"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	authhttp "credit-management/backend/internal/modules/auth/http"
	optionshttp "credit-management/backend/internal/modules/options/http"
	usershttp "credit-management/backend/internal/modules/users/http"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Logger         *slog.Logger
	AuthHandler    *authhttp.Handler
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

	return engine
}
