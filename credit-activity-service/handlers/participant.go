package handlers

import (
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ParticipantHandler struct {
	db        *gorm.DB
	validator *utils.Validator
}

func NewParticipantHandler(db *gorm.DB) *ParticipantHandler {
	return &ParticipantHandler{
		db:        db,
		validator: utils.NewValidator(),
	}
}

func (h *ParticipantHandler) AddParticipants(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("id")
	userType, _ := c.Get("user_type")

	var req models.AddParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 基础校验：ids 非空且不包含空值
	if len(req.UUIDs) == 0 {
		utils.SendBadRequest(c, "请提供要添加的用户ID列表")
		return
	}
	for _, id := range req.UUIDs {
		if id == "" {
			utils.SendBadRequest(c, "用户ID列表包含空值")
			return
		}
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" && userType != "teacher" {
		utils.SendForbidden(c, "权限不足，只有活动创建者、教师或管理员可以添加参与者")
		return
	}

	authToken := c.GetHeader("Authorization")
	for _, targetUserID := range req.UUIDs {
		if !h.isStudent(targetUserID, authToken) {
			log.Printf("AddParticipants validation failed: targetUserID=%s not a student or user lookup failed", targetUserID)
			utils.SendBadRequest(c, "只能添加学生用户作为参与者")
			return
		}
	}

	var participants []models.ActivityParticipant
	var addedCount int

	for _, targetUserID := range req.UUIDs {
		var existing models.ActivityParticipant
		err := h.db.Where("activity_id = ? AND user_id = ?", activityID, targetUserID).First(&existing).Error
		if err == nil {
			continue
		}

		participant := models.ActivityParticipant{
			ActivityID: activityID,
			UUID:       targetUserID,
			Credits:    req.Credits,
			JoinedAt:   time.Now(),
		}

		if err := h.db.Create(&participant).Error; err != nil {
			log.Printf("Failed to create participant: activity=%s user=%s err=%v", activityID, targetUserID, err)
			continue
		}

		participants = append(participants, participant)
		addedCount++
	}

	var responses []models.ParticipantResponse
	for _, participant := range participants {
		userInfo, err := h.getUserInfo(participant.UUID, authToken)
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

	utils.SendSuccessResponse(c, gin.H{
		"added_count":  addedCount,
		"participants": responses,
	})
}

func (h *ParticipantHandler) BatchSetCredits(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("id")
	userType, _ := c.Get("user_type")

	var req models.BatchCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "参数错误: "+err.Error())
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" && userType != "teacher" {
		utils.SendForbidden(c, "权限不足，只有活动创建者、教师或管理员可以设置学分")
		return
	}

	authToken := c.GetHeader("Authorization")
	var updatedParticipants []models.ParticipantResponse
	updatedCount := 0

	for userID, credits := range req.CreditsMap {
		var participant models.ActivityParticipant
		if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&participant).Error; err != nil {
			continue
		}

		participant.Credits = credits
		if err := h.db.Save(&participant).Error; err != nil {
			continue
		}

		userInfo, err := h.getUserInfo(participant.UUID, authToken)
		if err != nil {
			continue
		}

		response := models.ParticipantResponse{
			UUID:     participant.UUID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
			UserInfo: userInfo,
		}

		updatedParticipants = append(updatedParticipants, response)
		updatedCount++
	}

	utils.SendSuccessResponse(c, gin.H{
		"updated_count": updatedCount,
		"participants":  updatedParticipants,
	})
}

func (h *ParticipantHandler) SetSingleCredits(c *gin.Context) {
	activityID := c.Param("id")
	participantID := c.Param("uuid")
	userID, _ := c.Get("id")
	userType, _ := c.Get("user_type")

	var req models.SingleCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "参数错误: "+err.Error())
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" && userType != "teacher" {
		utils.SendForbidden(c, "权限不足，只有活动创建者、教师或管理员可以设置学分")
		return
	}

	var participant models.ActivityParticipant
	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, participantID).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "参与者不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	participant.Credits = req.Credits
	if err := h.db.Save(&participant).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	userInfo, err := h.getUserInfo(participant.UUID, c.GetHeader("Authorization"))
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	response := models.ParticipantResponse{
		UUID:     participant.UUID,
		Credits:  participant.Credits,
		JoinedAt: participant.JoinedAt,
		UserInfo: userInfo,
	}

	utils.SendSuccessResponse(c, response)
}

func (h *ParticipantHandler) RemoveParticipant(c *gin.Context) {
	activityID := c.Param("id")
	participantID := c.Param("uuid")
	userID, _ := c.Get("id")
	userType, _ := c.Get("user_type")

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" && userType != "teacher" {
		utils.SendForbidden(c, "权限不足，只有活动创建者、教师或管理员可以移除参与者")
		return
	}

	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, participantID).Delete(&models.ActivityParticipant{}).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "参与者移除成功"})
}

