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

// ActivityParticipant 活动参与者模型
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

// Application 学分申请模型（自动生成，只读）
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

// ActivityRequest 创建活动请求
type ActivityRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Category     string `json:"category"`
	Requirements string `json:"requirements"`

	InnovationDetail               *InnovationActivityDetail       `json:"innovation_detail,omitempty"`
	CompetitionDetail              *CompetitionActivityDetail      `json:"competition_detail,omitempty"`
	EntrepreneurshipProjectDetail  *EntrepreneurshipProjectDetail  `json:"entrepreneurship_project_detail,omitempty"`
	EntrepreneurshipPracticeDetail *EntrepreneurshipPracticeDetail `json:"entrepreneurship_practice_detail,omitempty"`
	PaperPatentDetail              *PaperPatentDetail              `json:"paper_patent_detail,omitempty"`
}

// BatchActivityRequest 批量创建活动请求
type BatchActivityRequest struct {
	Activities []ActivityRequest `json:"activities" binding:"required,min=1,max=10"`
}

// ActivityCreateResponse 活动创建响应
type ActivityCreateResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Status       string    `json:"status"`
	Category     string    `json:"category"`
	Requirements string    `json:"requirements"`
	OwnerID      string    `json:"owner_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ActivityUpdateRequest 活动更新请求
type ActivityUpdateRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	Category     *string `json:"category"`
	Requirements *string `json:"requirements"`

	InnovationDetail               *InnovationActivityDetail       `json:"innovation_detail,omitempty"`
	CompetitionDetail              *CompetitionActivityDetail      `json:"competition_detail,omitempty"`
	EntrepreneurshipProjectDetail  *EntrepreneurshipProjectDetail  `json:"entrepreneurship_project_detail,omitempty"`
	EntrepreneurshipPracticeDetail *EntrepreneurshipPracticeDetail `json:"entrepreneurship_practice_detail,omitempty"`
	PaperPatentDetail              *PaperPatentDetail              `json:"paper_patent_detail,omitempty"`
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
	ReviewerID     *string               `json:"reviewer_id"`
	ReviewComments string                `json:"review_comments"`
	ReviewedAt     *time.Time            `json:"reviewed_at"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	Participants   []ParticipantResponse `json:"participants"`
	Applications   []ApplicationResponse `json:"applications"`

	InnovationDetail               *InnovationActivityDetail       `json:"innovation_detail,omitempty"`
	CompetitionDetail              *CompetitionActivityDetail      `json:"competition_detail,omitempty"`
	EntrepreneurshipProjectDetail  *EntrepreneurshipProjectDetail  `json:"entrepreneurship_project_detail,omitempty"`
	EntrepreneurshipPracticeDetail *EntrepreneurshipPracticeDetail `json:"entrepreneurship_practice_detail,omitempty"`
	PaperPatentDetail              *PaperPatentDetail              `json:"paper_patent_detail,omitempty"`
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

// ActivitySearchRequest 活动搜索请求
type ActivitySearchRequest struct {
	Query     string `json:"query" form:"query"`           // 关键词搜索
	Category  string `json:"category" form:"category"`     // 活动类别
	Status    string `json:"status" form:"status"`         // 活动状态
	OwnerID   string `json:"owner_id" form:"owner_id"`     // 创建者ID
	StartDate string `json:"start_date" form:"start_date"` // 开始日期
	EndDate   string `json:"end_date" form:"end_date"`     // 结束日期
	Page      int    `json:"page" form:"page"`             // 页码
	PageSize  int    `json:"page_size" form:"page_size"`   // 每页数量
	SortBy    string `json:"sort_by" form:"sort_by"`       // 排序字段
	SortOrder string `json:"sort_order" form:"sort_order"` // 排序方向
}

// ApplicationSearchRequest 申请搜索请求
type ApplicationSearchRequest struct {
	Query      string `json:"query" form:"query"`             // 关键词搜索
	ActivityID string `json:"activity_id" form:"activity_id"` // 活动ID
	UserID     string `json:"user_id" form:"user_id"`         // 用户ID
	Status     string `json:"status" form:"status"`           // 申请状态
	StartDate  string `json:"start_date" form:"start_date"`   // 开始日期
	EndDate    string `json:"end_date" form:"end_date"`       // 结束日期
	MinCredits string `json:"min_credits" form:"min_credits"` // 最小学分
	MaxCredits string `json:"max_credits" form:"max_credits"` // 最大学分
	Page       int    `json:"page" form:"page"`               // 页码
	PageSize   int    `json:"page_size" form:"page_size"`     // 每页数量
	SortBy     string `json:"sort_by" form:"sort_by"`         // 排序字段
	SortOrder  string `json:"sort_order" form:"sort_order"`   // 排序方向
}

// ParticipantSearchRequest 参与者搜索请求
type ParticipantSearchRequest struct {
	Query      string `json:"query" form:"query"`             // 关键词搜索
	ActivityID string `json:"activity_id" form:"activity_id"` // 活动ID
	UserID     string `json:"user_id" form:"user_id"`         // 用户ID
	MinCredits string `json:"min_credits" form:"min_credits"` // 最小学分
	MaxCredits string `json:"max_credits" form:"max_credits"` // 最大学分
	Page       int    `json:"page" form:"page"`               // 页码
	PageSize   int    `json:"page_size" form:"page_size"`     // 每页数量
	SortBy     string `json:"sort_by" form:"sort_by"`         // 排序字段
	SortOrder  string `json:"sort_order" form:"sort_order"`   // 排序方向
}

