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
	UserServiceURL           string
	AuthServiceURL           string
	CreditActivityServiceURL string
}

func main() {
	// 获取服务URL配置
	config := ProxyConfig{
		UserServiceURL:           getEnv("USER_SERVICE_URL", "http://user-service:8084"),
		AuthServiceURL:           getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
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
			"version": "2.0.0",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证服务路由
		api.Any("/auth/*path", createProxyHandler(config.AuthServiceURL))

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

		// 统一用户服务路由（包含用户管理、学生信息、教师信息）
		api.POST("/users/register", createProxyHandler(config.UserServiceURL))
		api.POST("/users/teachers", createProxyHandler(config.UserServiceURL))
		api.POST("/users/students", createProxyHandler(config.UserServiceURL))
		api.GET("/users/stats", createProxyHandler(config.UserServiceURL))
		api.GET("/users/stats/students", createProxyHandler(config.UserServiceURL))
		api.GET("/users/stats/teachers", createProxyHandler(config.UserServiceURL))
		api.GET("/users/profile", createProxyHandler(config.UserServiceURL))
		api.PUT("/users/profile", createProxyHandler(config.UserServiceURL))
		api.GET("/users/:id", createProxyHandler(config.UserServiceURL))
		api.PUT("/users/:id", createProxyHandler(config.UserServiceURL))
		api.DELETE("/users/:id", createProxyHandler(config.UserServiceURL))

		// 学生相关路由（统一用户服务）
		api.POST("/students", createProxyHandler(config.UserServiceURL))
		api.PUT("/students/:id", createProxyHandler(config.UserServiceURL))
		api.DELETE("/students/:id", createProxyHandler(config.UserServiceURL))

		// 教师相关路由（统一用户服务）
		api.POST("/teachers", createProxyHandler(config.UserServiceURL))
		api.GET("/teachers/:id", createProxyHandler(config.UserServiceURL))
		api.PUT("/teachers/:id", createProxyHandler(config.UserServiceURL))
		api.DELETE("/teachers/:id", createProxyHandler(config.UserServiceURL))

		// 搜索路由（统一用户服务）
		api.GET("/search/users", createProxyHandler(config.UserServiceURL))

		// 学分活动服务路由 - 完整统一版本
		// 1. 活动管理基础路由
		api.GET("/activities/categories", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/templates", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/batch", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/stats", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.DELETE("/activities/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/submit", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/withdraw", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/review", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/pending", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/copy", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/save-template", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/deletable", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/batch-delete", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/import-csv", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/csv-template", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/export", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/report", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/batch", createProxyHandler(config.CreditActivityServiceURL))

		// 2. 活动参与者管理路由
		api.POST("/activities/:id/participants", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id/participants", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/:id/participants/batch-credits", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/:id/participants/:user_id/credits", createProxyHandler(config.CreditActivityServiceURL))
		api.DELETE("/activities/:id/participants/:user_id", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/participants/batch-remove", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id/participants/stats", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id/participants/export", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/leave", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id/my-activities", createProxyHandler(config.CreditActivityServiceURL))

		// 3. 活动附件管理路由
		api.GET("/activities/:id/attachments", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/attachments", createProxyHandler(config.CreditActivityServiceURL))
		api.POST("/activities/:id/attachments/batch", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id/attachments/:attachment_id/download", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/activities/:id/attachments/:attachment_id/preview", createProxyHandler(config.CreditActivityServiceURL))
		api.PUT("/activities/:id/attachments/:attachment_id", createProxyHandler(config.CreditActivityServiceURL))
		api.DELETE("/activities/:id/attachments/:attachment_id", createProxyHandler(config.CreditActivityServiceURL))

		// 4. 申请管理路由（独立路由组）
		api.GET("/applications", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/applications/:id", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/applications/stats", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/applications/export", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/applications/all", createProxyHandler(config.CreditActivityServiceURL))

		// 5. 统一检索API路由组
		api.GET("/search/activities", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/search/applications", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/search/participants", createProxyHandler(config.CreditActivityServiceURL))
		api.GET("/search/attachments", createProxyHandler(config.CreditActivityServiceURL))
	}

	// 默认路由 - 返回API信息
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Credit Management API Gateway",
			"version": "2.0.0",
			"services": gin.H{
				"auth_service":            config.AuthServiceURL,
				"user_service":            config.UserServiceURL,
				"credit_activity_service": config.CreditActivityServiceURL,
			},
			"endpoints": gin.H{
				"auth":        "/api/auth",
				"permissions": "/api/permissions",
				"users":       "/api/users",
				"students":    "/api/students",
				"teachers":    "/api/teachers",
				"search":      "/api/search",
				"activities":  "/api/activities",
				"health":      "/health",
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
	log.Printf("Credit Activity Service: %s", config.CreditActivityServiceURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start API Gateway:", err)
	}
}

// 创建代理处理器
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

		// 特殊处理：为预览和下载路由添加认证支持
		path := c.Request.URL.Path
		if strings.Contains(path, "/attachments/") && (strings.Contains(path, "/preview") || strings.Contains(path, "/download")) {
			// 检查是否有Authorization头
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				// 尝试从URL参数获取token
				token := c.Query("token")
				if token != "" {
					c.Request.Header.Set("Authorization", "Bearer "+token)
				}
			}
		}

		// 修改请求路径 - 保持原始路径
		c.Request.URL.Host = target.Host
		c.Request.URL.Scheme = target.Scheme

		// 执行代理请求
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
