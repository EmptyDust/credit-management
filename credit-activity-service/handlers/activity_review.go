package handlers

import (
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *ActivityHandler) SubmitActivity(c *gin.Context) {
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

	// 使用数据库基类获取活动
	activity, err := h.base.GetActivityByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.OwnerID != userID {
		utils.SendForbidden(c, "无权限提交此活动")
		return
	}

	if activity.Status != models.StatusDraft {
		utils.SendBadRequest(c, "只能提交草稿状态的活动")
		return
	}

	if err := h.db.Model(&activity).Update("status", models.StatusPendingReview).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"id":     activity.ID,
		"status": models.StatusPendingReview,
	})
}

func (h *ActivityHandler) ReviewActivity(c *gin.Context) {
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

	var req models.ActivityReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 使用数据库基类获取活动
	activity, err := h.base.GetActivityByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.Status != models.StatusPendingReview &&
		activity.Status != models.StatusApproved &&
		activity.Status != models.StatusRejected {
		utils.SendBadRequest(c, "只能审核待审核、已通过或已拒绝状态的活动")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":          req.Status,
		"reviewer_id":     userID.(string),
		"review_comments": req.ReviewComments,
		"reviewed_at":     &now,
	}

	if err := h.db.Model(&activity).Updates(updates).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"id":              activity.ID,
		"status":          req.Status,
		"reviewer_id":     userID.(string),
		"review_comments": req.ReviewComments,
		"reviewed_at":     now,
	})
}

func (h *ActivityHandler) GetPendingActivities(c *gin.Context) {
	// 使用统一的验证器处理分页参数
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("limit", "10"),
	)

	// 使用数据库基类获取待审核活动
	activities, total, err := h.base.GetPendingActivities(page, limit)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.ActivityResponse
	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

	for _, activity := range activities {
		response := h.enrichActivityResponse(activity, authToken)
		responses = append(responses, response)
	}

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ActivityHandler) WithdrawActivity(c *gin.Context) {
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

	// 使用数据库基类获取活动
	activity, err := h.base.GetActivityByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.OwnerID != userID {
		utils.SendForbidden(c, "无权限撤回此活动")
		return
	}

	if activity.Status != models.StatusPendingReview {
		utils.SendBadRequest(c, "只能撤回待审核状态的活动")
		return
	}

	if err := h.db.Model(&activity).Update("status", models.StatusDraft).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"id":     activity.ID,
		"status": models.StatusDraft,
	})
}

func (h *ActivityHandler) GetDeletableActivities(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	// 使用统一的验证器处理分页参数
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("limit", "10"),
	)

	// 使用数据库基类获取可删除的活动
	activities, total, err := h.base.GetDeletableActivities(userID.(string), page, limit)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.ActivityResponse
	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

	for _, activity := range activities {
		response := h.enrichActivityResponse(activity, authToken)
		responses = append(responses, response)
	}

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}
