package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct{}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// AuthRequired 认证必需中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误",
				"data":    nil,
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token不能为空",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 这里应该调用auth服务验证token
		// 暂时使用模拟数据
		userID := "user-uuid"
		userType := "student"

		// 设置用户信息到上下文
		c.Set("user_id", userID)
		c.Set("user_type", userType)

		c.Next()
	}
}

// PermissionMiddleware 权限中间件
type PermissionMiddleware struct{}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{}
}

// AllUsers 所有认证用户都可以访问
func (m *PermissionMiddleware) AllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 所有认证用户都可以访问
		c.Next()
	}
}

// TeacherOrAdmin 教师或管理员权限
func (m *PermissionMiddleware) TeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		if userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly 仅管理员权限
func (m *PermissionMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		if userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StudentOnly 仅学生权限
func (m *PermissionMiddleware) StudentOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		if userType != "student" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
