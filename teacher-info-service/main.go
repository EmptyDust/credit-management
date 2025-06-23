package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/teacher-info-service/handlers"
	"credit-management/teacher-info-service/models"
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
	err = db.AutoMigrate(&models.Teacher{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建处理器
	teacherHandler := handlers.NewTeacherHandler(db)

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
		// 教师相关路由
		teachers := api.Group("/teachers")
		{
			teachers.POST("", teacherHandler.CreateTeacher)                                 // 创建教师
			teachers.GET("/:username", teacherHandler.GetTeacher)                           // 根据用户名获取教师
			teachers.PUT("/:username", teacherHandler.UpdateTeacher)                        // 更新教师信息
			teachers.DELETE("/:username", teacherHandler.DeleteTeacher)                     // 删除教师
			teachers.GET("", teacherHandler.GetAllTeachers)                                 // 获取所有教师
			teachers.GET("/department/:department", teacherHandler.GetTeachersByDepartment) // 根据院系获取教师
			teachers.GET("/title/:title", teacherHandler.GetTeachersByTitle)                // 根据职称获取教师
			teachers.GET("/status/:status", teacherHandler.GetTeachersByStatus)             // 根据状态获取教师
			teachers.GET("/search", teacherHandler.SearchTeachers)                          // 搜索教师
			teachers.GET("/active", teacherHandler.GetActiveTeachers)                       // 获取活跃教师
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "teacher-info-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8085")
	log.Printf("Teacher Info Service starting on port %s", port)
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
