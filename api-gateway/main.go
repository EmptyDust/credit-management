package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// TODO: 路由转发到各微服务
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.Run(":8080")
} 