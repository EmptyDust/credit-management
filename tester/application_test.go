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

var applicationAPI = "http://localhost:8086/api/applications"

// 辅助函数：创建申请用于测试
func createTestApplication(t *testing.T) map[string]interface{} {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	application := map[string]interface{}{
		"affair_id":       1,
		"student_id":      "2021001",
		"applied_credits": 2.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 1,
				"content":   "https://example.com/certificate1.pdf",
				"file_name": "certificate1.pdf",
				"file_size": 1024000,
				"file_type": "application/pdf",
			},
		},
		"innovation_practice": map[string]interface{}{
			"company":     "测试公司",
			"project_id":  "SRTP2024001",
			"issuing_org": "测试机构",
			"date":        time.Now().Format("2006-01-02T15:04:05Z07:00"),
			"total_hours": 120,
		},
	}

	jsonData, _ := json.Marshal(application)
	resp, err := http.Post(applicationAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建申请失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("无法解析创建的申请响应: %v", err)
	}

	// 返回创建的申请ID
	return map[string]interface{}{
		"id": response["application_id"],
	}
}

func TestCreateApplication(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	application := map[string]interface{}{
		"affair_id":       2,
		"student_id":      "2021002",
		"applied_credits": 3.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 2,
				"content":   "https://example.com/certificate2.pdf",
				"file_name": "certificate2.pdf",
				"file_size": 2048000,
				"file_type": "application/pdf",
			},
		},
		"discipline_competition": map[string]interface{}{
			"level":            "国家级",
			"competition_name": "全国大学生数学竞赛_" + uniqueSuffix,
			"award_level":      "一等奖",
			"ranking":          1,
		},
	}

	jsonData, _ := json.Marshal(application)
	resp, err := http.Post(applicationAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建申请失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建申请成功")
}

func TestGetApplication(t *testing.T) {
	createdApplication := createTestApplication(t)
	applicationID := fmt.Sprintf("%.0f", createdApplication["id"])

	resp, err := http.Get(applicationAPI + "/" + applicationID)
	if err != nil {
		t.Fatalf("查询申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("查询申请失败，状态码: %d", resp.StatusCode)
	}

	var foundApplication map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&foundApplication); err != nil {
		t.Fatalf("无法解析查询的申请响应: %v", err)
	}

	if foundApplication["student_id"] != "2021001" {
		t.Errorf("期望学生ID 2021001, 得到 %s", foundApplication["student_id"])
	}
	fmt.Println("查询申请成功，学生ID:", foundApplication["student_id"])
}

func TestUpdateApplication(t *testing.T) {
	createdApplication := createTestApplication(t)
	applicationID := fmt.Sprintf("%.0f", createdApplication["id"])

	updateData := map[string]interface{}{
		"affair_id":       3,
		"applied_credits": 4.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 3,
				"content":   "https://example.com/updated_certificate.pdf",
				"file_name": "updated_certificate.pdf",
				"file_size": 3072000,
				"file_type": "application/pdf",
			},
		},
		"student_entrepreneurship": map[string]interface{}{
			"project_name":    "更新后的创业项目",
			"project_level":   "校级",
			"project_ranking": 2,
		},
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest(http.MethodPut, applicationAPI+"/"+applicationID, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("更新申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("更新申请失败，状态码: %d", resp.StatusCode)
	}

	fmt.Println("更新申请成功")
}

