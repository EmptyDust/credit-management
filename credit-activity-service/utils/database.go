package utils

import (
	"credit-management/credit-activity-service/models"

	"gorm.io/gorm"
)

// BaseHandler 数据库操作基类
type BaseHandler struct {
	db *gorm.DB
}

// NewBaseHandler 创建新的基类处理器
func NewBaseHandler(db *gorm.DB) *BaseHandler {
	return &BaseHandler{db: db}
}

// GetActivityByID 根据ID获取活动
func (h *BaseHandler) GetActivityByID(id string) (*models.CreditActivity, error) {
	var activity models.CreditActivity
	err := h.db.Where("id = ?", id).First(&activity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// GetActivityByIDWithParticipants 根据ID获取活动（包含参与者）
func (h *BaseHandler) GetActivityByIDWithParticipants(id string) (*models.CreditActivity, error) {
	var activity models.CreditActivity
	err := h.db.Preload("Participants").Where("id = ?", id).First(&activity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// CheckActivityExists 检查活动是否存在
func (h *BaseHandler) CheckActivityExists(id string) error {
	var count int64
	err := h.db.Model(&models.CreditActivity{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CheckUserParticipant 检查用户是否为活动参与者
func (h *BaseHandler) CheckUserParticipant(activityID, userID string) error {
	var count int64
	err := h.db.Model(&models.ActivityParticipant{}).
		Where("activity_id = ? AND id = ?", activityID, userID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetUserParticipatedActivities 获取用户参与的活动
func (h *BaseHandler) GetUserParticipatedActivities(userID string, page, limit int) ([]models.CreditActivity, int64, error) {
	var activities []models.CreditActivity
	var total int64

	// 获取总数
	err := h.db.Model(&models.CreditActivity{}).
		Joins("JOIN activity_participants ON credit_activities.id = activity_participants.activity_id").
		Where("activity_participants.id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	err = h.db.Model(&models.CreditActivity{}).
		Joins("JOIN activity_participants ON credit_activities.id = activity_participants.activity_id").
		Where("activity_participants.id = ?", userID).
		Offset((page - 1) * limit).
		Limit(limit).
		Order("credit_activities.created_at DESC").
		Find(&activities).Error

	return activities, total, err
}

// GetPendingActivities 获取待审核活动
func (h *BaseHandler) GetPendingActivities(page, limit int) ([]models.CreditActivity, int64, error) {
	var activities []models.CreditActivity
	var total int64

	err := h.db.Model(&models.CreditActivity{}).
		Where("status = ?", models.StatusPendingReview).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = h.db.Where("status = ?", models.StatusPendingReview).
		Offset((page - 1) * limit).
		Limit(limit).
		Order("created_at DESC").
		Find(&activities).Error

	return activities, total, err
}

// GetDeletableActivities 获取可删除的活动
func (h *BaseHandler) GetDeletableActivities(userID string, page, limit int) ([]models.CreditActivity, int64, error) {
	var activities []models.CreditActivity
	var total int64

	// 只能删除草稿状态的活动
	err := h.db.Model(&models.CreditActivity{}).
		Where("owner_id = ? AND status = ?", userID, models.StatusDraft).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = h.db.Where("owner_id = ? AND status = ?", userID, models.StatusDraft).
		Offset((page - 1) * limit).
		Limit(limit).
		Order("created_at DESC").
		Find(&activities).Error

	return activities, total, err
}

// SearchActivities 搜索活动
func (h *BaseHandler) SearchActivities(query, status, category, ownerID string, userID, userType string, page, limit int) ([]models.CreditActivity, int64, error) {
	var activities []models.CreditActivity
	var total int64

	dbQuery := h.db.Model(&models.CreditActivity{})

	// 权限过滤
	if userType == "student" {
		dbQuery = dbQuery.Where("owner_id = ? OR id IN (SELECT activity_id FROM activity_participants WHERE id = ?)", userID, userID)
	}

	// 搜索条件
	if query != "" {
		searchQuery := "%" + query + "%"
		dbQuery = dbQuery.Where(
			"title ILIKE ? OR description ILIKE ? OR category ILIKE ?",
			searchQuery, searchQuery, searchQuery,
		)
	}
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}
	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}
	if ownerID != "" {
		dbQuery = dbQuery.Where("owner_id = ?", ownerID)
	}

	// 获取总数
	err := dbQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取数据
	err = dbQuery.Offset((page - 1) * limit).
		Limit(limit).
		Order("created_at DESC").
		Find(&activities).Error

	return activities, total, err
}

// GetActivityParticipants 获取活动参与者
func (h *BaseHandler) GetActivityParticipants(activityID string, page, limit int) ([]models.ActivityParticipant, int64, error) {
	var participants []models.ActivityParticipant
	var total int64

	err := h.db.Model(&models.ActivityParticipant{}).
		Where("activity_id = ?", activityID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = h.db.Where("activity_id = ?", activityID).
		Offset((page - 1) * limit).
		Limit(limit).
		Order("joined_at DESC").
		Find(&participants).Error

	return participants, total, err
}

// GetApplications 获取申请列表
func (h *BaseHandler) GetApplications(userID, userType string, page, limit int) ([]models.Application, int64, error) {
	var applications []models.Application
	var total int64

	dbQuery := h.db.Model(&models.Application{})

	// 权限过滤
	if userType == "student" {
		dbQuery = dbQuery.Where("id = ?", userID)
	}

	err := dbQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = dbQuery.Offset((page - 1) * limit).
		Limit(limit).
		Order("created_at DESC").
		Find(&applications).Error

	return applications, total, err
}
