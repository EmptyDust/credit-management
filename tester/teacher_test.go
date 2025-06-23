package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const teacherServiceURL = "http://localhost:8083"

// Teacher 教师模型
type Teacher struct {
	Username   string    `json:"username"`
	Name       string    `json:"name"`
	Contact    string    `json:"contact"`
	Email      string    `json:"email"`
	Department string    `json:"department"`
	Title      string    `json:"title"`
	Specialty  string    `json:"specialty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TeacherRequest 教师创建请求
type TeacherRequest struct {
	Username   string `json:"username"`
	Name       string `json:"name"`
	Contact    string `json:"contact"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Title      string `json:"title"`
	Specialty  string `json:"specialty"`
}

// TeacherUpdateRequest 教师更新请求
type TeacherUpdateRequest struct {
	Name       string `json:"name"`
	Contact    string `json:"contact"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Title      string `json:"title"`
	Specialty  string `json:"specialty"`
	Status     string `json:"status"`
}

func TestTeacherService(t *testing.T) {
	fmt.Println("开始测试教师信息服务...")

	// 测试1: 创建教师
	t.Run("创建教师", func(t *testing.T) {
		teacherReq := TeacherRequest{
			Username:   "teacher001",
			Name:       "张教授",
			Contact:    "13800138001",
			Email:      "zhang@university.edu",
			Department: "计算机科学学院",
			Title:      "教授",
			Specialty:  "人工智能",
		}

		jsonData, _ := json.Marshal(teacherReq)
		resp, err := http.Post(teacherServiceURL+"/api/teachers", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("创建教师失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("期望状态码201，实际得到%d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		fmt.Printf("✓ 教师创建成功: %s\n", teacherReq.Username)
	})

	// 测试2: 获取教师信息
	t.Run("获取教师信息", func(t *testing.T) {
		resp, err := http.Get(teacherServiceURL + "/api/teachers/teacher001")
		if err != nil {
			t.Fatalf("获取教师信息失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		var teacher Teacher
		json.NewDecoder(resp.Body).Decode(&teacher)
		if teacher.Username != "teacher001" {
			t.Errorf("期望用户名teacher001，实际得到%s", teacher.Username)
		}
		fmt.Printf("✓ 教师信息获取成功: %s\n", teacher.Name)
	})

	// 测试3: 更新教师信息
	t.Run("更新教师信息", func(t *testing.T) {
		updateReq := TeacherUpdateRequest{
			Name:       "张教授（更新）",
			Contact:    "13800138002",
			Email:      "zhang.updated@university.edu",
			Department: "计算机科学学院",
			Title:      "教授",
			Specialty:  "人工智能与机器学习",
			Status:     "active",
		}

		jsonData, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", teacherServiceURL+"/api/teachers/teacher001", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("更新教师信息失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		fmt.Printf("✓ 教师信息更新成功\n")
	})

	// 测试4: 获取所有教师
	t.Run("获取所有教师", func(t *testing.T) {
		resp, err := http.Get(teacherServiceURL + "/api/teachers")
		if err != nil {
			t.Fatalf("获取所有教师失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		var teachers []Teacher
		json.NewDecoder(resp.Body).Decode(&teachers)
		fmt.Printf("✓ 获取所有教师成功，共%d名教师\n", len(teachers))
	})

	// 测试5: 根据院系获取教师
	t.Run("根据院系获取教师", func(t *testing.T) {
		resp, err := http.Get(teacherServiceURL + "/api/teachers/department/计算机科学学院")
		if err != nil {
			t.Fatalf("根据院系获取教师失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		var teachers []Teacher
		json.NewDecoder(resp.Body).Decode(&teachers)
		fmt.Printf("✓ 根据院系获取教师成功，计算机科学学院有%d名教师\n", len(teachers))
	})

	// 测试6: 根据职称获取教师
	t.Run("根据职称获取教师", func(t *testing.T) {
		resp, err := http.Get(teacherServiceURL + "/api/teachers/title/教授")
		if err != nil {
			t.Fatalf("根据职称获取教师失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		var teachers []Teacher
		json.NewDecoder(resp.Body).Decode(&teachers)
		fmt.Printf("✓ 根据职称获取教师成功，教授有%d名\n", len(teachers))
	})

	// 测试7: 搜索教师
	t.Run("搜索教师", func(t *testing.T) {
		resp, err := http.Get(teacherServiceURL + "/api/teachers/search?q=人工智能")
		if err != nil {
			t.Fatalf("搜索教师失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		var teachers []Teacher
		json.NewDecoder(resp.Body).Decode(&teachers)
		fmt.Printf("✓ 搜索教师成功，找到%d名相关教师\n", len(teachers))
	})

	// 测试8: 获取活跃教师
	t.Run("获取活跃教师", func(t *testing.T) {
		resp, err := http.Get(teacherServiceURL + "/api/teachers/active")
		if err != nil {
			t.Fatalf("获取活跃教师失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		var teachers []Teacher
		json.NewDecoder(resp.Body).Decode(&teachers)
		fmt.Printf("✓ 获取活跃教师成功，共%d名活跃教师\n", len(teachers))
	})

	// 测试9: 创建更多教师用于测试
	t.Run("创建更多教师", func(t *testing.T) {
		teachers := []TeacherRequest{
			{
				Username:   "teacher002",
				Name:       "李副教授",
				Contact:    "13800138003",
				Email:      "li@university.edu",
				Department: "计算机科学学院",
				Title:      "副教授",
				Specialty:  "软件工程",
			},
			{
				Username:   "teacher003",
				Name:       "王讲师",
				Contact:    "13800138004",
				Email:      "wang@university.edu",
				Department: "数学学院",
				Title:      "讲师",
				Specialty:  "应用数学",
			},
			{
				Username:   "teacher004",
				Name:       "陈教授",
				Contact:    "13800138005",
				Email:      "chen@university.edu",
				Department: "物理学院",
				Title:      "教授",
				Specialty:  "量子物理",
			},
		}

		for _, teacherReq := range teachers {
			jsonData, _ := json.Marshal(teacherReq)
			resp, err := http.Post(teacherServiceURL+"/api/teachers", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("创建教师失败: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				t.Errorf("创建教师%s失败，状态码: %d", teacherReq.Username, resp.StatusCode)
			}
		}

		fmt.Printf("✓ 成功创建%d名额外教师\n", len(teachers))
	})

	// 测试10: 删除教师
	t.Run("删除教师", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", teacherServiceURL+"/api/teachers/teacher001", nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("删除教师失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码200，实际得到%d", resp.StatusCode)
		}

		fmt.Printf("✓ 教师删除成功\n")
	})

	fmt.Println("教师信息服务测试完成！")
}
