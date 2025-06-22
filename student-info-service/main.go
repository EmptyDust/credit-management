package main

import (
	"os"
	"student-info-service/handlers"
	"student-info-service/models"
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
	db.AutoMigrate(&models.Student{})

	h := &handlers.StudentHandler{DB: db}
	r := gin.Default()
	r.POST("/students/register", h.Register)
	r.GET("/students/:studentNo", h.GetStudent)
	r.PUT("/students/:studentNo", h.UpdateStudent)

	r.Run(":8080")
} 