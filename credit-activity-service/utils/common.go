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
	"log"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetUserInfo(userID string, authToken ...string) (*models.UserInfo, error) {
	userServiceURL := GetEnv("USER_SERVICE_URL", "http://user-service:8084")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 优先使用内部服务通信头部
	useInternal := GetEnv("USER_SERVICE_INTERNAL", "true") == "true"
	internalName := GetEnv("INTERNAL_SERVICE_NAME", "credit-activity-service")
	serviceToken := GetEnv("USER_SERVICE_BEARER", "")

	// 如果未启用内部模式，则回退到上游的 Authorization
	var callerAuth string
	if !useInternal {
		if len(authToken) == 0 || strings.TrimSpace(authToken[0]) == "" {
			return nil, fmt.Errorf("缺少认证令牌 Authorization")
		}
		callerAuth = strings.TrimSpace(authToken[0])
	}

	studentURL := fmt.Sprintf("%s/api/search/users?query=%s&user_type=student&page=1&page_size=1", userServiceURL, userID)
	userInfo, err := searchUserByType(client, studentURL, useInternal, internalName, serviceToken, callerAuth)
	if err == nil {
		return userInfo, nil
	}
	log.Printf("GetUserInfo: student lookup failed for id=%s, err=%v", userID, err)

	teacherURL := fmt.Sprintf("%s/api/search/users?query=%s&user_type=teacher&page=1&page_size=1", userServiceURL, userID)
	userInfo, err = searchUserByType(client, teacherURL, useInternal, internalName, serviceToken, callerAuth)
	if err == nil {
		return userInfo, nil
	}
	log.Printf("GetUserInfo: teacher lookup failed for id=%s, err=%v", userID, err)

	return nil, fmt.Errorf("用户不存在或无法获取用户信息")
}

func searchUserByType(client *http.Client, apiURL string, useInternal bool, internalName, serviceToken, callerAuth string) (*models.UserInfo, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if useInternal {
		// 使用内部服务头，用户服务中间件将视为系统管理员
		req.Header.Set("X-Internal-Service", internalName)
		if strings.TrimSpace(serviceToken) != "" {
			// 可选：带上服务间 Bearer 以便审计
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(serviceToken)))
		}
		log.Printf("Calling user-service as internal service '%s'. URL: %s", internalName, apiURL)
	} else {
		trimmed := strings.TrimSpace(callerAuth)
		if trimmed == "" {
			return nil, fmt.Errorf("缺少认证令牌 Authorization")
		}
		req.Header.Set("Authorization", trimmed)
		log.Printf("Calling user-service with client Authorization. URL: %s", apiURL)
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

	log.Printf("User-service response status=%d body=%s", resp.StatusCode, string(body))

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

	userUUID, _ := user["uuid"].(string)
	username, _ := user["username"].(string)
	realName, _ := user["real_name"].(string)
	userType, _ := user["user_type"].(string)
	status, _ := user["status"].(string)
	college, _ := user["college"].(string)
	major, _ := user["major"].(string)
	class, _ := user["class"].(string)
	grade, _ := user["grade"].(string)
	department, _ := user["department"].(string)
	title, _ := user["title"].(string)

	if userType == "" {
		if strings.Contains(apiURL, "user_type=teacher") {
			userType = "teacher"
		} else {
			userType = "student"
		}
	}

	// 根据用户类型获取学号或工号
	var studentID string
	if userType == "student" {
		if sid, ok := user["student_id"].(string); ok && sid != "" {
			studentID = sid
		} else if sidPtr, ok := user["student_id"].(*string); ok && sidPtr != nil {
			studentID = *sidPtr
		}
	} else if userType == "teacher" {
		if tid, ok := user["teacher_id"].(string); ok && tid != "" {
			studentID = tid // 对于教师，工号也存储在StudentID字段中以便统一处理
		} else if tidPtr, ok := user["teacher_id"].(*string); ok && tidPtr != nil {
			studentID = *tidPtr
		}
	}

	return &models.UserInfo{
		UUID:       userUUID,
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
	}, nil
}

func IsStudent(userID string, authToken ...string) bool {
	userInfo, err := GetUserInfo(userID, authToken...)
	if err != nil {
		return false
	}
	return userInfo.UserType == "student"
}
