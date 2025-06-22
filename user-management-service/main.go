package main

import (
	"os"
	"user-management-service/handlers"
	"user-management-service/models"
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
	db.AutoMigrate(&models.User{})

	h := &handlers.UserHandler{DB: db}
	r := gin.Default()
	r.POST("/users/register", h.Register)
	r.POST("/users/login", h.Login)
	r.GET("/users/:username", h.GetUser)

	r.Run(":8080")
} 