package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"credit-management/auth-service/handlers"
	"credit-management/auth-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 连接数据库，带重试机制
func connectDatabase(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// 重试配置
	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := range maxRetries {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			log.Printf("Successfully connected to database on attempt %d", i+1)
			return db, nil
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
}

func main() {
	// 加载本地环境变量文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	// 数据库连接配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "credit_management")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// 连接数据库（带重试）
	log.Println("Connecting to database...")
	db, err := connectDatabase(dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化管理员用户
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := handlers.InitializeAdminUser(tx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal("Failed to run initializations:", err)
	}

	// Redis连接配置
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "password")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	// 连接Redis
	log.Println("Connecting to Redis...")
	redisClient := utils.NewRedisClient(redisAddr, redisPassword, 0)
	if redisClient == nil {
		log.Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// JWT密钥
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")

	// 创建处理器
	authHandler := handlers.NewAuthHandler(db, jwtSecret, redisClient)

	// 创建速率限制中间件（5次尝试/分钟）
	rateLimiter := utils.NewRateLimitMiddleware(redisClient, 5, time.Minute)

	// 设置Gin路由
	r := gin.Default()

	// 获取允许的前端域名
	corsAllowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// 检查请求来源是否在允许列表中
		if origin != "" {
			allowedOrigins := make(map[string]struct{})
			for _, allowedOrigin := range splitAndTrim(corsAllowedOrigins, ",") {
				allowedOrigins[allowedOrigin] = struct{}{}
			}
			if _, ok := allowedOrigins[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关路由
		auth := api.Group("/auth")
		{
			// 登录端点应用速率限制（基于IP和用户名）
			auth.POST("/login", rateLimiter.LimitByIP(), rateLimiter.LimitByUsername(), authHandler.Login)
			auth.POST("/validate-token", authHandler.ValidateToken)
			auth.POST("/validate-token-with-claims", authHandler.ValidateTokenWithClaims)
			auth.POST("/refresh-token", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8081")
	log.Printf("Auth service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// splitAndTrim 分割字符串并去除空格
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
