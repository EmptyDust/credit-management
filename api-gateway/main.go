package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// 辅助函数
func newScanner(r io.Reader) *bufio.Scanner {
	return bufio.NewScanner(r)
}

func execCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

func execCommandShell(command string) *exec.Cmd {
	return exec.Command("sh", "-c", command)
}

type ProxyConfig struct {
	UserServiceURL           string
	AuthServiceURL           string
	CreditActivityServiceURL string
	JWTSecret                string
}

// JWTClaims 自定义JWT claims结构
type JWTClaims struct {
	UUID     string `json:"uuid"`
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
		var tokenString string

		// 首先尝试从 Authorization header 获取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// 如果没有 Authorization header，尝试从 URL query 参数获取（用于 SSE）
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证令牌",
				"data":    nil,
			})
			c.Abort()
			return
		}

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
		if claims.UUID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token中缺少用户ID",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("uuid", claims.UUID)
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

func (m *PermissionMiddleware) RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}
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
		if _, ok := roleSet[userType.(string)]; !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
				"data":    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 原 ActivityOwnerOrTeacherOrAdmin函数: 活动所有者或教师或管理员权限
// 教师或管理员直接通过
// 对于学生，需要检查是否为活动所有者
// 由于API网关无法直接访问数据库，我们将这个检查放在具体的handler中
// 这个中间件主要用于路由级别的权限控制

