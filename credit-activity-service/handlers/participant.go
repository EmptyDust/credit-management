package handlers

import (
	"net/http"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ParticipantHandler 参与者处理器
type ParticipantHandler struct {
	db *gorm.DB
}

// NewParticipantHandler 创建参与者处理器
func NewParticipantHandler(db *gorm.DB) *ParticipantHandler {
	return &ParticipantHandler{db: db}
}

// AddParticipants 添加参与者
func (h *ParticipantHandler) AddParticipants(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	var req models.AddParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    "指定的活动不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败",
				"data":    err.Error(),
			})
		}
		return
	}

	// 权限检查：只有活动创建者和管理员可以添加参与者
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者或管理员可以添加参与者",
			"data":    nil,
		})
		return
	}

	// 获取当前用户的认证令牌
	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

	// 验证用户角色：只有学生可以参与活动
	for _, targetUserID := range req.UserIDs {
		if !h.isStudent(targetUserID, authToken) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "只能添加学生用户作为参与者",
				"data":    targetUserID,
			})
			return
		}
	}

	var participants []models.ActivityParticipant
	var addedCount int

	for _, targetUserID := range req.UserIDs {
		// 检查是否已经是参与者
		var existing models.ActivityParticipant
		err := h.db.Where("activity_id = ? AND user_id = ?", activityID, targetUserID).First(&existing).Error
		if err == nil {
			// 已经是参与者，跳过
			continue
		}

		participant := models.ActivityParticipant{
			ActivityID: activityID,
			UserID:     targetUserID,
			Credits:    req.Credits,
			JoinedAt:   time.Now(),
		}

		if err := h.db.Create(&participant).Error; err != nil {
			continue
		}

		participants = append(participants, participant)
		addedCount++
	}

	// 获取参与者信息（包含用户信息）
	var responses []models.ParticipantResponse
	for _, participant := range participants {
		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
		}

		// 获取用户信息
		if userInfo, err := h.getUserInfo(participant.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "参与者添加成功",
		"data": gin.H{
			"added_count":  addedCount,
			"participants": responses,
		},
	})
}

// BatchSetCredits 批量设置参与者学分
func (h *ParticipantHandler) BatchSetCredits(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	var req models.BatchCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    "指定的活动不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败",
				"data":    err.Error(),
			})
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者或管理员可以设置学分",
			"data":    nil,
		})
		return
	}

	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

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

		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
		}
		if userInfo, err := h.getUserInfo(participant.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}
		updatedParticipants = append(updatedParticipants, response)
		updatedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "批量设置学分成功",
		"data": gin.H{
			"updated_count": updatedCount,
			"participants":  updatedParticipants,
		},
	})
}

// SetSingleCredits 设置单个参与者学分
func (h *ParticipantHandler) SetSingleCredits(c *gin.Context) {
	activityID := c.Param("id")
	targetUserID := c.Param("user_id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	var req models.SingleCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    "指定的活动不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败",
				"data":    err.Error(),
			})
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者或管理员可以设置学分",
			"data":    nil,
		})
		return
	}

	var participant models.ActivityParticipant
	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, targetUserID).First(&participant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "参与者不存在",
			"data":    "指定的参与者不存在",
		})
		return
	}

	participant.Credits = req.Credits
	if err := h.db.Save(&participant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "设置学分失败",
			"data":    err.Error(),
		})
		return
	}

	response := models.ParticipantResponse{
		UserID:   participant.UserID,
		Credits:  participant.Credits,
		JoinedAt: participant.JoinedAt,
	}
	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}
	if userInfo, err := h.getUserInfo(participant.UserID, authToken); err == nil {
		response.UserInfo = userInfo
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "学分设置成功",
		"data":    response,
	})
}

// RemoveParticipant 删除参与者
func (h *ParticipantHandler) RemoveParticipant(c *gin.Context) {
	activityID := c.Param("id")
	targetUserID := c.Param("user_id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    "指定的活动不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败",
				"data":    err.Error(),
			})
		}
		return
	}

	// 只有活动创建者或管理员可以删除参与者
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者或管理员可以删除参与者",
			"data":    nil,
		})
		return
	}

	var participant models.ActivityParticipant
	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, targetUserID).First(&participant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "参与者不存在",
			"data":    "指定的参与者不存在",
		})
		return
	}

	if err := h.db.Delete(&participant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除参与者失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "参与者删除成功",
		"data": gin.H{
			"user_id":    targetUserID,
			"removed_at": time.Now(),
		},
	})
}

// LeaveActivity 退出活动
func (h *ParticipantHandler) LeaveActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")

	var participant models.ActivityParticipant
	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&participant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "参与者不存在",
			"data":    "指定的参与者不存在",
		})
		return
	}

	if err := h.db.Delete(&participant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "退出活动失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "退出活动成功",
		"data": gin.H{
			"user_id": userID,
			"left_at": time.Now(),
		},
	})
}

// GetActivityParticipants 获取活动参与者列表
func (h *ParticipantHandler) GetActivityParticipants(c *gin.Context) {
	activityID := c.Param("id")

	var participants []models.ActivityParticipant
	if err := h.db.Where("activity_id = ?", activityID).Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取参与者列表失败",
			"data":    err.Error(),
		})
		return
	}

	responses := make([]models.ParticipantResponse, 0, len(participants))
	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}
	for _, participant := range participants {
		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
		}
		if userInfo, err := h.getUserInfo(participant.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"participants": responses,
			"total":        len(responses),
		},
	})
}

// isStudent 检查用户是否为学生（使用真实用户服务）
func (h *ParticipantHandler) isStudent(userID string, authToken string) bool {
	return utils.IsStudent(userID, authToken)
}

// getUserInfo 获取用户信息（使用真实用户服务）
func (h *ParticipantHandler) getUserInfo(userID string, authToken string) (*models.UserInfo, error) {
	return utils.GetUserInfo(userID, authToken)
}
