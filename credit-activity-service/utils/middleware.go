package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HeaderAuthMiddleware 基于请求头的认证中间件
type HeaderAuthMiddleware struct{}

// NewHeaderAuthMiddleware 创建新的认证中间件
func NewHeaderAuthMiddleware() *HeaderAuthMiddleware {
	return &HeaderAuthMiddleware{}
}

// AuthRequired 需要认证的中间件
func (m *HeaderAuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		username := c.GetHeader("X-Username")
		userType := c.GetHeader("X-User-Type")

		if userID == "" || username == "" || userType == "" {
			SendUnauthorized(c)
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("id", userID)
		c.Set("username", username)
		c.Set("user_type", userType)

		c.Next()
	}
}

// PermissionMiddleware 权限控制中间件
type PermissionMiddleware struct{}

// NewPermissionMiddleware 创建新的权限中间件
func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{}
}

// AllUsers 所有认证用户都可以访问
func (m *PermissionMiddleware) AllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 所有认证用户都可以访问，无需额外检查
		c.Next()
	}
}

// StudentOnly 仅学生可以访问
func (m *PermissionMiddleware) StudentOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			SendUnauthorized(c)
			c.Abort()
			return
		}

		if userType != "student" {
			SendForbidden(c, "仅学生可以访问此功能")
			c.Abort()
			return
		}

		c.Next()
	}
}

// TeacherOrAdmin 教师或管理员可以访问
func (m *PermissionMiddleware) TeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			SendUnauthorized(c)
			c.Abort()
			return
		}

		if userType != "teacher" && userType != "admin" {
			SendForbidden(c, "需要教师或管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly 仅管理员可以访问
func (m *PermissionMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			SendUnauthorized(c)
			c.Abort()
			return
		}

		if userType != "admin" {
			SendForbidden(c, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ActivityOwnerOrTeacherOrAdmin 活动所有者、教师或管理员可以访问
func (m *PermissionMiddleware) ActivityOwnerOrTeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("id")
		if !exists {
			SendUnauthorized(c)
			c.Abort()
			return
		}

		userType, _ := c.Get("user_type")
		activityID := c.Param("id")

		// 教师或管理员可以直接访问
		if userType == "teacher" || userType == "admin" {
			c.Next()
			return
		}

		// 学生需要检查是否为活动所有者
		if userType == "student" {
			// 这里可以添加数据库查询来检查活动所有者
			// 暂时使用简单的参数检查
			if activityID != "" {
				c.Next()
				return
			}
		}

		SendForbidden(c, "无权限访问此资源")
		c.Abort()
	}
}

// CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-ID, X-Username, X-User-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 自定义日志格式
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			SendInternalServerError(c, fmt.Errorf("%s", err))
		} else {
			SendInternalServerError(c, fmt.Errorf("未知错误"))
		}
		c.Abort()
	})
}
