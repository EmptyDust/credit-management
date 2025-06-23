package main

import (
	"log"
	"os"

	"general-application-service/handlers"
	"general-application-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接
	dsn := "host=postgres user=postgres password=password dbname=credit_management port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移
	db.AutoMigrate(&models.Application{})

	// 创建Gin路由
	r := gin.Default()

	// 创建处理器
	applicationHandler := handlers.NewApplicationHandler(db)

	// 设置路由
	applications := r.Group("/api/applications")
	{
		applications.POST("/", applicationHandler.CreateApplication)
		applications.GET("/", applicationHandler.GetApplications)
		applications.GET("/:id", applicationHandler.GetApplication)
		applications.GET("/user/:userID", applicationHandler.GetApplicationsByUser)
		applications.GET("/student/:studentID", applicationHandler.GetApplicationsByStudent)
		applications.PUT("/:id", applicationHandler.UpdateApplication)
		applications.POST("/:id/review", applicationHandler.ReviewApplication)
		applications.DELETE("/:id", applicationHandler.DeleteApplication)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}
	r.Run(":" + port)
}
