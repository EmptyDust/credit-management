package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityHandler 活动处理器
type ActivityHandler struct {
	db *gorm.DB
}

// NewActivityHandler 创建活动处理器
func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{db: db}
}

// CreateActivity 创建活动
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req models.ActivityRequest
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

	// 业务验证
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动标题不能为空",
			"data":    nil,
		})
		return
	}

	if len(req.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动标题长度不能超过200个字符",
			"data":    nil,
		})
		return
	}

	// 验证活动类别
	if req.Category != "" {
		validCategories := models.GetActivityCategories()
		categoryValid := false
		for _, category := range validCategories {
			if category == req.Category {
				categoryValid = true
				break
			}
		}
		if !categoryValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的活动类别",
				"data":    nil,
			})
			return
		}
	}

	// 解析日期 - 支持多种格式
	var startDate, endDate time.Time
	var err error

	if req.StartDate != "" {
		// 尝试多种日期格式
		dateFormats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		parsed := false
		for _, format := range dateFormats {
			if startDate, err = time.Parse(format, req.StartDate); err == nil {
				parsed = true
				break
			}
		}

		if !parsed {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "开始日期格式错误，支持格式：YYYY-MM-DD、YYYY-MM-DD HH:mm:ss、YYYY-MM-DDTHH:mm:ss",
				"data":    nil,
			})
			return
		}
	}

	if req.EndDate != "" {
		// 尝试多种日期格式
		dateFormats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		parsed := false
		for _, format := range dateFormats {
			if endDate, err = time.Parse(format, req.EndDate); err == nil {
				parsed = true
				break
			}
		}

		if !parsed {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "结束日期格式错误，支持格式：YYYY-MM-DD、YYYY-MM-DD HH:mm:ss、YYYY-MM-DDTHH:mm:ss",
				"data":    nil,
			})
			return
		}
	}

	// 验证日期逻辑
	if !startDate.IsZero() && !endDate.IsZero() && startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始日期不能晚于结束日期",
			"data":    nil,
		})
		return
	}

	// 创建活动
	activity := models.CreditActivity{
		Title:        req.Title,
		Description:  req.Description,
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       models.StatusDraft,
		Category:     req.Category,
		Requirements: req.Requirements,
		OwnerID:      userID.(string),
	}

	if err := h.db.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	response := models.ActivityResponse{
		ID:             activity.ID,
		Title:          activity.Title,
		Description:    activity.Description,
		StartDate:      activity.StartDate,
		EndDate:        activity.EndDate,
		Status:         activity.Status,
		Category:       activity.Category,
		Requirements:   activity.Requirements,
		OwnerID:        activity.OwnerID,
		ReviewerID:     activity.ReviewerID,
		ReviewComments: activity.ReviewComments,
		ReviewedAt:     activity.ReviewedAt,
		CreatedAt:      activity.CreatedAt,
		UpdatedAt:      activity.UpdatedAt,
		Participants:   []models.ParticipantResponse{},
		Applications:   []models.ApplicationResponse{},
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "活动创建成功",
		"data":    response,
	})
}

