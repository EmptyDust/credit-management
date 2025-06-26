package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	db *gorm.DB
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

// SearchActivities 统一活动搜索API
func (h *SearchHandler) SearchActivities(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}
	userType, _ := c.Get("user_type")

	// 构建搜索请求
	var req models.ActivitySearchRequest

	// 从查询参数获取搜索条件
	req.Query = c.Query("query")
	req.Category = c.Query("category")
	req.Status = c.Query("status")
	req.OwnerID = c.Query("owner_id")
	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	req.Page = page
	req.PageSize = pageSize

	// 排序参数
	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 构建查询
	dbQuery := h.db.Model(&models.CreditActivity{})

	// 权限过滤：学生只能看到自己创建或参与的活动
	if userType == "student" {
		dbQuery = dbQuery.Where("owner_id = ? OR id IN (SELECT activity_id FROM activity_participants WHERE user_id = ?)", userID, userID)
	}

	// 应用搜索条件
	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		dbQuery = dbQuery.Where(
			"title ILIKE ? OR description ILIKE ? OR category ILIKE ? OR requirements ILIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery,
		)
	}

	if req.Category != "" {
		dbQuery = dbQuery.Where("category = ?", req.Category)
	}

	if req.Status != "" {
		dbQuery = dbQuery.Where("status = ?", req.Status)
	}

	if req.OwnerID != "" {
		dbQuery = dbQuery.Where("owner_id = ?", req.OwnerID)
	}

	if req.StartDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			dbQuery = dbQuery.Where("start_date >= ?", parsedDate)
		}
	}

	if req.EndDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			dbQuery = dbQuery.Where("end_date <= ?", parsedDate)
		}
	}

	// 获取总数
	var total int64
	dbQuery.Count(&total)

	// 排序
	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	// 获取分页数据
	offset := (req.Page - 1) * req.PageSize
	var activities []models.CreditActivity
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var responses []models.ActivityResponse
	for _, activity := range activities {
		response := models.ActivityResponse{
			ID:             activity.ID,
			Title:          activity.Title,
			Description:    activity.Description,
			StartDate:      activity.StartDate,
			EndDate:        activity.EndDate,
			Status:         activity.Status,
			Category:       activity.Category,
			Requirements:   activity.Requirements,
			OwnerID:        activity.OwnerID,
			ReviewerID:     activity.ReviewerID,
			ReviewComments: activity.ReviewComments,
			ReviewedAt:     activity.ReviewedAt,
			CreatedAt:      activity.CreatedAt,
			UpdatedAt:      activity.UpdatedAt,
			Participants:   []models.ParticipantResponse{},
			Applications:   []models.ApplicationResponse{},
		}
		responses = append(responses, response)
	}

	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "搜索成功",
		"data": models.SearchResponse{
			Data:       h.convertToInterfaceSlice(responses),
			Total:      total,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: totalPages,
			Filters:    req,
		},
	})
}

