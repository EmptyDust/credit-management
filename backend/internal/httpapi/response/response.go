package response

import (
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, data any) {
	requestID, _ := c.Get(middleware.RequestIDKey)
	c.JSON(status, gin.H{
		"code":       "OK",
		"message":    "success",
		"data":       data,
		"request_id": requestID,
	})
}

func OK(c *gin.Context, data any) {
	Success(c, http.StatusOK, data)
}

func Error(c *gin.Context, status int, code string, message string, details any) {
	requestID, _ := c.Get(middleware.RequestIDKey)
	payload := gin.H{
		"code":       code,
		"message":    message,
		"request_id": requestID,
	}
	if details != nil {
		payload["details"] = details
	}

	c.AbortWithStatusJSON(status, payload)
}
