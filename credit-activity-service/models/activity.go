package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 活动状态常量
const (
	StatusDraft         = "draft"
	StatusPendingReview = "pending_review"
	StatusApproved      = "approved"
	StatusRejected      = "rejected"
)

// 活动类别常量
const (
	CategoryInnovation       = "创新创业实践活动"
	CategoryCompetition      = "学科竞赛"
	CategoryEntrepreneurship = "大学生创业项目"
	CategoryPractice         = "创业实践项目"
	CategoryPaperPatent      = "论文专利"
)

// GetActivityCategories 获取活动类别列表
func GetActivityCategories() []string {
	return []string{
		"创新创业实践活动",
		"学科竞赛",
		"大学生创业项目",
		"创业实践项目",
		"论文专利",
	}
}

// CreditActivity 学分活动表
type CreditActivity struct {
	ID             string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title          string         `json:"title" gorm:"not null"`
	Description    string         `json:"description"`
	StartDate      time.Time      `json:"start_date"`
	EndDate        time.Time      `json:"end_date"`
	Status         string         `json:"status" gorm:"default:'draft';index"`
	Category       string         `json:"category"`
	Requirements   string         `json:"requirements"`
	OwnerID        string         `json:"owner_id" gorm:"type:uuid;not null;index"`
	ReviewerID     *string        `json:"reviewer_id" gorm:"type:uuid"`
	ReviewComments string         `json:"review_comments"`
	ReviewedAt     *time.Time     `json:"reviewed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联关系
	Participants []ActivityParticipant `json:"participants" gorm:"foreignKey:ActivityID"`
	Applications []Application         `json:"applications" gorm:"foreignKey:ActivityID"`
}

// BeforeCreate 在创建前自动生成UUID
func (ca *CreditActivity) BeforeCreate(tx *gorm.DB) error {
	if ca.ID == "" {
		ca.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (CreditActivity) TableName() string {
	return "credit_activities"
}

// ActivityParticipant 活动参与者表
type ActivityParticipant struct {
	ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	UserID     string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Credits    float64        `json:"credits" gorm:"type:decimal(5,2);not null;default:0"`
	JoinedAt   time.Time      `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联关系
	Activity CreditActivity `json:"activity" gorm:"foreignKey:ActivityID"`
	User     UserInfo       `json:"user" gorm:"foreignKey:UserID"`
}

// BeforeCreate 在创建前自动生成UUID
func (ap *ActivityParticipant) BeforeCreate(tx *gorm.DB) error {
	if ap.ID == "" {
		ap.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (ActivityParticipant) TableName() string {
	return "activity_participants"
}

// Application 申请表
type Application struct {
	ID             string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID     string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	UserID         string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Status         string         `json:"status" gorm:"default:'approved';index"`
	AppliedCredits float64        `json:"applied_credits" gorm:"type:decimal(5,2);not null"`
	AwardedCredits float64        `json:"awarded_credits" gorm:"type:decimal(5,2);not null"`
	SubmittedAt    time.Time      `json:"submitted_at" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联关系
	Activity CreditActivity `json:"activity" gorm:"foreignKey:ActivityID"`
}

// BeforeCreate 在创建前自动生成UUID
func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Application) TableName() string {
	return "applications"
}
