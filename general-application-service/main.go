package main

import (
	"os"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"general-application-service/models"
	"general-application-service/handlers"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.Application{})

	r := gin.Default()
	h := &handlers.ApplicationHandler{DB: db}

	r.POST("/applications", h.CreateApplication)
	r.GET("/applications/:id", h.GetApplication)
	r.PUT("/applications/:id/status", h.UpdateStatus)
	r.PUT("/applications/:id/credit", h.UpdateFinalCredit)

	r.Run(":8080")
} 