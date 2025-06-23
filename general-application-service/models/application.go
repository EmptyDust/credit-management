package models

import (
	"time"
)

// Application 申请模型
type Application struct {
	ID              int       `json:"id" gorm:"primaryKey;column:application_id"`
	AffairID        int       `json:"affair_id" gorm:"column:affair_id;not null"`
	StudentID       string    `json:"student_id" gorm:"column:student_id;not null"`
	SubmitTime      time.Time `json:"submit_time" gorm:"column:submit_time;not null"`
	Status          string    `json:"status" gorm:"column:status;not null;default:'pending'"` // pending, approved, rejected, cancelled
	ReviewerID      string    `json:"reviewer_id" gorm:"column:reviewer_id"`
	ReviewComment   string    `json:"review_comment" gorm:"column:review_comment"`
	AppliedCredits  float64   `json:"applied_credits" gorm:"column:applied_credits;not null"`
	ApprovedCredits float64   `json:"approved_credits" gorm:"column:approved_credits"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Affair   Affair  `json:"affair" gorm:"foreignKey:AffairID"`
	Student  Student `json:"student" gorm:"foreignKey:StudentID;references:StudentID"`
	Reviewer Teacher `json:"reviewer" gorm:"foreignKey:ReviewerID;references:Username"`

	// 证明材料
	ProofMaterials []ProofMaterial `json:"proof_materials" gorm:"foreignKey:ApplicationID"`

	// 学分申请细分表
	InnovationPracticeCredit       *InnovationPracticeCredit       `json:"innovation_practice_credit,omitempty"`
	DisciplineCompetitionCredit    *DisciplineCompetitionCredit    `json:"discipline_competition_credit,omitempty"`
	StudentEntrepreneurshipCredit  *StudentEntrepreneurshipCredit  `json:"student_entrepreneurship_credit,omitempty"`
	EntrepreneurshipPracticeCredit *EntrepreneurshipPracticeCredit `json:"entrepreneurship_practice_credit,omitempty"`
	PaperPatentCredit              *PaperPatentCredit              `json:"paper_patent_credit,omitempty"`
}

// TableName 指定表名
func (Application) TableName() string {
	return "applications"
}

// ProofMaterial 证明材料模型
type ProofMaterial struct {
	ID            int       `json:"id" gorm:"primaryKey;column:material_id"`
	ApplicationID int       `json:"application_id" gorm:"column:application_id;not null"`
	AffairID      int       `json:"affair_id" gorm:"column:affair_id;not null"`
	Content       string    `json:"content" gorm:"column:content;not null"` // 文件路径或URL
	FileName      string    `json:"file_name" gorm:"column:file_name"`
	FileSize      int64     `json:"file_size" gorm:"column:file_size"`
	FileType      string    `json:"file_type" gorm:"column:file_type"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联关系
	Application Application `json:"application" gorm:"foreignKey:ApplicationID"`
	Affair      Affair      `json:"affair" gorm:"foreignKey:AffairID"`
}

// TableName 指定表名
func (ProofMaterial) TableName() string {
	return "proof_materials"
}

// InnovationPracticeCredit 创新创业实践活动学分
type InnovationPracticeCredit struct {
	ApplicationID int       `json:"application_id" gorm:"primaryKey;column:application_id"`
	Company       string    `json:"company" gorm:"column:company"`         // 实习公司
	ProjectID     string    `json:"project_id" gorm:"column:project_id"`   // 课题编号
	IssuingOrg    string    `json:"issuing_org" gorm:"column:issuing_org"` // 发证机构
	Date          time.Time `json:"date" gorm:"column:date"`               // 活动/证书获得日期
	TotalHours    int       `json:"total_hours" gorm:"column:total_hours"` // 累计学时
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Application Application `json:"application" gorm:"foreignKey:ApplicationID"`
}

// TableName 指定表名
func (InnovationPracticeCredit) TableName() string {
	return "innovation_practice_credits"
}

// DisciplineCompetitionCredit 学科竞赛学分
type DisciplineCompetitionCredit struct {
	ApplicationID   int       `json:"application_id" gorm:"primaryKey;column:application_id"`
	Level           string    `json:"level" gorm:"column:level"`                       // 竞赛级别
	CompetitionName string    `json:"competition_name" gorm:"column:competition_name"` // 竞赛名称
	AwardLevel      string    `json:"award_level" gorm:"column:award_level"`           // 获奖等级
	Ranking         int       `json:"ranking" gorm:"column:ranking"`                   // 排名
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Application Application `json:"application" gorm:"foreignKey:ApplicationID"`
}

// TableName 指定表名
func (DisciplineCompetitionCredit) TableName() string {
	return "discipline_competition_credits"
}

// StudentEntrepreneurshipCredit 大学生创业项目学分
type StudentEntrepreneurshipCredit struct {
	ApplicationID  int       `json:"application_id" gorm:"primaryKey;column:application_id"`
	ProjectName    string    `json:"project_name" gorm:"column:project_name"`       // 项目名称
	ProjectLevel   string    `json:"project_level" gorm:"column:project_level"`     // 项目等级
	ProjectRanking int       `json:"project_ranking" gorm:"column:project_ranking"` // 项目排名
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Application Application `json:"application" gorm:"foreignKey:ApplicationID"`
}

