package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

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

		userID, ok := claims["uuid"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		username, _ := claims["username"].(string)
		userType, _ := claims["user_type"].(string)

		c.Set("uuid", userID)
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

// RateLimitMiddleware 速率限制中间件
type RateLimitMiddleware struct {
	redis  *RedisClient
	limit  int64         // 最大请求次数
	window time.Duration // 时间窗口
}

// NewRateLimitMiddleware 创建速率限制中间件
// limit: 在时间窗口内允许的最大请求次数
// window: 时间窗口（例如：1分钟）
func NewRateLimitMiddleware(redis *RedisClient, limit int64, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redis:  redis,
		limit:  limit,
		window: window,
	}
}

// LimitByIP 基于IP地址的速率限制
func (m *RateLimitMiddleware) LimitByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:ip:%s", clientIP)

		// 检查速率限制
		ctx := context.Background()
		count, exceeded, err := m.redis.IncrementRateLimit(ctx, key, m.limit, m.window)

		if err != nil {
			// 如果Redis出错，记录日志但不阻止请求
			c.Next()
			return
		}

		// 设置速率限制响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, m.limit-count)))

		if exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("请求过于频繁，请在%v后重试", m.window),
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LimitByUsername 基于用户名的速率限制（用于登录等场景）
func (m *RateLimitMiddleware) LimitByUsername() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求体中提取用户名
		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
		}

		// 使用ShouldBindBodyWith允许请求体被多次读取
		if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			c.Next()
			return
		}

		// 确定用户标识（优先使用username，然后是email，最后是phone）
		userIdentifier := req.Username
		if userIdentifier == "" {
			userIdentifier = req.Email
		}
		if userIdentifier == "" {
			userIdentifier = req.Phone
		}

		if userIdentifier == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("rate_limit:user:%s", userIdentifier)

		// 检查速率限制
		ctx := context.Background()
		count, exceeded, err := m.redis.IncrementRateLimit(ctx, key, m.limit, m.window)

		if err != nil {
			// 如果Redis出错，记录日志但不阻止请求
			c.Next()
			return
		}

		// 设置速率限制响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, m.limit-count)))

		if exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("登录尝试过于频繁，请在%v后重试", m.window),
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// max 返回两个int64中较大的值
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