// GetActivities 获取活动列表
func (h *ActivityHandler) GetActivities(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	// 获取查询参数
	query := c.Query("query")
	status := c.Query("status")
	category := c.Query("category")
	ownerID := c.Query("owner_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	dbQuery := h.db.Model(&models.CreditActivity{})

	// 权限过滤：学生只能看到自己创建或参与的活动，教师可以看到所有活动
	if userType == "student" {
		// 学生只能看到自己创建的活动或参与的活动
		dbQuery = dbQuery.Where("owner_id = ? OR id IN (SELECT activity_id FROM activity_participants WHERE user_id = ?)", userID, userID)
	}

	// 应用筛选条件
	if query != "" {
		// 关键词搜索：支持标题、描述、类别、要求的模糊搜索
		searchQuery := "%" + query + "%"
		dbQuery = dbQuery.Where(
			"title ILIKE ? OR description ILIKE ? OR category ILIKE ? OR requirements ILIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery,
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

	var activities []models.CreditActivity
	var total int64

	dbQuery.Count(&total)
	if err := dbQuery.Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取活动列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应
	var responses []models.ActivityResponse

	// 获取当前用户的认证令牌
	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

	for _, activity := range activities {
		response := h.enrichActivityResponse(activity, authToken)
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.PaginatedResponse{
			Data:       responses,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: (int(total) + limit - 1) / limit,
		},
	})
}

// GetActivity 获取活动详情
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
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

	userType, _ := c.Get("user_type")

	var activity models.CreditActivity
	if err := h.db.Preload("Participants").Where("id = ?", id).First(&activity).Error; err != nil {
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

	// 权限检查：学生只能查看自己创建或参与的活动
	if userType == "student" {
		if activity.OwnerID != userID {
			// 检查是否为参与者
			var participant models.ActivityParticipant
			if err := h.db.Where("activity_id = ? AND user_id = ?", id, userID).First(&participant).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "无权限查看此活动",
					"data":    nil,
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    activity,
	})
}

// UpdateActivity 更新活动
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
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

	userType, _ := c.Get("user_type")

	// 获取活动信息
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

	// 权限检查：只有活动创建者和管理员可以更新
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限更新此活动",
			"data":    nil,
		})
		return
	}

	// 状态检查：只有草稿状态的活动可以修改
	if activity.Status != models.StatusDraft && userType != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只有草稿状态的活动可以修改",
			"data":    nil,
		})
		return
	}

	var req models.ActivityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 验证更新数据
	if err := h.validateUpdateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 解析日期
	var startDate, endDate time.Time
	var err error

	if req.StartDate != nil {
		startDate, err = h.parseSingleDate(*req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "开始日期格式错误: " + err.Error(),
				"data":    nil,
			})
			return
		}
	}

	if req.EndDate != nil {
		endDate, err = h.parseSingleDate(*req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "结束日期格式错误: " + err.Error(),
				"data":    nil,
			})
			return
		}
	}

	// 验证日期逻辑
	if !startDate.IsZero() && !endDate.IsZero() && startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始日期不能晚于结束日期",
			"data":    nil,
		})
		return
	}

	// 更新活动信息
	updateFields := make(map[string]interface{})

	if req.Title != nil {
		if *req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "活动标题不能为空",
				"data":    nil,
			})
			return
		}
		activity.Title = *req.Title
		updateFields["title"] = *req.Title
	}

	if req.Description != nil {
		activity.Description = *req.Description
		updateFields["description"] = *req.Description
	}

	if req.StartDate != nil {
		activity.StartDate = startDate
		updateFields["start_date"] = startDate
	}

	if req.EndDate != nil {
		activity.EndDate = endDate
		updateFields["end_date"] = endDate
	}

	if req.Category != nil {
		activity.Category = *req.Category
		updateFields["category"] = *req.Category
	}

	if req.Requirements != nil {
		activity.Requirements = *req.Requirements
		updateFields["requirements"] = *req.Requirements
	}

	// 如果没有要更新的字段，返回错误
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "没有提供要更新的字段",
			"data":    nil,
		})
		return
	}

	// 更新活动
	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应数据
	response := models.ActivityResponse{
		ID:             activity.ID,
		Title:          activity.Title,
		Description:    activity.Description,
		StartDate:      activity.StartDate,
		EndDate:        activity.EndDate,
		Status:         activity.Status,
		Category:       activity.Category,
		Requirements:   activity.Requirements,
		OwnerID:        activity.OwnerID,
		ReviewerID:     activity.ReviewerID,
		ReviewComments: activity.ReviewComments,
		ReviewedAt:     activity.ReviewedAt,
		CreatedAt:      activity.CreatedAt,
		UpdatedAt:      activity.UpdatedAt,
		Participants:   []models.ParticipantResponse{},
		Applications:   []models.ApplicationResponse{},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "活动更新成功",
		"data":    response,
	})
}

// DeleteActivity 删除活动
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
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

	userType, _ := c.Get("user_type")

	// 使用存储过程删除活动（包含权限检查和级联删除）
	var result string
	err := h.db.Raw("SELECT delete_activity_with_permission_check(?, ?, ?)", id, userID, userType).Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 检查存储过程的返回结果
	if result != "活动删除成功" {
		// 根据返回的错误信息设置相应的HTTP状态码
		switch result {
		case "活动不存在或已删除":
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": result,
				"data":    nil,
			})
		case "无权限删除该活动":
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": result,
				"data":    nil,
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": result,
				"data":    nil,
			})
		}
		return
	}

	// 删除活动相关的物理文件（如果有的话）
	// 注意：存储过程已经处理了附件的逻辑删除，这里只需要处理物理文件
	var attachments []models.Attachment
	if err := h.db.Where("activity_id = ? AND deleted_at IS NOT NULL", id).Find(&attachments).Error; err == nil {
		for _, attachment := range attachments {
			// 检查是否有其他活动使用相同的文件
			var otherAttachmentsCount int64
			h.db.Model(&models.Attachment{}).
				Where("md5_hash = ? AND activity_id != ? AND deleted_at IS NULL", attachment.MD5Hash, id).
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "活动删除成功",
		"data": gin.H{
			"activity_id": id,
			"deleted_at":  time.Now(),
		},
	})
}

