package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// 用户管理服务路由
	userService := r.Group("/api/users")
	{
		userService.POST("/register", proxyToService("http://user-management-service:8081"))
		userService.POST("/login", proxyToService("http://user-management-service:8081"))
		userService.GET("/:id", proxyToService("http://user-management-service:8081"))
		userService.PUT("/:id", proxyToService("http://user-management-service:8081"))
		userService.DELETE("/:id", proxyToService("http://user-management-service:8081"))
	}

	// 学生信息服务路由
	studentService := r.Group("/api/students")
	{
		studentService.POST("/", proxyToService("http://student-info-service:8084"))
		studentService.GET("/", proxyToService("http://student-info-service:8084"))
		studentService.GET("/:id", proxyToService("http://student-info-service:8084"))
		studentService.GET("/user/:userID", proxyToService("http://student-info-service:8084"))
		studentService.PUT("/:id", proxyToService("http://student-info-service:8084"))
		studentService.DELETE("/:id", proxyToService("http://student-info-service:8084"))
	}

	// 教师信息服务路由
	teacherService := r.Group("/api/teachers")
	{
		teacherService.POST("/", proxyToService("http://teacher-info-service:8085"))
		teacherService.GET("/", proxyToService("http://teacher-info-service:8085"))
		teacherService.GET("/:id", proxyToService("http://teacher-info-service:8085"))
		teacherService.GET("/user/:userID", proxyToService("http://teacher-info-service:8085"))
		teacherService.PUT("/:id", proxyToService("http://teacher-info-service:8085"))
		teacherService.DELETE("/:id", proxyToService("http://teacher-info-service:8085"))
	}

	// 事项管理服务路由
	affairService := r.Group("/api/affairs")
	{
		affairService.POST("/", proxyToService("http://affair-management-service:8083"))
		affairService.GET("/", proxyToService("http://affair-management-service:8083"))
		affairService.GET("/:id", proxyToService("http://affair-management-service:8083"))
		affairService.PUT("/:id", proxyToService("http://affair-management-service:8083"))
		affairService.DELETE("/:id", proxyToService("http://affair-management-service:8083"))
	}

	// 通用申请服务路由
	applicationService := r.Group("/api/applications")
	{
		applicationService.POST("/", proxyToService("http://general-application-service:8086"))
		applicationService.GET("/", proxyToService("http://general-application-service:8086"))
		applicationService.GET("/:id", proxyToService("http://general-application-service:8086"))
		applicationService.GET("/user/:userID", proxyToService("http://general-application-service:8086"))
		applicationService.GET("/student/:studentID", proxyToService("http://general-application-service:8086"))
		applicationService.PUT("/:id", proxyToService("http://general-application-service:8086"))
		applicationService.POST("/:id/review", proxyToService("http://general-application-service:8086"))
		applicationService.DELETE("/:id", proxyToService("http://general-application-service:8086"))
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway starting on port %s", port)
	r.Run(":" + port)
}

// proxyToService 创建代理中间件
func proxyToService(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里应该实现实际的代理逻辑
		// 为了简化，我们直接返回一个响应
		c.JSON(200, gin.H{
			"message": "Proxied to " + targetURL,
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	}
}
