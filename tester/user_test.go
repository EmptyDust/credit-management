package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
)

var authToken string
var testUsername string

func init() {
	// 使用时间戳生成唯一用户名
	testUsername = "testuser" + strconv.FormatInt(time.Now().Unix(), 10)
	fmt.Printf("使用测试用户名: %s\n", testUsername)
}

func TestUserRegister(t *testing.T) {
	url := "http://localhost:8001/users/register"
	data := map[string]string{
		"username": testUsername,
		"password": "testpass123",
		"role":     "student",
	}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Errorf("用户注册请求失败: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("用户注册响应:", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Errorf("用户注册失败，状态码: %d", resp.StatusCode)
	}
}

func TestUserLogin(t *testing.T) {
	url := "http://localhost:8001/users/login"
	data := map[string]string{
		"username": testUsername,
		"password": "testpass123",
	}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Errorf("用户登录请求失败: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("用户登录响应:", string(body))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("用户登录失败，状态码: %d", resp.StatusCode)
		return
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)
	if token, ok := res["token"].(string); ok {
		authToken = token
		fmt.Println("获取到认证token:", token[:20]+"...")
	} else {
		t.Error("登录响应中没有找到token")
	}
}

func TestGetUser(t *testing.T) {
	if authToken == "" {
		t.Skip("跳过用户信息查询测试，因为登录失败或token为空")
		return
	}

	url := "http://localhost:8001/users/" + testUsername
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("用户信息查询请求失败: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("用户信息查询响应:", string(body))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("用户信息查询失败，状态码: %d", resp.StatusCode)
	}
}

// 集成测试，按顺序执行所有测试
func TestUserManagementIntegration(t *testing.T) {
	t.Run("Register", TestUserRegister)
	t.Run("Login", TestUserLogin)
	t.Run("GetUser", TestGetUser)
}
