package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"credit-management/auth-service/models"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

// AuthRequired 认证中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 解析token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 获取用户信息
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		username, _ := claims["username"].(string)
		userType, _ := claims["user_type"].(string)

		c.Set("user_id", userID)
		c.Set("username", username)
		c.Set("user_type", userType)

		c.Next()
	}
}

type PermissionMiddleware struct {
	db *gorm.DB
}

func NewPermissionMiddleware(db *gorm.DB) *PermissionMiddleware {
	return &PermissionMiddleware{db: db}
}

// RequirePermission 权限检查中间件
func (m *PermissionMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// 检查用户是否有指定权限
		var userPermission models.UserPermission
		var permission models.Permission

		// 先查找权限
		if err := m.db.Where("resource = ? AND action = ?", resource, action).First(&permission).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission not found"})
			c.Abort()
			return
		}

		// 检查用户直接权限
		if err := m.db.Where("user_id = ? AND permission_id = ?", userID, permission.ID).First(&userPermission).Error; err == nil {
			c.Next()
			return
		}

		// 检查用户角色权限
		var userRoles []models.UserRole
		if err := m.db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		for _, userRole := range userRoles {
			var rolePermission models.RolePermission
			if err := m.db.Where("role_id = ? AND permission_id = ?", userRole.RoleID, permission.ID).First(&rolePermission).Error; err == nil {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// RequireRole 角色检查中间件
func (m *PermissionMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// 查找角色
		var role models.Role
		if err := m.db.Where("name = ?", roleName).First(&role).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}

		// 检查用户是否有该角色
		var userRole models.UserRole
		if err := m.db.Where("user_id = ? AND role_id = ?", userID, role.ID).First(&userRole).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireUserType 用户类型检查中间件
func (m *PermissionMiddleware) RequireUserType(userType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userTypeFromToken, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		if userTypeFromToken != userType {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient user type"})
			c.Abort()
			return
		}

		c.Next()
	}
}