// TableName 指定表名
func (StudentEntrepreneurshipCredit) TableName() string {
	return "student_entrepreneurship_credits"
}

// EntrepreneurshipPracticeCredit 创业实践项目学分
type EntrepreneurshipPracticeCredit struct {
	ApplicationID int       `json:"application_id" gorm:"primaryKey;column:application_id"`
	CompanyName   string    `json:"company_name" gorm:"column:company_name"` // 公司名称
	LegalPerson   string    `json:"legal_person" gorm:"column:legal_person"` // 公司法人
	ShareRatio    float64   `json:"share_ratio" gorm:"column:share_ratio"`   // 占股比例
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Application Application `json:"application" gorm:"foreignKey:ApplicationID"`
}

// TableName 指定表名
func (EntrepreneurshipPracticeCredit) TableName() string {
	return "entrepreneurship_practice_credits"
}

// PaperPatentCredit 论文专利学分
type PaperPatentCredit struct {
	ApplicationID int       `json:"application_id" gorm:"primaryKey;column:application_id"`
	Title         string    `json:"title" gorm:"column:title"`       // 论文/专利/软件著作权名称
	Category      string    `json:"category" gorm:"column:category"` // 类别
	Ranking       int       `json:"ranking" gorm:"column:ranking"`   // 排名
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Application Application `json:"application" gorm:"foreignKey:ApplicationID"`
}

// TableName 指定表名
func (PaperPatentCredit) TableName() string {
	return "paper_patent_credits"
}

// 关联模型定义
type Affair struct {
	ID   int    `json:"id" gorm:"primaryKey;column:affair_id"`
	Name string `json:"name" gorm:"column:affair_name;not null"`
}

func (Affair) TableName() string {
	return "affairs"
}

type Student struct {
	Username  string `json:"username" gorm:"primaryKey;column:username"`
	StudentID string `json:"student_id" gorm:"column:student_id;unique;not null"`
	Name      string `json:"name" gorm:"column:name;not null"`
	College   string `json:"college" gorm:"column:college"`
	Major     string `json:"major" gorm:"column:major"`
	Class     string `json:"class" gorm:"column:class"`
	Contact   string `json:"contact" gorm:"column:contact"`
}

func (Student) TableName() string {
	return "students"
}

type Teacher struct {
	Username string `json:"username" gorm:"primaryKey;column:username"`
	Name     string `json:"name" gorm:"column:name;not null"`
	Contact  string `json:"contact" gorm:"column:contact"`
}

func (Teacher) TableName() string {
	return "teachers"
}

// ApplicationRequest 申请请求结构
type ApplicationRequest struct {
	AffairID       int     `json:"affair_id" binding:"required"`
	StudentID      string  `json:"student_id" binding:"required"`
	AppliedCredits float64 `json:"applied_credits" binding:"required"`

	// 证明材料
	ProofMaterials []ProofMaterialRequest `json:"proof_materials"`

	// 学分申请细分信息（根据事项类型选择对应的结构）
	InnovationPractice       *InnovationPracticeRequest       `json:"innovation_practice,omitempty"`
	DisciplineCompetition    *DisciplineCompetitionRequest    `json:"discipline_competition,omitempty"`
	StudentEntrepreneurship  *StudentEntrepreneurshipRequest  `json:"student_entrepreneurship,omitempty"`
	EntrepreneurshipPractice *EntrepreneurshipPracticeRequest `json:"entrepreneurship_practice,omitempty"`
	PaperPatent              *PaperPatentRequest              `json:"paper_patent,omitempty"`
}

type ProofMaterialRequest struct {
	AffairID int    `json:"affair_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
}

type InnovationPracticeRequest struct {
	Company    string    `json:"company"`
	ProjectID  string    `json:"project_id"`
	IssuingOrg string    `json:"issuing_org"`
	Date       time.Time `json:"date"`
	TotalHours int       `json:"total_hours"`
}

type DisciplineCompetitionRequest struct {
	Level           string `json:"level"`
	CompetitionName string `json:"competition_name"`
	AwardLevel      string `json:"award_level"`
	Ranking         int    `json:"ranking"`
}

type StudentEntrepreneurshipRequest struct {
	ProjectName    string `json:"project_name"`
	ProjectLevel   string `json:"project_level"`
	ProjectRanking int    `json:"project_ranking"`
}

type EntrepreneurshipPracticeRequest struct {
	CompanyName string  `json:"company_name"`
	LegalPerson string  `json:"legal_person"`
	ShareRatio  float64 `json:"share_ratio"`
}

type PaperPatentRequest struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Ranking  int    `json:"ranking"`
}

// ApplicationReviewRequest 申请审核请求
type ApplicationReviewRequest struct {
	Status          string  `json:"status" binding:"required"`
	ApprovedCredits float64 `json:"approved_credits"`
	ReviewComment   string  `json:"review_comment"`
	ReviewerID      string  `json:"reviewer_id" binding:"required"`
}
