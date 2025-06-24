package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/application-management-service/handlers"
	"credit-management/application-management-service/models"
	"credit-management/application-management-service/utils"
)

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "credit_management")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate database tables
	err = db.AutoMigrate(
		&models.Application{},
		&models.ProofMaterial{},
		&models.InnovationPracticeCredit{},
		&models.DisciplineCompetitionCredit{},
		&models.StudentEntrepreneurshipProjectCredit{},
		&models.EntrepreneurshipPracticeCredit{},
		&models.PaperPatentCredit{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create handlers and middleware
	applicationHandler := handlers.NewApplicationHandler(db)
	authMiddleware := utils.NewAuthMiddleware(getEnv("JWT_SECRET", "your-secret-key"))

	// Set up Gin router
	r := gin.Default()
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

	// API routes
	api := r.Group("/api")
	{
		applications := api.Group("/applications")
		applications.Use(authMiddleware.AuthRequired())
		{
			applications.POST("", applicationHandler.CreateApplication)
			applications.POST("/batch", applicationHandler.BatchCreateApplications)
			applications.GET(":id", applicationHandler.GetApplicationDetail)
			applications.GET(":id/detail", applicationHandler.GetApplicationDetail)
			applications.PUT(":id/details", applicationHandler.UpdateApplicationDetails)
			applications.POST(":id/submit", applicationHandler.SubmitApplication)
			applications.PUT(":id/status", applicationHandler.UpdateApplicationStatus)
			applications.GET("/user/:studentNumber", applicationHandler.GetUserApplications)
			applications.GET("", applicationHandler.GetAllApplications)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "application-management-service"})
	})

	// Start server
	port := getEnv("PORT", "8082")
	log.Printf("Application management service starting on port %s", port)
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
