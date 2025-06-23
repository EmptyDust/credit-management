package main

import (
	"log"
	"os"

	"affair-management-service/handlers"
	"affair-management-service/models"

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
	db.AutoMigrate(&models.Affair{})

	// 创建Gin路由
	r := gin.Default()

	// 创建处理器
	affairHandler := handlers.NewAffairHandler(db)

	// 设置路由
	affairs := r.Group("/api/affairs")
	{
		affairs.POST("/", affairHandler.CreateAffair)
		affairs.GET("/", affairHandler.GetAffairs)
		affairs.GET("/:id", affairHandler.GetAffair)
		affairs.PUT("/:id", affairHandler.UpdateAffair)
		affairs.DELETE("/:id", affairHandler.DeleteAffair)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	r.Run(":" + port)
}
