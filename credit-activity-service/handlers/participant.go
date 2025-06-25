package handlers

import (
	"net/http"
	"time"

	"credit-management/credit-activity-service/models"

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
			"success": false,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "活动不存在",
				"error":   "指定的活动不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取活动失败",
				"error":   err.Error(),
			})
		}
		return
	}

	// 权限检查：只有活动创建者和管理员可以添加参与者
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足",
			"error":   "只有活动创建者可以添加参与者",
		})
		return
	}

	// 验证用户角色：只有学生可以参与活动
	for _, userID := range req.UserIDs {
		// 这里应该调用用户服务验证用户角色
		// 暂时使用模拟验证，实际应该调用用户服务API
		if !h.isStudent(userID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "参与者限制",
				"error":   "只有学生可以参与活动",
			})
			return
		}
	}

	var participants []models.ActivityParticipant
	var addedCount int

	for _, userID := range req.UserIDs {
		// 检查是否已经是参与者
		var existing models.ActivityParticipant
		err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&existing).Error
		if err == nil {
			// 已经是参与者，跳过
			continue
		}

		participant := models.ActivityParticipant{
			ActivityID: activityID,
			UserID:     userID,
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
		if userInfo, err := h.getUserInfo(participant.UserID); err == nil {
			response.UserInfo = userInfo
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
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
			"success": false,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "活动不存在",
				"error":   "指定的活动不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取活动失败",
				"error":   err.Error(),
			})
		}
		return
	}

	// 权限检查：只有活动创建者和管理员可以设置学分
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足",
			"error":   "只有活动创建者可以设置学分",
		})
		return
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

		// 获取用户信息
		if userInfo, err := h.getUserInfo(participant.UserID); err == nil {
			response.UserInfo = userInfo
		}

		updatedParticipants = append(updatedParticipants, response)
		updatedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "学分设置成功",
		"data": gin.H{
			"updated_count": updatedCount,
			"participants":  updatedParticipants,
		},
	})
}

// SetSingleCredits 设置单个参与者学分
func (h *ParticipantHandler) SetSingleCredits(c *gin.Context) {
	activityID := c.Param("id")
	participantUserID := c.Param("user_id")
	if activityID == "" || participantUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID和用户ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	var req models.SingleCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 验证活动是否存在
	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：只有活动创建者和管理员可以设置学分
	if userType != "admin" && activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限设置学分",
			"data":    nil,
		})
		return
	}

	// 只有草稿状态的活动可以设置学分
	if activity.Status != models.StatusDraft && userType != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能为草稿状态的活动设置学分",
			"data":    nil,
		})
		return
	}

	// 更新参与者学分
	if err := h.db.Model(&models.ActivityParticipant{}).
		Where("activity_id = ? AND user_id = ?", activityID, participantUserID).
		Update("credits", req.Credits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "设置学分失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "学分设置成功",
		"data": gin.H{
			"user_id": participantUserID,
			"credits": req.Credits,
		},
	})
}

// RemoveParticipant 删除参与者
func (h *ParticipantHandler) RemoveParticipant(c *gin.Context) {
	activityID := c.Param("id")
	participantUserID := c.Param("user_id")
	if activityID == "" || participantUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID和用户ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 验证活动是否存在
	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：只有活动创建者和管理员可以删除参与者
	if userType != "admin" && activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限删除参与者",
			"data":    nil,
		})
		return
	}

	// 只有草稿状态的活动可以删除参与者
	if activity.Status != models.StatusDraft && userType != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能为草稿状态的活动删除参与者",
			"data":    nil,
		})
		return
	}

	// 删除参与者
	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, participantUserID).Delete(&models.ActivityParticipant{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除参与者失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "参与者删除成功",
		"data":    nil,
	})
}

// LeaveActivity 退出活动
func (h *ParticipantHandler) LeaveActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	// 只有学生可以退出活动
	if userType != "student" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足",
			"error":   "只有学生可以退出活动",
		})
		return
	}

	// 检查是否为活动参与者
	var participant models.ActivityParticipant
	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "参与者不存在",
				"error":   "您不是该活动的参与者",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取参与者信息失败",
				"error":   err.Error(),
			})
		}
		return
	}

	// 删除参与者
	if err := h.db.Delete(&participant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "退出活动失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "退出活动成功",
	})
}

// GetActivityParticipants 获取活动参与者列表
func (h *ParticipantHandler) GetActivityParticipants(c *gin.Context) {
	activityID := c.Param("id")

	var participants []models.ActivityParticipant
	if err := h.db.Where("activity_id = ?", activityID).Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取参与者列表失败",
			"error":   err.Error(),
		})
		return
	}

	var responses []models.ParticipantResponse
	for _, participant := range participants {
		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
		}

		// 获取用户信息
		if userInfo, err := h.getUserInfo(participant.UserID); err == nil {
			response.UserInfo = userInfo
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取参与者列表成功",
		"data": gin.H{
			"participants": responses,
			"total":        len(responses),
		},
	})
}

// isStudent 检查用户是否为学生（模拟实现）
func (h *ParticipantHandler) isStudent(userID string) bool {
	// 这里应该调用用户服务验证用户角色
	// 暂时返回true，实际应该调用用户服务API
	return true
}

// getUserInfo 获取用户信息（模拟实现）
func (h *ParticipantHandler) getUserInfo(userID string) (*models.UserInfo, error) {
	// 这里应该调用用户服务获取用户信息
	// 暂时返回模拟数据
	return &models.UserInfo{
		ID:        userID,
		Username:  "user",
		Name:      "用户",
		Role:      "student",
		StudentID: "2021001",
	}, nil
}
