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

	// 为列表构建轻量级响应，避免为每个活动加载全部参与者和申请详情
	if len(activities) == 0 {
		utils.SendPaginatedResponse(c, []models.ActivityResponse{}, total, page, limit)
		return
	}

	// 收集本页活动ID
	activityIDs := make([]string, 0, len(activities))
	for _, a := range activities {
		activityIDs = append(activityIDs, a.ID)
	}

	// 统计参与者数量
	type countResult struct {
		ActivityID string
		Count      int64
	}

	participantCounts := make([]countResult, 0, len(activities))
	if err := h.db.Table("activity_participants").
		Select("activity_id, COUNT(*) AS count").
		Where("activity_id IN ?", activityIDs).
		Group("activity_id").
		Scan(&participantCounts).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 统计申请数量
	applicationCounts := make([]countResult, 0, len(activities))
	if err := h.db.Table("applications").
		Select("activity_id, COUNT(*) AS count").
		Where("activity_id IN ?", activityIDs).
		Group("activity_id").
		Scan(&applicationCounts).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 构建 ID -> count 映射
	participantMap := make(map[string]int64, len(participantCounts))
	for _, item := range participantCounts {
		participantMap[item.ActivityID] = item.Count
	}

	applicationMap := make(map[string]int64, len(applicationCounts))
	for _, item := range applicationCounts {
		applicationMap[item.ActivityID] = item.Count
	}

	// 组装响应，仅包含基础字段和聚合计数
	responses := make([]models.ActivityResponse, 0, len(activities))
	for _, a := range activities {
		resp := models.ActivityResponse{
			ID:                a.ID,
			Title:             a.Title,
			Description:       a.Description,
			StartDate:         a.StartDate,
			EndDate:           a.EndDate,
			Status:            a.Status,
			Category:          a.Category,
			OwnerID:           a.OwnerID,
			ReviewerID:        a.ReviewerID,
			ReviewComments:    a.ReviewComments,
			ReviewedAt:        a.ReviewedAt,
			CreatedAt:         a.CreatedAt,
			UpdatedAt:         a.UpdatedAt,
			ParticipantsCount: participantMap[a.ID],
			ApplicationsCount: applicationMap[a.ID],
			// 列表页暂不返回 OwnerInfo / Participants / Applications / Details，减少数据量和外部调用
		}
		responses = append(responses, resp)
	}

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

	// 只有活动创建者、教师或管理员可以删除活动
	if userType != "admin" && userType != "teacher" && activity.OwnerID != userID {
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
