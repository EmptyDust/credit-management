package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// AuthRequired 认证中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少认证令牌", "data": nil})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证格式", "data": nil})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 解析JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证令牌: " + err.Error(), "data": nil})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证令牌已过期", "data": nil})
			c.Abort()
			return
		}

		// 提取claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证信息", "data": nil})
			c.Abort()
			return
		}

		// 将claims存储到上下文中
		c.Set("claims", claims)
		c.Next()
	}
}

type PermissionMiddleware struct{}

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

// AdminOnly 仅管理员可以访问
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

		// 调试日志
		fmt.Printf("DEBUG AdminOnly: user_type = %v, type = %T\n", userType, userType)

		if userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "需要管理员权限", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StudentOnly 仅学生可以访问
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

// TeacherOnly 仅教师可以访问
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

// StudentTeacherOrAdmin 学生、教师和管理员都可以访问
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

		// 检查用户类型是否为student、teacher或admin
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
