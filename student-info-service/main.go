package main

import (
	"log"
	"os"

	"student-info-service/handlers"
	"student-info-service/models"

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
	db.AutoMigrate(&models.Student{})

	// 创建Gin路由
	r := gin.Default()

	// 创建处理器
	studentHandler := handlers.NewStudentHandler(db)

	// 设置路由
	students := r.Group("/api/students")
	{
		students.POST("/", studentHandler.CreateStudent)
		students.GET("/", studentHandler.GetStudents)
		students.GET("/:id", studentHandler.GetStudent)
		students.GET("/user/:userID", studentHandler.GetStudentByUserID)
		students.PUT("/:id", studentHandler.UpdateStudent)
		students.DELETE("/:id", studentHandler.DeleteStudent)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	r.Run(":" + port)
}
