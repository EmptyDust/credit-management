package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// HeaderAuthMiddleware 认证中间件
type HeaderAuthMiddleware struct{}

// NewHeaderAuthMiddleware 创建认证中间件
func NewHeaderAuthMiddleware() *HeaderAuthMiddleware {
	return &HeaderAuthMiddleware{}
}

// AuthRequired 认证必需中间件
func (m *HeaderAuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取用户信息（由API网关传递）
		userID := c.GetHeader("X-User-ID")
		username := c.GetHeader("X-Username")
		userType := c.GetHeader("X-User-Type")

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少用户信息",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 创建claims对象
		claims := jwt.MapClaims{
			"user_id":   userID,
			"username":  username,
			"user_type": userType,
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", userID)
		c.Set("username", username)
		c.Set("user_type", userType)
		c.Set("claims", claims)

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

// ActivityOwnerOrTeacherOrAdmin 活动所有者或教师或管理员权限
func (m *PermissionMiddleware) ActivityOwnerOrTeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		userType, _ := c.Get("user_type")

		// 教师或管理员直接通过
		if userType == "teacher" || userType == "admin" {
			c.Next()
			return
		}

		// 对于学生，需要检查是否为活动所有者
		activityID := c.Param("id")
		if activityID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "活动ID不能为空",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 这里需要查询数据库检查是否为活动所有者
		// 由于中间件无法直接访问数据库，我们将这个检查放在具体的handler中
		// 这个中间件主要用于路由级别的权限控制
		c.Next()
	}
}
