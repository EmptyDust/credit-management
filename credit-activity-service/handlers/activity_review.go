package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/credit-activity-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	if activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限提交此活动",
			"data":    nil,
		})
		return
	}

	if activity.Status != models.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能提交草稿状态的活动",
			"data":    nil,
		})
		return
	}

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

	if activity.Status != models.StatusPendingReview &&
		activity.Status != models.StatusApproved &&
		activity.Status != models.StatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能审核待审核、已通过或已拒绝状态的活动",
			"data":    nil,
		})
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

func (h *ActivityHandler) GetPendingActivities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var activities []models.CreditActivity
	var total int64

	query := h.db.Where("status = ?", models.StatusPendingReview)

	query.Model(&models.CreditActivity{}).Count(&total)

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

func (h *ActivityHandler) WithdrawActivity(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", id).First(&activity).Error; err != nil {
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

	if activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者可以撤回活动",
			"data":    nil,
		})
		return
	}

	if activity.Status == models.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "草稿状态的活动无需撤回",
			"data":    nil,
		})
		return
	}

	activity.Status = models.StatusDraft
	activity.ReviewerID = nil
	activity.ReviewComments = ""
	activity.ReviewedAt = nil

	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "撤回活动失败",
			"data":    err.Error(),
		})
		return
	}

	now := time.Now()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "活动撤回成功",
		"data": gin.H{
			"id":           activity.ID,
			"status":       activity.Status,
			"withdrawn_at": now,
		},
	})
}
