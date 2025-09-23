package handlers

import (
	"strconv"
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
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}
	userType, _ := c.Get("user_type")

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

	dbQuery := h.db.Model(&models.CreditActivity{})

	// 权限过滤：学生只能看到自己创建或参与的活动
	if userType == "student" {
		dbQuery = dbQuery.Where("owner_id = ? OR id IN (SELECT activity_id FROM activity_participants WHERE id = ?)", userID, userID)
	}

	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		dbQuery = dbQuery.Where(
			"title ILIKE ? OR description ILIKE ? OR category ILIKE ?",
			searchQuery, searchQuery, searchQuery,
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

	var total int64
	dbQuery.Count(&total)

	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	offset := (req.Page - 1) * req.PageSize
	var activities []models.CreditActivity
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&activities).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

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

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *SearchHandler) SearchApplications(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}
	userType, _ := c.Get("user_type")

	authToken := c.GetHeader("Authorization")

	var req models.ApplicationSearchRequest

	req.Query = c.Query("query")
	req.ActivityID = c.Query("activity_id")
	req.UserID = c.Query("id")
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

	dbQuery := h.db.Model(&models.Application{}).Preload("Activity")

	if userType == "student" {
		dbQuery = dbQuery.Where("id = ?", userID)
	}

	if req.Query != "" {
		searchQuery := "%" + req.Query + "%"
		dbQuery = dbQuery.Where(
			"activity_id IN (SELECT id FROM credit_activities WHERE title ILIKE ? OR description ILIKE ?)",
			searchQuery, searchQuery,
		)
	}

	if req.ActivityID != "" {
		dbQuery = dbQuery.Where("activity_id = ?", req.ActivityID)
	}
	if req.UserID != "" {
		dbQuery = dbQuery.Where("id = ?", req.UserID)
	}
	if req.Status != "" {
		dbQuery = dbQuery.Where("status = ?", req.Status)
	}

	if req.StartDate != "" {
		if start, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			dbQuery = dbQuery.Where("submitted_at >= ?", start)
		}
	}
	if req.EndDate != "" {
		if end, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			dbQuery = dbQuery.Where("submitted_at <= ?", end.Add(24*time.Hour))
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

	var total int64
	dbQuery.Count(&total)

	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	offset := (req.Page - 1) * req.PageSize
	var applications []models.Application
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&applications).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.ApplicationResponse
	for _, app := range applications {
		userInfo, err := utils.GetUserInfo(app.UserID, authToken)
		if err != nil {
			continue
		}

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
			UserInfo: userInfo,
		}

		responses = append(responses, response)
	}

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *SearchHandler) SearchParticipants(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}
	userType, _ := c.Get("user_type")

	authToken := c.GetHeader("Authorization")

	var req models.ParticipantSearchRequest

	req.ActivityID = c.Query("activity_id")
	req.UserID = c.Query("id")
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

	dbQuery := h.db.Model(&models.ActivityParticipant{})

	if userType == "student" {
		dbQuery = dbQuery.Where("id = ?", userID)
	}

	if req.ActivityID != "" {
		dbQuery = dbQuery.Where("activity_id = ?", req.ActivityID)
	}

	if req.UserID != "" {
		dbQuery = dbQuery.Where("id = ?", req.UserID)
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

	var total int64
	dbQuery.Count(&total)

	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	offset := (req.Page - 1) * req.PageSize
	var participants []models.ActivityParticipant
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&participants).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.ParticipantResponse
	for _, participant := range participants {
		userInfo, err := utils.GetUserInfo(participant.UserID, authToken)
		if err != nil {
			continue
		}

		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
			UserInfo: userInfo,
		}

		responses = append(responses, response)
	}

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *SearchHandler) SearchAttachments(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}
	userType, _ := c.Get("user_type")

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

	dbQuery := h.db.Model(&models.Attachment{})

	// 权限过滤：学生只能看到自己创建或参与活动的附件
	if userType == "student" {
		dbQuery = dbQuery.Where("uploaded_by = ? OR activity_id IN (SELECT activity_id FROM activity_participants WHERE id = ?)", userID, userID)
	}

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

	var total int64
	dbQuery.Count(&total)

	orderClause := req.SortBy
	if req.SortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	offset := (req.Page - 1) * req.PageSize
	var attachments []models.Attachment
	if err := dbQuery.Offset(offset).Limit(req.PageSize).Order(orderClause).Find(&attachments).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.AttachmentResponse
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

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}
