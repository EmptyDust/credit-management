package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/student-info-service/handlers"
	"credit-management/student-info-service/models"
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
	err = db.AutoMigrate(&models.Student{}, &models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建处理器
	studentHandler := handlers.NewStudentHandler(db)

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
		// 学生相关路由
		students := api.Group("/students")
		{
			students.POST("", studentHandler.CreateStudent)                        // 创建学生
			students.GET("/:username", studentHandler.GetStudent)                  // 根据用户名获取学生
			students.GET("/id/:studentID", studentHandler.GetStudentByID)          // 根据学号获取学生
			students.PUT("/:username", studentHandler.UpdateStudent)               // 更新学生信息
			students.DELETE("/:username", studentHandler.DeleteStudent)            // 删除学生
			students.GET("", studentHandler.GetAllStudents)                        // 获取所有学生
			students.GET("/college/:college", studentHandler.GetStudentsByCollege) // 根据学院获取学生
			students.GET("/major/:major", studentHandler.GetStudentsByMajor)       // 根据专业获取学生
			students.GET("/class/:class", studentHandler.GetStudentsByClass)       // 根据班级获取学生
			students.GET("/status/:status", studentHandler.GetStudentsByStatus)    // 根据状态获取学生
			students.GET("/search", studentHandler.SearchStudents)                 // 搜索学生
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "student-info-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8082")
	log.Printf("Student Info Service starting on port %s", port)
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