func (h *ParticipantHandler) LeaveActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("id")

	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).Delete(&models.ActivityParticipant{}).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "成功退出活动"})
}

func (h *ParticipantHandler) GetActivityParticipants(c *gin.Context) {
	activityID := c.Param("id")
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("limit", "10"),
	)

	var participants []models.ActivityParticipant
	var total int64

	query := h.db.Model(&models.ActivityParticipant{}).Where("activity_id = ?", activityID)
	query.Count(&total)

	offset := (page - 1) * limit
	if err := h.db.Where("activity_id = ?", activityID).Offset(offset).Limit(limit).Order("joined_at DESC").Find(&participants).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.ParticipantResponse
	authToken := c.GetHeader("Authorization")
	for _, participant := range participants {
		userInfo, err := h.getUserInfo(participant.UUID, authToken)
		if err != nil {
			// 如果获取用户信息失败，创建一个基本的用户信息
			userInfo = &models.UserInfo{
				UUID:     participant.UUID,
				Username: userInfo.Username,
				RealName: "未知用户",
				UserType: "unknown",
				Status:   "unknown",
			}
		}

		response := models.ParticipantResponse{
			UUID:     participant.UUID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
			UserInfo: userInfo,
		}

		responses = append(responses, response)
	}

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ParticipantHandler) BatchRemoveParticipants(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("id")
	userType, _ := c.Get("user_type")

	var req models.BatchRemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "参数错误: "+err.Error())
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" && userType != "teacher" {
		utils.SendForbidden(c, "权限不足，只有活动创建者、教师或管理员可以批量移除参与者")
		return
	}

	removedCount := 0
	for _, participantID := range req.UUIDs {
		if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, participantID).Delete(&models.ActivityParticipant{}).Error; err == nil {
			removedCount++
		}
	}

	utils.SendSuccessResponse(c, gin.H{
		"removed_count":   removedCount,
		"total_requested": len(req.UUIDs),
	})
}

func (h *ParticipantHandler) GetParticipantStats(c *gin.Context) {
	activityID := c.Param("id")

	var stats struct {
		TotalParticipants int64   `json:"total_participants"`
		TotalCredits      float64 `json:"total_credits"`
		AverageCredits    float64 `json:"average_credits"`
	}

	h.db.Model(&models.ActivityParticipant{}).Where("activity_id = ?", activityID).Count(&stats.TotalParticipants)
	h.db.Model(&models.ActivityParticipant{}).Where("activity_id = ?", activityID).Select("COALESCE(SUM(credits), 0)").Scan(&stats.TotalCredits)

	if stats.TotalParticipants > 0 {
		stats.AverageCredits = stats.TotalCredits / float64(stats.TotalParticipants)
	}

	utils.SendSuccessResponse(c, stats)
}

func (h *ParticipantHandler) ExportParticipants(c *gin.Context) {
	activityID := c.Param("id")
	format := c.DefaultQuery("format", "json")

	var participants []models.ActivityParticipant
	if err := h.db.Where("activity_id = ?", activityID).Order("joined_at DESC").Find(&participants).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	switch format {
	case "json":
		var responses []models.ParticipantResponse
		authToken := c.GetHeader("Authorization")
		for _, participant := range participants {
			userInfo, err := h.getUserInfo(participant.UUID, authToken)
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
		utils.SendSuccessResponse(c, responses)
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=participants.csv")
		utils.SendSuccessResponse(c, gin.H{"message": "CSV导出功能待实现", "count": len(participants)})
	case "excel":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=participants.xlsx")
		utils.SendSuccessResponse(c, gin.H{"message": "Excel导出功能待实现", "count": len(participants)})
	default:
		utils.SendBadRequest(c, "不支持的导出格式")
	}
}

func (h *ParticipantHandler) GetUserParticipatedActivities(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("limit", "10"),
	)

	var participants []models.ActivityParticipant
	var total int64

	query := h.db.Where("user_id = ?", userID).Preload("Activity")
	query.Count(&total)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("joined_at DESC").Find(&participants).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.ParticipantActivityResponse
	for _, participant := range participants {
		response := models.ParticipantActivityResponse{
			ActivityID: participant.ActivityID,
			Credits:    participant.Credits,
			JoinedAt:   participant.JoinedAt,
			Activity: models.ActivityInfo{
				ID:          participant.Activity.ID,
				Title:       participant.Activity.Title,
				Description: participant.Activity.Description,
				Category:    participant.Activity.Category,
				StartDate:   participant.Activity.StartDate,
				EndDate:     participant.Activity.EndDate,
			},
		}
		responses = append(responses, response)
	}

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ParticipantHandler) isStudent(userID string, authToken string) bool {
	userInfo, err := utils.GetUserInfo(userID, authToken)
	return err == nil && userInfo != nil && userInfo.UserType == "student"
}

func (h *ParticipantHandler) getUserInfo(userID string, authToken string) (*models.UserInfo, error) {
	return utils.GetUserInfo(userID, authToken)
}
