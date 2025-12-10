package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"credit-management/credit-activity-service/models"

	"gorm.io/gorm"
)

func (h *ActivityHandler) handleStatusSideEffects(tx *gorm.DB, previousStatus, newStatus, activityID string) error {
	// 当活动第一次被审核通过时（从非 approved -> approved），为所有参与者生成申请记录
	if previousStatus != models.StatusApproved && newStatus == models.StatusApproved {
		return h.generateApplicationsForParticipants(tx, activityID)
	}

	// 当活动从已通过变为其他任何状态（如被拒绝、退回草稿等）时，撤销之前为该活动生成的申请
	// 这样可以避免活动已被拒绝但申请仍显示“已通过”的不一致问题
	if previousStatus == models.StatusApproved && newStatus != models.StatusApproved {
		return h.softDeleteApplications(tx, activityID)
	}

	// 其他状态流转暂时没有额外副作用
	return nil
}

func (h *ActivityHandler) generateApplicationsForParticipants(tx *gorm.DB, activityID string) error {
	var participants []models.ActivityParticipant
	if err := tx.Where("activity_id = ? AND deleted_at IS NULL", activityID).Find(&participants).Error; err != nil {
		return err
	}

	for _, participant := range participants {
		// 检查是否存在申请记录（包括软删除的）
		var existingApp models.Application
		err := tx.Unscoped().Where("activity_id = ? AND user_id = ?", activityID, participant.UUID).First(&existingApp).Error
		
		if err == nil {
			// 申请记录已存在
			if existingApp.DeletedAt.Valid {
				// 如果是软删除的记录，恢复它并更新字段
				// 使用 Unscoped().Update 来更新软删除的记录，设置 deleted_at 为 NULL
				if err := tx.Unscoped().Model(&models.Application{}).
					Where("id = ?", existingApp.ID).
					Updates(map[string]interface{}{
						"deleted_at":      nil,
						"status":          models.StatusApproved,
						"applied_credits": participant.Credits,
						"awarded_credits": participant.Credits,
					}).Error; err != nil {
					return err
				}
			}
			// 如果记录已存在且未删除，跳过
			continue
		} else if err != gorm.ErrRecordNotFound {
			// 其他错误
			return err
		}

		// 记录不存在，创建新的申请
		app := models.Application{
			ActivityID:     activityID,
			UUID:           participant.UUID,
			Status:         models.StatusApproved,
			AppliedCredits: participant.Credits,
			AwardedCredits: participant.Credits,
		}
		if err := tx.Create(&app).Error; err != nil {
			return err
		}
	}

	return nil
}

func (h *ActivityHandler) softDeleteApplications(tx *gorm.DB, activityID string) error {
	return tx.Where("activity_id = ?", activityID).Delete(&models.Application{}).Error
}

func (h *ActivityHandler) softDeleteActivityRelations(tx *gorm.DB, activityID string) ([]models.Attachment, error) {
	var attachments []models.Attachment
	if err := tx.Where("activity_id = ? AND deleted_at IS NULL", activityID).Find(&attachments).Error; err != nil {
		return nil, err
	}

	if err := tx.Where("activity_id = ?", activityID).Delete(&models.ActivityParticipant{}).Error; err != nil {
		return nil, err
	}
	if err := tx.Where("activity_id = ?", activityID).Delete(&models.Application{}).Error; err != nil {
		return nil, err
	}
	if err := tx.Where("activity_id = ?", activityID).Delete(&models.Attachment{}).Error; err != nil {
		return nil, err
	}

	return attachments, nil
}

func (h *ActivityHandler) cleanupAttachmentFiles(attachments []models.Attachment) {
	for _, attachment := range attachments {
		if attachment.MD5Hash == "" {
			continue
		}

		var otherCount int64
		if err := h.db.Model(&models.Attachment{}).
			Where("md5_hash = ? AND deleted_at IS NULL", attachment.MD5Hash).
			Count(&otherCount).Error; err != nil {
			continue
		}

		if otherCount > 0 {
			continue
		}

		filePath := filepath.Join("uploads/attachments", attachment.FileName)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("删除物理文件失败: %v\n", err)
		}
	}
}
