package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProxyConfig struct {
	UserServiceURL        string
	StudentServiceURL     string
	TeacherServiceURL     string
	AffairServiceURL      string
	ApplicationServiceURL string
}

func main() {
	// 获取服务URL配置
	config := ProxyConfig{
		UserServiceURL:        getEnv("USER_SERVICE_URL", "http://user-management-service:8081"),
		StudentServiceURL:     getEnv("STUDENT_SERVICE_URL", "http://student-info-service:8082"),
		TeacherServiceURL:     getEnv("TEACHER_SERVICE_URL", "http://teacher-info-service:8083"),
		AffairServiceURL:      getEnv("AFFAIR_SERVICE_URL", "http://affair-management-service:8087"),
		ApplicationServiceURL: getEnv("APPLICATION_SERVICE_URL", "http://general-application-service:8086"),
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
		// 用户管理服务路由
		api.Any("/users/*path", createProxyHandler(config.UserServiceURL))
		api.Any("/user/*path", createProxyHandler(config.UserServiceURL))

		// 学生信息服务路由
		api.Any("/students/*path", createProxyHandler(config.StudentServiceURL))
		api.Any("/student/*path", createProxyHandler(config.StudentServiceURL))

		// 教师信息服务路由
		api.Any("/teachers/*path", createProxyHandler(config.TeacherServiceURL))
		api.Any("/teacher/*path", createProxyHandler(config.TeacherServiceURL))

		// 事项管理服务路由
		api.Any("/affairs/*path", createProxyHandler(config.AffairServiceURL))
		api.Any("/affair/*path", createProxyHandler(config.AffairServiceURL))
		api.Any("/affair-students/*path", createProxyHandler(config.AffairServiceURL))

		// 申请服务路由
		api.Any("/applications/*path", createProxyHandler(config.ApplicationServiceURL))
		api.Any("/application/*path", createProxyHandler(config.ApplicationServiceURL))
	}

	// 默认路由 - 返回API信息
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Credit Management API Gateway",
			"version": "1.0.0",
			"services": gin.H{
				"user_service":        config.UserServiceURL,
				"student_service":     config.StudentServiceURL,
				"teacher_service":     config.TeacherServiceURL,
				"affair_service":      config.AffairServiceURL,
				"application_service": config.ApplicationServiceURL,
			},
			"endpoints": gin.H{
				"users":        "/api/users",
				"students":     "/api/students",
				"teachers":     "/api/teachers",
				"affairs":      "/api/affairs",
				"applications": "/api/applications",
				"health":       "/health",
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
	port := getEnv("PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
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
		// 解析目标URL
		target, err := url.Parse(targetURL)
		if err != nil {
			c.JSON(500, gin.H{"error": "Invalid target URL"})
			return
		}

		// 创建反向代理
		proxy := httputil.NewSingleHostReverseProxy(target)

		// 修改请求路径
		originalPath := c.Param("path")
		if originalPath == "" {
			originalPath = c.Request.URL.Path
		}

		// 移除前缀，只保留API路径
		path := strings.TrimPrefix(originalPath, "/api/")
		path = strings.TrimPrefix(path, "users/")
		path = strings.TrimPrefix(path, "user/")
		path = strings.TrimPrefix(path, "students/")
		path = strings.TrimPrefix(path, "student/")
		path = strings.TrimPrefix(path, "teachers/")
		path = strings.TrimPrefix(path, "teacher/")
		path = strings.TrimPrefix(path, "affairs/")
		path = strings.TrimPrefix(path, "affair/")
		path = strings.TrimPrefix(path, "affair-students/")
		path = strings.TrimPrefix(path, "applications/")
		path = strings.TrimPrefix(path, "application/")

		// 设置新的请求路径
		c.Request.URL.Path = "/api/" + path

		// 添加请求头
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Host)
		c.Request.Header.Set("X-Forwarded-Proto", "http")
		c.Request.Header.Set("X-Real-IP", c.ClientIP())

		// 执行代理请求
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
