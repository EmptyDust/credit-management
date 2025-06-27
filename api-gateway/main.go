package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ProxyConfig struct {
	UserServiceURL           string
	AuthServiceURL           string
	CreditActivityServiceURL string
	JWTSecret                string
}

// JWTClaims 自定义JWT claims结构
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

// AuthMiddleware JWT认证中间件
type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// AuthRequired 认证必需中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证令牌",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证格式",
				"data":    nil,
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 解析JWT token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌: " + err.Error(),
				"data":    nil,
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌已过期",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 提取claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证信息",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 验证用户ID
		if claims.UserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token中缺少用户ID",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_type", claims.UserType)
		c.Set("claims", claims)

		c.Next()
	}
}

// PermissionMiddleware 权限中间件
type PermissionMiddleware struct{}

func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{}
}

// AdminOnly 仅管理员权限
func (m *PermissionMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		if userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要管理员权限",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TeacherOrAdmin 教师或管理员权限
func (m *PermissionMiddleware) TeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		if userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要教师或管理员权限",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StudentOnly 仅学生权限
func (m *PermissionMiddleware) StudentOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		if userType != "student" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要学生权限",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	// 获取服务URL配置
	config := ProxyConfig{
		UserServiceURL:           getEnv("USER_SERVICE_URL", "http://user-service:8084"),
		AuthServiceURL:           getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		CreditActivityServiceURL: getEnv("CREDIT_ACTIVITY_SERVICE_URL", "http://credit-activity-service:8083"),
		JWTSecret:                getEnv("JWT_SECRET", "your-secret-key"),
	}

	// 创建中间件
	authMiddleware := NewAuthMiddleware(config.JWTSecret)
	permissionMiddleware := NewPermissionMiddleware()

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
			"version": "3.0.0",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证服务路由（无需认证）
		api.Any("/auth/*path", createProxyHandler(config.AuthServiceURL))

		// 权限管理服务路由（需要管理员权限）
		permissions := api.Group("/permissions")
		permissions.Use(authMiddleware.AuthRequired())
		permissions.Use(permissionMiddleware.AdminOnly())
		{
			permissions.POST("/init", createProxyHandler(config.AuthServiceURL))
			permissions.POST("/roles", createProxyHandler(config.AuthServiceURL))
			permissions.GET("/roles", createProxyHandler(config.AuthServiceURL))
			permissions.GET("/roles/:roleID", createProxyHandler(config.AuthServiceURL))
			permissions.PUT("/roles/:roleID", createProxyHandler(config.AuthServiceURL))
			permissions.DELETE("/roles/:roleID", createProxyHandler(config.AuthServiceURL))
			permissions.POST("", createProxyHandler(config.AuthServiceURL))
			permissions.GET("", createProxyHandler(config.AuthServiceURL))
			permissions.GET("/:id", createProxyHandler(config.AuthServiceURL))
			permissions.DELETE("/:id", createProxyHandler(config.AuthServiceURL))
			permissions.POST("/users/:userID/roles", createProxyHandler(config.AuthServiceURL))
			permissions.DELETE("/users/:userID/roles/:roleID", createProxyHandler(config.AuthServiceURL))
			permissions.POST("/users/:userID/permissions", createProxyHandler(config.AuthServiceURL))
			permissions.DELETE("/users/:userID/permissions/:permissionID", createProxyHandler(config.AuthServiceURL))
			permissions.POST("/roles/:roleID/permissions", createProxyHandler(config.AuthServiceURL))
			permissions.DELETE("/roles/:roleID/permissions/:permissionID", createProxyHandler(config.AuthServiceURL))
			permissions.GET("/users/:userID/roles", createProxyHandler(config.AuthServiceURL))
			permissions.GET("/users/:userID/permissions", createProxyHandler(config.AuthServiceURL))
		}

		// 统一用户服务路由（需要认证）
		users := api.Group("/users")
		users.Use(authMiddleware.AuthRequired())
		{
			// 注册路由（无需权限）
			users.POST("/register", createProxyHandler(config.UserServiceURL))

			// 管理员路由
			admin := users.Group("")
			admin.Use(permissionMiddleware.AdminOnly())
			{
				admin.POST("/teachers", createProxyHandler(config.UserServiceURL))
				admin.POST("/students", createProxyHandler(config.UserServiceURL))
				admin.PUT("/:id", createProxyHandler(config.UserServiceURL))
				admin.DELETE("/:id", createProxyHandler(config.UserServiceURL))
				admin.POST("/batch_delete", createProxyHandler(config.UserServiceURL))
				admin.POST("/batch_status", createProxyHandler(config.UserServiceURL))
				admin.POST("/reset_password", createProxyHandler(config.UserServiceURL))
				admin.GET("/export", createProxyHandler(config.UserServiceURL))
				admin.POST("/import-csv", createProxyHandler(config.UserServiceURL))
				admin.GET("/csv-template", createProxyHandler(config.UserServiceURL))
				admin.POST("/import", createProxyHandler(config.UserServiceURL))
				admin.GET("/excel-template", createProxyHandler(config.UserServiceURL))
			}

			// 教师或管理员路由
			teacherOrAdmin := users.Group("")
			teacherOrAdmin.Use(permissionMiddleware.TeacherOrAdmin())
			{
				teacherOrAdmin.GET("/stats/students", createProxyHandler(config.UserServiceURL))
				teacherOrAdmin.GET("/stats/teachers", createProxyHandler(config.UserServiceURL))
			}

			// 所有认证用户都可以访问的路由
			users.GET("/stats", createProxyHandler(config.UserServiceURL))
			users.GET("/profile", createProxyHandler(config.UserServiceURL))
			users.PUT("/profile", createProxyHandler(config.UserServiceURL))
			users.GET("/:id", createProxyHandler(config.UserServiceURL))
			users.POST("/change_password", createProxyHandler(config.UserServiceURL))
			users.GET("/activity", createProxyHandler(config.UserServiceURL))
			users.GET("/:id/activity", createProxyHandler(config.UserServiceURL))
		}

		// 学生相关路由（需要管理员权限）
		students := api.Group("/students")
		students.Use(authMiddleware.AuthRequired())
		students.Use(permissionMiddleware.AdminOnly())
		{
			students.POST("", createProxyHandler(config.UserServiceURL))
			students.PUT("/:id", createProxyHandler(config.UserServiceURL))
			students.DELETE("/:id", createProxyHandler(config.UserServiceURL))
		}

		// 教师相关路由（需要管理员权限）
		teachers := api.Group("/teachers")
		teachers.Use(authMiddleware.AuthRequired())
		teachers.Use(permissionMiddleware.AdminOnly())
		{
			teachers.POST("", createProxyHandler(config.UserServiceURL))
			teachers.GET("/:id", createProxyHandler(config.UserServiceURL))
			teachers.PUT("/:id", createProxyHandler(config.UserServiceURL))
			teachers.DELETE("/:id", createProxyHandler(config.UserServiceURL))
		}

		// 搜索路由（需要认证）
		search := api.Group("/search")
		search.Use(authMiddleware.AuthRequired())
		{
			search.GET("/users", createProxyHandler(config.UserServiceURL))
		}

		// 学分活动服务路由（需要认证）
		activities := api.Group("/activities")
		activities.Use(authMiddleware.AuthRequired())
		{
			// 基础路由（所有认证用户）
			activities.GET("/categories", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("/templates", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("/stats", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("/:id", createProxyHandler(config.CreditActivityServiceURL))
			activities.POST("/:id/submit", createProxyHandler(config.CreditActivityServiceURL))
			activities.POST("/:id/withdraw", createProxyHandler(config.CreditActivityServiceURL))
			activities.POST("/:id/copy", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("/:id/my-activities", createProxyHandler(config.CreditActivityServiceURL))

			// 教师或管理员路由
			teacherOrAdmin := activities.Group("")
			teacherOrAdmin.Use(permissionMiddleware.TeacherOrAdmin())
			{
				teacherOrAdmin.POST("", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.POST("/batch", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.PUT("/:id", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.POST("/:id/review", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.GET("/pending", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.POST("/:id/save-template", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.GET("/deletable", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.POST("/batch-delete", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.POST("/import", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.GET("/csv-template", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.GET("/excel-template", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.GET("/export", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.GET("/report", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.PUT("/batch", createProxyHandler(config.CreditActivityServiceURL))
			}

			// 管理员路由
			admin := activities.Group("")
			admin.Use(permissionMiddleware.AdminOnly())
			{
				admin.DELETE("/:id", createProxyHandler(config.CreditActivityServiceURL))
			}

			// 参与者管理路由
			participants := activities.Group("/:id/participants")
			{
				participants.GET("", createProxyHandler(config.CreditActivityServiceURL))
				participants.GET("/stats", createProxyHandler(config.CreditActivityServiceURL))
				participants.GET("/export", createProxyHandler(config.CreditActivityServiceURL))
				participants.POST("/leave", createProxyHandler(config.CreditActivityServiceURL))

				// 教师或管理员路由
				teacherOrAdminParticipants := participants.Group("")
				teacherOrAdminParticipants.Use(permissionMiddleware.TeacherOrAdmin())
				{
					teacherOrAdminParticipants.POST("", createProxyHandler(config.CreditActivityServiceURL))
					teacherOrAdminParticipants.PUT("/batch-credits", createProxyHandler(config.CreditActivityServiceURL))
					teacherOrAdminParticipants.PUT("/:user_id/credits", createProxyHandler(config.CreditActivityServiceURL))
					teacherOrAdminParticipants.DELETE("/:user_id", createProxyHandler(config.CreditActivityServiceURL))
					teacherOrAdminParticipants.POST("/batch-remove", createProxyHandler(config.CreditActivityServiceURL))
				}
			}

			// 附件管理路由
			attachments := activities.Group("/:id/attachments")
			{
				attachments.GET("", createProxyHandler(config.CreditActivityServiceURL))
				attachments.GET("/:attachment_id/download", createProxyHandler(config.CreditActivityServiceURL))
				attachments.GET("/:attachment_id/preview", createProxyHandler(config.CreditActivityServiceURL))

				// 教师或管理员路由
				teacherOrAdminAttachments := attachments.Group("")
				teacherOrAdminAttachments.Use(permissionMiddleware.TeacherOrAdmin())
				{
					teacherOrAdminAttachments.POST("", createProxyHandler(config.CreditActivityServiceURL))
					teacherOrAdminAttachments.POST("/batch", createProxyHandler(config.CreditActivityServiceURL))
					teacherOrAdminAttachments.PUT("/:attachment_id", createProxyHandler(config.CreditActivityServiceURL))
					teacherOrAdminAttachments.DELETE("/:attachment_id", createProxyHandler(config.CreditActivityServiceURL))
				}
			}
		}

		// 申请管理路由（需要认证）
		applications := api.Group("/applications")
		applications.Use(authMiddleware.AuthRequired())
		{
			applications.GET("", createProxyHandler(config.CreditActivityServiceURL))
			applications.GET("/:id", createProxyHandler(config.CreditActivityServiceURL))
			applications.GET("/stats", createProxyHandler(config.CreditActivityServiceURL))
			applications.GET("/export", createProxyHandler(config.CreditActivityServiceURL))
			applications.GET("/all", createProxyHandler(config.CreditActivityServiceURL))
		}

		// 统一检索API路由组（需要认证）
		searchActivities := api.Group("/search")
		searchActivities.Use(authMiddleware.AuthRequired())
		{
			searchActivities.GET("/activities", createProxyHandler(config.CreditActivityServiceURL))
			searchActivities.GET("/applications", createProxyHandler(config.CreditActivityServiceURL))
			searchActivities.GET("/participants", createProxyHandler(config.CreditActivityServiceURL))
			searchActivities.GET("/attachments", createProxyHandler(config.CreditActivityServiceURL))
		}
	}

	// 默认路由 - 返回API信息
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Credit Management API Gateway",
			"version": "3.0.0",
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

		// 特殊处理：为文件上传路由保留Content-Type和请求体
		if strings.Contains(path, "/import") || strings.Contains(path, "/upload") {
			// 确保multipart/form-data的Content-Type被保留
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "multipart/form-data") {
				c.Request.Header.Set("Content-Type", contentType)
				// 确保请求体被正确转发
				if c.Request.Body != nil {
					// 读取请求体并重新设置
					bodyBytes, err := io.ReadAll(c.Request.Body)
					if err == nil {
						c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					}
				}
			}
		}

		// 将用户信息传递给下游服务
		if userID, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}
		if username, exists := c.Get("username"); exists {
			c.Request.Header.Set("X-Username", username.(string))
		}
		if userType, exists := c.Get("user_type"); exists {
			c.Request.Header.Set("X-User-Type", userType.(string))
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
