package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"credit-management/credit-activity-service/models"

	"github.com/gin-gonic/gin"
)

// BatchDeleteActivities 批量删除活动
func (h *ActivityHandler) BatchDeleteActivities(c *gin.Context) {
	var req struct {
		ActivityIDs []string `json:"activity_ids" binding:"required"`
	}

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

	userType, _ := c.Get("user_type")

	// 使用存储过程批量删除活动
	var deletedCount int
	err := h.db.Raw("SELECT batch_delete_activities(?, ?, ?)", req.ActivityIDs, userID, userType).Scan(&deletedCount).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量删除活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 删除活动相关的物理文件
	for _, activityID := range req.ActivityIDs {
		var attachments []models.Attachment
		if err := h.db.Where("activity_id = ? AND deleted_at IS NOT NULL", activityID).Find(&attachments).Error; err == nil {
			for _, attachment := range attachments {
				// 检查是否有其他活动使用相同的文件
				var otherAttachmentsCount int64
				h.db.Model(&models.Attachment{}).
					Where("md5_hash = ? AND activity_id != ? AND deleted_at IS NULL", attachment.MD5Hash, activityID).
					Count(&otherAttachmentsCount)

				// 如果没有其他活动使用该文件，则删除物理文件
				if otherAttachmentsCount == 0 {
					filePath := filepath.Join("uploads/attachments", attachment.FileName)
					if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
						fmt.Printf("删除物理文件失败: %v\n", err)
					} else {
						fmt.Printf("彻底删除物理文件: %s\n", filePath)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "批量删除活动成功",
		"data": gin.H{
			"deleted_count": deletedCount,
			"total_count":   len(req.ActivityIDs),
			"deleted_at":    time.Now(),
		},
	})
}

// GetDeletableActivities 获取用户可删除的活动列表
func (h *ActivityHandler) GetDeletableActivities(c *gin.Context) {
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

	// 使用存储过程获取可删除的活动列表
	var activities []struct {
		ActivityID  string    `json:"activity_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		Category    string    `json:"category"`
		OwnerID     string    `json:"owner_id"`
		CreatedAt   time.Time `json:"created_at"`
		CanDelete   bool      `json:"can_delete"`
	}

	err := h.db.Raw("SELECT * FROM get_user_deletable_activities(?, ?)", userID, userType).Scan(&activities).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取可删除活动列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取可删除活动列表成功",
		"data": gin.H{
			"activities": activities,
			"total":      len(activities),
		},
	})
}

// BatchCreateActivities 批量创建活动
func (h *ActivityHandler) BatchCreateActivities(c *gin.Context) {
	var req models.BatchActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
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

	if len(req.Activities) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "批量创建活动数量不能超过10个",
			"data":    nil,
		})
		return
	}

	var createdActivities []models.ActivityCreateResponse
	var errors []string

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, activityReq := range req.Activities {
		if err := h.validateActivityRequest(activityReq); err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动: %s", i+1, err.Error()))
			continue
		}
		startDate, endDate, err := h.parseActivityDates(activityReq.StartDate, activityReq.EndDate)
		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动: %s", i+1, err.Error()))
			continue
		}
		activity := models.CreditActivity{
			Title:        activityReq.Title,
			Description:  activityReq.Description,
			StartDate:    startDate,
			EndDate:      endDate,
			Status:       models.StatusDraft,
			Category:     activityReq.Category,
			Requirements: activityReq.Requirements,
			OwnerID:      userID.(string),
		}
		if err := tx.Create(&activity).Error; err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动创建失败: %s", i+1, err.Error()))
			continue
		}
		// 创建详情表
		switch activityReq.Category {
		case "创新创业实践活动":
			if activityReq.InnovationDetail != nil {
				detail := activityReq.InnovationDetail
				detail.ActivityID = activity.ID
				tx.Create(detail)
			}
		case "学科竞赛":
			if activityReq.CompetitionDetail != nil {
				detail := activityReq.CompetitionDetail
				detail.ActivityID = activity.ID
				tx.Create(detail)
			}
		case "大学生创业项目":
			if activityReq.EntrepreneurshipProjectDetail != nil {
				detail := activityReq.EntrepreneurshipProjectDetail
				detail.ActivityID = activity.ID
				tx.Create(detail)
			}
		case "创业实践项目":
			if activityReq.EntrepreneurshipPracticeDetail != nil {
				detail := activityReq.EntrepreneurshipPracticeDetail
				detail.ActivityID = activity.ID
				tx.Create(detail)
			}
		case "论文专利":
			if activityReq.PaperPatentDetail != nil {
				detail := activityReq.PaperPatentDetail
				detail.ActivityID = activity.ID
				tx.Create(detail)
			}
		}
		response := models.ActivityCreateResponse{
			ID:           activity.ID,
			Title:        activity.Title,
			Description:  activity.Description,
			StartDate:    activity.StartDate,
			EndDate:      activity.EndDate,
			Status:       activity.Status,
			Category:     activity.Category,
			Requirements: activity.Requirements,
			OwnerID:      activity.OwnerID,
			CreatedAt:    activity.CreatedAt,
			UpdatedAt:    activity.UpdatedAt,
		}
		createdActivities = append(createdActivities, response)
	}

	if len(errors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "批量创建活动失败",
			"data": gin.H{
				"errors":             errors,
				"created_count":      0,
				"total_count":        len(req.Activities),
				"created_activities": []models.ActivityCreateResponse{},
			},
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "批量创建活动成功",
		"data": gin.H{
			"created_count":      len(createdActivities),
			"total_count":        len(req.Activities),
			"created_activities": createdActivities,
		},
	})
}

// BatchUpdateActivities 新增BatchUpdateActivities，支持主表和详情表的批量更新
func (h *ActivityHandler) BatchUpdateActivities(c *gin.Context) {
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

	var req struct {
		Updates []struct {
			ID   string                       `json:"id" binding:"required"`
			Main models.ActivityUpdateRequest `json:"main" binding:"required"`
		} `json:"updates" binding:"required,min=1,max=20"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	var errors []string
	var updatedActivities []models.ActivityCreateResponse
	tx := h.db.Begin()

	for i, upd := range req.Updates {
		var activity models.CreditActivity
		if err := tx.Where("id = ?", upd.ID).First(&activity).Error; err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动不存在", i+1))
			continue
		}

		// 权限检查：只有活动创建者和管理员可以更新活动
		if activity.OwnerID != userID && userType != "admin" {
			errors = append(errors, fmt.Sprintf("第%d个活动无权限更新", i+1))
			continue
		}

		// 状态检查：只有草稿状态的活动可以修改（管理员除外）
		if activity.Status != models.StatusDraft && userType != "admin" {
			errors = append(errors, fmt.Sprintf("第%d个活动状态不允许修改", i+1))
			continue
		}

		// 主表字段更新
		if upd.Main.Title != nil {
			activity.Title = *upd.Main.Title
		}
		if upd.Main.Description != nil {
			activity.Description = *upd.Main.Description
		}
		if upd.Main.StartDate != nil {
			if t, err := h.parseSingleDate(*upd.Main.StartDate); err == nil {
				activity.StartDate = t
			}
		}
		if upd.Main.EndDate != nil {
			if t, err := h.parseSingleDate(*upd.Main.EndDate); err == nil {
				activity.EndDate = t
			}
		}
		if upd.Main.Category != nil {
			activity.Category = *upd.Main.Category
		}
		if upd.Main.Requirements != nil {
			activity.Requirements = *upd.Main.Requirements
		}
		if err := tx.Save(&activity).Error; err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动主表更新失败", i+1))
			continue
		}

		// 记录成功更新的活动
		updatedActivities = append(updatedActivities, models.ActivityCreateResponse{
			ID:        activity.ID,
			Title:     activity.Title,
			Status:    activity.Status,
			CreatedAt: activity.CreatedAt,
		})

		// 详情表字段更新
		switch activity.Category {
		case "创新创业实践活动":
			if upd.Main.InnovationDetail != nil {
				var detail models.InnovationActivityDetail
				tx.Where("activity_id = ?", activity.ID).First(&detail)
				if detail.ID != "" {
					tx.Model(&detail).Updates(upd.Main.InnovationDetail)
				} else {
					detail = *upd.Main.InnovationDetail
					detail.ActivityID = activity.ID
					tx.Create(&detail)
				}
			}
		case "学科竞赛":
			if upd.Main.CompetitionDetail != nil {
				var detail models.CompetitionActivityDetail
				tx.Where("activity_id = ?", activity.ID).First(&detail)
				if detail.ID != "" {
					tx.Model(&detail).Updates(upd.Main.CompetitionDetail)
				} else {
					detail = *upd.Main.CompetitionDetail
					detail.ActivityID = activity.ID
					tx.Create(&detail)
				}
			}
		case "大学生创业项目":
			if upd.Main.EntrepreneurshipProjectDetail != nil {
				var detail models.EntrepreneurshipProjectDetail
				tx.Where("activity_id = ?", activity.ID).First(&detail)
				if detail.ID != "" {
					tx.Model(&detail).Updates(upd.Main.EntrepreneurshipProjectDetail)
				} else {
					detail = *upd.Main.EntrepreneurshipProjectDetail
					detail.ActivityID = activity.ID
					tx.Create(&detail)
				}
			}
		case "创业实践项目":
			if upd.Main.EntrepreneurshipPracticeDetail != nil {
				var detail models.EntrepreneurshipPracticeDetail
				tx.Where("activity_id = ?", activity.ID).First(&detail)
				if detail.ID != "" {
					tx.Model(&detail).Updates(upd.Main.EntrepreneurshipPracticeDetail)
				} else {
					detail = *upd.Main.EntrepreneurshipPracticeDetail
					detail.ActivityID = activity.ID
					tx.Create(&detail)
				}
			}
		case "论文专利":
			if upd.Main.PaperPatentDetail != nil {
				var detail models.PaperPatentDetail
				tx.Where("activity_id = ?", activity.ID).First(&detail)
				if detail.ID != "" {
					tx.Model(&detail).Updates(upd.Main.PaperPatentDetail)
				} else {
					detail = *upd.Main.PaperPatentDetail
					detail.ActivityID = activity.ID
					tx.Create(&detail)
				}
			}
		}
	}

	if len(errors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "批量更新部分失败",
			"data": gin.H{
				"updated_count":      len(updatedActivities),
				"total_count":        len(req.Updates),
				"updated_activities": updatedActivities,
				"errors":             errors,
			},
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "批量更新成功",
		"data": gin.H{
			"updated_count":      len(updatedActivities),
			"total_count":        len(req.Updates),
			"updated_activities": updatedActivities,
			"errors":             []string{},
		},
	})
}
