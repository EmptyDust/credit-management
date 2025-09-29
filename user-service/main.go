package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"credit-management/user-service/handlers"
	// "credit-management/user-service/middleware"
	"credit-management/user-service/routers"
)

func connectDatabase(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			log.Printf("数据库连接成功，第%d次尝试", i+1)
			return db, nil
		}

		log.Printf("数据库连接失败（尝试 %d/%d）: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("%v后重试...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("数据库连接失败，已尝试%d次: %v", maxRetries, err)
}

func main() {
	// 加载本地环境变量文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "credit_management")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	log.Println("正在连接数据库...")
	db, err := connectDatabase(dsn)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	userHandler := handlers.NewUserHandler(db)

	r := routers.RegisterRouters(userHandler)

	port := getEnv("PORT", "8084")
	log.Printf("用户服务启动，监听端口：%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务启动失败：", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
