package router

import (
	"log/slog"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func New(logger *slog.Logger) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.RequestID(), middleware.Recovery(logger))

	engine.GET("/healthz", func(c *gin.Context) {
		requestID, _ := c.Get(middleware.RequestIDKey)
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"service":    "backend",
			"request_id": requestID,
		})
	})

	return engine
}
