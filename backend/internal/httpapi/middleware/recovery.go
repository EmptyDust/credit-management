package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		requestID, _ := c.Get(RequestIDKey)
		logger.Error("panic recovered", "request_id", requestID, "panic", recovered)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":       "INTERNAL_SERVER_ERROR",
			"message":    "internal server error",
			"request_id": requestID,
		})
	})
}
