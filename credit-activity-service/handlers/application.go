package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApplicationHandler 申请处理器
type ApplicationHandler struct {
	db *gorm.DB
}

// NewApplicationHandler 创建申请处理器
func NewApplicationHandler(db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{db: db}
}

// GetUserApplications 获取用户申请列表
func (h *ApplicationHandler) GetUserApplications(c *gin.Context) {
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

	// 获取认证令牌
	authToken := c.GetHeader("Authorization")

	// 获取查询参数
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var applications []models.Application
	var total int64

	query := h.db.Model(&models.Application{}).Where("user_id = ?", userID)

	// 应用筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	query.Count(&total)

	// 获取分页数据
	if err := query.Preload("Activity").Offset(offset).Limit(limit).Order("created_at DESC").Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取申请列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var responses []models.ApplicationResponse
	for _, app := range applications {
		response := models.ApplicationResponse{
			ID:             app.ID,
			ActivityID:     app.ActivityID,
			UserID:         app.UserID,
			Status:         app.Status,
			AppliedCredits: app.AppliedCredits,
			AwardedCredits: app.AwardedCredits,
			SubmittedAt:    app.SubmittedAt,
			CreatedAt:      app.CreatedAt,
			UpdatedAt:      app.UpdatedAt,
			Activity: models.ActivityInfo{
				ID:          app.Activity.ID,
				Title:       app.Activity.Title,
				Description: app.Activity.Description,
				Category:    app.Activity.Category,
				StartDate:   app.Activity.StartDate,
				EndDate:     app.Activity.EndDate,
			},
		}

		// 获取用户信息
		if userInfo, err := utils.GetUserInfo(app.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}

		responses = append(responses, response)
	}

	totalPages := (int(total) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.PaginatedResponse{
			Data:       responses,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// GetApplication 获取申请详情
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "申请ID不能为空",
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

	// 获取认证令牌
	authToken := c.GetHeader("Authorization")

	userType, _ := c.Get("user_type")

	var application models.Application
	if err := h.db.Preload("Activity").Where("id = ?", id).First(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "申请不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取申请失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：学生只能查看自己的申请
	if userType == "student" && application.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限查看此申请",
			"data":    nil,
		})
		return
	}

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
		Activity: models.ActivityInfo{
			ID:          application.Activity.ID,
			Title:       application.Activity.Title,
			Description: application.Activity.Description,
			Category:    application.Activity.Category,
			StartDate:   application.Activity.StartDate,
			EndDate:     application.Activity.EndDate,
		},
	}

	// 获取用户信息
	if userInfo, err := utils.GetUserInfo(application.UserID, authToken); err == nil {
		response.UserInfo = userInfo
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// GetAllApplications 获取所有申请（教师/管理员）
func (h *ApplicationHandler) GetAllApplications(c *gin.Context) {
	// 获取认证令牌
	authToken := c.GetHeader("Authorization")

	// 获取查询参数
	activityID := c.Query("activity_id")
	userID := c.Query("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var applications []models.Application
	var total int64

	query := h.db.Model(&models.Application{})

	// 应用筛选条件
	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 统计总数
	query.Count(&total)

	// 获取分页数据
	if err := query.Preload("Activity").Offset(offset).Limit(limit).Order("created_at DESC").Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取申请列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	var responses []models.ApplicationResponse
	for _, app := range applications {
		var userInfo *models.UserInfo
		userInfo, err := utils.GetUserInfo(app.UserID, authToken)
		if err != nil {
			userInfo = nil // 记录日志可选
		}
		response := models.ApplicationResponse{
			ID:             app.ID,
			ActivityID:     app.ActivityID,
			UserID:         app.UserID,
			Status:         app.Status,
			AppliedCredits: app.AppliedCredits,
			AwardedCredits: app.AwardedCredits,
			SubmittedAt:    app.SubmittedAt,
			CreatedAt:      app.CreatedAt,
			UpdatedAt:      app.UpdatedAt,
			Activity: models.ActivityInfo{
				ID:          app.Activity.ID,
				Title:       app.Activity.Title,
				Description: app.Activity.Description,
				Category:    app.Activity.Category,
				StartDate:   app.Activity.StartDate,
				EndDate:     app.Activity.EndDate,
			},
			UserInfo: userInfo,
		}
		responses = append(responses, response)
	}

	totalPages := (int(total) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.PaginatedResponse{
			Data:       responses,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// GetApplicationStats 获取申请统计
func (h *ApplicationHandler) GetApplicationStats(c *gin.Context) {
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

	var stats models.ApplicationStats

	// 统计申请数量和学分
	h.db.Model(&models.Application{}).Where("user_id = ?", userID).Count(&stats.TotalApplications)
	h.db.Model(&models.Application{}).Where("user_id = ?", userID).Select("COALESCE(SUM(applied_credits), 0)").Scan(&stats.TotalCredits)
	h.db.Model(&models.Application{}).Where("user_id = ?", userID).Select("COALESCE(SUM(awarded_credits), 0)").Scan(&stats.AwardedCredits)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// ExportApplications 导出申请数据
func (h *ApplicationHandler) ExportApplications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	// 获取查询参数
	format := c.DefaultQuery("format", "csv")
	activityID := c.Query("activity_id")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := h.db.Model(&models.Application{})

	// 权限过滤：学生只能导出自己的申请，教师/管理员可以导出所有申请
	if userType == "student" {
		query = query.Where("user_id = ?", userID)
	}

	// 应用筛选条件
	if activityID != "" {
		query = query.Where("activity_id = ?", activityID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		if start, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("submitted_at >= ?", start)
		}
	}
	if endDate != "" {
		if end, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("submitted_at <= ?", end.Add(24*time.Hour))
		}
	}

	var applications []models.Application
	if err := query.Order("submitted_at DESC").Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取申请数据失败",
			"data":    err.Error(),
		})
		return
	}

	// 根据格式生成导出文件
	switch format {
	case "csv":
		h.exportToCSV(c, applications)
	case "excel":
		h.exportToExcel(c, applications)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的导出格式，支持的格式：csv, excel",
			"data":    nil,
		})
	}
}

// exportToCSV 导出为CSV格式
func (h *ApplicationHandler) exportToCSV(c *gin.Context, applications []models.Application) {
	// 设置响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=applications.csv")

	// 写入CSV头部
	c.Writer.WriteString("申请ID,活动ID,用户ID,状态,申请学分,获得学分,提交时间,创建时间\n")

	// 写入数据
	for _, app := range applications {
		line := fmt.Sprintf("%s,%s,%s,%s,%.2f,%.2f,%s,%s\n",
			app.ID,
			app.ActivityID,
			app.UserID,
			app.Status,
			app.AppliedCredits,
			app.AwardedCredits,
			app.SubmittedAt.Format("2006-01-02 15:04:05"),
			app.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		c.Writer.WriteString(line)
	}
}

// exportToExcel 导出为Excel格式
func (h *ApplicationHandler) exportToExcel(c *gin.Context, applications []models.Application) {
	// 这里应该使用Excel库生成Excel文件
	// 暂时返回CSV格式
	h.exportToCSV(c, applications)
}
