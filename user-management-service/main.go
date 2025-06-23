package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/user-management-service/handlers"
	"credit-management/user-management-service/models"
	"credit-management/user-management-service/utils"
)

func main() {
	// 数据库连接配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "credit_management")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// JWT密钥
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")

	// 创建处理器
	userHandler := handlers.NewUserHandler(db, jwtSecret)

	// 创建中间件
	authMiddleware := utils.NewAuthMiddleware(jwtSecret)

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

	// 添加全局中间件，打印每个请求的完整路径和方法
	r.Use(func(c *gin.Context) {
		log.Printf("[DEBUG] %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// API路由组
	api := r.Group("/api")
	{
		// 用户相关路由
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register) // 用户注册
			users.GET("/stats", userHandler.GetUserStats) // 获取用户统计信息

			// 需要认证的路由
			auth := users.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				auth.GET("/profile", userHandler.GetUser)    // 获取当前用户信息
				auth.PUT("/profile", userHandler.UpdateUser) // 更新当前用户信息
			}

			// 管理员路由
			admin := users.Group("")
			admin.Use(authMiddleware.AuthRequired())
			{
				admin.GET("/:username", userHandler.GetUser)             // 获取指定用户信息
				admin.PUT("/:username", userHandler.UpdateUser)          // 更新指定用户信息
				admin.DELETE("/:username", userHandler.DeleteUser)       // 删除用户
				admin.GET("", userHandler.GetAllUsers)                   // 获取所有用户
				admin.GET("/type/:userType", userHandler.GetUsersByType) // 根据用户类型获取用户
			}
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-management-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8080")
	log.Printf("User management service starting on port %s", port)
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
