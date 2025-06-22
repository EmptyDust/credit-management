package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// 这些结构体需要与 credit-service 中的定义保持一致
type AppCompetition struct {
	CompetitionLevel string `json:"competition_level"`
	CompetitionName  string `json:"competition_name"`
	AwardLevel       string `json:"award_level"`
	Ranking          int    `json:"ranking"`
}

type Proof struct {
	FileUrl     string `json:"file_url"`
	Description string `json:"description"`
}

type CreateApplicationRequest struct {
	ApplicantUserID    uint            `json:"applicant_user_id"`
	ApplicationType    string          `json:"application_type"`
	CompetitionDetails *AppCompetition `json:"competition_details,omitempty"`
	Proofs             []Proof         `json:"proofs"`
}

func main() {
	fmt.Println("正在准备测试请求...")

	// 1. 准备测试数据 (一个学科竞赛申请)
	// 假设系统中存在 ID 为 1 的用户
	reqData := CreateApplicationRequest{
		ApplicantUserID: 1,
		ApplicationType: "competition",
		CompetitionDetails: &AppCompetition{
			CompetitionLevel: "国家级",
			CompetitionName:  "“挑战杯”全国大学生课外学术科技作品竞赛",
			AwardLevel:       "一等奖",
			Ranking:          1,
		},
		Proofs: []Proof{
			{
				FileUrl:     "http://example.com/proof1.pdf",
				Description: "获奖证书扫描件",
			},
		},
	}

	// 2. 将数据序列化为 JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		fmt.Printf("JSON 序列化失败: %v\n", err)
		return
	}

	// 3. 发送 POST 请求
	url := "http://localhost:8081/applications"
	fmt.Printf("向 %s 发送 POST 请求...\n", url)
	fmt.Printf("请求体: %s\n", string(jsonData))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 4. 读取并打印响应
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应体失败: %v\n", err)
		return
	}
	fmt.Printf("响应体: %s\n", string(body))

	// 等待用户输入以退出
	fmt.Println("\n测试完成，按 Enter 键退出。")
	fmt.Scanln()
} 