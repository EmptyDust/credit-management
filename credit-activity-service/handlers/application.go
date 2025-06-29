package handlers

import (
	"fmt"
	"strconv"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicationHandler struct {
	db *gorm.DB
}

func NewApplicationHandler(db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{db: db}
}

func (h *ApplicationHandler) GetUserApplications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	authToken := c.GetHeader("Authorization")

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var applications []models.Application
	var total int64

	query := h.db.Model(&models.Application{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	if err := query.Preload("Activity").Offset(offset).Limit(limit).Order("created_at DESC").Find(&applications).Error; err != nil {
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

	totalPages := (int(total) + limit - 1) / limit

	utils.SendSuccessResponse(c, models.PaginatedResponse{
		Data:       responses,
		Total:      int64(len(responses)), // 使用实际返回的记录数
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.SendBadRequest(c, "申请ID不能为空")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	authToken := c.GetHeader("Authorization")

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

	if userType == "student" && application.UserID != userID {
		utils.SendForbidden(c, "无权限查看此申请")
		return
	}

	userInfo, err := utils.GetUserInfo(application.UserID, authToken)
	if err != nil {
		utils.SendNotFound(c, "申请关联的用户不存在")
		return
	}

	response := models.ApplicationResponse{
		ID:             application.ID,
		ActivityID:     application.ActivityID,
		UserID:         application.UserID,
		Status:         application.Status,
		AppliedCredits: application.AppliedCredits,
		AwardedCredits: application.AwardedCredits,
		SubmittedAt:    application.SubmittedAt,
		CreatedAt:      application.CreatedAt,
		UpdatedAt:      application.UpdatedAt,
		Activity: models.ActivityInfo{
			ID:          application.Activity.ID,
			Title:       application.Activity.Title,
			Description: application.Activity.Description,
			Category:    application.Activity.Category,
			StartDate:   application.Activity.StartDate,
			EndDate:     application.Activity.EndDate,
		},
		UserInfo: userInfo,
	}

	utils.SendSuccessResponse(c, response)
}

func (h *ApplicationHandler) GetAllApplications(c *gin.Context) {
	authToken := c.GetHeader("Authorization")

	activityID := c.Query("activity_id")
	userID := c.Query("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var applications []models.Application
	var total int64

	query := h.db.Model(&models.Application{})

	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	query.Count(&total)

	if err := query.Preload("Activity").Offset(offset).Limit(limit).Order("created_at DESC").Find(&applications).Error; err != nil {
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

	totalPages := (int(total) + limit - 1) / limit

	utils.SendSuccessResponse(c, models.PaginatedResponse{
		Data:       responses,
		Total:      int64(len(responses)), // 使用实际返回的记录数
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *ApplicationHandler) GetApplicationStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var stats models.ApplicationStats

	h.db.Model(&models.Application{}).Where("user_id = ?", userID).Count(&stats.TotalApplications)
	h.db.Model(&models.Application{}).Where("user_id = ?", userID).Select("COALESCE(SUM(applied_credits), 0)").Scan(&stats.TotalCredits)
	h.db.Model(&models.Application{}).Where("user_id = ?", userID).Select("COALESCE(SUM(awarded_credits), 0)").Scan(&stats.AwardedCredits)

	utils.SendSuccessResponse(c, stats)
}

func (h *ApplicationHandler) ExportApplications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	format := c.DefaultQuery("format", "csv")
	activityID := c.Query("activity_id")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := h.db.Model(&models.Application{})

	if userType == "student" {
		query = query.Where("user_id = ?", userID)
	}

	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		if start, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("submitted_at >= ?", start)
		}
	}
	if endDate != "" {
		if end, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("submitted_at <= ?", end.Add(24*time.Hour))
		}
	}

	var applications []models.Application
	if err := query.Order("submitted_at DESC").Find(&applications).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	switch format {
	case "csv":
		h.exportToCSV(c, applications)
	case "excel":
		h.exportToExcel(c, applications)
	default:
		utils.SendBadRequest(c, "不支持的导出格式，支持的格式：csv, excel")
	}
}

func (h *ApplicationHandler) exportToCSV(c *gin.Context, applications []models.Application) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=applications.csv")

	c.Writer.WriteString("申请ID,活动ID,用户ID,状态,申请学分,获得学分,提交时间,创建时间\n")

	for _, app := range applications {
		line := fmt.Sprintf("%s,%s,%s,%s,%.2f,%.2f,%s,%s\n",
			app.ID,
			app.ActivityID,
			app.UserID,
			app.Status,
			app.AppliedCredits,
			app.AwardedCredits,
			app.SubmittedAt.Format("2006-01-02 15:04:05"),
			app.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		c.Writer.WriteString(line)
	}
}

func (h *ApplicationHandler) exportToExcel(c *gin.Context, applications []models.Application) {
	// 这里应该使用Excel库生成Excel文件
	// 暂时返回CSV格式
	h.exportToCSV(c, applications)
}
