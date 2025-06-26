package models

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
