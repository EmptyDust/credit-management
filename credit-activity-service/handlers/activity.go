package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/credit-activity-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityHandler 活动处理器
type ActivityHandler struct {
	db *gorm.DB
}

// NewActivityHandler 创建活动处理器
func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{db: db}
}

// CreateActivity 创建活动
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req models.ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
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

	// 解析日期
	var startDate, endDate time.Time
	var err error
	if req.StartDate != "" {
		startDate, err = time.Parse("2006-01-02T15:04:05Z", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "开始日期格式错误",
				"data":    nil,
			})
			return
		}
	}
	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02T15:04:05Z", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "结束日期格式错误",
				"data":    nil,
			})
			return
		}
	}

	// 创建活动
	activity := models.CreditActivity{
		Title:        req.Title,
		Description:  req.Description,
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       models.StatusDraft,
		Category:     req.Category,
		Requirements: req.Requirements,
		OwnerID:      userID.(string),
	}

	if err := h.db.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    activity,
	})
}

// GetActivities 获取活动列表
func (h *ActivityHandler) GetActivities(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	// 获取查询参数
	status := c.Query("status")
	category := c.Query("category")
	ownerID := c.Query("owner_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	query := h.db.Model(&models.CreditActivity{})

	// 权限过滤：学生只能看到自己创建或参与的活动，教师可以看到所有活动
	if userType == "student" {
		// 学生只能看到自己创建的活动或参与的活动
		query = query.Where("owner_id = ? OR id IN (SELECT activity_id FROM activity_participants WHERE user_id = ?)", userID, userID)
	}

	// 应用筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}

	var activities []models.CreditActivity
	var total int64

	query.Count(&total)
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取活动列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 构建响应
	var responses []models.ActivityResponse
	for _, activity := range activities {
		response := h.enrichActivityResponse(activity)
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取活动列表成功",
		"data": gin.H{
			"activities":  responses,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetActivity 获取活动详情
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
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

	var activity models.CreditActivity
	if err := h.db.Preload("Participants").Where("id = ?", id).First(&activity).Error; err != nil {
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

	// 权限检查：学生只能查看自己创建或参与的活动
	if userType == "student" {
		if activity.OwnerID != userID {
			// 检查是否为参与者
			var participant models.ActivityParticipant
			if err := h.db.Where("activity_id = ? AND user_id = ?", id, userID).First(&participant).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "无权限查看此活动",
					"data":    nil,
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    activity,
	})
}

// UpdateActivity 更新活动
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", id).First(&activity).Error; err != nil {
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

	// 权限检查：只有活动创建者和管理员可以更新
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足",
			"error":   "只有活动创建者可以更新活动",
		})
		return
	}

	// 只有草稿状态的活动可以修改
	if activity.Status != models.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "操作失败",
			"error":   "只有草稿状态的活动可以修改",
		})
		return
	}

	var req models.ActivityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 更新活动信息
	if req.Title != "" {
		activity.Title = req.Title
	}
	if req.Description != "" {
		activity.Description = req.Description
	}
	if !req.StartDate.IsZero() {
		activity.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		activity.EndDate = req.EndDate
	}
	if req.Category != "" {
		activity.Category = req.Category
	}
	if req.Requirements != "" {
		activity.Requirements = req.Requirements
	}

	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新活动失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "活动更新成功",
		"data": gin.H{
			"id":           activity.ID,
			"title":        activity.Title,
			"description":  activity.Description,
			"start_date":   activity.StartDate,
			"end_date":     activity.EndDate,
			"status":       activity.Status,
			"category":     activity.Category,
			"requirements": activity.Requirements,
			"owner_id":     activity.OwnerID,
			"updated_at":   activity.UpdatedAt,
		},
	})
}

// DeleteActivity 删除活动
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
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

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", id).First(&activity).Error; err != nil {
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

	// 权限检查：只有活动创建者和管理员可以删除活动
	if userType != "admin" && activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限删除此活动",
			"data":    nil,
		})
		return
	}

	// 删除活动（会级联删除参与者和申请）
	if err := h.db.Delete(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "活动删除成功",
		"data":    nil,
	})
}

// SubmitActivity 提交活动审核
func (h *ActivityHandler) SubmitActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
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

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", id).First(&activity).Error; err != nil {
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

	// 权限检查：只有活动创建者可以提交审核
	if activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限提交此活动",
			"data":    nil,
		})
		return
	}

	// 只有草稿状态的活动可以提交审核
	if activity.Status != models.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能提交草稿状态的活动",
			"data":    nil,
		})
		return
	}

	// 更新状态为待审核
	if err := h.db.Model(&activity).Update("status", models.StatusPendingReview).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交审核失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "活动已提交审核",
		"data": gin.H{
			"id":     activity.ID,
			"status": models.StatusPendingReview,
		},
	})
}

