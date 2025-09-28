package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseDetail struct {
	ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// 创新创业实践活动详情表
// Table: innovation_activity_details
// 字段：事项、实习公司、课题编号、发证机构、日期、累计学时

type InnovationActivityDetail struct {
	BaseDetail
	Item       string    `json:"item"`
	Company    string    `json:"company"`
	ProjectNo  string    `json:"project_no"`
	Issuer     string    `json:"issuer"`
	Date       time.Time `json:"date" gorm:"type:date"`
	TotalHours float64   `json:"total_hours" gorm:"type:numeric(6,2)"`
}

func (i *InnovationActivityDetail) UnmarshalJSON(data []byte) error {
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

	if aux.Date != "" && aux.Date != "0001-01-01T00:00:00Z" {
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

func (i *InnovationActivityDetail) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

func (InnovationActivityDetail) TableName() string {
	return "innovation_activity_details"
}

// 学科竞赛学分详情表
// Table: competition_activity_details
// 字段：竞赛级别、竞赛名称、获奖等级、排名

type CompetitionActivityDetail struct {
	BaseDetail
	Level       string `json:"level"`
	Competition string `json:"competition"`
	AwardLevel  string `json:"award_level"`
	Rank        string `json:"rank"`
}

func (c *CompetitionActivityDetail) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

func (CompetitionActivityDetail) TableName() string {
	return "competition_activity_details"
}

// 大学生创业项目详情表
// Table: entrepreneurship_project_details
// 字段：项目名称、项目等级、项目排名

type EntrepreneurshipProjectDetail struct {
	BaseDetail
	ProjectName  string `json:"project_name"`
	ProjectLevel string `json:"project_level"`
	ProjectRank  string `json:"project_rank"`
}

func (e *EntrepreneurshipProjectDetail) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

func (EntrepreneurshipProjectDetail) TableName() string {
	return "entrepreneurship_project_details"
}

// 创业实践项目详情表
// Table: entrepreneurship_practice_details
// 字段：公司名称、公司法人、本人占股比例

type EntrepreneurshipPracticeDetail struct {
	BaseDetail
	CompanyName  string  `json:"company_name"`
	LegalPerson  string  `json:"legal_person"`
	SharePercent float64 `json:"share_percent"`
}

func (e *EntrepreneurshipPracticeDetail) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

func (EntrepreneurshipPracticeDetail) TableName() string {
	return "entrepreneurship_practice_details"
}

// 论文专利详情表
// Table: paper_patent_details
// 字段：论文/专利/软件著作权名称、类别、排名

type PaperPatentDetail struct {
	BaseDetail
	Title    string `json:"title"`
	Category string `json:"category"`
	Rank     string `json:"rank"`
}

func (p *PaperPatentDetail) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (PaperPatentDetail) TableName() string {
	return "paper_patent_details"
}