func TestReviewApplication(t *testing.T) {
	createdApplication := createTestApplication(t)
	applicationID := fmt.Sprintf("%.0f", createdApplication["id"])

	reviewData := map[string]interface{}{
		"status":           "approved",
		"approved_credits": 2.5,
		"review_comment":   "审核通过，认定2.5学分",
		"reviewer_id":      "teacher001",
	}
	jsonData, _ := json.Marshal(reviewData)

	req, _ := http.NewRequest(http.MethodPost, applicationAPI+"/"+applicationID+"/review", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("审核申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("审核申请失败，状态码: %d", resp.StatusCode)
	}

	fmt.Println("审核申请成功")
}

func TestDeleteApplication(t *testing.T) {
	createdApplication := createTestApplication(t)
	applicationID := fmt.Sprintf("%.0f", createdApplication["id"])

	req, _ := http.NewRequest(http.MethodDelete, applicationAPI+"/"+applicationID, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("删除申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("删除申请失败，状态码: %d", resp.StatusCode)
	}
	fmt.Println("删除申请成功")

	// 验证是否真的被删除
	resp, err = http.Get(applicationAPI + "/" + applicationID)
	if err != nil {
		t.Fatalf("验证删除时请求失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("期望状态码 404, 得到 %d", resp.StatusCode)
	}
	fmt.Println("验证删除成功")
}

func TestGetAllApplications(t *testing.T) {
	// 先创建几个申请
	createTestApplication(t)
	createTestApplication(t)

	resp, err := http.Get(applicationAPI)
	if err != nil {
		t.Fatalf("获取所有申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取所有申请失败，状态码: %d", resp.StatusCode)
	}

	var applications []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&applications); err != nil {
		t.Fatalf("无法解析所有申请列表响应: %v", err)
	}

	if len(applications) < 2 {
		t.Errorf("期望至少有2个申请, 得到 %d", len(applications))
	}
	fmt.Printf("获取所有申请成功，共 %d 个申请\n", len(applications))
}

func TestGetApplicationsByStudent(t *testing.T) {
	// 先创建一个申请
	createTestApplication(t)

	resp, err := http.Get(applicationAPI + "/student/2021001")
	if err != nil {
		t.Fatalf("根据学生ID获取申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("根据学生ID获取申请失败，状态码: %d", resp.StatusCode)
	}

	var applications []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&applications); err != nil {
		t.Fatalf("无法解析根据学生ID获取的申请响应: %v", err)
	}

	if len(applications) < 1 {
		t.Errorf("期望至少有1个申请, 得到 %d", len(applications))
	}
	fmt.Printf("根据学生ID获取申请成功，共 %d 个申请\n", len(applications))
}

func TestGetApplicationsByStatus(t *testing.T) {
	// 先创建一个申请
	createTestApplication(t)

	resp, err := http.Get(applicationAPI + "/status/pending")
	if err != nil {
		t.Fatalf("根据状态获取申请请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("根据状态获取申请失败，状态码: %d", resp.StatusCode)
	}

	var applications []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&applications); err != nil {
		t.Fatalf("无法解析根据状态获取的申请响应: %v", err)
	}

	if len(applications) < 1 {
		t.Errorf("期望至少有1个申请, 得到 %d", len(applications))
	}
	fmt.Printf("根据状态获取申请成功，共 %d 个申请\n", len(applications))
}

// 测试不同类型的学分申请
func TestCreateInnovationPracticeApplication(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	application := map[string]interface{}{
		"affair_id":       1,
		"student_id":      "2021003",
		"applied_credits": 2.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 1,
				"content":   "https://example.com/innovation_cert.pdf",
				"file_name": "innovation_cert.pdf",
				"file_size": 1536000,
				"file_type": "application/pdf",
			},
		},
		"innovation_practice": map[string]interface{}{
			"company":     "创新科技公司_" + uniqueSuffix,
			"project_id":  "SRTP2024002",
			"issuing_org": "创新机构",
			"date":        time.Now().Format("2006-01-02T15:04:05Z07:00"),
			"total_hours": 160,
		},
	}

	jsonData, _ := json.Marshal(application)
	resp, err := http.Post(applicationAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建创新创业实践活动申请失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建创新创业实践活动申请失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建创新创业实践活动申请成功")
}

func TestCreateDisciplineCompetitionApplication(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	application := map[string]interface{}{
		"affair_id":       2,
		"student_id":      "2021004",
		"applied_credits": 3.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 2,
				"content":   "https://example.com/competition_cert.pdf",
				"file_name": "competition_cert.pdf",
				"file_size": 2048000,
				"file_type": "application/pdf",
			},
		},
		"discipline_competition": map[string]interface{}{
			"level":            "国际级",
			"competition_name": "国际数学建模竞赛_" + uniqueSuffix,
			"award_level":      "特等奖",
			"ranking":          1,
		},
	}

	jsonData, _ := json.Marshal(application)
	resp, err := http.Post(applicationAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建学科竞赛申请失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建学科竞赛申请失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建学科竞赛申请成功")
}

func TestCreateStudentEntrepreneurshipApplication(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	application := map[string]interface{}{
		"affair_id":       3,
		"student_id":      "2021005",
		"applied_credits": 4.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 3,
				"content":   "https://example.com/entrepreneurship_cert.pdf",
				"file_name": "entrepreneurship_cert.pdf",
				"file_size": 2560000,
				"file_type": "application/pdf",
			},
		},
		"student_entrepreneurship": map[string]interface{}{
			"project_name":    "智能校园管理系统_" + uniqueSuffix,
			"project_level":   "国家级",
			"project_ranking": 1,
		},
	}

	jsonData, _ := json.Marshal(application)
	resp, err := http.Post(applicationAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建大学生创业项目申请失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建大学生创业项目申请失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建大学生创业项目申请成功")
}

func TestCreateEntrepreneurshipPracticeApplication(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	application := map[string]interface{}{
		"affair_id":       4,
		"student_id":      "2021006",
		"applied_credits": 5.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 4,
				"content":   "https://example.com/practice_cert.pdf",
				"file_name": "practice_cert.pdf",
				"file_size": 3072000,
				"file_type": "application/pdf",
			},
		},
		"entrepreneurship_practice": map[string]interface{}{
			"company_name": "创业实践公司_" + uniqueSuffix,
			"legal_person": "张三",
			"share_ratio":  30.5,
		},
	}

	jsonData, _ := json.Marshal(application)
	resp, err := http.Post(applicationAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建创业实践项目申请失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建创业实践项目申请失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建创业实践项目申请成功")
}

func TestCreatePaperPatentApplication(t *testing.T) {
	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	application := map[string]interface{}{
		"affair_id":       5,
		"student_id":      "2021007",
		"applied_credits": 6.0,
		"proof_materials": []map[string]interface{}{
			{
				"affair_id": 5,
				"content":   "https://example.com/paper_cert.pdf",
				"file_name": "paper_cert.pdf",
				"file_size": 3584000,
				"file_type": "application/pdf",
			},
		},
		"paper_patent": map[string]interface{}{
			"title":    "基于深度学习的图像识别算法研究_" + uniqueSuffix,
			"category": "学术论文-核心",
			"ranking":  1,
		},
	}

	jsonData, _ := json.Marshal(application)
	resp, err := http.Post(applicationAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建论文专利申请失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建论文专利申请失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Println("创建论文专利申请成功")
}
