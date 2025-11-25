package handlers

import (
	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req models.ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}

	if err := h.validateActivityRequest(req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	startDate, endDate, err := utils.ParseDateRange(req.StartDate, req.EndDate)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	activity := models.CreditActivity{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      models.StatusDraft,
		Category:    req.Category,
		OwnerID:     userID,
		Details:     datatypes.JSONMap(req.Details),
	}

	if err := h.db.Create(&activity).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 详情改为写入 JSONB

	response := h.enrichActivityResponse(activity, "")
	utils.SendCreatedResponse(c, "活动创建成功", response)
}

func (h *ActivityHandler) GetActivities(c *gin.Context) {
	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}
	userType := c.GetString("user_type")
	if userType == "" {
		utils.SendUnauthorized(c)
		return
	}

	query := c.Query("query")
	status := c.Query("status")
	category := c.Query("category")
	ownerID := c.Query("owner_id")

	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", c.DefaultQuery("limit", "10")),
	)

	activities, total, err := h.base.SearchActivities(query, status, category, ownerID, userID, userType, page, limit)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	responses := h.buildActivityResponses(activities, c.GetHeader("Authorization"))
	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id := c.Param("id")
	if err := h.validator.ValidateUUID(id); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}

	userType := c.GetString("user_type")
	if userType == "" {
		utils.SendUnauthorized(c)
		return
	}

	activity, err := h.base.GetActivityByIDWithParticipants(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	// 权限检查
	if userType == "student" && activity.OwnerID != userID {
		if err := h.base.CheckUserParticipant(id, userID); err != nil {
			utils.SendForbidden(c, "无权限查看此活动")
			return
		}
	}

	response := h.enrichActivityResponse(*activity, c.GetHeader("Authorization"))
	utils.SendSuccessResponse(c, response)
}

func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	if err := h.validator.ValidateUUID(id); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	userType, _ := c.Get("user_type")

	activity, err := h.base.GetActivityByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if userType == "student" && activity.OwnerID != userID {
		utils.SendForbidden(c, "无权限修改此活动")
		return
	}

	var req models.ActivityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.validateUpdateRequest(req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	updates := h.buildUpdateMap(req)
	if err := h.db.Model(&models.CreditActivity{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	updatedActivity, err := h.base.GetActivityByID(id)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	response := h.enrichActivityResponse(*updatedActivity, "")
	utils.SendSuccessResponse(c, response)
}

func (h *ActivityHandler) buildUpdateMap(req models.ActivityUpdateRequest) map[string]interface{} {
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Details != nil {
		updates["details"] = datatypes.JSONMap(req.Details)
	}

	if req.StartDate != nil || req.EndDate != nil {
		startDateStr := ""
		endDateStr := ""
		if req.StartDate != nil {
			startDateStr = *req.StartDate
		}
		if req.EndDate != nil {
			endDateStr = *req.EndDate
		}

		if startDate, endDate, err := utils.ParseDateRange(startDateStr, endDateStr); err == nil {
			if !startDate.IsZero() {
				updates["start_date"] = startDate
			}
			if !endDate.IsZero() {
				updates["end_date"] = endDate
			}
		}
	}

	return updates
}

func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	if err := h.validator.ValidateUUID(id); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	userID := c.GetString("id")
	if userID == "" {
		utils.SendUnauthorized(c)
		return
	}
	userType := c.GetString("user_type")
	if userType == "" {
		utils.SendUnauthorized(c)
		return
	}

	activity, err := h.base.GetActivityByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if userType != "admin" && activity.OwnerID != userID {
		utils.SendForbidden(c, "无权限删除该活动")
		return
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	attachments, err := h.softDeleteActivityRelations(tx, activity.ID)
	if err != nil {
		tx.Rollback()
		utils.SendInternalServerError(c, err)
		return
	}

	if err := tx.Delete(&models.CreditActivity{ID: activity.ID}).Error; err != nil {
		tx.Rollback()
		utils.SendInternalServerError(c, err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	h.cleanupAttachmentFiles(attachments)

	utils.SendSuccessResponse(c, gin.H{"message": "活动删除成功"})
}

func (h *ActivityHandler) buildActivityResponses(activities []models.CreditActivity, authToken string) []models.ActivityResponse {
	var responses []models.ActivityResponse
	for _, activity := range activities {
		response := h.enrichActivityResponse(activity, authToken)
		responses = append(responses, response)
	}
	return responses
}
