package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SearchHandler struct {
	db        *gorm.DB
	validator *utils.Validator
}

func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{
		db:        db,
		validator: utils.NewValidator(),
	}
}

func (h *SearchHandler) SearchActivities(c *gin.Context) {
	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}
	userType := c.GetString("user_type")

	var req models.ActivitySearchRequest

	req.Query = c.Query("query")
	req.Category = c.Query("category")
	req.Status = c.Query("status")
	req.OwnerID = c.Query("owner_id")
	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")

	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", "10"),
	)
	req.Page = page
	req.PageSize = limit

	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 构建基础查询
	dbQuery := h.db.Model(&models.CreditActivity{})

	// 权限过滤：学生只能看到自己创建或参与的活动
	if userType == "student" {
		dbQuery = h.applyStudentPermissionFilter(dbQuery, userID)
	}

	// 应用搜索条件
	dbQuery = h.applySearchConditions(dbQuery, req)

	// 获取总数
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 应用排序和分页
	var activities []models.CreditActivity
	query := dbQuery.
		Order(fmt.Sprintf("%s %s", req.SortBy, strings.ToUpper(req.SortOrder))).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize)

	if err := query.Find(&activities).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 转换为响应格式
	responses := h.buildActivityResponses(activities)

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

// applyStudentPermissionFilter 应用学生权限过滤
func (h *SearchHandler) applyStudentPermissionFilter(query *gorm.DB, userID string) *gorm.DB {
	return query.Where(
		"owner_id = ? OR id IN (SELECT activity_id FROM activity_participants WHERE user_id = ?)",
		userID, userID,
	)
}

// applySearchConditions 应用搜索条件
func (h *SearchHandler) applySearchConditions(query *gorm.DB, req models.ActivitySearchRequest) *gorm.DB {
	// 文本搜索
	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		query = query.Where(
			"title ILIKE ? OR description ILIKE ? OR category ILIKE ?",
			searchQuery, searchQuery, searchQuery,
		)
	}

	// 分类过滤
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	// 状态过滤
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 所有者过滤
	if req.OwnerID != "" {
		query = query.Where("owner_id = ?", req.OwnerID)
	}

	// 开始日期过滤
	if req.StartDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			query = query.Where("start_date >= ?", parsedDate)
		}
	}

	// 结束日期过滤
	if req.EndDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			query = query.Where("end_date <= ?", parsedDate)
		}
	}

	return query
}

// buildActivityResponses 构建活动响应列表
func (h *SearchHandler) buildActivityResponses(activities []models.CreditActivity) []models.ActivityResponse {
	responses := make([]models.ActivityResponse, 0, len(activities))

	for _, activity := range activities {
		response := models.ActivityResponse{
			ID:             activity.ID,
			Title:          activity.Title,
			Description:    activity.Description,
			StartDate:      activity.StartDate,
			EndDate:        activity.EndDate,
			Status:         activity.Status,
			Category:       activity.Category,
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

	return responses
}

func (h *SearchHandler) SearchApplications(c *gin.Context) {
	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}
	userType := c.GetString("user_type")
	authToken := c.GetHeader("Authorization")

	var req models.ApplicationSearchRequest

	req.Query = c.Query("query")
	req.ActivityID = c.Query("activity_id")
	req.UUID = c.Query("id")
	req.Status = c.Query("status")
	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")
	req.MinCredits = c.Query("min_credits")
	req.MaxCredits = c.Query("max_credits")

	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", "10"),
	)
	req.Page = page
	req.PageSize = limit

	req.SortBy = c.DefaultQuery("sort_by", "submitted_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 构建基础查询
	dbQuery := h.db.Model(&models.Application{}).Preload("Activity")

	// 应用权限过滤
	if userType == "student" {
		dbQuery = dbQuery.Where("user_id = ?", userID)
	}

	// 应用搜索条件
	dbQuery = h.applyApplicationSearchConditions(dbQuery, req)

	// 获取总数
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 应用排序和分页
	var applications []models.Application
	query := dbQuery.
		Order(fmt.Sprintf("%s %s", req.SortBy, strings.ToUpper(req.SortOrder))).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize)

	if err := query.Find(&applications).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 转换为响应格式
	responses := h.buildApplicationResponses(applications, authToken)

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

// applyApplicationSearchConditions 应用申请搜索条件
func (h *SearchHandler) applyApplicationSearchConditions(query *gorm.DB, req models.ApplicationSearchRequest) *gorm.DB {
	// 文本搜索
	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		query = query.Where(
			"activity_id IN (SELECT id FROM credit_activities WHERE title ILIKE ? OR description ILIKE ?)",
			searchQuery, searchQuery,
		)
	}

	// 活动ID过滤
	if req.ActivityID != "" {
		query = query.Where("activity_id = ?", req.ActivityID)
	}

	// 用户ID过滤
	if req.UUID != "" {
		query = query.Where("user_id = ?", req.UUID)
	}

	// 状态过滤
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 开始日期过滤
	if req.StartDate != "" {
		if start, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			query = query.Where("submitted_at >= ?", start)
		}
	}

	// 结束日期过滤
	if req.EndDate != "" {
		if end, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			query = query.Where("submitted_at <= ?", end.Add(24*time.Hour))
		}
	}

	// 最小学分过滤
	if req.MinCredits != "" {
		if minCredits, err := strconv.ParseFloat(req.MinCredits, 64); err == nil {
			query = query.Where("awarded_credits >= ?", minCredits)
		}
	}

	// 最大学分过滤
	if req.MaxCredits != "" {
		if maxCredits, err := strconv.ParseFloat(req.MaxCredits, 64); err == nil {
			query = query.Where("awarded_credits <= ?", maxCredits)
		}
	}

	return query
}

