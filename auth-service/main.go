package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/auth-service/handlers"
	"credit-management/auth-service/models"
	"credit-management/auth-service/utils"
)

// 连接数据库，带重试机制
func connectDatabase(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	
	// 重试配置
	maxRetries := 30
	retryInterval := 2 * time.Second
	
	for i := 0; i < maxRetries; i++ {
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

	// 自动迁移数据库表
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.UserPermission{},
		&models.RolePermission{},
		&models.PermissionGroup{},
		&models.PermissionGroupPermission{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 在一个事务中完成所有初始化
	log.Println("Initializing data...")
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := handlers.InitializeAdminUser(tx); err != nil {
			return err
		}
		// 初始化权限和角色
		permissionHandler := handlers.NewPermissionHandler(tx)
		permissionHandler.InitializePermissions(nil) // 传入nil context，因为这是初始化调用
		return nil
	})
	if err != nil {
		log.Fatal("Failed to run initializations:", err)
	}

	// JWT密钥
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")

	// 创建处理器
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	permissionHandler := handlers.NewPermissionHandler(db)

	// 创建中间件
	authMiddleware := utils.NewAuthMiddleware(jwtSecret)
	permissionMiddleware := utils.NewPermissionMiddleware(db)

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

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关路由
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)                  // 用户登录
			auth.POST("/validate-token", authHandler.ValidateToken) // 验证JWT token
			auth.POST("/refresh-token", authHandler.RefreshToken)   // 刷新token
			auth.POST("/logout", authHandler.Logout)                // 用户登出
			auth.GET("/validate-permission", authHandler.ValidatePermission) // 验证用户权限
		}

		// 权限管理路由
		permissions := api.Group("/permissions")
		{
			// 初始化权限（不需要权限验证）
			permissions.POST("/init", permissionHandler.InitializePermissions)
			
			// 需要权限验证的路由
			permissions.Use(authMiddleware.AuthRequired(), permissionMiddleware.RequirePermission("permission", "manage"))
			{
				// 角色管理
				permissions.POST("/roles", permissionHandler.CreateRole)           // 创建角色
				permissions.GET("/roles", permissionHandler.GetRoles)              // 获取所有角色
				permissions.GET("/roles/:roleID", permissionHandler.GetRole)       // 获取指定角色
				permissions.PUT("/roles/:roleID", permissionHandler.UpdateRole)    // 更新角色
				permissions.DELETE("/roles/:roleID", permissionHandler.DeleteRole) // 删除角色

				// 权限管理
				permissions.POST("", permissionHandler.CreatePermission)       // 创建权限
				permissions.GET("", permissionHandler.GetPermissions)          // 获取所有权限
				permissions.GET("/:id", permissionHandler.GetPermission)       // 获取指定权限
				permissions.DELETE("/:id", permissionHandler.DeletePermission) // 删除权限

				// 用户权限分配
				permissions.POST("/users/:userID/roles", permissionHandler.AssignRole)                             // 分配角色给用户
				permissions.DELETE("/users/:userID/roles/:roleID", permissionHandler.RemoveRole)                   // 移除用户角色
				permissions.POST("/users/:userID/permissions", permissionHandler.AssignPermission)                 // 分配权限给用户
				permissions.DELETE("/users/:userID/permissions/:permissionID", permissionHandler.RemovePermission) // 移除用户权限

				// 角色权限管理
				permissions.POST("/roles/:roleID/permissions", permissionHandler.AssignPermissionToRole)                   // 分配权限给角色
				permissions.DELETE("/roles/:roleID/permissions/:permissionID", permissionHandler.RemovePermissionFromRole) // 移除角色权限

				// 查询
				permissions.GET("/users/:userID/roles", permissionHandler.GetUserRoles)             // 获取用户角色
				permissions.GET("/users/:userID/permissions", permissionHandler.GetUserPermissions) // 获取用户权限
			}
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