// SubmitActivity 提交活动审核
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

	// 权限检查：只有活动创建者可以提交审核
	if activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限提交此活动",
			"data":    nil,
		})
		return
	}

	// 只有草稿状态的活动可以提交审核
	if activity.Status != models.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能提交草稿状态的活动",
			"data":    nil,
		})
		return
	}

	// 更新状态为待审核
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

// ReviewActivity 审核活动
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

	// 只有待审核状态的活动可以审核
	if activity.Status != models.StatusPendingReview {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能审核待审核状态的活动",
			"data":    nil,
		})
		return
	}

	// 更新审核信息
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

// GetPendingActivities 获取待审核活动
func (h *ActivityHandler) GetPendingActivities(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var activities []models.CreditActivity
	var total int64

	query := h.db.Where("status = ?", models.StatusPendingReview)

	// 统计总数
	query.Model(&models.CreditActivity{}).Count(&total)

	// 获取分页数据
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

// GetActivityStats 获取活动统计
func (h *ActivityHandler) GetActivityStats(c *gin.Context) {
	var stats models.ActivityStats

	// 统计各种状态的活动数量
	h.db.Model(&models.CreditActivity{}).Count(&stats.TotalActivities)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusDraft).Count(&stats.DraftCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusPendingReview).Count(&stats.PendingCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusApproved).Count(&stats.ApprovedCount)
	h.db.Model(&models.CreditActivity{}).Where("status = ?", models.StatusRejected).Count(&stats.RejectedCount)

	// 统计参与者总数
	h.db.Model(&models.ActivityParticipant{}).Count(&stats.TotalParticipants)

	// 统计总学分
	h.db.Model(&models.ActivityParticipant{}).Select("COALESCE(SUM(credits), 0)").Scan(&stats.TotalCredits)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetActivityCategories 获取活动类别
func (h *ActivityHandler) GetActivityCategories(c *gin.Context) {
	categories := models.GetActivityCategories()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"categories":  categories,
			"count":       len(categories),
			"description": "活动类别列表",
		},
	})
}

// WithdrawActivity 撤回活动
func (h *ActivityHandler) WithdrawActivity(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

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

	// 权限检查：只有活动创建者可以撤回
	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，只有活动创建者可以撤回活动",
			"data":    nil,
		})
		return
	}

	// 检查活动状态：只有非草稿状态的活动可以撤回
	if activity.Status == models.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "草稿状态的活动无需撤回",
			"data":    nil,
		})
		return
	}

	// 撤回活动到草稿状态
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

// enrichActivityResponse 丰富活动响应信息
func (h *ActivityHandler) enrichActivityResponse(activity models.CreditActivity, authToken string) models.ActivityResponse {
	response := models.ActivityResponse{
		ID:             activity.ID,
		Title:          activity.Title,
		Description:    activity.Description,
		StartDate:      activity.StartDate,
		EndDate:        activity.EndDate,
		Status:         activity.Status,
		Category:       activity.Category,
		Requirements:   activity.Requirements,
		OwnerID:        activity.OwnerID,
		ReviewerID:     activity.ReviewerID,
		ReviewComments: activity.ReviewComments,
		ReviewedAt:     activity.ReviewedAt,
		CreatedAt:      activity.CreatedAt,
		UpdatedAt:      activity.UpdatedAt,
	}

	// 获取参与者信息
	var participants []models.ActivityParticipant
	h.db.Where("activity_id = ?", activity.ID).Find(&participants)

	var participantResponses []models.ParticipantResponse
	for _, participant := range participants {
		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
		}

		// 获取用户信息
		if userInfo, err := h.getUserInfo(participant.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}

		participantResponses = append(participantResponses, response)
	}
	response.Participants = participantResponses

	// 获取申请信息
	var applications []models.Application
	h.db.Where("activity_id = ?", activity.ID).Find(&applications)

	var applicationResponses []models.ApplicationResponse
	for _, application := range applications {
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
		}

		// 获取活动信息
		response.Activity = models.ActivityInfo{
			ID:          activity.ID,
			Title:       activity.Title,
			Description: activity.Description,
			Category:    activity.Category,
			StartDate:   activity.StartDate,
			EndDate:     activity.EndDate,
		}

		applicationResponses = append(applicationResponses, response)
	}
	response.Applications = applicationResponses

	return response
}

