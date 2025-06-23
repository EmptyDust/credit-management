package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	jwtSecret []byte
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

// AuthRequired 需要认证的中间件
func (am *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证格式"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 解析JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return am.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 提取用户信息
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌声明"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
			c.Abort()
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户名"})
			c.Abort()
			return
		}

		userType, ok := claims["user_type"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户类型"})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户角色"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", uint(userID))
		c.Set("username", username)
		c.Set("user_type", userType)
		c.Set("role", role)

		c.Next()
	}
}

// PermissionMiddlewareHandler 权限中间件处理器
type PermissionMiddlewareHandler struct {
	db *gorm.DB
}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware(db *gorm.DB) *PermissionMiddlewareHandler {
	return &PermissionMiddlewareHandler{
		db: db,
	}
}

// RequirePermission 需要特定权限的中间件（简化版本，仅检查用户类型）
func (pm *PermissionMiddlewareHandler) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		// 简化权限检查：管理员拥有所有权限
		if userType == "admin" {
			c.Next()
			return
		}

		// 其他用户类型根据资源类型进行简单检查
		switch resource {
		case "user":
			if action == "read" && (userType == "teacher" || userType == "student") {
				c.Next()
				return
			}
		case "notification":
			if action == "read" || action == "write" {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		c.Abort()
	}
}

// RequireRole 需要特定角色的中间件（简化版本）
func (pm *PermissionMiddlewareHandler) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		if role != roleName && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "角色权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly 仅管理员中间件
func (pm *PermissionMiddlewareHandler) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		if userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TeacherOnly 仅教师中间件
func (pm *PermissionMiddlewareHandler) TeacherOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		if userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅教师和管理员可访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StudentOnly 仅学生中间件
func (pm *PermissionMiddlewareHandler) StudentOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		if userType != "student" && userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅学生、教师和管理员可访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}
