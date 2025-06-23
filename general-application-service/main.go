package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/general-application-service/handlers"
	"credit-management/general-application-service/models"
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
		&models.Application{},
		&models.ProofMaterial{},
		&models.InnovationPracticeCredit{},
		&models.DisciplineCompetitionCredit{},
		&models.StudentEntrepreneurshipCredit{},
		&models.EntrepreneurshipPracticeCredit{},
		&models.PaperPatentCredit{},
		&models.Affair{},
		&models.Student{},
		&models.Teacher{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建处理器
	applicationHandler := handlers.NewApplicationHandler(db)

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
		// 申请相关路由
		applications := api.Group("/applications")
		{
			applications.POST("", applicationHandler.CreateApplication)                          // 创建申请
			applications.GET("/:id", applicationHandler.GetApplication)                          // 获取单个申请
			applications.PUT("/:id", applicationHandler.UpdateApplication)                       // 更新申请
			applications.DELETE("/:id", applicationHandler.DeleteApplication)                    // 删除申请
			applications.POST("/:id/review", applicationHandler.ReviewApplication)               // 审核申请
			applications.GET("", applicationHandler.GetAllApplications)                          // 获取所有申请
			applications.GET("/student/:studentID", applicationHandler.GetApplicationsByStudent) // 根据学生ID获取申请
			applications.GET("/status/:status", applicationHandler.GetApplicationsByStatus)      // 根据状态获取申请
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "general-application-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8086")
	log.Printf("General Application Service starting on port %s", port)
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
