package handlers

import (
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

	status := c.Query("status")
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("limit", "10"),
	)

	applications, total, err := h.getApplicationsWithPagination(
		h.db.Where("user_id = ?", userID),
		status,
		page,
		limit,
	)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	responses := h.buildApplicationResponses(applications, c.GetHeader("Authorization"))
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
	utils.SendSuccessResponse(c, response)
}

func (h *ApplicationHandler) GetAllApplications(c *gin.Context) {
	activityID := c.Query("activity_id")
	userID := c.Query("id")
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("limit", "10"),
	)

	query := h.db.Model(&models.Application{})
	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}
	if userID != "" {
		query = query.Where("id = ?", userID)
	}

	applications, total, err := h.getApplicationsWithPagination(query, "", page, limit)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	responses := h.buildApplicationResponses(applications, c.GetHeader("Authorization"))
	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ApplicationHandler) getApplicationsWithPagination(query *gorm.DB, status string, page, limit int) ([]models.Application, int64, error) {
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var applications []models.Application
	offset := (page - 1) * limit
	err := query.Preload("Activity").Offset(offset).Limit(limit).Order("created_at DESC").Find(&applications).Error

	return applications, total, err
}

func (h *ApplicationHandler) buildApplicationResponses(applications []models.Application, authToken string) []models.ApplicationResponse {
	var responses []models.ApplicationResponse
	for _, app := range applications {
		response := h.buildApplicationResponse(app, authToken)
		responses = append(responses, response)
	}
	return responses
}

func (h *ApplicationHandler) buildApplicationResponse(app models.Application, authToken string) models.ApplicationResponse {
	userInfo, _ := utils.GetUserInfo(app.UUID, authToken)

	return models.ApplicationResponse{
		ID:             app.ID,
		ActivityID:     app.ActivityID,
		UUID:           app.UUID,
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

	h.db.Model(&models.Application{}).Where("id = ?", userID).Count(&stats.TotalApplications)
	h.db.Model(&models.Application{}).Where("id = ? AND status = ?", userID, "pending").Count(&stats.PendingCount)
	h.db.Model(&models.Application{}).Where("id = ? AND status = ?", userID, "approved").Count(&stats.ApprovedCount)
	h.db.Model(&models.Application{}).Where("id = ? AND status = ?", userID, "rejected").Count(&stats.RejectedCount)
	h.db.Model(&models.Application{}).Where("id = ? AND status = ?", userID, "approved").Select("COALESCE(SUM(awarded_credits), 0)").Scan(&stats.TotalCredits)

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
		query = query.Where("id = ?", userID)
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
