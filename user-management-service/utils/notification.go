package utils

import (
	"fmt"
	"strings"
	"time"

	"credit-management/user-management-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotificationManager 通知管理器
type NotificationManager struct {
	db *gorm.DB
}

// NewNotificationManager 创建通知管理器
func NewNotificationManager(db *gorm.DB) *NotificationManager {
	return &NotificationManager{db: db}
}

// CreateNotification 创建通知
func (nm *NotificationManager) CreateNotification(userID uint, title, content, notificationType string) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:  userID,
		Title:   title,
		Content: content,
		Type:    notificationType,
		IsRead:  false,
	}

	err := nm.db.Create(notification).Error
	return notification, err
}

// GetUserNotifications 获取用户通知
func (nm *NotificationManager) GetUserNotifications(userID uint, page, pageSize int, isRead *bool) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := nm.db.Where("user_id = ?", userID)

	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	// 获取总数
	err := query.Model(&models.Notification{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications).Error

	return notifications, total, err
}

// MarkAsRead 标记通知为已读
func (nm *NotificationManager) MarkAsRead(notificationID, userID uint) error {
	now := time.Now()
	return nm.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

// MarkAllAsRead 标记用户所有通知为已读
func (nm *NotificationManager) MarkAllAsRead(userID uint) error {
	now := time.Now()
	return nm.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

// DeleteNotification 删除通知
func (nm *NotificationManager) DeleteNotification(notificationID, userID uint) error {
	return nm.db.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{}).Error
}

// GetUnreadCount 获取未读通知数量
func (nm *NotificationManager) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := nm.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// SendSystemNotification 发送系统通知
func (nm *NotificationManager) SendSystemNotification(title, content string, userIDs []uint) error {
	notifications := make([]models.Notification, 0, len(userIDs))
	
	for _, userID := range userIDs {
		notification := models.Notification{
			UserID:  userID,
			Title:   title,
			Content: content,
			Type:    "system",
			IsRead:  false,
		}
		notifications = append(notifications, notification)
	}
	
	return nm.db.Create(&notifications).Error
}

// SendBatchNotification 批量发送通知
func (nm *NotificationManager) SendBatchNotification(notifications []models.NotificationRequest) error {
	notificationModels := make([]models.Notification, 0, len(notifications))

	for _, req := range notifications {
		notification := models.Notification{
			UserID:  req.UserID,
			Title:   req.Title,
			Content: req.Content,
			Type:    req.Type,
			IsRead:  false,
		}
		notificationModels = append(notificationModels, notification)
	}

	return nm.db.Create(&notificationModels).Error
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	Title   string                 `json:"title"`
	Content string                 `json:"content"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
}

// SendTemplateNotification 发送模板通知
func (nm *NotificationManager) SendTemplateNotification(userID uint, template NotificationTemplate, data map[string]interface{}) error {
	// 替换模板变量
	content := template.Content
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}
	
	notification := &models.Notification{
		UserID:  userID,
		Title:   template.Title,
		Content: content,
		Type:    template.Type,
		IsRead:  false,
	}
	
	return nm.db.Create(notification).Error
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalNotifications int64 `json:"total_notifications"`
	UnreadCount        int64 `json:"unread_count"`
	ReadCount          int64 `json:"read_count"`
	TodayCount         int64 `json:"today_count"`
	WeekCount          int64 `json:"week_count"`
	MonthCount         int64 `json:"month_count"`
}

// GetNotificationStats 获取通知统计
func (nm *NotificationManager) GetNotificationStats(userID uint) (*NotificationStats, error) {
	stats := &NotificationStats{}

	// 总通知数
	err := nm.db.Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalNotifications).Error
	if err != nil {
		return nil, err
	}

	// 未读数
	err = nm.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&stats.UnreadCount).Error
	if err != nil {
		return nil, err
	}

	// 已读数
	stats.ReadCount = stats.TotalNotifications - stats.UnreadCount

	// 今日通知数
	today := time.Now().Truncate(24 * time.Hour)
	err = nm.db.Model(&models.Notification{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Count(&stats.TodayCount).Error
	if err != nil {
		return nil, err
	}

	// 本周通知数
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	err = nm.db.Model(&models.Notification{}).
		Where("user_id = ? AND created_at >= ?", userID, weekStart).
		Count(&stats.WeekCount).Error
	if err != nil {
		return nil, err
	}

	// 本月通知数
	monthStart := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	monthStart = monthStart.Truncate(24 * time.Hour)
	err = nm.db.Model(&models.Notification{}).
		Where("user_id = ? AND created_at >= ?", userID, monthStart).
		Count(&stats.MonthCount).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// NotificationMiddleware 通知中间件
func NotificationMiddleware(nm *NotificationManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// 获取未读通知数量
		unreadCount, err := nm.GetUnreadCount(userID.(uint))
		if err == nil {
			c.Set("unread_notifications", unreadCount)
		}

		c.Next()
	}
}

// 预定义的通知模板
var (
	// 用户注册成功通知
	UserRegisteredTemplate = NotificationTemplate{
		Title:   "欢迎加入创新创业学分管理平台",
		Content: "恭喜您成功注册账号！您的用户名是：{{username}}。请及时完善个人信息并开始使用平台功能。",
		Type:    "welcome",
	}

	// 申请提交成功通知
	ApplicationSubmittedTemplate = NotificationTemplate{
		Title:   "学分申请已提交",
		Content: "您的{{affair_name}}申请已成功提交，申请编号：{{application_id}}。请耐心等待审核结果。",
		Type:    "application",
	}

	// 申请审核通过通知
	ApplicationApprovedTemplate = NotificationTemplate{
		Title:   "学分申请审核通过",
		Content: "恭喜！您的{{affair_name}}申请已通过审核，认定学分：{{credit}}分。",
		Type:    "success",
	}

	// 申请审核拒绝通知
	ApplicationRejectedTemplate = NotificationTemplate{
		Title:   "学分申请审核未通过",
		Content: "很抱歉，您的{{affair_name}}申请未通过审核。拒绝原因：{{reason}}。请根据反馈意见修改后重新提交。",
		Type:    "warning",
	}

	// 文件上传成功通知
	FileUploadedTemplate = NotificationTemplate{
		Title:   "文件上传成功",
		Content: "文件《{{filename}}》上传成功，文件大小：{{filesize}}。",
		Type:    "info",
	}

	// 系统维护通知
	SystemMaintenanceTemplate = NotificationTemplate{
		Title:   "系统维护通知",
		Content: "系统将于{{start_time}}进行维护，预计维护时间：{{duration}}。维护期间可能影响正常使用，请提前做好准备。",
		Type:    "system",
	}
)
