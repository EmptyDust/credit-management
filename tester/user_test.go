package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func testUserRegister() {
	url := "http://localhost:8080/users/register"
	data := map[string]string{
		"username": "testuser1",
		"password": "testpass123",
		"role":     "student",
	}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("用户注册请求失败:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("用户注册响应:", string(body))
}

func testUserLogin() string {
	url := "http://localhost:8080/users/login"
	data := map[string]string{
		"username": "testuser1",
		"password": "testpass123",
	}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("用户登录请求失败:", err)
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("用户登录响应:", string(body))
	var res map[string]interface{}
	json.Unmarshal(body, &res)
	if token, ok := res["token"].(string); ok {
		return token
	}
	return ""
}

func testGetUser(token string) {
	url := "http://localhost:8080/users/testuser1"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("用户信息查询请求失败:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("用户信息查询响应:", string(body))
}

func main() {
	testUserRegister()
	token := testUserLogin()
	if token != "" {
		testGetUser(token)
	}
} 