// ReviewActivity 审核活动
func (h *ActivityHandler) ReviewActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
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

	var req models.ActivityReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", id).First(&activity).Error; err != nil {
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

	// 只有待审核状态的活动可以审核
	if activity.Status != models.StatusPendingReview {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能审核待审核状态的活动",
			"data":    nil,
		})
		return
	}

	// 更新审核信息
	now := time.Now()
	updates := map[string]interface{}{
		"status":          req.Status,
		"reviewer_id":     userID.(string),
		"review_comments": req.ReviewComments,
		"reviewed_at":     &now,
	}

	if err := h.db.Model(&activity).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "审核活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "审核完成",
		"data": gin.H{
			"id":              activity.ID,
			"status":          req.Status,
			"reviewer_id":     userID.(string),
			"review_comments": req.ReviewComments,
			"reviewed_at":     now,
		},
	})
}

// GetPendingActivities 获取待审核活动
func (h *ActivityHandler) GetPendingActivities(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var activities []models.CreditActivity
	var total int64

	query := h.db.Where("status = ?", models.StatusPendingReview)

	// 统计总数
	query.Model(&models.CreditActivity{}).Count(&total)

	// 获取分页数据
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取待审核活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	totalPages := (int(total) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.PaginatedResponse{
			Data:       activities,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// GetActivityStats 获取活动统计
func (h *ActivityHandler) GetActivityStats(c *gin.Context) {
	var stats models.ActivityStats

	// 统计各种状态的活动数量
	h.db.Model(&models.CreditActivity{}).Count(&stats.TotalActivities)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusDraft).Count(&stats.DraftCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusPendingReview).Count(&stats.PendingCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusApproved).Count(&stats.ApprovedCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusRejected).Count(&stats.RejectedCount)

	// 统计参与者总数
	h.db.Model(&models.ActivityParticipant{}).Count(&stats.TotalParticipants)

	// 统计总学分
	h.db.Model(&models.ActivityParticipant{}).Select("COALESCE(SUM(credits), 0)").Scan(&stats.TotalCredits)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetActivityCategories 获取活动类别
func (h *ActivityHandler) GetActivityCategories(c *gin.Context) {
	categories := models.GetActivityCategories()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"categories":  categories,
			"count":       len(categories),
			"description": "活动类别列表",
		},
	})
}

// WithdrawActivity 撤回活动
func (h *ActivityHandler) WithdrawActivity(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", id).First(&activity).Error; err != nil {
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

	// 权限检查：只有活动创建者可以撤回
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足",
			"error":   "只有活动创建者可以撤回活动",
		})
		return
	}

	// 检查活动状态：只有非草稿状态的活动可以撤回
	if activity.Status == models.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "操作失败",
			"error":   "草稿状态的活动无需撤回",
		})
		return
	}

	// 撤回活动到草稿状态
	activity.Status = models.StatusDraft
	activity.ReviewerID = ""
	activity.ReviewComments = ""
	activity.ReviewedAt = nil

	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "撤回活动失败",
			"error":   err.Error(),
		})
		return
	}

	now := time.Now()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "活动撤回成功",
		"data": gin.H{
			"id":           activity.ID,
			"status":       activity.Status,
			"withdrawn_at": now,
		},
	})
}

// enrichActivityResponse 丰富活动响应信息
func (h *ActivityHandler) enrichActivityResponse(activity models.CreditActivity) models.ActivityResponse {
	response := models.ActivityResponse{
		ID:             activity.ID,
		Title:          activity.Title,
		Description:    activity.Description,
		StartDate:      activity.StartDate,
		EndDate:        activity.EndDate,
		Status:         activity.Status,
		Category:       activity.Category,
		Requirements:   activity.Requirements,
		OwnerID:        activity.OwnerID,
		ReviewerID:     activity.ReviewerID,
		ReviewComments: activity.ReviewComments,
		ReviewedAt:     activity.ReviewedAt,
		CreatedAt:      activity.CreatedAt,
		UpdatedAt:      activity.UpdatedAt,
	}

	// 获取参与者信息
	var participants []models.ActivityParticipant
	h.db.Where("activity_id = ?", activity.ID).Find(&participants)

	var participantResponses []models.ParticipantResponse
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

		participantResponses = append(participantResponses, response)
	}
	response.Participants = participantResponses

	// 获取申请信息
	var applications []models.Application
	h.db.Where("activity_id = ?", activity.ID).Find(&applications)

	var applicationResponses []models.ApplicationResponse
	for _, application := range applications {
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
		}

		// 获取活动信息
		response.Activity = models.ActivityInfo{
			ID:          activity.ID,
			Title:       activity.Title,
			Description: activity.Description,
			Category:    activity.Category,
			StartDate:   activity.StartDate,
			EndDate:     activity.EndDate,
		}

		applicationResponses = append(applicationResponses, response)
	}
	response.Applications = applicationResponses

	return response
}

// getUserInfo 获取用户信息（模拟实现）
func (h *ActivityHandler) getUserInfo(userID string) (*models.UserInfo, error) {
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
