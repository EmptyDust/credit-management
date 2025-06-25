package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 活动类别常量
const (
	CategoryInnovation  = "创新创业" // 创新创业实践活动
	CategoryCompetition = "学科竞赛" // 学科竞赛
	CategoryVolunteer   = "志愿服务" // 志愿服务
	CategoryAcademic    = "学术研究" // 学术研究
	CategoryCultural    = "文体活动" // 文体活动
)

// 获取所有活动类别
func GetActivityCategories() []string {
	return []string{
		CategoryInnovation,
		CategoryCompetition,
		CategoryVolunteer,
		CategoryAcademic,
		CategoryCultural,
	}
}

// 活动状态常量
const (
	StatusDraft         = "draft"          // 草稿
	StatusPendingReview = "pending_review" // 待审核
	StatusApproved      = "approved"       // 已通过
	StatusRejected      = "rejected"       // 已拒绝
)

// CreditActivity 学分活动模型
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
	ReviewerID     string         `json:"reviewer_id" gorm:"type:uuid"`
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

// ActivityParticipant 活动参与者模型
type ActivityParticipant struct {
	ActivityID string    `json:"activity_id" gorm:"primaryKey;type:uuid"`
	UserID     string    `json:"user_id" gorm:"primaryKey;type:uuid"`
	Credits    float64   `json:"credits" gorm:"not null;default:0"`
	JoinedAt   time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ActivityParticipant) TableName() string {
	return "activity_participants"
}

// Application 学分申请模型（自动生成，只读）
type Application struct {
	ID             string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID     string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	UserID         string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Status         string         `json:"status" gorm:"default:'approved';index"`
	AppliedCredits float64        `json:"applied_credits" gorm:"not null"`
	AwardedCredits float64        `json:"awarded_credits" gorm:"not null"`
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

// ActivityRequest 创建活动请求
type ActivityRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Category     string `json:"category"`
	Requirements string `json:"requirements"`
}

// ActivityUpdateRequest 活动更新请求
type ActivityUpdateRequest struct {
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Category     string    `json:"category"`
	Requirements string    `json:"requirements"`
}

// ActivityReviewRequest 审核活动请求
type ActivityReviewRequest struct {
	Status         string `json:"status" binding:"required,oneof=approved rejected"`
	ReviewComments string `json:"review_comments"`
}

// ActivityResponse 活动响应
type ActivityResponse struct {
	ID             string                `json:"id"`
	Title          string                `json:"title"`
	Description    string                `json:"description"`
	StartDate      time.Time             `json:"start_date"`
	EndDate        time.Time             `json:"end_date"`
	Status         string                `json:"status"`
	Category       string                `json:"category"`
	Requirements   string                `json:"requirements"`
	OwnerID        string                `json:"owner_id"`
	ReviewerID     string                `json:"reviewer_id"`
	ReviewComments string                `json:"review_comments"`
	ReviewedAt     *time.Time            `json:"reviewed_at"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	Participants   []ParticipantResponse `json:"participants"`
	Applications   []ApplicationResponse `json:"applications"`
}

// ApplicationResponse 申请响应
type ApplicationResponse struct {
	ID             string       `json:"id"`
	ActivityID     string       `json:"activity_id"`
	UserID         string       `json:"user_id"`
	Status         string       `json:"status"`
	AppliedCredits float64      `json:"applied_credits"`
	AwardedCredits float64      `json:"awarded_credits"`
	SubmittedAt    time.Time    `json:"submitted_at"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Activity       ActivityInfo `json:"activity"`
}

// ActivityInfo 活动信息（用于申请响应）
type ActivityInfo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

// ParticipantRequest 添加参与者请求
type ParticipantRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
	Credits float64  `json:"credits" binding:"required,gt=0"`
}

// AddParticipantsRequest 添加参与者请求
type AddParticipantsRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
	Credits float64  `json:"credits" binding:"required,min=0"`
}

// BatchCreditsRequest 批量设置学分请求
type BatchCreditsRequest struct {
	CreditsMap map[string]float64 `json:"credits_map" binding:"required"`
}

// SingleCreditsRequest 单个设置学分请求
type SingleCreditsRequest struct {
	Credits float64 `json:"credits" binding:"required,min=0"`
}

// ParticipantResponse 参与者响应
type ParticipantResponse struct {
	UserID   string    `json:"user_id"`
	Credits  float64   `json:"credits"`
	JoinedAt time.Time `json:"joined_at"`
	UserInfo *UserInfo `json:"user_info,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	StudentID string `json:"student_id"`
}

// ActivityStats 活动统计
type ActivityStats struct {
	TotalActivities   int64   `json:"total_activities"`
	DraftCount        int64   `json:"draft_count"`
	PendingCount      int64   `json:"pending_count"`
	ApprovedCount     int64   `json:"approved_count"`
	RejectedCount     int64   `json:"rejected_count"`
	TotalParticipants int64   `json:"total_participants"`
	TotalCredits      float64 `json:"total_credits"`
}

// ApplicationStats 申请统计
type ApplicationStats struct {
	TotalApplications int64   `json:"total_applications"`
	TotalCredits      float64 `json:"total_credits"`
	AwardedCredits    float64 `json:"awarded_credits"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// StandardResponse 标准响应
type StandardResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
