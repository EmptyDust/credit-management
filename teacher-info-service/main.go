package main

import (
	"os"
	"teacher-info-service/handlers"
	"teacher-info-service/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.Teacher{})

	h := &handlers.TeacherHandler{DB: db}
	r := gin.Default()
	r.POST("/teachers/register", h.Register)
	r.GET("/teachers/:username", h.GetTeacher)
	r.PUT("/teachers/:username", h.UpdateTeacher)

	r.Run(":8080")
} 