// AttachmentSearchRequest 附件搜索请求
type AttachmentSearchRequest struct {
	Query        string `json:"query" form:"query"`                 // 关键词搜索
	ActivityID   string `json:"activity_id" form:"activity_id"`     // 活动ID
	UploaderID   string `json:"uploader_id" form:"uploader_id"`     // 上传者ID
	FileType     string `json:"file_type" form:"file_type"`         // 文件类型
	FileCategory string `json:"file_category" form:"file_category"` // 文件类别
	MinSize      string `json:"min_size" form:"min_size"`           // 最小文件大小
	MaxSize      string `json:"max_size" form:"max_size"`           // 最大文件大小
	Page         int    `json:"page" form:"page"`                   // 页码
	PageSize     int    `json:"page_size" form:"page_size"`         // 每页数量
	SortBy       string `json:"sort_by" form:"sort_by"`             // 排序字段
	SortOrder    string `json:"sort_order" form:"sort_order"`       // 排序方向
}

// SearchResponse 统一搜索响应
type SearchResponse struct {
	Data       []interface{} `json:"data"`        // 搜索结果数据
	Total      int64         `json:"total"`       // 总记录数
	Page       int           `json:"page"`        // 当前页码
	PageSize   int           `json:"page_size"`   // 每页数量
	TotalPages int           `json:"total_pages"` // 总页数
	Filters    interface{}   `json:"filters"`     // 应用的筛选条件
}

// 创新创业实践活动详情表
// Table: innovation_activity_details
// 字段：申请编号、学生学号、事项、实习公司、课题编号、发证机构、日期、累计学时、认定学分、审核者ID、审核意见

type InnovationActivityDetail struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID    string    `json:"activity_id" gorm:"type:uuid;not null;index"`
	ApplicationNo string    `json:"application_no"`
	StudentNo     string    `json:"student_no"`
	Item          string    `json:"item"`
	Company       string    `json:"company"`
	ProjectNo     string    `json:"project_no"`
	Issuer        string    `json:"issuer"`
	Date          string    `json:"date"`
	TotalHours    float64   `json:"total_hours"`
	Credits       float64   `json:"credits"`
	ReviewerID    string    `json:"reviewer_id"`
	ReviewComment string    `json:"review_comment"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 学科竞赛学分详情表
// Table: competition_activity_details
// 字段：申请编号、学生学号、竞赛级别、竞赛名称、获奖等级、排名、认定学分、审核者ID、审核意见

type CompetitionActivityDetail struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID    string    `json:"activity_id" gorm:"type:uuid;not null;index"`
	ApplicationNo string    `json:"application_no"`
	StudentNo     string    `json:"student_no"`
	Level         string    `json:"level"`
	Competition   string    `json:"competition"`
	AwardLevel    string    `json:"award_level"`
	Rank          string    `json:"rank"`
	Credits       float64   `json:"credits"`
	ReviewerID    string    `json:"reviewer_id"`
	ReviewComment string    `json:"review_comment"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 大学生创业项目详情表
// Table: entrepreneurship_project_details
// 字段：申请编号、学生学号、项目名称、项目等级、项目排名、认定学分、审核者ID、审核意见

type EntrepreneurshipProjectDetail struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID    string    `json:"activity_id" gorm:"type:uuid;not null;index"`
	ApplicationNo string    `json:"application_no"`
	StudentNo     string    `json:"student_no"`
	ProjectName   string    `json:"project_name"`
	ProjectLevel  string    `json:"project_level"`
	ProjectRank   string    `json:"project_rank"`
	Credits       float64   `json:"credits"`
	ReviewerID    string    `json:"reviewer_id"`
	ReviewComment string    `json:"review_comment"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 创业实践项目详情表
// Table: entrepreneurship_practice_details
// 字段：申请编号、学生学号、公司名称、公司法人、本人占股比例、认定学分、审核者ID、审核意见

type EntrepreneurshipPracticeDetail struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID    string    `json:"activity_id" gorm:"type:uuid;not null;index"`
	ApplicationNo string    `json:"application_no"`
	StudentNo     string    `json:"student_no"`
	CompanyName   string    `json:"company_name"`
	LegalPerson   string    `json:"legal_person"`
	SharePercent  float64   `json:"share_percent"`
	Credits       float64   `json:"credits"`
	ReviewerID    string    `json:"reviewer_id"`
	ReviewComment string    `json:"review_comment"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 论文专利详情表
// Table: paper_patent_details
// 字段：申请编号、论文/专利/软件著作权名称、类别、排名、认定学分、审核者ID、审核意见

type PaperPatentDetail struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID    string    `json:"activity_id" gorm:"type:uuid;not null;index"`
	ApplicationNo string    `json:"application_no"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	Rank          string    `json:"rank"`
	Credits       float64   `json:"credits"`
	ReviewerID    string    `json:"reviewer_id"`
	ReviewComment string    `json:"review_comment"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
