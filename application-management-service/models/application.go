package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Application (申请)
type Application struct {
	ID              string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AffairID        string         `json:"affair_id" gorm:"not null;type:uuid"`        // 事项ID
	StudentNumber   string         `json:"student_number" gorm:"not null"`   // 学生学号
	SubmissionTime  time.Time      `json:"submission_time" gorm:"autoCreateTime"`
	Status          string         `json:"status" gorm:"default:'unsubmitted'"`     // unsubmitted/pending/approved/rejected
	ReviewerID      string         `json:"reviewer_id;type:uuid"`                      // 审核者ID (Teacher)
	ReviewComment   string         `json:"review_comment"`
	AppliedCredits  float64        `json:"applied_credits"`
	ApprovedCredits float64        `json:"approved_credits"`
	ReviewTime      *time.Time     `json:"review_time"`                      // 审核时间
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// CreateApplicationRequest 定义了创建新申请的请求体
type CreateApplicationRequest struct {
	AffairID      string                 `json:"affair_id" binding:"required"`
	StudentNumber string                 `json:"student_number" binding:"required"`
	Details       map[string]interface{} `json:"details" binding:"required"`
}

// BatchCreateApplicationsRequest 批量创建申请（事务创建时调用）
type BatchCreateApplicationsRequest struct {
	AffairID      string   `json:"affair_id" binding:"required"`
	CreatorID     string   `json:"creator_id" binding:"required"`
	Participants  []string `json:"participants" binding:"required"`
}

// UpdateApplicationRequest 更新申请详情
type UpdateApplicationRequest struct {
	Details       map[string]interface{} `json:"details" binding:"required"`
	AppliedCredits float64               `json:"applied_credits"`
}

// ProofMaterial (证明材料)
type ProofMaterial struct {
	ID            string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ApplicationID string `json:"application_id;type:uuid"`
	Content       string `json:"content"` // 存储位置链接 (URL/Path)
}

// BeforeCreate 在创建前自动生成UUID
func (pm *ProofMaterial) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == "" {
		pm.ID = uuid.New().String()
	}
	return nil
}

// --- 学分申请事项细分表 ---

// InnovationPracticeCredit 创新创业实践活动学分
type InnovationPracticeCredit struct {
	ApplicationID  string `json:"application_id" gorm:"primaryKey;type:uuid"`
	Internship     string `json:"internship"`     // 实习单位
	ProjectID      string `json:"project_id"`     // 项目编号
	CertifyingBody string `json:"certifying_body"` // 认证机构
	Date           string `json:"date"`           // 实践日期
	Hours          int    `json:"hours"`          // 实践时长
}

// DisciplineCompetitionCredit 学科竞赛学分
type DisciplineCompetitionCredit struct {
	ApplicationID string `json:"application_id" gorm:"primaryKey;type:uuid"`
	Level         string `json:"level"`         // 竞赛级别（国家级/省级/校级）
	Name          string `json:"name"`          // 竞赛名称
	Award         string `json:"award"`         // 获奖等级
	Ranking       int    `json:"ranking"`       // 排名
}

// StudentEntrepreneurshipProjectCredit 大学生创业项目学分
type StudentEntrepreneurshipProjectCredit struct {
	ApplicationID string `json:"application_id" gorm:"primaryKey;type:uuid"`
	ProjectName   string `json:"project_name"`   // 项目名称
	ProjectLevel  string `json:"project_level"`  // 项目级别
	ProjectRank   int    `json:"project_rank"`   // 项目排名
}

// EntrepreneurshipPracticeCredit 创业实践项目学分
type EntrepreneurshipPracticeCredit struct {
	ApplicationID string  `json:"application_id" gorm:"primaryKey;type:uuid"`
	CompanyName   string  `json:"company_name"`   // 公司名称
	CompanyRep    string  `json:"company_rep"`    // 公司代表
	ShareRatio    float64 `json:"share_ratio"`    // 持股比例
}

// PaperPatentCredit 论文专利学分
type PaperPatentCredit struct {
	ApplicationID string `json:"application_id" gorm:"primaryKey;type:uuid"`
	Name          string `json:"name"`          // 论文/专利名称
	Category      string `json:"category"`      // 类别（论文/专利）
	Ranking       int    `json:"ranking"`       // 排名/影响因子
}

// ApplicationDetail 申请详情联合查询结果
type ApplicationDetail struct {
	Application                    `json:"application"`
	InnovationPracticeCredit       *InnovationPracticeCredit       `json:"innovation_practice,omitempty"`
	DisciplineCompetitionCredit    *DisciplineCompetitionCredit    `json:"discipline_competition,omitempty"`
	StudentEntrepreneurshipProjectCredit *StudentEntrepreneurshipProjectCredit `json:"student_entrepreneurship,omitempty"`
	EntrepreneurshipPracticeCredit *EntrepreneurshipPracticeCredit `json:"entrepreneurship_practice,omitempty"`
	PaperPatentCredit              *PaperPatentCredit              `json:"paper_patent,omitempty"`
}
