package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	authServiceURL string
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(authServiceURL string) *AuthMiddleware {
	return &AuthMiddleware{authServiceURL: authServiceURL}
}

// AuthRequired 需要认证的中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 调用auth-service验证权限
		userInfo, err := m.validatePermission(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证失败: " + err.Error(), "data": nil})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", userInfo["user_id"])
		c.Set("username", userInfo["username"])
		c.Set("user_type", userInfo["user_type"])
		c.Set("role", userInfo["role"])
		c.Set("status", userInfo["status"])
		c.Set("real_name", userInfo["real_name"])

		c.Next()
	}
}

// validatePermission 调用auth-service验证权限
func (m *AuthMiddleware) validatePermission(c *gin.Context) (map[string]interface{}, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("未提供认证令牌")
	}

	// 创建请求到auth-service
	req, err := http.NewRequest("GET", m.authServiceURL+"/api/auth/validate-permission", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求auth-service失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var response struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("权限验证失败: %s", response.Message)
	}

	return response.Data, nil
}

// PermissionMiddleware 权限中间件
type PermissionMiddleware struct{}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{}
}

// AdminOnly 仅管理员中间件
func (pm *PermissionMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
			c.Abort()
			return
		}

		if userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "仅管理员可访问", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StudentOrAdmin 学生或管理员中间件
func (pm *PermissionMiddleware) StudentOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
			c.Abort()
			return
		}

		if userType != "student" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "仅学生和管理员可访问", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StudentTeacherOrAdmin 学生、教师或管理员中间件
func (pm *PermissionMiddleware) StudentTeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
			c.Abort()
			return
		}

		if userType != "student" && userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "仅学生、教师和管理员可访问", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TeacherOrAdmin 教师或管理员中间件
func (pm *PermissionMiddleware) TeacherOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证", "data": nil})
			c.Abort()
			return
		}

		if userType != "teacher" && userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "仅教师和管理员可访问", "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}