func main() {
	// 加载本地环境变量文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	// 测试数据模式：enabled 时开放，无需任何鉴权；其他值或未设置时完全禁用
	testDataMode := getEnv("TEST_DATA_MODE", "disabled")
	testDataEnabled := strings.EqualFold(testDataMode, "enabled")

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
		// 开发者工具路由（仅管理员）
		devtools := api.Group("/devtools")
		devtools.Use(authMiddleware.AuthRequired())
		devtools.Use(permissionMiddleware.RequireRoles("admin"))
		{
			devtools.GET("/services", getDockerServices)
			devtools.GET("/logs/:service", streamServiceLogs)
		}

		// 认证服务路由（无需认证）
		api.Any("/auth/*path", createProxyHandler(config.AuthServiceURL))

		// 配置选项（透传到 user-service）
		api.GET("/config/options", createProxyHandler(config.UserServiceURL))

		// 测试数据相关接口
		// 是否可用由 API Gateway 基于 TEST_DATA_MODE 环境变量控制：
		// - TEST_DATA_MODE=enabled: 完全开放，无需任何鉴权
		// - 其它或未设置: 完全禁用，返回 403
		api.POST("/test-data/departments-from-options", func(c *gin.Context) {
			if !testDataEnabled {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "测试数据功能未启用，请在 API Gateway 中设置 TEST_DATA_MODE=enabled 后再试",
					"data":    nil,
				})
				return
			}
			createProxyHandler(config.UserServiceURL)(c)
		})

		// 权限管理服务路由
		permissions := api.Group("/permissions")
		permissions.Use(authMiddleware.AuthRequired())
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

		// 学生注册路由（无需认证）
		publicStudents := api.Group("/students")
		{
			publicStudents.POST("/register", createProxyHandler(config.UserServiceURL))
		}

		// 统一用户服务路由
		users := api.Group("/users")
		users.Use(authMiddleware.AuthRequired())
		{

			// 管理员路由
			admin := users.Group("")
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

		// 学生相关路由（需要认证，管理员权限由 user-service 自己校验）
		students := api.Group("/students")
		students.Use(authMiddleware.AuthRequired())
		{
			students.POST("", createProxyHandler(config.UserServiceURL))
			students.PUT("/:id", createProxyHandler(config.UserServiceURL))
			students.DELETE("/:id", createProxyHandler(config.UserServiceURL))
		}

		// 教师相关路由（需要认证，管理员权限由 user-service 自己校验）
		teachers := api.Group("/teachers")
		teachers.Use(authMiddleware.AuthRequired())
		{
			teachers.POST("", createProxyHandler(config.UserServiceURL))
			teachers.PUT("/:id", createProxyHandler(config.UserServiceURL))
			teachers.DELETE("/:id", createProxyHandler(config.UserServiceURL))
		}

		// 搜索路由（需要认证）
		search := api.Group("/search")
		search.Use(authMiddleware.AuthRequired())
		{
			search.GET("/users", createProxyHandler(config.UserServiceURL))
		}

		// 学分活动服务路由（需要认证，角色/权限检查下沉到 credit-activity-service）
		activities := api.Group("/activities")
		activities.Use(authMiddleware.AuthRequired())
		{
			// 公共选项（无需认证）
			r.GET("/api/activities/config/options", createProxyHandler(config.CreditActivityServiceURL))

			// 基础路由（所有认证用户）
			activities.GET("/categories", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("/templates", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("/stats", createProxyHandler(config.CreditActivityServiceURL))
			activities.GET("/:id", createProxyHandler(config.CreditActivityServiceURL))
			activities.POST("/:id/submit", createProxyHandler(config.CreditActivityServiceURL))
			activities.POST("/:id/withdraw", createProxyHandler(config.CreditActivityServiceURL))
			activities.POST("/:id/copy", createProxyHandler(config.CreditActivityServiceURL))
			activities.POST("", createProxyHandler(config.CreditActivityServiceURL))
			activities.PUT("/:id", createProxyHandler(config.CreditActivityServiceURL))

			// 教师或管理员路由（仅在 credit-activity-service 内部做角色检查）
			teacherOrAdmin := activities.Group("")
			{
				teacherOrAdmin.POST("/batch", createProxyHandler(config.CreditActivityServiceURL))
				teacherOrAdmin.PUT("/batch", createProxyHandler(config.CreditActivityServiceURL))
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
			}

			// 管理员路由（仅在 credit-activity-service 内部做角色检查）
			admin := activities.Group("")
			{
				admin.DELETE("/:id", createProxyHandler(config.CreditActivityServiceURL))
			}

			// 参与者管理路由
			participants := activities.Group("/:id/participants")
			{
				participants.GET("", createProxyHandler(config.CreditActivityServiceURL))
				participants.GET("/stats", createProxyHandler(config.CreditActivityServiceURL))
				participants.GET("/export", createProxyHandler(config.CreditActivityServiceURL))
				participants.GET("/my-activities", createProxyHandler(config.CreditActivityServiceURL))

				// 活动所有者或教师或管理员路由
				ownerOrTeacherOrAdminParticipants := participants.Group("")
				{
					ownerOrTeacherOrAdminParticipants.POST("", createProxyHandler(config.CreditActivityServiceURL))
					ownerOrTeacherOrAdminParticipants.PUT("/batch-credits", createProxyHandler(config.CreditActivityServiceURL))
					ownerOrTeacherOrAdminParticipants.PUT("/:uuid/credits", createProxyHandler(config.CreditActivityServiceURL))
					ownerOrTeacherOrAdminParticipants.DELETE("/:uuid", createProxyHandler(config.CreditActivityServiceURL))
					ownerOrTeacherOrAdminParticipants.POST("/batch-remove", createProxyHandler(config.CreditActivityServiceURL))
				}

				// 学生路由
				studentOnly := participants.Group("")
				{
					studentOnly.POST("/leave", createProxyHandler(config.CreditActivityServiceURL))
				}
			}

			// 附件管理路由
			attachments := activities.Group("/:id/attachments")
			attachments.Use(authMiddleware.AuthRequired())
			{
				attachments.GET("", createProxyHandler(config.CreditActivityServiceURL))
				attachments.GET("/:attachment_id/download", createProxyHandler(config.CreditActivityServiceURL))
				attachments.GET("/:attachment_id/preview", createProxyHandler(config.CreditActivityServiceURL))

				// 活动所有者或教师或管理员路由
				ownerOrTeacherOrAdminAttachments := attachments.Group("")
				{
					ownerOrTeacherOrAdminAttachments.POST("", createProxyHandler(config.CreditActivityServiceURL))
					ownerOrTeacherOrAdminAttachments.POST("/batch", createProxyHandler(config.CreditActivityServiceURL))
					ownerOrTeacherOrAdminAttachments.PUT("/:attachment_id", createProxyHandler(config.CreditActivityServiceURL))
					ownerOrTeacherOrAdminAttachments.DELETE("/:attachment_id", createProxyHandler(config.CreditActivityServiceURL))
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

			// 教师或管理员路由
			teacherOrAdmin := applications.Group("")
			teacherOrAdmin.Use(permissionMiddleware.RequireRoles("teacher", "admin"))
			{
				teacherOrAdmin.GET("/all", createProxyHandler(config.CreditActivityServiceURL))
			}
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
		if userID, exists := c.Get("uuid"); exists {
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

// Docker 服务配置
var dockerServices = []struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}{
	{ID: "credit_management_gateway", Name: "API Gateway"},
	{ID: "credit_management_auth", Name: "Auth Service"},
	{ID: "credit_management_user", Name: "User Service"},
	{ID: "credit_management_credit_activity", Name: "Activity Service"},
	{ID: "credit_management_postgres", Name: "PostgreSQL"},
	{ID: "credit_management_redis", Name: "Redis"},
	{ID: "credit_management_frontend", Name: "Frontend"},
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Health string `json:"health,omitempty"`
}

// 获取 Docker 服务列表和状态
func getDockerServices(c *gin.Context) {
	services := make([]ServiceStatus, 0, len(dockerServices))

	for _, svc := range dockerServices {
		status := ServiceStatus{
			ID:     svc.ID,
			Name:   svc.Name,
			Status: "unknown",
		}

		// 尝试检查容器状态（使用 docker inspect）
		cmd := fmt.Sprintf("docker inspect --format='{{.State.Status}}' %s 2>/dev/null || echo 'not_found'", svc.ID)
		out, err := execCommand(cmd)
		if err == nil {
			out = strings.TrimSpace(out)
			if out != "not_found" && out != "" {
				status.Status = out
			}
		}

		services = append(services, status)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    services,
	})
}

// 流式传输服务日志
func streamServiceLogs(c *gin.Context) {
	service := c.Param("service")
	tail := c.DefaultQuery("tail", "100")
	follow := c.DefaultQuery("follow", "true")

	// 验证服务名
	validService := false
	for _, svc := range dockerServices {
		if svc.ID == service {
			validService = true
			break
		}
	}

	if !validService {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的服务名",
			"data":    nil,
		})
		return
	}

	// 设置 SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建 context 用于取消
	ctx := c.Request.Context()

	// 构建 docker logs 命令
	args := []string{"logs", "--tail", tail}
	if follow == "true" {
		args = append(args, "-f")
	}
	args = append(args, "--timestamps", service)

	cmd := execCommandContext(ctx, "docker", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "无法获取日志流: " + err.Error()})
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "无法获取错误流: " + err.Error()})
		return
	}

	if err := cmd.Start(); err != nil {
		c.SSEvent("error", gin.H{"message": "启动日志命令失败: " + err.Error()})
		return
	}

	// 创建一个 channel 用于合并 stdout 和 stderr
	logChan := make(chan string, 100)
	done := make(chan struct{})

	// 读取 stdout
	go func() {
		scanner := newScanner(stdout)
		for scanner.Scan() {
			select {
			case logChan <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	// 读取 stderr
	go func() {
		scanner := newScanner(stderr)
		for scanner.Scan() {
			select {
			case logChan <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	// 等待命令结束
	go func() {
		cmd.Wait()
		close(done)
	}()

	// 发送日志
	c.Stream(func(w io.Writer) bool {
		select {
		case line := <-logChan:
			// 解析日志行，提取时间戳和内容
			logEntry := parseLogLine(line, service)
			c.SSEvent("log", logEntry)
			return true
		case <-done:
			return false
		case <-ctx.Done():
			return false
		}
	})
}

// LogEntry 日志条目
type LogEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Message   string `json:"message"`
}

// 解析日志行
func parseLogLine(line string, service string) LogEntry {
	entry := LogEntry{
		ID:      fmt.Sprintf("%d-%s", time.Now().UnixNano(), randString(6)),
		Service: service,
		Level:   "INFO",
		Message: line,
	}

	// 尝试解析时间戳（Docker logs --timestamps 格式）
	if len(line) > 30 && line[4] == '-' && line[10] == 'T' {
		entry.Timestamp = line[:30]
		entry.Message = strings.TrimSpace(line[31:])
	} else {
		entry.Timestamp = time.Now().Format("2006-01-02T15:04:05.000000000Z")
	}

	// 简单的日志级别检测
	msgLower := strings.ToLower(entry.Message)
	if strings.Contains(msgLower, "error") || strings.Contains(msgLower, "err") || strings.Contains(msgLower, "fail") {
		entry.Level = "ERROR"
	} else if strings.Contains(msgLower, "warn") {
		entry.Level = "WARN"
	} else if strings.Contains(msgLower, "debug") {
		entry.Level = "DEBUG"
	}

	return entry
}

// 执行命令并返回输出
func execCommand(command string) (string, error) {
	cmd := execCommandShell(command)
	out, err := cmd.Output()
	return string(out), err
}

// 生成随机字符串
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
