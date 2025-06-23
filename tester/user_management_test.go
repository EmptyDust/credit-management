package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8081/api"

// 测试用户注册
func TestUserRegister(t *testing.T) {
	// 测试数据
	userData := map[string]interface{}{
		"username":  "testuser",
		"password":  "password123",
		"email":     "test@example.com",
		"phone":     "13800138000",
		"real_name": "测试用户",
		"user_type": "student",
	}

	jsonData, _ := json.Marshal(userData)
	resp, err := http.Post(baseURL+"/users/register", "application/json", bytes.NewBuffer(jsonData))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "用户注册成功", result["message"])

	resp.Body.Close()
}

// 测试用户登录
func TestUserLogin(t *testing.T) {
	loginData := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/users/login", "application/json", bytes.NewBuffer(jsonData))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.NotEmpty(t, result["token"])

	resp.Body.Close()
}

// 测试获取用户信息
func TestGetUser(t *testing.T) {
	resp, err := http.Get(baseURL + "/users/testuser")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "testuser", result["username"])

	resp.Body.Close()
}

// 测试文件上传
func TestFileUpload(t *testing.T) {
	// 创建测试文件
	file, _ := os.CreateTemp("", "test.txt")
	defer os.Remove(file.Name())
	file.WriteString("test content")
	file.Close()

	// 创建multipart请求
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	part, _ := writer.CreateFormFile("file", "test.txt")
	file, _ = os.Open(file.Name())
	io.Copy(part, file)
	file.Close()

	// 添加其他字段
	writer.WriteField("category", "document")
	writer.WriteField("description", "测试文件")
	writer.WriteField("is_public", "false")
	writer.Close()

	// 发送请求（需要先登录获取token）
	req, _ := http.NewRequest("POST", baseURL+"/files/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE") // 需要替换为实际的token

	client := &http.Client{}
	resp, err := client.Do(req)

	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// 测试权限管理
func TestPermissionManagement(t *testing.T) {
	// 测试创建角色
	roleData := map[string]interface{}{
		"name":        "test_role",
		"description": "测试角色",
		"is_system":   false,
	}

	jsonData, _ := json.Marshal(roleData)
	req, _ := http.NewRequest("POST", baseURL+"/permissions/roles", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE") // 需要替换为实际的token

	client := &http.Client{}
	resp, err := client.Do(req)

	if err == nil {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}
}

// 测试通知系统
func TestNotificationSystem(t *testing.T) {
	// 测试获取用户通知
	req, _ := http.NewRequest("GET", baseURL+"/notifications", nil)
	req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE") // 需要替换为实际的token

	client := &http.Client{}
	resp, err := client.Do(req)

	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// 测试用户统计
func TestUserStats(t *testing.T) {
	resp, err := http.Get(baseURL + "/users/stats")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.NotNil(t, result["total_users"])

	resp.Body.Close()
}

// 测试文件统计
func TestFileStats(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/files/stats", nil)
	req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE") // 需要替换为实际的token

	client := &http.Client{}
	resp, err := client.Do(req)

	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// 测试通知统计
func TestNotificationStats(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/notifications/stats", nil)
	req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE") // 需要替换为实际的token

	client := &http.Client{}
	resp, err := client.Do(req)

	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// 性能测试
func BenchmarkUserRegister(b *testing.B) {
	userData := map[string]interface{}{
		"username":  "benchuser",
		"password":  "password123",
		"email":     "bench@example.com",
		"phone":     "13800138001",
		"real_name": "性能测试用户",
		"user_type": "student",
	}

	jsonData, _ := json.Marshal(userData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userData["username"] = fmt.Sprintf("benchuser%d", i)
		userData["email"] = fmt.Sprintf("bench%d@example.com", i)
		jsonData, _ = json.Marshal(userData)

		resp, err := http.Post(baseURL+"/users/register", "application/json", bytes.NewBuffer(jsonData))
		if err == nil {
			resp.Body.Close()
		}
	}
}

// 并发测试
func TestConcurrentUserRegistration(t *testing.T) {
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			userData := map[string]interface{}{
				"username":  fmt.Sprintf("concurrentuser%d", id),
				"password":  "password123",
				"email":     fmt.Sprintf("concurrent%d@example.com", id),
				"phone":     fmt.Sprintf("13800138%03d", id),
				"real_name": fmt.Sprintf("并发用户%d", id),
				"user_type": "student",
			}

			jsonData, _ := json.Marshal(userData)
			resp, err := http.Post(baseURL+"/users/register", "application/json", bytes.NewBuffer(jsonData))

			if err == nil {
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// 主测试函数
func TestMain(m *testing.M) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 运行测试
	code := m.Run()

	// 清理测试数据
	cleanupTestData()

	os.Exit(code)
}

// 清理测试数据
func cleanupTestData() {
	// 这里可以添加清理测试数据的逻辑
	// 比如删除测试用户、文件等
	fmt.Println("Cleaning up test data...")
}