// getUserInfo 获取用户信息（使用真实用户服务）
func (h *ActivityHandler) getUserInfo(userID string, authToken string) (*models.UserInfo, error) {
	return utils.GetUserInfo(userID, authToken)
}

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

	// 验证批量创建数量
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

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, activityReq := range req.Activities {
		// 验证活动数据
		if err := h.validateActivityRequest(activityReq); err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动: %s", i+1, err.Error()))
			continue
		}

		// 解析日期
		startDate, endDate, err := h.parseActivityDates(activityReq.StartDate, activityReq.EndDate)
		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d个活动: %s", i+1, err.Error()))
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
			errors = append(errors, fmt.Sprintf("第%d个活动创建失败: %s", i+1, err.Error()))
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

	// 如果有错误，回滚事务
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
		"message": "批量创建活动成功",
		"data": gin.H{
			"created_count":      len(createdActivities),
			"total_count":        len(req.Activities),
			"created_activities": createdActivities,
		},
	})
}

// validateActivityRequest 验证活动请求数据
func (h *ActivityHandler) validateActivityRequest(req models.ActivityRequest) error {
	if req.Title == "" {
		return fmt.Errorf("活动标题不能为空")
	}

	if len(req.Title) > 200 {
		return fmt.Errorf("活动标题长度不能超过200个字符")
	}

	if req.Category != "" {
		validCategories := models.GetActivityCategories()
		categoryValid := false
		for _, category := range validCategories {
			if category == req.Category {
				categoryValid = true
				break
			}
		}
		if !categoryValid {
			return fmt.Errorf("无效的活动类别")
		}
	}

	return nil
}

// parseActivityDates 解析活动日期
func (h *ActivityHandler) parseActivityDates(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	var err error

	dateFormats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	if startDateStr != "" {
		parsed := false
		for _, format := range dateFormats {
			if startDate, err = time.Parse(format, startDateStr); err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			return time.Time{}, time.Time{}, fmt.Errorf("开始日期格式错误")
		}
	}

	if endDateStr != "" {
		parsed := false
		for _, format := range dateFormats {
			if endDate, err = time.Parse(format, endDateStr); err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			return time.Time{}, time.Time{}, fmt.Errorf("结束日期格式错误")
		}
	}

	// 验证日期逻辑
	if !startDate.IsZero() && !endDate.IsZero() && startDate.After(endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期不能晚于结束日期")
	}

	return startDate, endDate, nil
}

// GetActivityTemplates 获取活动模板
func (h *ActivityHandler) GetActivityTemplates(c *gin.Context) {
	templates := []gin.H{
		{
			"name":         "创新创业活动",
			"category":     models.CategoryInnovation,
			"title":        "创新创业实践活动",
			"description":  "参与创新创业项目，提升创新能力和实践技能",
			"requirements": "需要提交项目计划书和成果展示",
		},
		{
			"name":         "学科竞赛",
			"category":     models.CategoryCompetition,
			"title":        "学科竞赛活动",
			"description":  "参加各类学科竞赛，提升专业能力和竞争意识",
			"requirements": "需要获得竞赛证书或奖项证明",
		},
		{
			"name":         "志愿服务",
			"category":     models.CategoryVolunteer,
			"title":        "志愿服务活动",
			"description":  "参与社会志愿服务，培养社会责任感和奉献精神",
			"requirements": "需要志愿服务时长证明",
		},
		{
			"name":         "学术研究",
			"category":     models.CategoryAcademic,
			"title":        "学术研究活动",
			"description":  "参与学术研究项目，提升科研能力和学术素养",
			"requirements": "需要提交研究报告或论文",
		},
		{
			"name":         "文体活动",
			"category":     models.CategoryCultural,
			"title":        "文体活动",
			"description":  "参与文化体育活动，提升综合素质和团队协作能力",
			"requirements": "需要活动参与证明",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    templates,
	})
}

