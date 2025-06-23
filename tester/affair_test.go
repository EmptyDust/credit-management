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

var affairAPI = "http://localhost:8087/api/affairs"
var affairStudentAPI = "http://localhost:8087/api/affair-students"

// 辅助函数：创建事项用于测试
func createTestAffair(t *testing.T) map[string]interface{} {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	affair := map[string]interface{}{
		"name":        "测试事项_" + uniqueSuffix,
		"description": "这是一个测试事项",
		"category":    "创新创业",
		"max_credits": 5.0,
	}

	jsonData, _ := json.Marshal(affair)
	resp, err := http.Post(affairAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建事项失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("无法解析创建的事项响应: %v", err)
	}

	return map[string]interface{}{
		"id": response["affair_id"],
	}
}

func TestCreateAffair(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	affair := map[string]interface{}{
		"name":        "SRTP项目_" + uniqueSuffix,
		"description": "大学生科研训练计划项目",
		"category":    "科研训练",
		"max_credits": 3.0,
	}

	jsonData, _ := json.Marshal(affair)
	resp, err := http.Post(affairAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建事项失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建事项成功")
}

func TestGetAffair(t *testing.T) {
	createdAffair := createTestAffair(t)
	affairID := fmt.Sprintf("%.0f", createdAffair["id"])

	resp, err := http.Get(affairAPI + "/" + affairID)
	if err != nil {
		t.Fatalf("查询事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("查询事项失败，状态码: %d", resp.StatusCode)
	}

	var foundAffair map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&foundAffair); err != nil {
		t.Fatalf("无法解析查询的事项响应: %v", err)
	}

	if foundAffair["name"] == "" {
		t.Errorf("事项名称不能为空")
	}
	fmt.Println("查询事项成功:", foundAffair["name"])
}

func TestUpdateAffair(t *testing.T) {
	createdAffair := createTestAffair(t)
	affairID := fmt.Sprintf("%.0f", createdAffair["id"])

	updateData := map[string]interface{}{
		"description": "更新后的描述",
		"max_credits": 6.0,
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest(http.MethodPut, affairAPI+"/"+affairID, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("更新事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("更新事项失败，状态码: %d", resp.StatusCode)
	}

	fmt.Println("更新事项成功")
}

func TestDeleteAffair(t *testing.T) {
	createdAffair := createTestAffair(t)
	affairID := fmt.Sprintf("%.0f", createdAffair["id"])

	req, _ := http.NewRequest(http.MethodDelete, affairAPI+"/"+affairID, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("删除事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("删除事项失败，状态码: %d", resp.StatusCode)
	}
	fmt.Println("删除事项成功")

	// 验证是否真的被删除
	resp, err = http.Get(affairAPI + "/" + affairID)
	if err != nil {
		t.Fatalf("验证删除时请求失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("期望状态码 404, 得到 %d", resp.StatusCode)
	}
	fmt.Println("验证删除成功")
}

func TestGetAllAffairs(t *testing.T) {
	// 先创建几个事项
	createTestAffair(t)
	createTestAffair(t)

	resp, err := http.Get(affairAPI)
	if err != nil {
		t.Fatalf("获取所有事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取所有事项失败，状态码: %d", resp.StatusCode)
	}

	var affairs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&affairs); err != nil {
		t.Fatalf("无法解析所有事项列表响应: %v", err)
	}

	if len(affairs) < 2 {
		t.Errorf("期望至少有2个事项, 得到 %d", len(affairs))
	}
	fmt.Printf("获取所有事项成功，共 %d 个事项\n", len(affairs))
}

func TestGetAffairsByCategory(t *testing.T) {
	// 先创建一个事项
	createTestAffair(t)

	resp, err := http.Get(affairAPI + "/category/创新创业")
	if err != nil {
		t.Fatalf("根据类别获取事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("根据类别获取事项失败，状态码: %d", resp.StatusCode)
	}

	var affairs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&affairs); err != nil {
		t.Fatalf("无法解析根据类别获取的事项响应: %v", err)
	}

	if len(affairs) < 1 {
		t.Errorf("期望至少有1个事项, 得到 %d", len(affairs))
	}
	fmt.Printf("根据类别获取事项成功，共 %d 个事项\n", len(affairs))
}

func TestGetActiveAffairs(t *testing.T) {
	// 先创建一个事项
	createTestAffair(t)

	resp, err := http.Get(affairAPI + "/active")
	if err != nil {
		t.Fatalf("获取活跃事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取活跃事项失败，状态码: %d", resp.StatusCode)
	}

	var affairs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&affairs); err != nil {
		t.Fatalf("无法解析活跃事项响应: %v", err)
	}

	if len(affairs) < 1 {
		t.Errorf("期望至少有1个活跃事项, 得到 %d", len(affairs))
	}
	fmt.Printf("获取活跃事项成功，共 %d 个事项\n", len(affairs))
}

// 事项-学生关系测试
func TestAddStudentToAffair(t *testing.T) {
	createdAffair := createTestAffair(t)
	affairID := fmt.Sprintf("%.0f", createdAffair["id"])

	affairStudent := map[string]interface{}{
		"affair_id":           affairID,
		"student_id":          "2021001",
		"is_main_responsible": true,
	}

	jsonData, _ := json.Marshal(affairStudent)
	resp, err := http.Post(affairStudentAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("添加学生到事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("添加学生到事项失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("添加学生到事项成功")
}

func TestGetStudentsByAffair(t *testing.T) {
	createdAffair := createTestAffair(t)
	affairID := fmt.Sprintf("%.0f", createdAffair["id"])

	// 先添加一个学生
	affairStudent := map[string]interface{}{
		"affair_id":           affairID,
		"student_id":          "2021002",
		"is_main_responsible": false,
	}
	jsonData, _ := json.Marshal(affairStudent)
	http.Post(affairStudentAPI, "application/json", bytes.NewBuffer(jsonData))

	resp, err := http.Get(affairStudentAPI + "/affair/" + affairID)
	if err != nil {
		t.Fatalf("获取事项下的学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取事项下的学生失败，状态码: %d", resp.StatusCode)
	}

	var affairStudents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&affairStudents); err != nil {
		t.Fatalf("无法解析事项下的学生响应: %v", err)
	}

	if len(affairStudents) < 1 {
		t.Errorf("期望至少有1个学生, 得到 %d", len(affairStudents))
	}
	fmt.Printf("获取事项下的学生成功，共 %d 个学生\n", len(affairStudents))
}

func TestGetAffairsByStudent(t *testing.T) {
	createdAffair := createTestAffair(t)
	affairID := fmt.Sprintf("%.0f", createdAffair["id"])

	// 先添加一个学生
	affairStudent := map[string]interface{}{
		"affair_id":           affairID,
		"student_id":          "2021003",
		"is_main_responsible": true,
	}
	jsonData, _ := json.Marshal(affairStudent)
	http.Post(affairStudentAPI, "application/json", bytes.NewBuffer(jsonData))

	resp, err := http.Get(affairStudentAPI + "/student/2021003")
	if err != nil {
		t.Fatalf("获取学生参与的事项请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取学生参与的事项失败，状态码: %d", resp.StatusCode)
	}

	var affairStudents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&affairStudents); err != nil {
		t.Fatalf("无法解析学生参与的事项响应: %v", err)
	}

	if len(affairStudents) < 1 {
		t.Errorf("期望至少有1个事项, 得到 %d", len(affairStudents))
	}
	fmt.Printf("获取学生参与的事项成功，共 %d 个事项\n", len(affairStudents))
}

func TestRemoveStudentFromAffair(t *testing.T) {
	createdAffair := createTestAffair(t)
	affairID := fmt.Sprintf("%.0f", createdAffair["id"])

	// 先添加一个学生
	affairStudent := map[string]interface{}{
		"affair_id":           affairID,
		"student_id":          "2021004",
		"is_main_responsible": false,
	}
	jsonData, _ := json.Marshal(affairStudent)
	http.Post(affairStudentAPI, "application/json", bytes.NewBuffer(jsonData))

	// 移除学生
	req, _ := http.NewRequest(http.MethodDelete, affairStudentAPI+"/"+affairID+"/2021004", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("移除学生请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("移除学生失败，状态码: %d", resp.StatusCode)
	}
	fmt.Println("移除学生成功")

	// 验证是否真的被移除
	resp, err = http.Get(affairStudentAPI + "/affair/" + affairID)
	if err != nil {
		t.Fatalf("验证移除时请求失败: %v", err)
	}
	defer resp.Body.Close()

	var affairStudents []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&affairStudents)

	// 检查是否还有该学生
	for _, student := range affairStudents {
		if student["student_id"] == "2021004" {
			t.Errorf("学生应该已被移除")
			break
		}
	}
	fmt.Println("验证移除成功")
}
