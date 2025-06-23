package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"student-info-service/models"
	"testing"
	"time"
)

var studentAPI = "http://localhost:8006/students"
var testStudentNo string

func init() {
	// 使用时间戳生成唯一学号
	testStudentNo = "2024" + strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
}

// 辅助函数：创建学生用于测试
func createTestStudent(t *testing.T) models.Student {
	// 每次都生成唯一学号和用户名
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	student := models.Student{
		UserID:    uint(time.Now().UnixNano()), // 唯一UserID
		Username:  "student_" + uniqueSuffix,
		Name:      "测试学生",
		StudentNo: "sn_" + uniqueSuffix,
		College:   "计算机学院",
		Major:     "软件工程",
		Class:     "软工2001",
		Contact:   "13800138000",
	}

	jsonData, _ := json.Marshal(student)
	resp, err := http.Post(studentAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建学生失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var createdStudent models.Student
	if err := json.NewDecoder(resp.Body).Decode(&createdStudent); err != nil {
		t.Fatalf("无法解析创建的学生响应: %v", err)
	}

	return createdStudent
}

func TestCreateStudent(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	student := models.Student{
		UserID:    uint(time.Now().UnixNano()), // 唯一UserID
		Username:  "student_create_" + uniqueSuffix,
		Name:      "新生",
		StudentNo: "new_" + uniqueSuffix,
		College:   "信息学院",
	}

	jsonData, _ := json.Marshal(student)
	resp, err := http.Post(studentAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建学生失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建学生成功")
}

func TestGetStudent(t *testing.T) {
	createdStudent := createTestStudent(t)

	resp, err := http.Get(studentAPI + "/" + createdStudent.StudentNo)
	if err != nil {
		t.Fatalf("查询学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("查询学生失败，状态码: %d", resp.StatusCode)
	}

	var foundStudent models.Student
	if err := json.NewDecoder(resp.Body).Decode(&foundStudent); err != nil {
		t.Fatalf("无法解析查询的学生响应: %v", err)
	}

	if foundStudent.StudentNo != createdStudent.StudentNo {
		t.Errorf("期望学号 %s, 得到 %s", createdStudent.StudentNo, foundStudent.StudentNo)
	}
	fmt.Println("查询学生成功:", foundStudent.Name)
}

func TestUpdateStudent(t *testing.T) {
	createdStudent := createTestStudent(t)

	updateData := map[string]string{
		"college": "人工智能学院",
		"contact": "13900139000",
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest(http.MethodPut, studentAPI+"/"+createdStudent.StudentNo, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("更新学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("更新学生失败，状态码: %d", resp.StatusCode)
	}

	var updatedStudent models.Student
	if err := json.NewDecoder(resp.Body).Decode(&updatedStudent); err != nil {
		t.Fatalf("无法解析更新的学生响应: %v", err)
	}

	if updatedStudent.College != updateData["college"] {
		t.Errorf("期望学院 %s, 得到 %s", updateData["college"], updatedStudent.College)
	}
	fmt.Println("更新学生成功，新学院:", updatedStudent.College)
}

func TestDeleteStudent(t *testing.T) {
	createdStudent := createTestStudent(t)

	req, _ := http.NewRequest(http.MethodDelete, studentAPI+"/"+createdStudent.StudentNo, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("删除学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("删除学生失败，状态码: %d", resp.StatusCode)
	}
	fmt.Println("删除学生成功")

	// 验证是否真的被删除
	resp, err = http.Get(studentAPI + "/" + createdStudent.StudentNo)
	if err != nil {
		t.Fatalf("验证删除时请求失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("期望状态码 404, 得到 %d", resp.StatusCode)
	}
	fmt.Println("验证删除成功")
}

func TestGetAllStudents(t *testing.T) {
	// 先创建几个学生
	createTestStudent(t)
	createTestStudent(t)

	resp, err := http.Get(studentAPI)
	if err != nil {
		t.Fatalf("获取所有学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取所有学生失败，状态码: %d", resp.StatusCode)
	}

	var students []models.Student
	if err := json.NewDecoder(resp.Body).Decode(&students); err != nil {
		t.Fatalf("无法解析所有学生列表响应: %v", err)
	}

	if len(students) < 2 {
		t.Errorf("期望至少有2个学生, 得到 %d", len(students))
	}
	fmt.Printf("获取所有学生成功，共 %d 个学生\n", len(students))
}
