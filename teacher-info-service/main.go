package main

import (
	"log"
	"os"

	"teacher-info-service/handlers"
	"teacher-info-service/models"

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
	db.AutoMigrate(&models.Teacher{})

	// 创建Gin路由
	r := gin.Default()

	// 创建处理器
	teacherHandler := handlers.NewTeacherHandler(db)

	// 设置路由
	teachers := r.Group("/api/teachers")
	{
		teachers.POST("/", teacherHandler.CreateTeacher)
		teachers.GET("/", teacherHandler.GetTeachers)
		teachers.GET("/:id", teacherHandler.GetTeacher)
		teachers.GET("/user/:userID", teacherHandler.GetTeacherByUserID)
		teachers.PUT("/:id", teacherHandler.UpdateTeacher)
		teachers.DELETE("/:id", teacherHandler.DeleteTeacher)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}
	r.Run(":" + port)
}