// buildApplicationResponses 构建申请响应列表
func (h *SearchHandler) buildApplicationResponses(applications []models.Application, authToken string) []models.ApplicationResponse {
	responses := make([]models.ApplicationResponse, 0, len(applications))

	for _, app := range applications {
		userInfo, err := utils.GetUserInfo(app.UUID, authToken)
		if err != nil {
			continue
		}

		response := models.ApplicationResponse{
			ID:             app.ID,
			ActivityID:     app.ActivityID,
			UserID:         app.UUID,
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
			UserInfo: userInfo,
		}

		responses = append(responses, response)
	}

	return responses
}

func (h *SearchHandler) SearchParticipants(c *gin.Context) {
	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}
	userType := c.GetString("user_type")
	authToken := c.GetHeader("Authorization")

	var req models.ParticipantSearchRequest

	req.ActivityID = c.Query("activity_id")
	req.UUID = c.Query("id")
	req.MinCredits = c.Query("min_credits")
	req.MaxCredits = c.Query("max_credits")

	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", "10"),
	)
	req.Page = page
	req.PageSize = limit

	req.SortBy = c.DefaultQuery("sort_by", "joined_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 构建基础查询
	dbQuery := h.db.Model(&models.ActivityParticipant{})

	// 应用权限过滤
	if userType == "student" {
		dbQuery = dbQuery.Where("user_id = ?", userID)
	}

	// 应用搜索条件
	dbQuery = h.applyParticipantSearchConditions(dbQuery, req)

	// 获取总数
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 应用排序和分页
	var participants []models.ActivityParticipant
	query := dbQuery.
		Order(fmt.Sprintf("%s %s", req.SortBy, strings.ToUpper(req.SortOrder))).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize)

	if err := query.Find(&participants).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 转换为响应格式
	responses := h.buildParticipantResponses(participants, authToken)

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

// applyParticipantSearchConditions 应用参与者搜索条件
func (h *SearchHandler) applyParticipantSearchConditions(query *gorm.DB, req models.ParticipantSearchRequest) *gorm.DB {
	// 活动ID过滤
	if req.ActivityID != "" {
		query = query.Where("activity_id = ?", req.ActivityID)
	}

	// 用户ID过滤
	if req.UUID != "" {
		query = query.Where("user_id = ?", req.UUID)
	}

	// 最小学分过滤
	if req.MinCredits != "" {
		if minCredits, err := strconv.ParseFloat(req.MinCredits, 64); err == nil {
			query = query.Where("credits >= ?", minCredits)
		}
	}

	// 最大学分过滤
	if req.MaxCredits != "" {
		if maxCredits, err := strconv.ParseFloat(req.MaxCredits, 64); err == nil {
			query = query.Where("credits <= ?", maxCredits)
		}
	}

	return query
}