// SearchApplications 统一申请搜索API
func (h *SearchHandler) SearchApplications(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}
	userType, _ := c.Get("user_type")

	// 获取认证令牌
	authToken := c.GetHeader("Authorization")

	// 构建搜索请求
	var req models.ApplicationSearchRequest

	// 从查询参数获取搜索条件
	req.Query = c.Query("query")
	req.ActivityID = c.Query("activity_id")
	req.UserID = c.Query("user_id")
	req.Status = c.Query("status")
	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")
	req.MinCredits = c.Query("min_credits")
	req.MaxCredits = c.Query("max_credits")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	req.Page = page
	req.PageSize = pageSize

	// 排序参数
	req.SortBy = c.DefaultQuery("sort_by", "submitted_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 构建查询
	dbQuery := h.db.Model(&models.Application{}).Preload("Activity")

	// 权限过滤：学生只能看到自己的申请
	if userType == "student" {
		dbQuery = dbQuery.Where("user_id = ?", userID)
	}

	// 应用搜索条件
	if req.Query != "" {
		// 通过活动信息搜索
		searchQuery := "%" + req.Query + "%"
		dbQuery = dbQuery.Joins("JOIN credit_activities ON applications.activity_id = credit_activities.id").
			Where("credit_activities.title ILIKE ? OR credit_activities.description ILIKE ? OR credit_activities.category ILIKE ?",
				searchQuery, searchQuery, searchQuery)
	}

	if req.ActivityID != "" {
		dbQuery = dbQuery.Where("activity_id = ?", req.ActivityID)
	}

	if req.UserID != "" {
		dbQuery = dbQuery.Where("user_id = ?", req.UserID)
	}

	if req.Status != "" {
		dbQuery = dbQuery.Where("status = ?", req.Status)
	}

	if req.StartDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			dbQuery = dbQuery.Where("submitted_at >= ?", parsedDate)
		}
	}

	if req.EndDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			dbQuery = dbQuery.Where("submitted_at <= ?", parsedDate.Add(24*time.Hour))
		}
	}

	if req.MinCredits != "" {
		if minCredits, err := strconv.ParseFloat(req.MinCredits, 64); err == nil {
			dbQuery = dbQuery.Where("awarded_credits >= ?", minCredits)
		}
	}

	if req.MaxCredits != "" {
		if maxCredits, err := strconv.ParseFloat(req.MaxCredits, 64); err == nil {
			dbQuery = dbQuery.Where("awarded_credits <= ?", maxCredits)
		}
	}

	// 获取总数
	var total int64
	dbQuery.Count(&total)

	// 排序
	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	// 获取分页数据
	offset := (req.Page - 1) * req.PageSize
	var applications []models.Application
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索申请失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var responses []models.ApplicationResponse
	for _, app := range applications {
		response := models.ApplicationResponse{
			ID:             app.ID,
			ActivityID:     app.ActivityID,
			UserID:         app.UserID,
			Status:         app.Status,
			AppliedCredits: app.AppliedCredits,
			AwardedCredits: app.AwardedCredits,
			SubmittedAt:    app.SubmittedAt,
			CreatedAt:      app.CreatedAt,
			UpdatedAt:      app.UpdatedAt,
			Activity: models.ActivityInfo{
				ID:          app.Activity.ID,
				Title:       app.Activity.Title,
				Description: app.Activity.Description,
				Category:    app.Activity.Category,
				StartDate:   app.Activity.StartDate,
				EndDate:     app.Activity.EndDate,
			},
		}

		// 获取用户信息
		if userInfo, err := utils.GetUserInfo(app.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}

		responses = append(responses, response)
	}

	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "搜索成功",
		"data": models.SearchResponse{
			Data:       h.convertToInterfaceSlice(responses),
			Total:      total,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: totalPages,
			Filters:    req,
		},
	})
}

// SearchParticipants 统一参与者搜索API
func (h *SearchHandler) SearchParticipants(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}
	userType, _ := c.Get("user_type")

	// 获取认证令牌
	authToken := c.GetHeader("Authorization")

	// 构建搜索请求
	var req models.ParticipantSearchRequest

	// 从查询参数获取搜索条件
	req.Query = c.Query("query")
	req.ActivityID = c.Query("activity_id")
	req.UserID = c.Query("user_id")
	req.MinCredits = c.Query("min_credits")
	req.MaxCredits = c.Query("max_credits")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	req.Page = page
	req.PageSize = pageSize

	// 排序参数
	req.SortBy = c.DefaultQuery("sort_by", "joined_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 构建查询
	dbQuery := h.db.Model(&models.ActivityParticipant{})

	// 权限过滤：学生只能看到自己参与的活动
	if userType == "student" {
		dbQuery = dbQuery.Where("user_id = ?", userID)
	}

	// 应用搜索条件
	if req.ActivityID != "" {
		dbQuery = dbQuery.Where("activity_id = ?", req.ActivityID)
	}

	if req.UserID != "" {
		dbQuery = dbQuery.Where("user_id = ?", req.UserID)
	}

	if req.MinCredits != "" {
		if minCredits, err := strconv.ParseFloat(req.MinCredits, 64); err == nil {
			dbQuery = dbQuery.Where("credits >= ?", minCredits)
		}
	}

	if req.MaxCredits != "" {
		if maxCredits, err := strconv.ParseFloat(req.MaxCredits, 64); err == nil {
			dbQuery = dbQuery.Where("credits <= ?", maxCredits)
		}
	}

	// 获取总数
	var total int64
	dbQuery.Count(&total)

	// 排序
	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	// 获取分页数据
	offset := (req.Page - 1) * req.PageSize
	var participants []models.ActivityParticipant
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索参与者失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var responses []models.ParticipantResponse
	for _, participant := range participants {
		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
		}

		// 获取用户信息
		if userInfo, err := utils.GetUserInfo(participant.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}

		responses = append(responses, response)
	}

	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "搜索成功",
		"data": models.SearchResponse{
			Data:       h.convertToInterfaceSlice(responses),
			Total:      total,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: totalPages,
			Filters:    req,
		},
	})
}

