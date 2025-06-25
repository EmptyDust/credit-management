package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

type ProxyConfig struct {
	UserServiceURL           string
	AuthServiceURL           string
	StudentServiceURL        string
	TeacherServiceURL        string
	CreditActivityServiceURL string
}

func main() {
	// 获取服务URL配置
	config := ProxyConfig{
		UserServiceURL:           getEnv("USER_SERVICE_URL", "http://user-service:8084"),
		AuthServiceURL:           getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		StudentServiceURL:        getEnv("STUDENT_SERVICE_URL", "http://student-service:8085"),
		TeacherServiceURL:        getEnv("TEACHER_SERVICE_URL", "http://teacher-service:8086"),
		CreditActivityServiceURL: getEnv("CREDIT_ACTIVITY_SERVICE_URL", "http://credit-activity-service:8083"),
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
		api.GET("/auth/validate-permission", createProxyHandler(config.AuthServiceURL))

		// 权限管理服务路由
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
		api.POST("/users/teachers", createProxyHandler(config.UserServiceURL))
		api.GET("/users/stats", createProxyHandler(config.UserServiceURL))
		api.GET("/users/profile", createProxyHandler(config.UserServiceURL))
		api.PUT("/users/profile", createProxyHandler(config.UserServiceURL))
		api.GET("/users", createProxyHandler(config.UserServiceURL))
		api.GET("/users/type/:userType", createProxyHandler(config.UserServiceURL))
		api.GET("/users/:id", createProxyHandler(config.UserServiceURL))
		api.PUT("/users/:id", createProxyHandler(config.UserServiceURL))
		api.DELETE("/users/:id", createProxyHandler(config.UserServiceURL))
		api.Any("/notifications/*path", createProxyHandler(config.UserServiceURL))

		// 学生信息服务路由
		api.POST("/students", createProxyHandler(config.StudentServiceURL))
		api.GET("/students", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/search", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/search/username", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/college/:college", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/major/:major", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/class/:class", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/status/:status", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/:id", createProxyHandler(config.StudentServiceURL))
		api.PUT("/students/:id", createProxyHandler(config.StudentServiceURL))
		api.DELETE("/students/:id", createProxyHandler(config.StudentServiceURL))
		api.GET("/students/user/:userID", createProxyHandler(config.StudentServiceURL))
		api.DELETE("/students/user/:user_id", createProxyHandler(config.StudentServiceURL))

		// 教师信息服务路由
		api.POST("/teachers", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/:id", createProxyHandler(config.TeacherServiceURL))
		api.PUT("/teachers/:id", createProxyHandler(config.TeacherServiceURL))
		api.DELETE("/teachers/:id", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/department/:department", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/title/:title", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/status/:status", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/search", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/search/username", createProxyHandler(config.TeacherServiceURL))
		api.GET("/teachers/active", createProxyHandler(config.TeacherServiceURL))
		api.DELETE("/teachers/user/:user_id", createProxyHandler(config.TeacherServiceURL))

		// 学分活动服务路由
		api.GET("/activities/categories", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/stats", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.DELETE("/activities/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/submit", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/withdraw", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/review", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/pending", createProxyHandler(config.CreditActivityServiceURL))

		// 学分活动参与者管理
		api.POST("/activities/:id/participants", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id/participants", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/:id/participants/batch-credits", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/:id/participants/:user_id/credits", createProxyHandler(config.CreditActivityServiceURL))
		api.DELETE("/activities/:id/participants/:user_id", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/leave", createProxyHandler(config.CreditActivityServiceURL))

		// 学分活动申请管理
		api.GET("/activities/applications", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/applications/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/applications/all", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/applications/export", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/applications/stats", createProxyHandler(config.CreditActivityServiceURL))
	}

	// 默认路由 - 返回API信息
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Credit Management API Gateway",
			"version": "1.0.0",
			"services": gin.H{
				"auth_service":            config.AuthServiceURL,
				"user_service":            config.UserServiceURL,
				"student_service":         config.StudentServiceURL,
				"teacher_service":         config.TeacherServiceURL,
				"credit_activity_service": config.CreditActivityServiceURL,
			},
			"endpoints": gin.H{
				"auth":          "/api/auth",
				"permissions":   "/api/permissions",
				"users":         "/api/users",
				"notifications": "/api/notifications",
				"students":      "/api/students",
				"teachers":      "/api/teachers",
				"activities":    "/api/activities",
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
	port := getEnv("PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
	log.Printf("Auth Service: %s", config.AuthServiceURL)
	log.Printf("User Service: %s", config.UserServiceURL)
	log.Printf("Student Service: %s", config.StudentServiceURL)
	log.Printf("Teacher Service: %s", config.TeacherServiceURL)
	log.Printf("Credit Activity Service: %s", config.CreditActivityServiceURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

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
		originalPath := c.Request.URL.Path
		c.Request.URL.Path = originalPath
		c.Request.URL.Host = target.Host
		c.Request.URL.Scheme = target.Scheme

		// 执行代理请求
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
