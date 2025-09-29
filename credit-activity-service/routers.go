package main

import (
	"credit-management/credit-activity-service/handlers"

	"github.com/gin-gonic/gin"
)

// registerActivityOptionsRoute registers the public activity options endpoint
func registerActivityOptionsRoute(r *gin.Engine) {
	r.GET("/api/activities/config/options", handlers.GetActivityOptions)
}
