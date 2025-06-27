package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type HeaderAuthMiddleware struct{}

func NewHeaderAuthMiddleware() *HeaderAuthMiddleware {
	return &HeaderAuthMiddleware{}
}

// AuthRequired 认证中间件（从请求头获取用户信息）
func (m *HeaderAuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为内部服务通信
		internalService := c.GetHeader("X-Internal-Service")
		if internalService != "" {
			// 内部服务通信，设置系统用户信息
			c.Set("user_id", "system")
			c.Set("username", "system")
			c.Set("user_type", "admin")
			c.Set("claims", jwt.MapClaims{
				"user_id":   "system",
				"username":  "system",
				"user_type": "admin",
			})
			c.Next()
			return
		}

		// 从请求头获取用户信息（由API网关传递）
		userID := c.GetHeader("X-User-ID")
		username := c.GetHeader("X-Username")
		userType := c.GetHeader("X-User-Type")

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少用户信息", "data": nil})
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

type PermissionMiddleware struct{}

func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{}
}

func (m *PermissionMiddleware) AllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (m *PermissionMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
			c.Abort()
			return
		}

		claimsMap, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			c.Abort()
			return
		}

		userType, exists := claimsMap["user_type"]
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "缺少用户类型信息", "data": nil})
			c.Abort()
			return
		}

		if userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要管理员权限", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *PermissionMiddleware) StudentOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
			c.Abort()
			return
		}

		claimsMap, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			c.Abort()
			return
		}

		userType, exists := claimsMap["user_type"]
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "缺少用户类型信息", "data": nil})
			c.Abort()
			return
		}

		if userType != "student" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要学生权限", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *PermissionMiddleware) TeacherOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
			c.Abort()
			return
		}

		claimsMap, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			c.Abort()
			return
		}

		userType, exists := claimsMap["user_type"]
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "缺少用户类型信息", "data": nil})
			c.Abort()
			return
		}

		if userType != "teacher" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要教师权限", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *PermissionMiddleware) StudentTeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
			c.Abort()
			return
		}

		claimsMap, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			c.Abort()
			return
		}

		userType, exists := claimsMap["user_type"]
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "缺少用户类型信息", "data": nil})
			c.Abort()
			return
		}

		if userType != "student" && userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TeacherOrAdmin 教师和管理员可以访问
func (m *PermissionMiddleware) TeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
			c.Abort()
			return
		}

		claimsMap, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			c.Abort()
			return
		}

		userType, exists := claimsMap["user_type"]
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "缺少用户类型信息", "data": nil})
			c.Abort()
			return
		}

		// 检查用户类型是否为teacher或admin
		if userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要教师或管理员权限", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}
