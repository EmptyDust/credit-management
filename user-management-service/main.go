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
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.UserPermission{},
		&models.RolePermission{},
		&models.PermissionGroup{},
		&models.PermissionGroupPermission{},
		&models.UserFile{},
		&models.Notification{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// JWT密钥
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")

	// 创建处理器
	userHandler := handlers.NewUserHandler(db, jwtSecret)
	fileHandler := handlers.NewFileHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
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

	// 静态文件服务
	r.Static("/uploads", "./uploads")
	r.Static("/avatars", "./avatars")

	// API路由组
	api := r.Group("/api")
	{
		// 用户相关路由
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)            // 用户注册
			users.POST("/login", userHandler.Login)                  // 用户登录
			users.POST("/validate-token", userHandler.ValidateToken) // 验证JWT token
			users.GET("/stats", userHandler.GetUserStats)            // 获取用户统计信息

			// 需要认证的路由
			auth := users.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				auth.GET("/profile", userHandler.GetUser)      // 获取当前用户信息
				auth.PUT("/profile", userHandler.UpdateUser)   // 更新当前用户信息
				auth.POST("/avatar", userHandler.UploadAvatar) // 上传头像
			}

			// 管理员路由
			admin := users.Group("")
			admin.Use(authMiddleware.AuthRequired(), permissionMiddleware.RequirePermission("user", "manage"))
			{
				admin.GET("/:username", userHandler.GetUser)             // 获取指定用户信息
				admin.PUT("/:username", userHandler.UpdateUser)          // 更新指定用户信息
				admin.DELETE("/:username", userHandler.DeleteUser)       // 删除用户
				admin.GET("", userHandler.GetAllUsers)                   // 获取所有用户
				admin.GET("/type/:userType", userHandler.GetUsersByType) // 根据用户类型获取用户
			}
		}

		// 文件管理路由
		files := api.Group("/files")
		files.Use(authMiddleware.AuthRequired())
		{
			files.POST("/upload", fileHandler.UploadFile)            // 上传文件
			files.GET("/download/:fileID", fileHandler.DownloadFile) // 下载文件
			files.GET("/:fileID", fileHandler.GetFile)               // 获取文件信息
			files.GET("", fileHandler.GetUserFiles)                  // 获取用户文件列表
			files.PUT("/:fileID", fileHandler.UpdateFile)            // 更新文件信息
			files.DELETE("/:fileID", fileHandler.DeleteFile)         // 删除文件
			files.GET("/public", fileHandler.GetPublicFiles)         // 获取公开文件
			files.GET("/stats", fileHandler.GetFileStats)            // 获取文件统计信息
		}

		// 通知管理路由
		notifications := api.Group("/notifications")
		notifications.Use(authMiddleware.AuthRequired())
		{
			notifications.GET("", notificationHandler.GetUserNotifications)        // 获取用户通知
			notifications.GET("/:id", notificationHandler.GetNotification)         // 获取指定通知
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)         // 标记通知为已读
			notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)      // 标记所有通知为已读
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)   // 删除通知
			notifications.GET("/unread-count", notificationHandler.GetUnreadCount) // 获取未读通知数量
			notifications.GET("/stats", notificationHandler.GetNotificationStats)  // 获取通知统计信息
		}

		// 管理员通知路由
		adminNotifications := api.Group("/admin/notifications")
		adminNotifications.Use(authMiddleware.AuthRequired(), permissionMiddleware.RequirePermission("notification", "manage"))
		{
			adminNotifications.POST("", notificationHandler.CreateNotification)              // 创建通知
			adminNotifications.POST("/system", notificationHandler.SendSystemNotification)   // 发送系统通知
			adminNotifications.POST("/batch", notificationHandler.SendBatchNotification)     // 批量发送通知
			adminNotifications.GET("", notificationHandler.GetAllNotifications)              // 获取所有通知
			adminNotifications.DELETE("/:id", notificationHandler.DeleteNotificationByAdmin) // 管理员删除通知
			adminNotifications.GET("/stats", notificationHandler.GetSystemNotificationStats) // 获取系统通知统计
		}

		// 权限管理路由
		permissions := api.Group("/permissions")
		permissions.Use(authMiddleware.AuthRequired(), permissionMiddleware.RequirePermission("permission", "manage"))
		{
			// 角色管理
			permissions.POST("/roles", permissionHandler.CreateRole)       // 创建角色
			permissions.GET("/roles", permissionHandler.GetRoles)          // 获取所有角色
			permissions.GET("/roles/:id", permissionHandler.GetRole)       // 获取指定角色
			permissions.PUT("/roles/:id", permissionHandler.UpdateRole)    // 更新角色
			permissions.DELETE("/roles/:id", permissionHandler.DeleteRole) // 删除角色

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

		// 初始化权限
		api.POST("/init-permissions", permissionHandler.InitializePermissions)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-management-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8081")
	log.Printf("User Management Service starting on port %s", port)
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
