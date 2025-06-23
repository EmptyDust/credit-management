package models

import (
	"time"

	"gorm.io/gorm"
)

// Application (申请)
type Application struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	AffairID        uint           `json:"affair_id" gorm:"not null"`        // 事项ID
	StudentNumber   string         `json:"student_number" gorm:"not null"`   // 学生学号
	SubmissionTime  time.Time      `json:"submission_time" gorm:"autoCreateTime"`
	Status          string         `json:"status" gorm:"default:'待审核'"`
	ReviewerID      uint           `json:"reviewer_id"`                      // 审核者ID (Teacher)
	ReviewComment   string         `json:"review_comment"`
	AppliedCredits  float64        `json:"applied_credits"`
	ApprovedCredits float64        `json:"approved_credits"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateApplicationRequest 定义了创建新申请的请求体
type CreateApplicationRequest struct {
	AffairID      uint                   `json:"affair_id" binding:"required"`
	StudentNumber string                 `json:"student_number" binding:"required"`
	Details       map[string]interface{} `json:"details" binding:"required"`
}

// ProofMaterial (证明材料)
type ProofMaterial struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ApplicationID uint   `json:"application_id"`
	Content       string `json:"content"` // 存储位置链接 (URL/Path)
}

// --- 学分申请事项细分表 ---

// InnovationPracticeCredit 创新创业实践活动学分
type InnovationPracticeCredit struct {
	ApplicationID  uint   `json:"application_id" gorm:"primaryKey"`
	Internship     string `json:"internship"`
	ProjectID      string `json:"project_id"`
	CertifyingBody string `json:"certifying_body"`
	Date           string `json:"date"`
	Hours          int    `json:"hours"`
}

// DisciplineCompetitionCredit 学科竞赛学分
type DisciplineCompetitionCredit struct {
	ApplicationID uint   `json:"application_id" gorm:"primaryKey"`
	Level         string `json:"level"`
	Name          string `json:"name"`
	Award         string `json:"award"`
	Ranking       int    `json:"ranking"`
}

// StudentEntrepreneurshipProjectCredit 大学生创业项目学分
type StudentEntrepreneurshipProjectCredit struct {
	ApplicationID uint   `json:"application_id" gorm:"primaryKey"`
	ProjectName   string `json:"project_name"`
	ProjectLevel  string `json:"project_level"`
	ProjectRank   int    `json:"project_rank"`
}

// EntrepreneurshipPracticeCredit 创业实践项目学分
type EntrepreneurshipPracticeCredit struct {
	ApplicationID uint    `json:"application_id" gorm:"primaryKey"`
	CompanyName   string  `json:"company_name"`
	CompanyRep    string  `json:"company_rep"`
	ShareRatio    float64 `json:"share_ratio"`
}

// PaperPatentCredit 论文专利学分
type PaperPatentCredit struct {
	ApplicationID uint   `json:"application_id" gorm:"primaryKey"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	Ranking       int    `json:"ranking"`
}