// SearchAttachments 统一附件搜索API
func (h *SearchHandler) SearchAttachments(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}
	userType, _ := c.Get("user_type")

	// 构建搜索请求
	var req models.AttachmentSearchRequest

	// 从查询参数获取搜索条件
	req.Query = c.Query("query")
	req.ActivityID = c.Query("activity_id")
	req.UploaderID = c.Query("uploader_id")
	req.FileType = c.Query("file_type")
	req.FileCategory = c.Query("file_category")
	req.MinSize = c.Query("min_size")
	req.MaxSize = c.Query("max_size")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	req.Page = page
	req.PageSize = pageSize

	// 排序参数
	req.SortBy = c.DefaultQuery("sort_by", "uploaded_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 构建查询
	dbQuery := h.db.Model(&models.Attachment{})

	// 权限过滤：学生只能看到自己创建或参与活动的附件
	if userType == "student" {
		dbQuery = dbQuery.Where("uploaded_by = ? OR activity_id IN (SELECT activity_id FROM activity_participants WHERE user_id = ?)", userID, userID)
	}

	// 应用搜索条件
	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		dbQuery = dbQuery.Where("file_name ILIKE ? OR original_name ILIKE ? OR description ILIKE ?",
			searchQuery, searchQuery, searchQuery)
	}

	if req.ActivityID != "" {
		dbQuery = dbQuery.Where("activity_id = ?", req.ActivityID)
	}

	if req.UploaderID != "" {
		dbQuery = dbQuery.Where("uploaded_by = ?", req.UploaderID)
	}

	if req.FileType != "" {
		dbQuery = dbQuery.Where("file_type = ?", req.FileType)
	}

	if req.FileCategory != "" {
		dbQuery = dbQuery.Where("file_category = ?", req.FileCategory)
	}

	if req.MinSize != "" {
		if minSize, err := strconv.ParseInt(req.MinSize, 10, 64); err == nil {
			dbQuery = dbQuery.Where("file_size >= ?", minSize)
		}
	}

	if req.MaxSize != "" {
		if maxSize, err := strconv.ParseInt(req.MaxSize, 10, 64); err == nil {
			dbQuery = dbQuery.Where("file_size <= ?", maxSize)
		}
	}

	// 获取总数
	var total int64
	dbQuery.Count(&total)

	// 排序
	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	// 获取分页数据
	offset := (req.Page - 1) * req.PageSize
	var attachments []models.Attachment
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&attachments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索附件失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var responses []models.AttachmentResponse

	// 获取认证令牌用于调用用户服务
	authToken := c.GetHeader("Authorization")

	for _, attachment := range attachments {
		response := models.AttachmentResponse{
			ID:            attachment.ID,
			ActivityID:    attachment.ActivityID,
			FileName:      attachment.FileName,
			OriginalName:  attachment.OriginalName,
			FileSize:      attachment.FileSize,
			FileType:      attachment.FileType,
			FileCategory:  attachment.FileCategory,
			Description:   attachment.Description,
			UploadedBy:    attachment.UploadedBy,
			UploadedAt:    attachment.UploadedAt,
			DownloadCount: attachment.DownloadCount,
		}

		// 获取上传者信息
		if userInfo, err := utils.GetUserInfo(attachment.UploadedBy, authToken); err == nil {
			response.Uploader = *userInfo
		}

		responses = append(responses, response)
	}

	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "搜索成功",
		"data": models.SearchResponse{
			Data:       h.convertToInterfaceSlice(responses),
			Total:      total,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: totalPages,
			Filters:    req,
		},
	})
}

// convertToInterfaceSlice 将任意类型的切片转换为interface{}切片
func (h *SearchHandler) convertToInterfaceSlice(slice interface{}) []interface{} {
	switch v := slice.(type) {
	case []models.ActivityResponse:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []models.ApplicationResponse:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []models.ParticipantResponse:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []models.AttachmentResponse:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	default:
		return []interface{}{}
	}
}
