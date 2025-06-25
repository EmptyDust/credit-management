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

	"credit-management/teacher-info-service/handlers"
	"credit-management/teacher-info-service/models"
	"credit-management/teacher-info-service/utils"
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
	err = db.AutoMigrate(&models.Teacher{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建处理器
	teacherHandler := handlers.NewTeacherHandler(db)

	// 创建中间件
	authServiceURL := getEnv("AUTH_SERVICE_URL", "http://auth-service:8081")
	authMiddleware := utils.NewAuthMiddleware(authServiceURL)
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

	// API路由组
	api := r.Group("/api")
	{
		// 教师信息管理路由
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
					studentTeacherOrAdmin.GET("", teacherHandler.GetAllTeachers)                  // 获取所有教师（根据角色返回不同数据）
					studentTeacherOrAdmin.GET(":id", teacherHandler.GetTeacher)                    // 获取指定教师
					studentTeacherOrAdmin.GET("/department/:department", teacherHandler.GetTeachersByDepartment) // 按部门获取
					studentTeacherOrAdmin.GET("/title/:title", teacherHandler.GetTeachersByTitle)           // 按职称获取
					studentTeacherOrAdmin.GET("/status/:status", teacherHandler.GetTeachersByStatus)        // 按状态获取
					studentTeacherOrAdmin.GET("/search", teacherHandler.SearchTeachers)             // 搜索教师
					studentTeacherOrAdmin.GET("/active", teacherHandler.GetActiveTeachers)          // 获取活跃教师
				}

				// 仅管理员可以访问的路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("", teacherHandler.CreateTeacher)                    // 创建教师
					admin.PUT(":id", teacherHandler.UpdateTeacher)                 // 更新教师
					admin.DELETE(":id", teacherHandler.DeleteTeacher)              // 删除教师
					admin.DELETE("/user/:user_id", teacherHandler.DeleteTeacherByUserID) // 按用户ID删除教师档案
				}

				// 所有认证用户都可以访问的路由
				auth.GET("/search/username", teacherHandler.GetTeacherByUsername)  // 按用户名搜索
			}
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "teacher-info-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8086")
	log.Printf("Teacher info service starting on port %s", port)
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