// buildParticipantResponses 构建参与者响应列表
func (h *SearchHandler) buildParticipantResponses(participants []models.ActivityParticipant, authToken string) []models.ParticipantResponse {
	responses := make([]models.ParticipantResponse, 0, len(participants))

	for _, participant := range participants {
		userInfo, err := utils.GetUserInfo(participant.UUID, authToken)
		if err != nil {
			continue
		}

		response := models.ParticipantResponse{
			UUID:     participant.UUID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
			UserInfo: userInfo,
		}

		responses = append(responses, response)
	}

	return responses
}

func (h *SearchHandler) SearchAttachments(c *gin.Context) {
	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}
	userType := c.GetString("user_type")

	var req models.AttachmentSearchRequest

	req.Query = c.Query("query")
	req.ActivityID = c.Query("activity_id")
	req.UploaderID = c.Query("uploader_id")
	req.FileType = c.Query("file_type")
	req.FileCategory = c.Query("file_category")
	req.MinSize = c.Query("min_size")
	req.MaxSize = c.Query("max_size")

	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", "10"),
	)
	req.Page = page
	req.PageSize = limit

	req.SortBy = c.DefaultQuery("sort_by", "uploaded_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// 构建基础查询
	dbQuery := h.db.Model(&models.Attachment{})

	// 应用权限过滤
	if userType == "student" {
		dbQuery = h.applyAttachmentPermissionFilter(dbQuery, userID)
	}

	// 应用搜索条件
	dbQuery = h.applyAttachmentSearchConditions(dbQuery, req)

	// 获取总数
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 应用排序和分页
	var attachments []models.Attachment
	query := dbQuery.
		Order(fmt.Sprintf("%s %s", req.SortBy, strings.ToUpper(req.SortOrder))).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize)

	if err := query.Find(&attachments).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 转换为响应格式
	responses := h.buildAttachmentResponses(attachments)

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

// applyAttachmentPermissionFilter 应用附件权限过滤
func (h *SearchHandler) applyAttachmentPermissionFilter(query *gorm.DB, userID string) *gorm.DB {
	return query.Where(
		"uploaded_by = ? OR activity_id IN (SELECT activity_id FROM activity_participants WHERE user_id = ?)",
		userID, userID,
	)
}

// applyAttachmentSearchConditions 应用附件搜索条件
func (h *SearchHandler) applyAttachmentSearchConditions(query *gorm.DB, req models.AttachmentSearchRequest) *gorm.DB {
	// 文本搜索
	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		query = query.Where(
			"file_name ILIKE ? OR original_name ILIKE ? OR description ILIKE ?",
			searchQuery, searchQuery, searchQuery,
		)
	}

	// 活动ID过滤
	if req.ActivityID != "" {
		query = query.Where("activity_id = ?", req.ActivityID)
	}

	// 上传者ID过滤
	if req.UploaderID != "" {
		query = query.Where("uploaded_by = ?", req.UploaderID)
	}

	// 文件类型过滤
	if req.FileType != "" {
		query = query.Where("file_type = ?", req.FileType)
	}

	// 文件分类过滤
	if req.FileCategory != "" {
		query = query.Where("file_category = ?", req.FileCategory)
	}

	// 最小文件大小过滤
	if req.MinSize != "" {
		if minSize, err := strconv.ParseInt(req.MinSize, 10, 64); err == nil {
			query = query.Where("file_size >= ?", minSize)
		}
	}

	// 最大文件大小过滤
	if req.MaxSize != "" {
		if maxSize, err := strconv.ParseInt(req.MaxSize, 10, 64); err == nil {
			query = query.Where("file_size <= ?", maxSize)
		}
	}

	return query
}

// buildAttachmentResponses 构建附件响应列表
func (h *SearchHandler) buildAttachmentResponses(attachments []models.Attachment) []models.AttachmentResponse {
	responses := make([]models.AttachmentResponse, 0, len(attachments))

	for _, attachment := range attachments {
		response := models.AttachmentResponse{
			ID:            attachment.ID,
			ActivityID:    attachment.ActivityID,
			FileName:      attachment.FileName,
			OriginalName:  attachment.OriginalName,
			FileType:      attachment.FileType,
			FileSize:      attachment.FileSize,
			FileCategory:  attachment.FileCategory,
			Description:   attachment.Description,
			UploadedBy:    attachment.UploadedBy,
			UploadedAt:    attachment.UploadedAt,
			DownloadCount: attachment.DownloadCount,
			// DownloadURL and Uploader can be set if needed
		}
		responses = append(responses, response)
	}

	return responses
}
