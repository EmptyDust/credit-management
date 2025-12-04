package handlers

import (
	"log"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicationHandler struct {
	db        *gorm.DB
	validator *utils.Validator
}

func NewApplicationHandler(db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{
		db:        db,
		validator: utils.NewValidator(),
	}
}

func (h *ApplicationHandler) GetUserApplications(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	log.Printf("[GetUserApplications] userID=%v", userID)

	status := c.Query("status")
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", "10"),
	)

	// 必须显式指定 Model，否则 GORM 无法推断表名，会报 "Table not set" 错误
	applications, total, err := h.getApplicationsWithPagination(
		h.db.Model(&models.Application{}).Where("user_id = ?", userID),
		status,
		page,
		limit,
	)
	if err != nil {
		log.Printf("[GetUserApplications] query error: %+v", err)
		utils.SendInternalServerError(c, err)
		return
	}

	// 获取用户信息（学生查看自己的申请，用户信息应该相同）
	responses := h.buildApplicationResponsesWithUserInfo(applications, c.GetHeader("Authorization"))
	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id := c.Param("id")
	if err := h.validator.ValidateUUID(id); err != nil {
		utils.SendBadRequest(c, "申请ID不能为空")
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	userType, _ := c.Get("user_type")

	var application models.Application
	if err := h.db.Preload("Activity").Where("id = ?", id).First(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "申请不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if userType == "student" && application.UUID != userID {
		utils.SendForbidden(c, "无权限查看此申请")
		return
	}

	response := h.buildApplicationResponse(application, c.GetHeader("Authorization"))
	// 获取用户信息
	if userInfo, err := utils.GetUserInfo(application.UUID, c.GetHeader("Authorization")); err == nil {
		response.UserInfo = userInfo
	} else {
		log.Printf("[GetApplication] failed to get user info for user_id=%s: %v", application.UUID, err)
	}
	utils.SendSuccessResponse(c, response)
}

func (h *ApplicationHandler) GetAllApplications(c *gin.Context) {
	activityID := c.Query("activity_id")
	userID := c.Query("id")
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", "10"),
	)

	query := h.db.Model(&models.Application{})
	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}
	if userID != "" {
		// 根据用户过滤时应使用 user_id 字段
		query = query.Where("user_id = ?", userID)
	}

	applications, total, err := h.getApplicationsWithPagination(query, "", page, limit)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 获取所有申请的用户信息
	responses := h.buildApplicationResponsesWithUserInfo(applications, c.GetHeader("Authorization"))
	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ApplicationHandler) getApplicationsWithPagination(query *gorm.DB, status string, page, limit int) ([]models.Application, int64, error) {
	// 确保始终设置了 Model，避免出现 "Table not set" 的 GORM 错误
	query = query.Model(&models.Application{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var applications []models.Application
	offset := (page - 1) * limit
	err := query.
		Preload("Activity").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&applications).Error

	return applications, total, err
}

func (h *ApplicationHandler) buildApplicationResponses(applications []models.Application, authToken string) []models.ApplicationResponse {
	responses := make([]models.ApplicationResponse, 0, len(applications))
	for _, app := range applications {
		responses = append(responses, h.buildApplicationResponse(app, authToken))
	}
	return responses
}

func (h *ApplicationHandler) buildApplicationResponsesWithUserInfo(applications []models.Application, authToken string) []models.ApplicationResponse {
	responses := make([]models.ApplicationResponse, 0, len(applications))
	for _, app := range applications {
		response := h.buildApplicationResponse(app, authToken)
		// 获取用户信息
		if userInfo, err := utils.GetUserInfo(app.UUID, authToken); err == nil {
			response.UserInfo = userInfo
		} else {
			log.Printf("[buildApplicationResponsesWithUserInfo] failed to get user info for user_id=%s: %v", app.UUID, err)
		}
		responses = append(responses, response)
	}
	return responses
}

func (h *ApplicationHandler) buildApplicationResponse(app models.Application, authToken string) models.ApplicationResponse {
	return models.ApplicationResponse{
		ID:         app.ID,
		ActivityID: app.ActivityID,
		UserID:     app.UUID,
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
		// 列表接口默认不附带 UserInfo，避免对用户服务的高频调用
	}
}

func (h *ApplicationHandler) GetApplicationStats(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var stats struct {
		TotalApplications int64   `json:"total_applications"`
		PendingCount      int64   `json:"pending_count"`
		ApprovedCount     int64   `json:"approved_count"`
		RejectedCount     int64   `json:"rejected_count"`
		TotalCredits      float64 `json:"total_credits"`
	}

	// 注意：这里统计的是当前登录用户的申请，字段为 user_id 而不是 id
	h.db.Model(&models.Application{}).Where("user_id = ?", userID).Count(&stats.TotalApplications)
	h.db.Model(&models.Application{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&stats.PendingCount)
	h.db.Model(&models.Application{}).Where("user_id = ? AND status = ?", userID, "approved").Count(&stats.ApprovedCount)
	h.db.Model(&models.Application{}).Where("user_id = ? AND status = ?", userID, "rejected").Count(&stats.RejectedCount)
	h.db.Model(&models.Application{}).Where("user_id = ? AND status = ?", userID, "approved").Select("COALESCE(SUM(awarded_credits), 0)").Scan(&stats.TotalCredits)

	utils.SendSuccessResponse(c, stats)
}

func (h *ApplicationHandler) ExportApplications(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	activityID := c.Query("activity_id")
	userID := c.Query("id")

	query := h.db.Model(&models.Application{}).Preload("Activity")
	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}
	if userID != "" {
		// 按用户过滤时同样应该使用 user_id
		query = query.Where("user_id = ?", userID)
	}

	var applications []models.Application
	if err := query.Order("created_at DESC").Find(&applications).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	switch format {
	case "json":
		responses := h.buildApplicationResponses(applications, c.GetHeader("Authorization"))
		utils.SendSuccessResponse(c, responses)
	case "csv":
		h.exportToCSV(c, applications)
	case "excel":
		h.exportToExcel(c, applications)
	default:
		utils.SendBadRequest(c, "不支持的导出格式")
	}
}

func (h *ApplicationHandler) exportToCSV(c *gin.Context, applications []models.Application) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=applications.csv")

	// CSV导出逻辑
	utils.SendSuccessResponse(c, gin.H{"message": "CSV导出功能待实现", "count": len(applications)})
}

func (h *ApplicationHandler) exportToExcel(c *gin.Context, applications []models.Application) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=applications.xlsx")

	// Excel导出逻辑
	utils.SendSuccessResponse(c, gin.H{"message": "Excel导出功能待实现", "count": len(applications)})
}
