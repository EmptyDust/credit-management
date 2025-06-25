package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"credit-management/credit-activity-service/models"
)

// getEnv 获取环境变量
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetUserInfo 获取用户信息（使用真实用户服务）
func GetUserInfo(userID string, authToken ...string) (*models.UserInfo, error) {
	// 获取用户服务URL（从环境变量或使用默认值）
	userServiceURL := GetEnv("USER_SERVICE_URL", "http://user-service:8084")
	if userServiceURL == "" {
		userServiceURL = "http://localhost:8084" // 本地开发默认值
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/api/users/%s", userServiceURL, userID)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 如果有认证令牌，添加到请求头
	if len(authToken) > 0 && authToken[0] != "" {
		// 如果传递的是完整的Authorization头，提取token部分
		token := authToken[0]
		if after, ok := strings.CutPrefix(token, "Bearer "); ok {
			token = after
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求用户服务失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("用户服务返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			UserID    string `json:"user_id"`
			Username  string `json:"username"`
			RealName  string `json:"real_name"`
			UserType  string `json:"user_type"`
			Status    string `json:"status"`
			StudentID string `json:"student_id,omitempty"`
			College   string `json:"college,omitempty"`
			Major     string `json:"major,omitempty"`
			Class     string `json:"class,omitempty"`
			Grade     string `json:"grade,omitempty"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查业务状态码
	if response.Code != 0 {
		return nil, fmt.Errorf("用户服务返回业务错误: %s", response.Message)
	}

	// 构建用户信息
	userInfo := &models.UserInfo{
		ID:        response.Data.UserID,
		Username:  response.Data.Username,
		Name:      response.Data.RealName,
		Role:      response.Data.UserType,
		StudentID: response.Data.StudentID,
	}

	return userInfo, nil
}

// IsStudent 检查用户是否为学生
func IsStudent(userID string, authToken ...string) bool {
	userInfo, err := GetUserInfo(userID, authToken...)
	if err != nil {
		// 如果获取用户信息失败，记录错误并返回false
		fmt.Printf("获取用户信息失败: %v\n", err)
		return false
	}

	// 检查用户类型是否为student
	return userInfo != nil && userInfo.Role == "student"
}
