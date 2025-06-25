package main

import (
	"fmt"
	"log"
	"os"

	"credit-management/credit-activity-service/handlers"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("=== 学分活动服务启动 ===")

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	log.Println("Gin模式已设置")

	// 初始化数据库连接
	log.Println("正在连接数据库...")
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("数据库连接成功")

	// 自动迁移数据库表
	log.Println("正在迁移数据库表...")
	if err := autoMigrate(db); err != nil {
		log.Printf("Warning: AutoMigrate failed, but continuing with init.sql schema: %v", err)
	} else {
		log.Println("数据库表迁移成功")
	}

	// 创建触发器
	log.Println("正在创建触发器...")
	if err := createTriggers(db); err != nil {
		log.Printf("Warning: Trigger creation failed, but continuing: %v", err)
	} else {
		log.Println("触发器创建成功")
	}

	// 创建处理器
	activityHandler := handlers.NewActivityHandler(db)
	participantHandler := handlers.NewParticipantHandler(db)
	applicationHandler := handlers.NewApplicationHandler(db)
	attachmentHandler := handlers.NewAttachmentHandler(db)

	// 创建中间件
	authMiddleware := utils.NewAuthMiddleware()
	permissionMiddleware := utils.NewPermissionMiddleware()

	// 创建路由
	log.Println("正在创建路由...")
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API路由组
	api := r.Group("/api")
	{
		// 活动管理路由组
		activities := api.Group("/activities")
		{
			// 获取活动类别（无需认证）
			activities.GET("/categories", activityHandler.GetActivityCategories)
			// 获取活动模板（无需认证）
			activities.GET("/templates", activityHandler.GetActivityTemplates)

			// 需要认证的路由
			auth := activities.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 所有认证用户都可以访问的路由
				allUsers := auth.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("", activityHandler.GetActivities)                    // 获取活动列表
					allUsers.GET("/stats", activityHandler.GetActivityStats)           // 获取活动统计
					allUsers.POST("", activityHandler.CreateActivity)                  // 创建活动
					allUsers.POST("/batch", activityHandler.BatchCreateActivities)     // 批量创建活动
					allUsers.GET("/:id", activityHandler.GetActivity)                  // 获取活动详情
					allUsers.PUT("/:id", activityHandler.UpdateActivity)               // 更新活动
					allUsers.DELETE("/:id", activityHandler.DeleteActivity)            // 删除活动
					allUsers.POST("/:id/submit", activityHandler.SubmitActivity)       // 提交活动审核
					allUsers.POST("/:id/withdraw", activityHandler.WithdrawActivity)   // 撤回活动
					allUsers.GET("/deletable", activityHandler.GetDeletableActivities) // 获取可删除的活动列表
				}

				// 教师和管理员可以访问的路由
				teacherOrAdmin := auth.Group("")
				teacherOrAdmin.Use(permissionMiddleware.TeacherOrAdmin())
				{
					teacherOrAdmin.POST("/:id/review", activityHandler.ReviewActivity)          // 审核活动
					teacherOrAdmin.GET("/pending", activityHandler.GetPendingActivities)        // 获取待审核活动
					teacherOrAdmin.POST("/batch-delete", activityHandler.BatchDeleteActivities) // 批量删除活动
				}
			}

			// 参与者管理路由
			participants := activities.Group(":id")
			participants.Use(authMiddleware.AuthRequired())
			{
				// 活动创建者和管理员可以访问的路由
				ownerOrAdmin := participants.Group("")
				ownerOrAdmin.Use(permissionMiddleware.AllUsers()) // 这里需要在handler中检查权限
				{
					ownerOrAdmin.POST("/participants", participantHandler.AddParticipants)                  // 添加参与者
					ownerOrAdmin.PUT("/participants/batch-credits", participantHandler.BatchSetCredits)     // 批量设置学分
					ownerOrAdmin.PUT("/participants/:user_id/credits", participantHandler.SetSingleCredits) // 设置单个学分
					ownerOrAdmin.DELETE("/participants/:user_id", participantHandler.RemoveParticipant)     // 删除参与者
					ownerOrAdmin.GET("/participants", participantHandler.GetActivityParticipants)           // 获取参与者列表
				}

				// 学生退出活动路由
				studentOnly := participants.Group("")
				studentOnly.Use(permissionMiddleware.StudentOnly())
				{
					studentOnly.POST("/leave", participantHandler.LeaveActivity) // 退出活动
				}

				// 附件管理路由
				// 所有认证用户都可以访问的路由
				allUsers := participants.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("/attachments", attachmentHandler.GetAttachments)                             // 获取附件列表
					allUsers.GET("/attachments/:attachment_id/download", attachmentHandler.DownloadAttachment) // 下载附件
				}

				// 活动创建者、参与者和管理员可以访问的路由
				participantsOrAdmin := participants.Group("")
				participantsOrAdmin.Use(permissionMiddleware.AllUsers()) // 这里需要在handler中检查权限
				{
					participantsOrAdmin.POST("/attachments", attachmentHandler.UploadAttachment)             // 上传单个附件
					participantsOrAdmin.POST("/attachments/batch", attachmentHandler.BatchUploadAttachments) // 批量上传附件
				}

				// 上传者和管理员可以访问的路由
				uploaderOrAdmin := participants.Group("")
				uploaderOrAdmin.Use(permissionMiddleware.AllUsers()) // 这里需要在handler中检查权限
				{
					uploaderOrAdmin.PUT("/attachments/:attachment_id", attachmentHandler.UpdateAttachment)    // 更新附件信息
					uploaderOrAdmin.DELETE("/attachments/:attachment_id", attachmentHandler.DeleteAttachment) // 删除附件
				}
			}
		}

		// 申请管理路由组
		applications := api.Group("/applications")
		applications.Use(authMiddleware.AuthRequired())
		{
			// 所有认证用户都可以访问的路由
			allUsers := applications.Group("")
			allUsers.Use(permissionMiddleware.AllUsers())
			{
				allUsers.GET("", applicationHandler.GetUserApplications)       // 获取用户申请列表
				allUsers.GET("/:id", applicationHandler.GetApplication)        // 获取申请详情
				allUsers.GET("/stats", applicationHandler.GetApplicationStats) // 获取申请统计
				allUsers.GET("/export", applicationHandler.ExportApplications) // 导出申请数据
			}

			// 教师和管理员可以访问的路由
			teacherOrAdmin := applications.Group("")
			teacherOrAdmin.Use(permissionMiddleware.TeacherOrAdmin())
			{
				teacherOrAdmin.GET("/all", applicationHandler.GetAllApplications) // 获取所有申请
			}
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "credit-activity-service"})
	})

	// 启动服务器
	port := getEnv("PORT", "8083")
	log.Printf("Credit Activity Service starting on port %s", port)
	log.Println("服务启动完成，等待请求...")
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// initDatabase 初始化数据库连接
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

	// 测试连接
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

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	// 只进行表结构更新，不删除表，忽略错误
	log.Println("Skipping AutoMigrate to use init.sql schema")
	return nil
}

