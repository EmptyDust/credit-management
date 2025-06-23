package handlers

import (
	"net/http"
	"strconv"
	"time"

	"credit-management/user-management-service/models"
	"credit-management/user-management-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db                  *gorm.DB
	notificationManager *utils.NotificationManager
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	notificationManager := utils.NewNotificationManager(db)

	return &NotificationHandler{
		db:                  db,
		notificationManager: notificationManager,
	}
}

// GetUserNotifications 获取用户通知列表
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	isReadStr := c.Query("is_read")

	var isRead *bool
	if isReadStr != "" {
		if isReadStr == "true" {
			read := true
			isRead = &read
		} else if isReadStr == "false" {
			read := false
			isRead = &read
		}
	}

	notifications, total, err := h.notificationManager.GetUserNotifications(userID.(uint), page, pageSize, isRead)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"total_pages":   (int(total) + pageSize - 1) / pageSize,
	})
}

// GetNotification 获取通知详情
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	var notification models.Notification
	if err := h.db.Where("id = ? AND user_id = ?", notificationID, userID.(uint)).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询通知失败: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, notification)
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	if err := h.notificationManager.MarkAsRead(uint(notificationID), userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记已读失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "标记已读成功"})
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	if err := h.notificationManager.MarkAllAsRead(userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记已读失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有通知已标记为已读"})
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	if err := h.notificationManager.DeleteNotification(uint(notificationID), userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知删除成功"})
}

// GetUnreadCount 获取未读通知数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	count, err := h.notificationManager.GetUnreadCount(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取未读数量失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// GetNotificationStats 获取通知统计信息
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	stats, err := h.notificationManager.GetNotificationStats(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateNotification 创建通知（管理员功能）
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	notification, err := h.notificationManager.CreateNotification(req.UserID, req.Title, req.Content, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// SendSystemNotification 发送系统通知（管理员功能）
func (h *NotificationHandler) SendSystemNotification(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		UserIDs []uint `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if err := h.notificationManager.SendSystemNotification(req.Title, req.Content, req.UserIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送系统通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "系统通知发送成功"})
}

// SendBatchNotification 批量发送通知（管理员功能）
func (h *NotificationHandler) SendBatchNotification(c *gin.Context) {
	var req []models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if err := h.notificationManager.SendBatchNotification(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量发送通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "批量通知发送成功"})
}

// GetAllNotifications 获取所有通知（管理员功能）
func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID := c.Query("user_id")
	notificationType := c.Query("type")
	isRead := c.Query("is_read")

	query := h.db.Model(&models.Notification{}).Preload("User")

	if userID != "" {
		if userIDUint, err := strconv.ParseUint(userID, 10, 32); err == nil {
			query = query.Where("user_id = ?", userIDUint)
		}
	}

	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	if isRead != "" {
		if isRead == "true" {
			query = query.Where("is_read = ?", true)
		} else if isRead == "false" {
			query = query.Where("is_read = ?", false)
		}
	}

	var total int64
	query.Count(&total)

	var notifications []models.Notification
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifications).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"total_pages":   (int(total) + pageSize - 1) / pageSize,
	})
}

// DeleteNotificationByAdmin 管理员删除通知
func (h *NotificationHandler) DeleteNotificationByAdmin(c *gin.Context) {
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	var notification models.Notification
	if err := h.db.First(&notification, notificationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询通知失败: " + err.Error()})
		}
		return
	}

	if err := h.db.Delete(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知删除成功"})
}

// GetSystemNotificationStats 获取系统通知统计（管理员功能）
func (h *NotificationHandler) GetSystemNotificationStats(c *gin.Context) {
	var stats struct {
		TotalNotifications int64            `json:"total_notifications"`
		UnreadCount        int64            `json:"unread_count"`
		ReadCount          int64            `json:"read_count"`
		TodayCount         int64            `json:"today_count"`
		WeekCount          int64            `json:"week_count"`
		MonthCount         int64            `json:"month_count"`
		TypeStats          map[string]int64 `json:"type_stats"`
	}

	// 总通知数
	h.db.Model(&models.Notification{}).Count(&stats.TotalNotifications)

	// 未读数
	h.db.Model(&models.Notification{}).Where("is_read = ?", false).Count(&stats.UnreadCount)

	// 已读数
	stats.ReadCount = stats.TotalNotifications - stats.UnreadCount

	// 今日通知数
	today := time.Now().Truncate(24 * time.Hour)
	h.db.Model(&models.Notification{}).Where("created_at >= ?", today).Count(&stats.TodayCount)

	// 本周通知数
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	h.db.Model(&models.Notification{}).Where("created_at >= ?", weekStart).Count(&stats.WeekCount)

	// 本月通知数
	monthStart := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	monthStart = monthStart.Truncate(24 * time.Hour)
	h.db.Model(&models.Notification{}).Where("created_at >= ?", monthStart).Count(&stats.MonthCount)

	// 各类型通知统计
	stats.TypeStats = make(map[string]int64)
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	h.db.Model(&models.Notification{}).Select("type, count(*) as count").Group("type").Scan(&typeStats)

	for _, ts := range typeStats {
		stats.TypeStats[ts.Type] = ts.Count
	}

	c.JSON(http.StatusOK, stats)
}
