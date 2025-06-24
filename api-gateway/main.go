package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

type ProxyConfig struct {
	UserServiceURL        string
	AuthServiceURL        string
	StudentServiceURL     string
	TeacherServiceURL     string
	AffairServiceURL      string
	ApplicationServiceURL string
}

func main() {
	// 获取服务URL配置
	config := ProxyConfig{
		UserServiceURL:        getEnv("USER_SERVICE_URL", "http://user-management-service:8080"),
		AuthServiceURL:        getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		StudentServiceURL:     getEnv("STUDENT_SERVICE_URL", "http://student-info-service:8084"),
		TeacherServiceURL:     getEnv("TEACHER_SERVICE_URL", "http://teacher-info-service:8085"),
		AffairServiceURL:      getEnv("AFFAIR_SERVICE_URL", "http://affair-management-service:8087"),
		ApplicationServiceURL: getEnv("APPLICATION_SERVICE_URL", "http://application-management-service:8082"),
	}

	// 设置Gin路由
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 添加日志中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "api-gateway",
			"version": "1.0.0",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证服务路由
		api.Any("/auth/*path", createProxyHandler(config.AuthServiceURL))

		// 权限管理服务路由
		// api.Any("/permissions/*path", createProxyHandler(config.AuthServiceURL))
		// 明确列出权限相关路由，代理到权限服务
		api.POST("/permissions/init", createProxyHandler(config.AuthServiceURL))
		api.POST("/permissions/roles", createProxyHandler(config.AuthServiceURL))
		api.GET("/permissions/roles", createProxyHandler(config.AuthServiceURL))
		api.GET("/permissions/roles/:roleID", createProxyHandler(config.AuthServiceURL))
		api.PUT("/permissions/roles/:roleID", createProxyHandler(config.AuthServiceURL))
		api.DELETE("/permissions/roles/:roleID", createProxyHandler(config.AuthServiceURL))
		api.POST("/permissions", createProxyHandler(config.AuthServiceURL))
		api.GET("/permissions", createProxyHandler(config.AuthServiceURL))
		api.GET("/permissions/:id", createProxyHandler(config.AuthServiceURL))
		api.DELETE("/permissions/:id", createProxyHandler(config.AuthServiceURL))
		api.POST("/permissions/users/:userID/roles", createProxyHandler(config.AuthServiceURL))
		api.DELETE("/permissions/users/:userID/roles/:roleID", createProxyHandler(config.AuthServiceURL))
		api.POST("/permissions/users/:userID/permissions", createProxyHandler(config.AuthServiceURL))
		api.DELETE("/permissions/users/:userID/permissions/:permissionID", createProxyHandler(config.AuthServiceURL))
		api.POST("/permissions/roles/:roleID/permissions", createProxyHandler(config.AuthServiceURL))
		api.DELETE("/permissions/roles/:roleID/permissions/:permissionID", createProxyHandler(config.AuthServiceURL))
		api.GET("/permissions/users/:userID/roles", createProxyHandler(config.AuthServiceURL))
		api.GET("/permissions/users/:userID/permissions", createProxyHandler(config.AuthServiceURL))

		// 用户管理服务路由
		api.POST("/users/register", createProxyHandler(config.UserServiceURL))
		api.GET("/users/stats", createProxyHandler(config.UserServiceURL))
		api.GET("/users/profile", createProxyHandler(config.UserServiceURL))
		api.PUT("/users/profile", createProxyHandler(config.UserServiceURL))
		api.GET("/users", createProxyHandler(config.UserServiceURL))
		api.GET("/users/type/:userType", createProxyHandler(config.UserServiceURL))
		api.GET("/users/:username", createProxyHandler(config.UserServiceURL))
		api.PUT("/users/:username", createProxyHandler(config.UserServiceURL))
		api.DELETE("/users/:username", createProxyHandler(config.UserServiceURL))
		api.Any("/notifications/*path", createProxyHandler(config.UserServiceURL))

		// 事项管理服务路由
		api.GET("/affairs", createProxyHandler(config.AffairServiceURL))
		api.POST("/affairs", createProxyHandler(config.AffairServiceURL))
		api.GET("/affairs/:id/participants", createProxyHandler(config.AffairServiceURL))
		api.GET("/affairs/:id/applications", createProxyHandler(config.AffairServiceURL))
		api.GET("/affairs/:id", createProxyHandler(config.AffairServiceURL))
		api.PUT("/affairs/:id", createProxyHandler(config.AffairServiceURL))
		api.DELETE("/affairs/:id", createProxyHandler(config.AffairServiceURL))

		// 学生信息服务路由
		api.POST("/students", createProxyHandler(config.StudentServiceURL))
		api.GET("/students", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/:studentID", createProxyHandler(config.StudentServiceURL))
		api.PUT("/students/:studentID", createProxyHandler(config.StudentServiceURL))
		api.DELETE("/students/:studentID", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/college/:college", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/major/:major", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/class/:class", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/status/:status", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/search", createProxyHandler(config.StudentServiceURL))

		// 教师信息服务路由
		api.POST("/teachers", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/:username", createProxyHandler(config.TeacherServiceURL))
		api.PUT("/teachers/:username", createProxyHandler(config.TeacherServiceURL))
		api.DELETE("/teachers/:username", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/department/:department", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/title/:title", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/status/:status", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/search", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/active", createProxyHandler(config.TeacherServiceURL))

		// 申请管理服务路由
		api.POST("/applications", createProxyHandler(config.ApplicationServiceURL))
		api.POST("/applications/batch", createProxyHandler(config.ApplicationServiceURL))
		api.GET("/applications/:id/detail", createProxyHandler(config.ApplicationServiceURL))
		api.PUT("/applications/:id/details", createProxyHandler(config.ApplicationServiceURL))
		api.POST("/applications/:id/submit", createProxyHandler(config.ApplicationServiceURL))
		api.PUT("/applications/:id/status", createProxyHandler(config.ApplicationServiceURL))
		api.GET("/applications/:id", createProxyHandler(config.ApplicationServiceURL))
		api.GET("/applications/user/:studentNumber", createProxyHandler(config.ApplicationServiceURL))
		api.GET("/applications", createProxyHandler(config.ApplicationServiceURL))
	}

	// 默认路由 - 返回API信息
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Credit Management API Gateway",
			"version": "1.0.0",
			"services": gin.H{
				"auth_service":        config.AuthServiceURL,
				"user_service":        config.UserServiceURL,
				"student_service":     config.StudentServiceURL,
				"teacher_service":     config.TeacherServiceURL,
				"affair_service":      config.AffairServiceURL,
				"application_service": config.ApplicationServiceURL,
			},
			"endpoints": gin.H{
				"auth":          "/api/auth",
				"permissions":   "/api/permissions",
				"users":         "/api/users",
				"notifications": "/api/notifications",
				"students":      "/api/students",
				"teachers":      "/api/teachers",
				"affairs":       "/api/affairs",
				"applications":  "/api/applications",
				"health":        "/health",
			},
		})
	})

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "Not Found",
			"message": "The requested endpoint does not exist",
			"path":    c.Request.URL.Path,
		})
	})

	// 启动服务器
	port := getEnv("PORT", "8000")
	log.Printf("API Gateway starting on port %s", port)
	log.Printf("Auth Service: %s", config.AuthServiceURL)
	log.Printf("User Service: %s", config.UserServiceURL)
	log.Printf("Student Service: %s", config.StudentServiceURL)
	log.Printf("Teacher Service: %s", config.TeacherServiceURL)
	log.Printf("Affair Service: %s", config.AffairServiceURL)
	log.Printf("Application Service: %s", config.ApplicationServiceURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// createProxyHandler 创建代理处理器
func createProxyHandler(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Proxying request for %s to %s", c.Request.URL.Path, targetURL)
		target, err := url.Parse(targetURL)
		if err != nil {
			c.JSON(500, gin.H{"error": "Invalid target URL"})
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		c.Request.Header.Set("X-Forwarded-Host", c.Request.Host)
		c.Request.Header.Set("X-Forwarded-Proto", "http")
		c.Request.Header.Set("X-Real-IP", c.ClientIP())

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
