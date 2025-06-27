package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ParticipantHandler struct {
	db *gorm.DB
}

func NewParticipantHandler(db *gorm.DB) *ParticipantHandler {
	return &ParticipantHandler{db: db}
}

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

	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者或管理员可以添加参与者",
			"data":    nil,
		})
		return
	}

	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

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
		var existing models.ActivityParticipant
		err := h.db.Where("activity_id = ? AND user_id = ?", activityID, targetUserID).First(&existing).Error
		if err == nil {
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

	var responses []models.ParticipantResponse
	for _, participant := range participants {
		userInfo, err := h.getUserInfo(participant.UserID, authToken)
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "参与者添加成功",
		"data": gin.H{
			"added_count":  addedCount,
			"participants": responses,
		},
	})
}

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

		userInfo, err := h.getUserInfo(participant.UserID, authToken)
		if err != nil {
			continue
		}

		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
			UserInfo: userInfo,
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

	userInfo, err := h.getUserInfo(participant.UserID, authToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "参与者关联的用户不存在",
			"data":    nil,
		})
		return
	}
	response.UserInfo = userInfo

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "学分设置成功",
		"data":    response,
	})
}

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
		userInfo, err := h.getUserInfo(participant.UserID, authToken)
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"participants": responses,
			"total":        len(responses),
		},
	})
}

func (h *ParticipantHandler) isStudent(userID string, authToken string) bool {
	return utils.IsStudent(userID, authToken)
}

func (h *ParticipantHandler) getUserInfo(userID string, authToken string) (*models.UserInfo, error) {
	return utils.GetUserInfo(userID, authToken)
}

func (h *ParticipantHandler) BatchRemoveParticipants(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	var req struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

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
				"message": "获取活动失败",
				"data":    err.Error(),
			})
		}
		return
	}

	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者或管理员可以批量删除参与者",
			"data":    nil,
		})
		return
	}

	if err := h.db.Where("activity_id = ? AND user_id IN ?", activityID, req.UserIDs).Delete(&models.ActivityParticipant{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量删除参与者失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "批量删除参与者成功",
		"data": gin.H{
			"removed_count": len(req.UserIDs),
			"removed_users": req.UserIDs,
		},
	})
}

func (h *ParticipantHandler) GetParticipantStats(c *gin.Context) {
	activityID := c.Param("id")

	var stats struct {
		TotalParticipants int64   `json:"total_participants"`
		TotalCredits      float64 `json:"total_credits"`
		AvgCredits        float64 `json:"avg_credits"`
		MaxCredits        float64 `json:"max_credits"`
		MinCredits        float64 `json:"min_credits"`
	}

	if err := h.db.Model(&models.ActivityParticipant{}).
		Where("activity_id = ?", activityID).
		Select("COUNT(*) as total_participants, SUM(credits) as total_credits, AVG(credits) as avg_credits, MAX(credits) as max_credits, MIN(credits) as min_credits").
		Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取统计信息失败",
			"data":    err.Error(),
		})
		return
	}

	var recentParticipants int64
	h.db.Model(&models.ActivityParticipant{}).
		Where("activity_id = ? AND joined_at >= ?", activityID, time.Now().AddDate(0, 0, -7)).
		Count(&recentParticipants)

	statsData := gin.H{
		"total_participants":  stats.TotalParticipants,
		"total_credits":       stats.TotalCredits,
		"avg_credits":         stats.AvgCredits,
		"max_credits":         stats.MaxCredits,
		"min_credits":         stats.MinCredits,
		"recent_participants": recentParticipants,
		"activity_id":         activityID,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取统计信息成功",
		"data":    statsData,
	})
}

func (h *ParticipantHandler) ExportParticipants(c *gin.Context) {
	activityID := c.Param("id")
	format := c.DefaultQuery("format", "json")

	var participants []models.ActivityParticipant
	if err := h.db.Where("activity_id = ?", activityID).Order("joined_at DESC").Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取参与者数据失败",
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
		userInfo, err := h.getUserInfo(participant.UserID, authToken)
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

	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "导出成功",
			"data": gin.H{
				"activity_id":  activityID,
				"participants": responses,
				"total_count":  len(responses),
				"export_time":  time.Now(),
			},
		})
	case "csv":
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "导出成功",
			"data": gin.H{
				"message":     "CSV导出功能待实现",
				"total_count": len(responses),
				"activity_id": activityID,
			},
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的导出格式",
			"data":    nil,
		})
	}
}

func (h *ParticipantHandler) GetUserParticipatedActivities(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var participants []models.ActivityParticipant
	if err := h.db.Where("user_id = ?", userID).Offset((page - 1) * pageSize).Limit(pageSize).Order("joined_at DESC").Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取参与活动列表失败",
			"data":    err.Error(),
		})
		return
	}

	var total int64
	h.db.Model(&models.ActivityParticipant{}).Where("user_id = ?", userID).Count(&total)

	var activities []gin.H
	for _, participant := range participants {
		var activity models.CreditActivity
		if err := h.db.Where("id = ?", participant.ActivityID).First(&activity).Error; err == nil {
			activities = append(activities, gin.H{
				"activity_id": activity.ID,
				"title":       activity.Title,
				"category":    activity.Category,
				"status":      activity.Status,
				"start_date":  activity.StartDate,
				"end_date":    activity.EndDate,
				"credits":     participant.Credits,
				"joined_at":   participant.JoinedAt,
			})
		}
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取成功",
		"data": gin.H{
			"activities":  activities,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}
