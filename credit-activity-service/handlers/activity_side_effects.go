package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"credit-management/credit-activity-service/models"

	"gorm.io/gorm"
)

func (h *ActivityHandler) handleStatusSideEffects(tx *gorm.DB, previousStatus, newStatus, activityID string) error {
	if previousStatus != models.StatusApproved && newStatus == models.StatusApproved {
		return h.generateApplicationsForParticipants(tx, activityID)
	}

	if previousStatus != models.StatusDraft && newStatus == models.StatusDraft {
		return h.softDeleteApplications(tx, activityID)
	}

	return nil
}

func (h *ActivityHandler) generateApplicationsForParticipants(tx *gorm.DB, activityID string) error {
	var participants []models.ActivityParticipant
	if err := tx.Where("activity_id = ? AND deleted_at IS NULL", activityID).Find(&participants).Error; err != nil {
		return err
	}

	for _, participant := range participants {
		var existing int64
		if err := tx.Model(&models.Application{}).
			Where("activity_id = ? AND user_id = ? AND deleted_at IS NULL", activityID, participant.UUID).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			continue
		}

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
