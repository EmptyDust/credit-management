package main

import (
	"fmt"
	"log"
	"os"

	"credit-management/credit-activity-service/handlers"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载本地环境变量文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	log.Println("=== 学分活动服务启动 ===")

	gin.SetMode(gin.ReleaseMode)
	log.Println("Gin模式已设置")

	log.Println("正在创建必要的目录...")
	if err := createDirectories(); err != nil {
		log.Printf("Warning: Failed to create directories, but continuing: %v", err)
	} else {
		log.Println("目录创建成功")
	}

	log.Println("正在连接数据库...")
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("数据库连接成功")

	activityHandler := handlers.NewActivityHandler(db)
	participantHandler := handlers.NewParticipantHandler(db)
	applicationHandler := handlers.NewApplicationHandler(db)
	attachmentHandler := handlers.NewAttachmentHandler(db)
	searchHandler := handlers.NewSearchHandler(db)

	authMiddleware := utils.NewHeaderAuthMiddleware()
	permissionMiddleware := utils.NewPermissionMiddleware()

	log.Println("正在创建路由...")
	r := gin.New()

	// 使用新的中间件
	r.Use(utils.RecoveryMiddleware())
	r.Use(utils.LoggingMiddleware())
	r.Use(utils.CORSMiddleware())

	// 注册公共配置接口（无需鉴权）
	registerActivityOptionsRoute(r)

	api := r.Group("/api")
	{
		activities := api.Group("/activities")
		{
			activities.GET("/categories", activityHandler.GetActivityCategories)
			activities.GET("/templates", activityHandler.GetActivityTemplates)

			auth := activities.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				allUsers := auth.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("", activityHandler.GetActivities)
					allUsers.GET("/stats", activityHandler.GetActivityStats)
					allUsers.GET("/:id", activityHandler.GetActivity)
					allUsers.POST("/:id/submit", activityHandler.SubmitActivity)
					allUsers.POST("/:id/withdraw", activityHandler.WithdrawActivity)
					allUsers.GET("/deletable", activityHandler.GetDeletableActivities)
					allUsers.POST("/:id/copy", activityHandler.CopyActivity)
					allUsers.POST("/:id/save-template", activityHandler.SaveAsTemplate)
					allUsers.POST("/import", activityHandler.ImportActivities)
					allUsers.GET("/csv-template", activityHandler.GetCSVTemplate)
					allUsers.GET("/excel-template", activityHandler.GetExcelTemplate)
					allUsers.POST("", activityHandler.CreateActivity)
					allUsers.PUT("/:id", activityHandler.UpdateActivity)
				}

				teacherOrAdmin := auth.Group("")
				teacherOrAdmin.Use(permissionMiddleware.TeacherOrAdmin())
				{
					teacherOrAdmin.POST("/batch", activityHandler.BatchCreateActivities)
					teacherOrAdmin.PUT("/batch", activityHandler.BatchUpdateActivities)
					teacherOrAdmin.POST("/:id/review", activityHandler.ReviewActivity)
					teacherOrAdmin.GET("/pending", activityHandler.GetPendingActivities)
					teacherOrAdmin.POST("/batch-delete", activityHandler.BatchDeleteActivities)
					teacherOrAdmin.GET("/export", activityHandler.ExportActivities)
					teacherOrAdmin.GET("/report", activityHandler.GetActivityReport)
				}

				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.DELETE("/:id", activityHandler.DeleteActivity)
				}
			}

			participants := activities.Group(":id")
			participants.Use(authMiddleware.AuthRequired())
			{
				allUsers := participants.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("/participants", participantHandler.GetActivityParticipants)
					allUsers.GET("/participants/stats", participantHandler.GetParticipantStats)
					allUsers.GET("/participants/export", participantHandler.ExportParticipants)
					allUsers.GET("/my-activities", participantHandler.GetUserParticipatedActivities)
				}

				ownerOrTeacherOrAdmin := participants.Group("")
				ownerOrTeacherOrAdmin.Use(permissionMiddleware.ActivityOwnerOrTeacherOrAdmin())
				{
					ownerOrTeacherOrAdmin.POST("/participants", participantHandler.AddParticipants)
					ownerOrTeacherOrAdmin.PUT("/participants/batch-credits", participantHandler.BatchSetCredits)
					ownerOrTeacherOrAdmin.PUT("/participants/:uuid/credits", participantHandler.SetSingleCredits)
					ownerOrTeacherOrAdmin.DELETE("/participants/:uuid", participantHandler.RemoveParticipant)
					ownerOrTeacherOrAdmin.POST("/participants/batch-remove", participantHandler.BatchRemoveParticipants)
				}

				studentOnly := participants.Group("")
				studentOnly.Use(permissionMiddleware.StudentOnly())
				{
					studentOnly.POST("/participants/leave", participantHandler.LeaveActivity)
				}
			}

			// 附件管理路由（单独抽出，保证所有认证用户都能访问预览/下载）
			attachments := activities.Group(":id/attachments")
			attachments.Use(authMiddleware.AuthRequired())
			{
				allUsers := attachments.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("", attachmentHandler.GetAttachments)
					allUsers.GET("/:attachment_id/download", attachmentHandler.DownloadAttachment)
					allUsers.GET("/:attachment_id/preview", attachmentHandler.PreviewAttachment)
				}

				ownerOrTeacherOrAdmin := attachments.Group("")
				ownerOrTeacherOrAdmin.Use(permissionMiddleware.ActivityOwnerOrTeacherOrAdmin())
				{
					ownerOrTeacherOrAdmin.POST("", attachmentHandler.UploadAttachment)
					ownerOrTeacherOrAdmin.POST("/batch", attachmentHandler.BatchUploadAttachments)
					ownerOrTeacherOrAdmin.PUT("/:attachment_id", attachmentHandler.UpdateAttachment)
					ownerOrTeacherOrAdmin.DELETE("/:attachment_id", attachmentHandler.DeleteAttachment)
				}
			}
		}

		applications := api.Group("/applications")
		applications.Use(authMiddleware.AuthRequired())
		{
			allUsers := applications.Group("")
			allUsers.Use(permissionMiddleware.AllUsers())
			{
				allUsers.GET("", applicationHandler.GetUserApplications)
				allUsers.GET("/:id", applicationHandler.GetApplication)
				allUsers.GET("/stats", applicationHandler.GetApplicationStats)
				allUsers.GET("/export", applicationHandler.ExportApplications)
			}

			teacherOrAdmin := applications.Group("")
			teacherOrAdmin.Use(permissionMiddleware.TeacherOrAdmin())
			{
				teacherOrAdmin.GET("/all", applicationHandler.GetAllApplications)
			}
		}

		search := api.Group("/search")
		search.Use(authMiddleware.AuthRequired())
		{
			allUsers := search.Group("")
			allUsers.Use(permissionMiddleware.AllUsers())
			{
				allUsers.GET("/activities", searchHandler.SearchActivities)
				allUsers.GET("/applications", searchHandler.SearchApplications)
				allUsers.GET("/participants", searchHandler.SearchParticipants)
				allUsers.GET("/attachments", searchHandler.SearchAttachments)
			}
		}
	}

	r.GET("/health", func(c *gin.Context) {
		utils.SendSuccessResponse(c, gin.H{"status": "ok", "service": "credit-activity-service"})
	})

	port := getEnv("PORT", "8083")
	log.Printf("Credit Activity Service starting on port %s", port)
	log.Println("服务启动完成，等待请求...")
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDatabase() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "credit_management")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, port, user, password, dbname)

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully")
	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createDirectories() error {
	dirs := []string{
		"uploads",
		"uploads/attachments",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return nil
}
