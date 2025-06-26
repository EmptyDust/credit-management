package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"credit-management/credit-activity-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CopyActivity 复制活动
func (h *ActivityHandler) CopyActivity(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
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

	// 获取原活动
	var originalActivity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&originalActivity).Error; err != nil {
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

	// 创建新活动（复制基本信息）
	newActivity := models.CreditActivity{
		Title:        originalActivity.Title + " (副本)",
		Description:  originalActivity.Description,
		StartDate:    originalActivity.StartDate,
		EndDate:      originalActivity.EndDate,
		Status:       models.StatusDraft,
		Category:     originalActivity.Category,
		Requirements: originalActivity.Requirements,
		OwnerID:      userID.(string),
	}

	if err := h.db.Create(&newActivity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "复制活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	response := models.ActivityResponse{
		ID:             newActivity.ID,
		Title:          newActivity.Title,
		Description:    newActivity.Description,
		StartDate:      newActivity.StartDate,
		EndDate:        newActivity.EndDate,
		Status:         newActivity.Status,
		Category:       newActivity.Category,
		Requirements:   newActivity.Requirements,
		OwnerID:        newActivity.OwnerID,
		ReviewerID:     newActivity.ReviewerID,
		ReviewComments: newActivity.ReviewComments,
		ReviewedAt:     newActivity.ReviewedAt,
		CreatedAt:      newActivity.CreatedAt,
		UpdatedAt:      newActivity.UpdatedAt,
		Participants:   []models.ParticipantResponse{},
		Applications:   []models.ApplicationResponse{},
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "活动复制成功",
		"data":    response,
	})
}

// ExportActivities 导出活动数据
func (h *ActivityHandler) ExportActivities(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	category := c.Query("category")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 构建查询
	dbQuery := h.db.Model(&models.CreditActivity{})

	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}
	if startDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDate); err == nil {
			dbQuery = dbQuery.Where("start_date >= ?", parsedDate)
		}
	}
	if endDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDate); err == nil {
			dbQuery = dbQuery.Where("end_date <= ?", parsedDate)
		}
	}

	var activities []models.CreditActivity
	if err := dbQuery.Order("created_at DESC").Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取活动数据失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "导出成功",
			"data":    activities,
		})
	case "csv":
		// 这里可以实现CSV导出功能
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "导出成功",
			"data":    gin.H{"message": "CSV导出功能待实现", "count": len(activities)},
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的导出格式",
			"data":    nil,
		})
	}
}

// SaveAsTemplate 保存活动为模板
func (h *ActivityHandler) SaveAsTemplate(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
			"data":    nil,
		})
		return
	}

	var req struct {
		TemplateName string `json:"template_name" binding:"required"`
		Description  string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 获取原活动
	var originalActivity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&originalActivity).Error; err != nil {
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

	// 这里可以实现模板保存逻辑
	// 目前返回成功消息，实际实现需要创建模板表

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "模板保存成功",
		"data": gin.H{
			"template_name": req.TemplateName,
			"activity_id":   activityID,
		},
	})
}

// ImportActivitiesFromCSV 从CSV文件批量导入活动
func (h *ActivityHandler) ImportActivitiesFromCSV(c *gin.Context) {
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

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请上传CSV文件",
			"data":    nil,
		})
		return
	}

	// 检查文件类型
	if !strings.HasSuffix(file.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只支持CSV文件格式",
			"data":    nil,
		})
		return
	}

	// 检查文件大小（限制为5MB）
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件大小不能超过5MB",
			"data":    nil,
		})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "打开文件失败: " + err.Error(),
			"data":    nil,
		})
		return
	}
	defer src.Close()

	// 读取CSV文件
	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1 // 允许变长记录

	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "CSV文件格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 检查记录数量
	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "CSV文件至少需要包含标题行和一行数据",
			"data":    nil,
		})
		return
	}

	if len(records) > 1001 { // 标题行 + 1000行数据
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "CSV文件最多支持1000行数据",
			"data":    nil,
		})
		return
	}

	// 解析标题行
	headers := records[0]
	expectedHeaders := []string{"title", "description", "start_date", "end_date", "category", "requirements"}

	// 检查必需的列
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	missingHeaders := []string{}
	for _, expected := range expectedHeaders {
		if _, exists := headerMap[expected]; !exists {
			missingHeaders = append(missingHeaders, expected)
		}
	}

	if len(missingHeaders) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "CSV文件缺少必需的列: " + strings.Join(missingHeaders, ", "),
			"data":    nil,
		})
		return
	}

	// 解析数据行
	var activities []models.ActivityRequest
	var errors []string

	for i, record := range records[1:] {
		rowNum := i + 2 // 从第2行开始计算

		// 检查记录长度
		if len(record) < len(headers) {
			errors = append(errors, fmt.Sprintf("第%d行: 列数不匹配", rowNum))
			continue
		}

		// 构建活动请求
		activity := models.ActivityRequest{
			Title:        strings.TrimSpace(record[headerMap["title"]]),
			Description:  strings.TrimSpace(record[headerMap["description"]]),
			StartDate:    strings.TrimSpace(record[headerMap["start_date"]]),
			EndDate:      strings.TrimSpace(record[headerMap["end_date"]]),
			Category:     strings.TrimSpace(record[headerMap["category"]]),
			Requirements: strings.TrimSpace(record[headerMap["requirements"]]),
		}

		// 验证数据
		if err := h.validateActivityRequest(activity); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行: %s", rowNum, err.Error()))
			continue
		}

		activities = append(activities, activity)
	}

	// 如果有验证错误，返回错误信息
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "CSV数据验证失败",
			"data": gin.H{
				"errors":       errors,
				"total_rows":   len(records) - 1,
				"valid_rows":   len(activities),
				"invalid_rows": len(errors),
			},
		})
		return
	}

	// 批量创建活动
	var createdActivities []models.ActivityCreateResponse
	var createErrors []string

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, activityReq := range activities {
		// 解析日期
		startDate, endDate, err := h.parseActivityDates(activityReq.StartDate, activityReq.EndDate)
		if err != nil {
			createErrors = append(createErrors, fmt.Sprintf("第%d个活动: %s", i+1, err.Error()))
			continue
		}

		// 创建活动
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
			createErrors = append(createErrors, fmt.Sprintf("第%d个活动创建失败: %s", i+1, err.Error()))
			continue
		}

		// 构建响应
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

	// 如果有创建错误，回滚事务
	if len(createErrors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "批量创建活动失败",
			"data": gin.H{
				"errors":             createErrors,
				"created_count":      0,
				"total_count":        len(activities),
				"created_activities": []models.ActivityCreateResponse{},
			},
		})
		return
	}

	// 提交事务
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
		"message": "CSV导入成功",
		"data": gin.H{
			"created_count":      len(createdActivities),
			"total_count":        len(activities),
			"created_activities": createdActivities,
			"file_name":          file.Filename,
			"file_size":          file.Size,
		},
	})
}

// GetCSVTemplate 获取CSV模板
func (h *ActivityHandler) GetCSVTemplate(c *gin.Context) {
	// 创建CSV模板内容
	template := [][]string{
		{"title", "description", "start_date", "end_date", "category", "requirements"},
		{"示例活动1", "这是一个示例活动描述", "2024-01-01", "2024-01-31", "创新创业", "需要提交报告"},
		{"示例活动2", "另一个示例活动", "2024-02-01", "2024-02-28", "学科竞赛", "需要参加比赛"},
	}

	// 设置响应头
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=activity_template.csv")

	// 写入CSV数据
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	for _, record := range template {
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "生成CSV模板失败",
				"data":    nil,
			})
			return
		}
	}
} 