// validateUpdateRequest 验证活动更新请求数据
func (h *ActivityHandler) validateUpdateRequest(req models.ActivityUpdateRequest) error {
	if req.Title != nil && *req.Title == "" {
		return fmt.Errorf("活动标题不能为空")
	}

	if req.Category != nil {
		validCategories := models.GetActivityCategories()
		categoryValid := false
		for _, category := range validCategories {
			if category == *req.Category {
				categoryValid = true
				break
			}
		}
		if !categoryValid {
			return fmt.Errorf("无效的活动类别")
		}
	}

	return nil
}

// parseSingleDate 解析单个日期
func (h *ActivityHandler) parseSingleDate(dateStr string) (time.Time, error) {
	var date time.Time
	var err error

	dateFormats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range dateFormats {
		if date, err = time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("日期格式错误")
}

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

// GetActivityReport 获取活动统计报表
func (h *ActivityHandler) GetActivityReport(c *gin.Context) {
	reportType := c.DefaultQuery("type", "monthly")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var start, end time.Time
	var err error

	// 解析日期范围
	if startDate != "" && endDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "开始日期格式错误",
				"data":    nil,
			})
			return
		}
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "结束日期格式错误",
				"data":    nil,
			})
			return
		}
	} else {
		// 默认查询最近30天
		end = time.Now()
		start = end.AddDate(0, 0, -30)
	}

	var report interface{}

	switch reportType {
	case "monthly":
		report = h.generateMonthlyReport(start, end)
	case "category":
		report = h.generateCategoryReport(start, end)
	case "status":
		report = h.generateStatusReport(start, end)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的报表类型",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取报表成功",
		"data":    report,
	})
}

// generateMonthlyReport 生成月度报表
func (h *ActivityHandler) generateMonthlyReport(start, end time.Time) map[string]interface{} {
	var result []map[string]interface{}

	// 按月份统计活动数量
	rows, err := h.db.Raw(`
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as total_activities,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_activities,
			COUNT(CASE WHEN status = 'pending_review' THEN 1 END) as pending_activities,
			COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected_activities
		FROM credit_activities 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`, start, end).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var month time.Time
			var total, approved, pending, rejected int64
			rows.Scan(&month, &total, &approved, &pending, &rejected)
			result = append(result, map[string]interface{}{
				"month":               month.Format("2006-01"),
				"total_activities":    total,
				"approved_activities": approved,
				"pending_activities":  pending,
				"rejected_activities": rejected,
			})
		}
	}

	return map[string]interface{}{
		"type":       "monthly",
		"start_date": start.Format("2006-01-02"),
		"end_date":   end.Format("2006-01-02"),
		"data":       result,
	}
}

// generateCategoryReport 生成分类报表
func (h *ActivityHandler) generateCategoryReport(start, end time.Time) map[string]interface{} {
	var result []map[string]interface{}

	rows, err := h.db.Raw(`
		SELECT 
			category,
			COUNT(*) as total_activities,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_activities,
			AVG(EXTRACT(EPOCH FROM (end_date - start_date))/86400) as avg_duration_days
		FROM credit_activities 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY category
		ORDER BY total_activities DESC
	`, start, end).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var category string
			var total, approved int64
			var avgDuration float64
			rows.Scan(&category, &total, &approved, &avgDuration)
			result = append(result, map[string]interface{}{
				"category":            category,
				"total_activities":    total,
				"approved_activities": approved,
				"avg_duration_days":   avgDuration,
			})
		}
	}

	return map[string]interface{}{
		"type":       "category",
		"start_date": start.Format("2006-01-02"),
		"end_date":   end.Format("2006-01-02"),
		"data":       result,
	}
}

// generateStatusReport 生成状态报表
func (h *ActivityHandler) generateStatusReport(start, end time.Time) map[string]interface{} {
	var result []map[string]interface{}

	rows, err := h.db.Raw(`
		SELECT 
			status,
			COUNT(*) as count,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as recent_count
		FROM credit_activities 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY status
		ORDER BY count DESC
	`, start, end).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count, recentCount int64
			rows.Scan(&status, &count, &recentCount)
			result = append(result, map[string]interface{}{
				"status":       status,
				"count":        count,
				"recent_count": recentCount,
			})
		}
	}

	return map[string]interface{}{
		"type":       "status",
		"start_date": start.Format("2006-01-02"),
		"end_date":   end.Format("2006-01-02"),
		"data":       result,
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
