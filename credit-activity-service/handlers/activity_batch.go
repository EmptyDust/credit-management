package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
)

func (h *ActivityHandler) BatchDeleteActivities(c *gin.Context) {
	var req struct {
		ActivityIDs []string `json:"activity_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	userType, _ := c.Get("user_type")

	var deletedCount int
	err := h.db.Raw("SELECT batch_delete_activities(?, ?, ?)", req.ActivityIDs, userID, userType).Scan(&deletedCount).Error

	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	for _, activityID := range req.ActivityIDs {
		var attachments []models.Attachment
		if err := h.db.Where("activity_id = ? AND deleted_at IS NOT NULL", activityID).Find(&attachments).Error; err == nil {
			for _, attachment := range attachments {
				var otherAttachmentsCount int64
				h.db.Model(&models.Attachment{}).
					Where("md5_hash = ? AND activity_id != ? AND deleted_at IS NULL", attachment.MD5Hash, activityID).
					Count(&otherAttachmentsCount)

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

	utils.SendSuccessResponse(c, gin.H{
		"message":       "批量删除活动成功",
		"deleted_count": deletedCount,
		"total_count":   len(req.ActivityIDs),
		"deleted_at":    time.Now(),
	})
}

func (h *ActivityHandler) BatchCreateActivities(c *gin.Context) {
	var req models.BatchActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	if len(req.Activities) > 10 {
		utils.SendBadRequest(c, "批量创建活动数量不能超过10个")
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
			Title:       activityReq.Title,
			Description: activityReq.Description,
			StartDate:   startDate,
			EndDate:     endDate,
			Status:      models.StatusDraft,
			Category:    activityReq.Category,
			OwnerID:     userID.(string),
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
			ID:          activity.ID,
			Title:       activity.Title,
			Description: activity.Description,
			StartDate:   activity.StartDate,
			EndDate:     activity.EndDate,
			Status:      activity.Status,
			Category:    activity.Category,
			OwnerID:     activity.OwnerID,
			CreatedAt:   activity.CreatedAt,
			UpdatedAt:   activity.UpdatedAt,
		}
		createdActivities = append(createdActivities, response)
	}

	if len(errors) > 0 {
		tx.Rollback()
		utils.SendBadRequest(c, "批量创建活动失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"message":            "批量创建活动成功",
		"created_count":      len(createdActivities),
		"total_count":        len(req.Activities),
		"created_activities": createdActivities,
	})
}

func (h *ActivityHandler) BatchUpdateActivities(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
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
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
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

		if activity.OwnerID != userID && userType != "admin" {
			errors = append(errors, fmt.Sprintf("第%d个活动无权限更新", i+1))
			continue
		}

		if activity.Status != models.StatusDraft && userType != "admin" {
			errors = append(errors, fmt.Sprintf("第%d个活动状态不允许修改", i+1))
			continue
		}

		if upd.Main.Title != nil {
			activity.Title = *upd.Main.Title
		}
		if upd.Main.Description != nil {
			activity.Description = *upd.Main.Description
		}

		var newStartDate, newEndDate time.Time
		var err error

		if upd.Main.StartDate != nil {
			newStartDate, err = h.parseSingleDate(*upd.Main.StartDate)
			if err != nil {
				errors = append(errors, fmt.Sprintf("第%d个活动开始日期格式错误: %s", i+1, err.Error()))
				continue
			}
		}

		if upd.Main.EndDate != nil {
			newEndDate, err = h.parseSingleDate(*upd.Main.EndDate)
			if err != nil {
				errors = append(errors, fmt.Sprintf("第%d个活动结束日期格式错误: %s", i+1, err.Error()))
				continue
			}
		}

		var compareStartDate, compareEndDate time.Time

		if upd.Main.StartDate != nil {
			compareStartDate = newStartDate
		} else {
			compareStartDate = activity.StartDate
		}

		if upd.Main.EndDate != nil {
			compareEndDate = newEndDate
		} else {
			compareEndDate = activity.EndDate
		}

		if !compareStartDate.IsZero() && !compareEndDate.IsZero() && compareStartDate.After(compareEndDate) {
			errors = append(errors, fmt.Sprintf("第%d个活动开始日期不能晚于结束日期", i+1))
			continue
		}

		// 更新日期字段
		if upd.Main.StartDate != nil {
			activity.StartDate = newStartDate
		}
		if upd.Main.EndDate != nil {
			activity.EndDate = newEndDate
		}

		if upd.Main.Category != nil {
			activity.Category = *upd.Main.Category
		}

		if err := tx.Save(&activity).Error; err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动主表更新失败", i+1))
			continue
		}

		updatedActivities = append(updatedActivities, models.ActivityCreateResponse{
			ID:        activity.ID,
			Title:     activity.Title,
			Status:    activity.Status,
			CreatedAt: activity.CreatedAt,
		})

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
		utils.SendBadRequestWithData(c, "批量更新部分失败", gin.H{
			"updated_count":      len(updatedActivities),
			"total_count":        len(req.Updates),
			"updated_activities": updatedActivities,
			"errors":             errors,
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"updated_count":      len(updatedActivities),
		"total_count":        len(req.Updates),
		"updated_activities": updatedActivities,
		"errors":             []string{},
	})
}
