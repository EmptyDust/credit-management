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

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetUserInfo(userID string, authToken ...string) (*models.UserInfo, error) {
	userServiceURL := GetEnv("USER_SERVICE_URL", "http://user-service:8084")
	if userServiceURL == "" {
		userServiceURL = "http://localhost:8084"
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	studentURL := fmt.Sprintf("%s/api/search/users?query=%s&user_type=student&page=1&page_size=1", userServiceURL, userID)
	userInfo, err := searchUserByType(client, studentURL, authToken...)
	if err == nil {
		return userInfo, nil
	}

	teacherURL := fmt.Sprintf("%s/api/search/users?query=%s&user_type=teacher&page=1&page_size=1", userServiceURL, userID)
	userInfo, err = searchUserByType(client, teacherURL, authToken...)
	if err == nil {
		return userInfo, nil
	}

	return nil, fmt.Errorf("用户不存在或无法获取用户信息")
}

func searchUserByType(client *http.Client, apiURL string, authToken ...string) (*models.UserInfo, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if len(authToken) > 0 && authToken[0] != "" {
		req.Header.Set("Authorization", authToken[0])
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("用户服务返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Users []struct {
				UserID     string `json:"user_id"`
				Username   string `json:"username"`
				RealName   string `json:"real_name"`
				UserType   string `json:"user_type,omitempty"`
				Status     string `json:"status,omitempty"`
				StudentID  string `json:"student_id,omitempty"`
				College    string `json:"college,omitempty"`
				Major      string `json:"major,omitempty"`
				Class      string `json:"class,omitempty"`
				Grade      string `json:"grade,omitempty"`
				Department string `json:"department,omitempty"`
				Title      string `json:"title,omitempty"`
			} `json:"users"`
			Total int64 `json:"total"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("用户服务返回错误: %s", response.Message)
	}

	if len(response.Data.Users) == 0 {
		return nil, fmt.Errorf("用户不存在")
	}

	user := response.Data.Users[0]

	userType := "student"
	if strings.Contains(apiURL, "user_type=teacher") {
		userType = "teacher"
	}

	return &models.UserInfo{
		UserID:     user.UserID,
		Username:   user.Username,
		RealName:   user.RealName,
		UserType:   userType,
		Status:     user.Status,
		StudentID:  user.StudentID,
		College:    user.College,
		Major:      user.Major,
		Class:      user.Class,
		Grade:      user.Grade,
		Department: user.Department,
		Title:      user.Title,
		// 向后兼容字段
		ID:   user.UserID,
		Name: user.RealName,
		Role: userType,
	}, nil
}

func IsStudent(userID string, authToken ...string) bool {
	userInfo, err := GetUserInfo(userID, authToken...)
	if err != nil {
		return false
	}
	return userInfo.UserType == "student"
}
