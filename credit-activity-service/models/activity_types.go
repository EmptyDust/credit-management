package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 创新创业实践活动详情表
// Table: innovation_activity_details
// 字段：事项、实习公司、课题编号、发证机构、日期、累计学时

type InnovationActivityDetail struct {
	ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	Item       string         `json:"item"`
	Company    string         `json:"company"`
	ProjectNo  string         `json:"project_no"`
	Issuer     string         `json:"issuer"`
	Date       time.Time      `json:"date" gorm:"type:date"`
	TotalHours float64        `json:"total_hours" gorm:"type:numeric(6,2)"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// UnmarshalJSON 自定义JSON反序列化方法
func (i *InnovationActivityDetail) UnmarshalJSON(data []byte) error {
	// 创建一个临时结构体来处理JSON反序列化
	type Alias InnovationActivityDetail
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理Date字段
	if aux.Date != "" && aux.Date != "0001-01-01T00:00:00Z" {
		// 尝试解析日期字符串
		dateFormats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		for _, format := range dateFormats {
			if parsedDate, err := time.Parse(format, aux.Date); err == nil {
				i.Date = parsedDate
				break
			}
		}
	}

	return nil
}

// BeforeCreate 在创建前自动生成UUID
func (i *InnovationActivityDetail) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (InnovationActivityDetail) TableName() string {
	return "innovation_activity_details"
}

// 学科竞赛学分详情表
// Table: competition_activity_details
// 字段：竞赛级别、竞赛名称、获奖等级、排名

type CompetitionActivityDetail struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID  string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	Level       string         `json:"level"`
	Competition string         `json:"competition"`
	AwardLevel  string         `json:"award_level"`
	Rank        string         `json:"rank"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (c *CompetitionActivityDetail) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (CompetitionActivityDetail) TableName() string {
	return "competition_activity_details"
}

// 大学生创业项目详情表
// Table: entrepreneurship_project_details
// 字段：项目名称、项目等级、项目排名

type EntrepreneurshipProjectDetail struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID   string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	ProjectName  string         `json:"project_name"`
	ProjectLevel string         `json:"project_level"`
	ProjectRank  string         `json:"project_rank"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (e *EntrepreneurshipProjectDetail) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (EntrepreneurshipProjectDetail) TableName() string {
	return "entrepreneurship_project_details"
}

// 创业实践项目详情表
// Table: entrepreneurship_practice_details
// 字段：公司名称、公司法人、本人占股比例

type EntrepreneurshipPracticeDetail struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID   string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	CompanyName  string         `json:"company_name"`
	LegalPerson  string         `json:"legal_person"`
	SharePercent float64        `json:"share_percent"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (e *EntrepreneurshipPracticeDetail) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (EntrepreneurshipPracticeDetail) TableName() string {
	return "entrepreneurship_practice_details"
}

// 论文专利详情表
// Table: paper_patent_details
// 字段：论文/专利/软件著作权名称、类别、排名

type PaperPatentDetail struct {
	ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	Name       string         `json:"name"`
	Category   string         `json:"category"`
	Rank       string         `json:"rank"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate 在创建前自动生成UUID
func (p *PaperPatentDetail) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (PaperPatentDetail) TableName() string {
	return "paper_patent_details"
}
