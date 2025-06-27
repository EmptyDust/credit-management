package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type RemoteAuthMiddleware struct {
	authServiceURL string
}

func NewRemoteAuthMiddleware(authServiceURL string) *RemoteAuthMiddleware {
	return &RemoteAuthMiddleware{
		authServiceURL: authServiceURL,
	}
}

// AuthRequired 认证中间件（调用auth服务验证）
func (m *RemoteAuthMiddleware) AuthRequired() gin.HandlerFunc {
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

		// 调用auth服务验证token
		claims, err := m.validateTokenWithAuthService(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证失败: " + err.Error(), "data": nil})
			c.Abort()
			return
		}

		// 将claims存储到上下文中
		c.Set("claims", claims)
		c.Next()
	}
}

// validateTokenWithAuthService 调用auth服务验证token
func (m *RemoteAuthMiddleware) validateTokenWithAuthService(authHeader string) (jwt.MapClaims, error) {
	// 创建请求到auth服务
	req, err := http.NewRequest("POST", m.authServiceURL+"/api/auth/validate-token-with-claims", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求auth服务失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Valid   bool                   `json:"valid"`
			Claims  map[string]interface{} `json:"claims"`
			Message string                 `json:"message"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK || response.Code != 0 {
		return nil, fmt.Errorf("auth服务验证失败: %s", response.Message)
	}

	// 检查token是否有效
	if !response.Data.Valid {
		return nil, fmt.Errorf("token无效: %s", response.Data.Message)
	}

	// 将claims转换为jwt.MapClaims
	claims := jwt.MapClaims{}
	for key, value := range response.Data.Claims {
		claims[key] = value
	}

	return claims, nil
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