// createTriggers 创建数据库触发器
func createTriggers(db *gorm.DB) error {
	// 创建触发器函数
	triggerFunction := `
	CREATE OR REPLACE FUNCTION generate_applications_on_activity_approval()
	RETURNS TRIGGER AS $$
	BEGIN
		-- 只有当状态从非approved变为approved时才触发
		IF OLD.status != 'approved' AND NEW.status = 'approved' THEN
			-- 为所有参与者生成申请
			INSERT INTO applications (id, activity_id, user_id, status, applied_credits, awarded_credits, submitted_at, created_at, updated_at)
			SELECT 
				gen_random_uuid(),
				ap.activity_id,
				ap.user_id,
				'approved',
				ap.credits,
				ap.credits,
				NOW(),
				NOW(),
				NOW()
			FROM activity_participants ap
			WHERE ap.activity_id = NEW.id;
		END IF;
		
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;
	`

	// 删除已存在的触发器
	db.Exec(`DROP TRIGGER IF EXISTS trigger_generate_applications ON credit_activities;`)

	// 创建触发器函数
	if err := db.Exec(triggerFunction).Error; err != nil {
		return err
	}

	// 创建触发器
	trigger := `
	CREATE TRIGGER trigger_generate_applications
		AFTER UPDATE ON credit_activities
		FOR EACH ROW
		EXECUTE FUNCTION generate_applications_on_activity_approval();
	`

	if err := db.Exec(trigger).Error; err != nil {
		return err
	}

	log.Println("Database triggers created successfully")
	return nil
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
