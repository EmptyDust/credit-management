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

	"credit-management/student-info-service/handlers"
	"credit-management/student-info-service/models"
	"credit-management/student-info-service/utils"
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
	err = db.AutoMigrate(&models.Student{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建处理器
	studentHandler := handlers.NewStudentHandler(db)

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
		// 学生信息管理路由
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
					studentTeacherOrAdmin.GET("", studentHandler.GetAllStudents)                    // 获取所有学生（根据角色返回不同数据）
					studentTeacherOrAdmin.GET("/search", studentHandler.SearchStudents)             // 搜索学生
					studentTeacherOrAdmin.GET("/college/:college", studentHandler.GetStudentsByCollege)     // 按学院获取
					studentTeacherOrAdmin.GET("/major/:major", studentHandler.GetStudentsByMajor)           // 按专业获取
					studentTeacherOrAdmin.GET("/class/:class", studentHandler.GetStudentsByClass)           // 按班级获取
					studentTeacherOrAdmin.GET("/status/:status", studentHandler.GetStudentsByStatus)        // 按状态获取
					studentTeacherOrAdmin.GET(":id", studentHandler.GetStudentByID)                    // 获取指定学生
					studentTeacherOrAdmin.GET("/user/:userID", studentHandler.GetStudentByUserID)   // 按用户ID获取学生
				}

				// 仅管理员可以访问的路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("", studentHandler.CreateStudent)                    // 创建学生
					admin.PUT(":id", studentHandler.UpdateStudentByID)                 // 更新学生
					admin.DELETE(":id", studentHandler.DeleteStudentByID)              // 删除学生
					admin.DELETE("/user/:user_id", studentHandler.DeleteStudentByUserID) // 按用户ID删除学生档案
				}

				// 所有认证用户都可以访问的路由
				auth.GET("/search/username", studentHandler.GetStudentByUsername)  // 按用户名搜索
			}
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "student-info-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8085")
	log.Printf("Student info service starting on port %s", port)
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
