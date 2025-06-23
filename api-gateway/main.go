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
		api.Any("/permissions/*path", createProxyHandler(config.AuthServiceURL))
		api.POST("/init-permissions", createProxyHandler(config.AuthServiceURL))

		// 用户管理服务路由（register用POST方法注册）
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
		api.Any("/affairs/*path", createProxyHandler(config.AffairServiceURL))

		// 学生信息服务路由
		api.POST("/students", createProxyHandler(config.StudentServiceURL))
		api.Any("/students/*path", createProxyHandler(config.StudentServiceURL))

		// 教师信息服务路由
		api.POST("/teachers", createProxyHandler(config.TeacherServiceURL))
		api.Any("/teachers/*path", createProxyHandler(config.TeacherServiceURL))

		// 申请管理服务路由
		api.POST("/applications", createProxyHandler(config.ApplicationServiceURL))
		api.Any("/applications/*path", createProxyHandler(config.ApplicationServiceURL))
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
