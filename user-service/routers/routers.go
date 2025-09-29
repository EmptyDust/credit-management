package routers

import (
	"credit-management/user-service/handlers"
	"credit-management/user-service/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouters(userHandler *handlers.UserHandler) *gin.Engine {
	authMiddleware := middleware.NewHeaderAuthMiddleware()
	permissionMiddleware := middleware.NewPermissionMiddleware()

	r := gin.Default()

	// 添加CORS中间件
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api")
	{
		// 公共配置选项（无需认证）
		api.GET("/config/options", handlers.GetOptions)

		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)

			// 需要认证的路由
			auth := users.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 所有认证用户都可以访问的路由
				allUsers := auth.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("/profile", userHandler.GetUser)    // 获取当前用户信息
					allUsers.PUT("/profile", userHandler.UpdateUser) // 更新当前用户信息
					allUsers.GET("/stats", userHandler.GetUserStats) // 获取用户统计信息
					allUsers.GET("/:id", userHandler.GetUser)        // 获取指定用户信息

					// 新增：用户自助修改密码
					allUsers.POST("/change_password", userHandler.ChangePassword) // 修改自己密码

					// 新增：获取用户活动记录（预留）
					allUsers.GET("/activity", userHandler.GetUserActivity)     // 当前用户活动
					allUsers.GET("/:id/activity", userHandler.GetUserActivity) // 指定用户活动（管理员/教师）
				}

				// 管理员路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("/teachers", userHandler.CreateTeacher) // 管理员创建教师
					admin.POST("/students", userHandler.CreateStudent) // 管理员创建学生
					admin.PUT("/:id", userHandler.UpdateUser)          // 更新指定用户信息
					admin.DELETE("/:id", userHandler.DeleteUser)       // 删除用户

					// 新增：批量删除、批量状态、重置密码、导出
					admin.POST("/batch_delete", userHandler.BatchDeleteUsers)      // 批量删除
					admin.POST("/batch_status", userHandler.BatchUpdateUserStatus) // 批量状态
					admin.POST("/reset_password", userHandler.ResetPassword)       // 重置密码
					admin.GET("/export", userHandler.ExportUsers)                  // 导出用户数据

					// 新增：CSV导入功能
					admin.POST("/import-csv", userHandler.ImportUsersFromCSV)      // 从CSV导入用户
					admin.GET("/csv-template", userHandler.GetUserCSVTemplate)     // 获取CSV模板
					admin.POST("/import", userHandler.ImportUsers)                 // 通用导入接口（支持Excel和CSV）
					admin.GET("/excel-template", userHandler.GetUserExcelTemplate) // 获取Excel模板

					// 学生、教师和管理员可以访问的路由（基于角色的权限控制）
					studentTeacherOrAdmin := auth.Group("")
					studentTeacherOrAdmin.Use(permissionMiddleware.StudentTeacherOrAdmin())
					{
						studentTeacherOrAdmin.GET("/stats/students", userHandler.GetStudentStats) // 获取学生统计信息
						studentTeacherOrAdmin.GET("/stats/teachers", userHandler.GetTeacherStats) // 获取教师统计信息
					}
				}
			}
		}

		// 学生相关路由
		students := api.Group("/students")
		{
			// 需要认证的路由
			auth := students.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 仅管理员可以访问的路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("", userHandler.CreateStudent)   // 创建学生
					admin.PUT(":id", userHandler.UpdateUser)    // 更新学生
					admin.DELETE(":id", userHandler.DeleteUser) // 删除学生
				}
			}
		}

		// 教师相关路由
		teachers := api.Group("/teachers")
		{
			// 需要认证的路由
			auth := teachers.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 仅管理员可以访问的路由
				admin := auth.Group("")
				admin.Use(permissionMiddleware.AdminOnly())
				{
					admin.POST("", userHandler.CreateTeacher)   // 创建教师
					admin.PUT(":id", userHandler.UpdateUser)    // 更新教师
					admin.DELETE(":id", userHandler.DeleteUser) // 删除教师
				}
			}
		}

		// 搜索相关路由
		search := api.Group("/search")
		{
			// 需要认证的路由
			auth := search.Group("")
			auth.Use(authMiddleware.AuthRequired())
			{
				// 所有认证用户都可以访问的路由
				allUsers := auth.Group("")
				allUsers.Use(permissionMiddleware.AllUsers())
				{
					allUsers.GET("/users", userHandler.SearchUsers) // 搜索用户（使用视图进行权限控制）
				}
			}
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	return r
}
