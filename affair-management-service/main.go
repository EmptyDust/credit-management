package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/affair-management-service/handlers"
	"credit-management/affair-management-service/models"
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
		&models.Affair{},
		&models.AffairStudent{},
		&models.Student{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建处理器
	affairHandler := handlers.NewAffairHandler(db)

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
		// 事项相关路由
		affairs := api.Group("/affairs")
		{
			affairs.POST("", affairHandler.CreateAffair)                           // 创建事项
			affairs.GET("/:id", affairHandler.GetAffair)                           // 获取单个事项
			affairs.PUT("/:id", affairHandler.UpdateAffair)                        // 更新事项
			affairs.DELETE("/:id", affairHandler.DeleteAffair)                     // 删除事项
			affairs.GET("", affairHandler.GetAllAffairs)                           // 获取所有事项
			affairs.GET("/category/:category", affairHandler.GetAffairsByCategory) // 根据类别获取事项
			affairs.GET("/active", affairHandler.GetActiveAffairs)                 // 获取活跃事项
		}

		// 事项-学生关系路由
		affairStudents := api.Group("/affair-students")
		{
			affairStudents.POST("", affairHandler.AddStudentToAffair)                             // 为学生添加事项
			affairStudents.DELETE("/:affairID/:studentID", affairHandler.RemoveStudentFromAffair) // 从事项中移除学生
			affairStudents.GET("/affair/:affairID", affairHandler.GetStudentsByAffair)            // 获取事项下的所有学生
			affairStudents.GET("/student/:studentID", affairHandler.GetAffairsByStudent)          // 获取学生参与的所有事项
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "affair-management-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8087")
	log.Printf("Affair Management Service starting on port %s", port)
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
