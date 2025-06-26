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

	"credit-management/user-service/handlers"
	"credit-management/user-service/models"
	"credit-management/user-service/utils"
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

// createDatabaseViews 创建数据库视图
func createDatabaseViews(db *gorm.DB) error {
	views := []string{
		`CREATE OR REPLACE VIEW student_basic_info AS
		SELECT 
			user_id,
			username,
			real_name,
			student_id,
			college,
			major,
			class,
			grade,
			status,
			avatar,
			register_time,
			created_at,
			updated_at
		FROM users 
		WHERE user_type = 'student' AND deleted_at IS NULL`,

		`CREATE OR REPLACE VIEW teacher_basic_info AS
		SELECT 
			user_id,
			username,
			real_name,
			department,
			title,
			status,
			avatar,
			register_time,
			created_at,
			updated_at
		FROM users 
		WHERE user_type = 'teacher' AND deleted_at IS NULL`,

		`CREATE OR REPLACE VIEW student_detail_info AS
		SELECT 
			user_id,
			username,
			email,
			phone,
			real_name,
			student_id,
			college,
			major,
			class,
			grade,
			status,
			avatar,
			last_login_at,
			register_time,
			created_at,
			updated_at
		FROM users 
		WHERE user_type = 'student' AND deleted_at IS NULL`,

		`CREATE OR REPLACE VIEW teacher_detail_info AS
		SELECT 
			user_id,
			username,
			email,
			phone,
			real_name,
			department,
			title,
			specialty,
			status,
			avatar,
			last_login_at,
			register_time,
			created_at,
			updated_at
		FROM users 
		WHERE user_type = 'teacher' AND deleted_at IS NULL`,

		`CREATE OR REPLACE VIEW user_stats_view AS
		SELECT 
			user_type,
			status,
			COUNT(*) as count,
			DATE(created_at) as created_date
		FROM users 
		WHERE deleted_at IS NULL
		GROUP BY user_type, status, DATE(created_at)`,

		`CREATE OR REPLACE VIEW student_stats_view AS
		SELECT 
			college,
			major,
			grade,
			status,
			COUNT(*) as count
		FROM users 
		WHERE user_type = 'student' AND deleted_at IS NULL
		GROUP BY college, major, grade, status`,

		`CREATE OR REPLACE VIEW teacher_stats_view AS
		SELECT 
			department,
			title,
			status,
			COUNT(*) as count
		FROM users 
		WHERE user_type = 'teacher' AND deleted_at IS NULL
		GROUP BY department, title, status`,
	}

	for _, view := range views {
		if err := db.Exec(view).Error; err != nil {
			return fmt.Errorf("failed to create view: %v", err)
		}
	}

	return nil
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

	// 检查数据库表是否已存在（通过初始化脚本创建）
	log.Println("Checking database tables...")
	var tableExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists)

	if !tableExists {
		log.Println("Tables not found, creating database tables...")
		err = db.AutoMigrate(&models.User{})
		if err != nil {
			log.Fatal("Failed to create database tables:", err)
		}
	} else {
		log.Println("Database tables already exist, skipping AutoMigrate")
	}

	// 创建数据库视图（如果不存在）
	log.Println("Creating database views...")
	err = createDatabaseViews(db)
	if err != nil {
		log.Fatal("Failed to create database views:", err)
	}

	// JWT密钥
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")

	// 创建处理器
	userHandler := handlers.NewUserHandler(db)

	// 创建中间件
	authMiddleware := utils.NewAuthMiddleware(jwtSecret)
	permissionMiddleware := utils.NewPermissionMiddleware()

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
			// 公开路由
			users.POST("/register", userHandler.Register) // 用户注册（仅限学生）

			// 需要认证的路由
			auth := users.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 所有认证用户都可以访问的路由
				allUsers := auth.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("/profile", userHandler.GetUser)    // 获取当前用户信息
					allUsers.PUT("/profile", userHandler.UpdateUser) // 更新当前用户信息
					allUsers.GET("/stats", userHandler.GetUserStats) // 获取用户统计信息
					allUsers.GET("/:id", userHandler.GetUser)        // 获取指定用户信息

					// 新增：用户自助修改密码
					allUsers.POST("/change_password", userHandler.ChangePassword) // 修改自己密码

					// 新增：获取用户活动记录（预留）
					allUsers.GET("/activity", userHandler.GetUserActivity)     // 当前用户活动
					allUsers.GET("/:id/activity", userHandler.GetUserActivity) // 指定用户活动（管理员/教师）
				}

				// 管理员路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("/teachers", userHandler.CreateTeacher)       // 管理员创建教师
					admin.POST("/students", userHandler.CreateStudent)       // 管理员创建学生
					admin.PUT("/:id", userHandler.UpdateUser)                // 更新指定用户信息
					admin.DELETE("/:id", userHandler.DeleteUser)             // 删除用户
					admin.GET("", userHandler.GetAllUsers)                   // 获取所有用户
					admin.GET("/type/:userType", userHandler.GetUsersByType) // 根据用户类型获取用户

					// 新增：批量删除、批量状态、重置密码、导出
					admin.POST("/batch_delete", userHandler.BatchDeleteUsers)      // 批量删除
					admin.POST("/batch_status", userHandler.BatchUpdateUserStatus) // 批量状态
					admin.POST("/reset_password", userHandler.ResetPassword)       // 重置密码
					admin.GET("/export", userHandler.ExportUsers)                  // 导出用户数据

					// 新增：CSV导入功能
					admin.POST("/import-csv", userHandler.ImportUsersFromCSV)  // 从CSV导入用户
					admin.GET("/csv-template", userHandler.GetUserCSVTemplate) // 获取CSV模板
				}

				// 学生、教师和管理员可以访问的路由（基于角色的权限控制）
				studentTeacherOrAdmin := auth.Group("")
				studentTeacherOrAdmin.Use(permissionMiddleware.StudentTeacherOrAdmin())
				{
					// 删除重复的 /type/:userType 路由，因为 GetUsersByType 方法内部已经实现了权限控制
				}
			}
		}

		// 学生相关路由
		students := api.Group("/students")
		{
			// 需要认证的路由
			auth := students.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 学生、教师和管理员可以访问的路由
				studentTeacherOrAdmin := auth.Group("")
				studentTeacherOrAdmin.Use(permissionMiddleware.StudentTeacherOrAdmin())
				{
					studentTeacherOrAdmin.GET("", userHandler.GetStudents)           // 获取所有学生
					studentTeacherOrAdmin.GET("/stats", userHandler.GetStudentStats) // 获取学生统计信息
				}

				// 仅管理员可以访问的路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("", userHandler.CreateStudent)   // 创建学生
					admin.PUT(":id", userHandler.UpdateUser)    // 更新学生
					admin.DELETE(":id", userHandler.DeleteUser) // 删除学生
				}
			}
		}

		// 教师相关路由
		teachers := api.Group("/teachers")
		{
			// 需要认证的路由
			auth := teachers.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 学生、教师和管理员可以访问的路由
				studentTeacherOrAdmin := auth.Group("")
				studentTeacherOrAdmin.Use(permissionMiddleware.StudentTeacherOrAdmin())
				{
					studentTeacherOrAdmin.GET("", userHandler.GetTeachers)           // 获取所有教师
					studentTeacherOrAdmin.GET("/stats", userHandler.GetTeacherStats) // 获取教师统计信息
				}

				// 仅管理员可以访问的路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("", userHandler.CreateTeacher)   // 创建教师
					admin.PUT(":id", userHandler.UpdateUser)    // 更新教师
					admin.DELETE(":id", userHandler.DeleteUser) // 删除教师
				}
			}
		}

		// 搜索路由
		search := api.Group("/search")
		{
			search.Use(authMiddleware.AuthRequired())
			{
				search.GET("/users", userHandler.SearchUsers) // 通用用户搜索
			}
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8084")
	log.Printf("User service starting on port %s", port)
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
