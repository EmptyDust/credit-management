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
	userInfo, err := searchUserByType(client, studentURL)
	if err == nil {
		return userInfo, nil
	}

	teacherURL := fmt.Sprintf("%s/api/search/users?query=%s&user_type=teacher&page=1&page_size=1", userServiceURL, userID)
	userInfo, err = searchUserByType(client, teacherURL)
	if err == nil {
		return userInfo, nil
	}

	return nil, fmt.Errorf("用户不存在或无法获取用户信息")
}

func searchUserByType(client *http.Client, apiURL string) (*models.UserInfo, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 总是使用内部服务通信，不传递JWT token
	req.Header.Set("X-Internal-Service", "credit-activity-service")
	req.Header.Set("X-User-ID", "system")
	req.Header.Set("X-Username", "system")
	req.Header.Set("X-User-Type", "admin")
	fmt.Printf("使用内部服务通信，URL: %s\n", apiURL)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	fmt.Printf("用户服务响应状态码: %d, 响应体: %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("用户服务返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Users []map[string]interface{} `json:"users"`
			Total int64                    `json:"total"`
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

	// 从 map[string]interface{} 中提取用户信息
	userID, _ := user["user_id"].(string)
	username, _ := user["username"].(string)
	realName, _ := user["real_name"].(string)
	userType, _ := user["user_type"].(string)
	status, _ := user["status"].(string)
	studentID, _ := user["student_id"].(string)
	college, _ := user["college"].(string)
	major, _ := user["major"].(string)
	class, _ := user["class"].(string)
	grade, _ := user["grade"].(string)
	department, _ := user["department"].(string)
	title, _ := user["title"].(string)

	// 如果 userType 为空，根据 URL 判断
	if userType == "" {
		if strings.Contains(apiURL, "user_type=teacher") {
			userType = "teacher"
		} else {
			userType = "student"
		}
	}

	fmt.Printf("成功获取用户信息: %s (%s)\n", username, userType)

	return &models.UserInfo{
		UserID:     userID,
		Username:   username,
		RealName:   realName,
		UserType:   userType,
		Status:     status,
		StudentID:  studentID,
		College:    college,
		Major:      major,
		Class:      class,
		Grade:      grade,
		Department: department,
		Title:      title,
		// 向后兼容字段
		ID:   userID,
		Name: realName,
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
