package models

import "time"

// ActivityRequest 创建活动请求
type ActivityRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Category    string `json:"category"`

	Details map[string]any `json:"details"`
}

// BatchActivityRequest 批量创建活动请求
type BatchActivityRequest struct {
	Activities []ActivityRequest `json:"activities" binding:"required,min=1,max=10"`
}

// ActivityCreateResponse 活动创建响应
type ActivityCreateResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"`
	Category    string    `json:"category"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ActivityUpdateRequest 活动更新请求
type ActivityUpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Category    *string `json:"category"`

	Details map[string]any `json:"details"`
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
	OwnerID        string                `json:"owner_id"`
	OwnerInfo      *UserInfo             `json:"owner_info,omitempty"`
	ReviewerID     *string               `json:"reviewer_id"`
	ReviewComments string                `json:"review_comments"`
	ReviewedAt     *time.Time            `json:"reviewed_at"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	Participants   []ParticipantResponse `json:"participants"`
	Applications   []ApplicationResponse `json:"applications"`
	Details        map[string]any        `json:"details"`
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
	UserInfo       *UserInfo    `json:"user_info,omitempty"`
}

// ActivityInfo 活动信息
type ActivityInfo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

// ParticipantRequest 参与者请求
type ParticipantRequest struct {
	UUIDs   []string `json:"ids" binding:"required"`
	Credits float64  `json:"credits" binding:"required,gt=0"`
}

// AddParticipantsRequest 添加参与者请求
type AddParticipantsRequest struct {
	UUIDs   []string `json:"ids" binding:"required"`
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
	UUID     string    `json:"id"`
	Credits  float64   `json:"credits"`
	JoinedAt time.Time `json:"joined_at"`
	UserInfo *UserInfo `json:"user_info,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	UUID       string `json:"id"`
	Username   string `json:"username"`
	RealName   string `json:"real_name"`
	UserType   string `json:"user_type"`
	Status     string `json:"status"`
	StudentID  string `json:"student_id,omitempty"`
	College    string `json:"college,omitempty"`
	Major      string `json:"major,omitempty"`
	Class      string `json:"class,omitempty"`
	Grade      string `json:"grade,omitempty"`
	Department string `json:"department,omitempty"`
	Title      string `json:"title,omitempty"`
}

// BatchRemoveRequest 批量移除参与者请求
type BatchRemoveRequest struct {
	UUIDs []string `json:"ids" binding:"required"`
}

// ParticipantActivityResponse 参与者活动响应
type ParticipantActivityResponse struct {
	ActivityID string       `json:"activity_id"`
	Credits    float64      `json:"credits"`
	JoinedAt   time.Time    `json:"joined_at"`
	Activity   ActivityInfo `json:"activity"`
}
