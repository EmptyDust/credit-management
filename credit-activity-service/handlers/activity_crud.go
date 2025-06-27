package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"credit-management/credit-activity-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

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

	var startDate, endDate time.Time
	var err error

	if req.StartDate != "" {
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

	if !startDate.IsZero() && !endDate.IsZero() && startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始日期不能晚于结束日期",
			"data":    nil,
		})
		return
	}

	activity := models.CreditActivity{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      models.StatusDraft,
		Category:    req.Category,
		OwnerID:     userID.(string),
	}

	if err := h.db.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 创建详情表
	switch req.Category {
	case "创新创业实践活动":
		if req.InnovationDetail != nil {
			detail := req.InnovationDetail
			detail.ActivityID = activity.ID
			// 处理Date字段
			if detail.Date.IsZero() && detail.Date.String() == "" && req.InnovationDetail.Date.String() != "" {
				parsedDate, err := time.Parse("2006-01-02", req.InnovationDetail.Date.String())
				if err == nil {
					detail.Date = parsedDate
				}
			}
			h.db.Create(detail)
		}
	case "学科竞赛":
		if req.CompetitionDetail != nil {
			detail := req.CompetitionDetail
			detail.ActivityID = activity.ID
			h.db.Create(detail)
		}
	case "大学生创业项目":
		if req.EntrepreneurshipProjectDetail != nil {
			detail := req.EntrepreneurshipProjectDetail
			detail.ActivityID = activity.ID
			h.db.Create(detail)
		}
	case "创业实践项目":
		if req.EntrepreneurshipPracticeDetail != nil {
			detail := req.EntrepreneurshipPracticeDetail
			detail.ActivityID = activity.ID
			h.db.Create(detail)
		}
	case "论文专利":
		if req.PaperPatentDetail != nil {
			detail := req.PaperPatentDetail
			detail.ActivityID = activity.ID
			h.db.Create(detail)
		}
	}

	// 构建响应
	response := h.enrichActivityResponse(activity, "")

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "活动创建成功",
		"data":    response,
	})
}

func (h *ActivityHandler) GetActivities(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	query := c.Query("query")
	status := c.Query("status")
	category := c.Query("category")
	ownerID := c.Query("owner_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSizeStr := c.DefaultQuery("page_size", "")
	limit := 10
	if pageSizeStr != "" {
		limit, _ = strconv.Atoi(pageSizeStr)
	} else {
		limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	dbQuery := h.db.Model(&models.CreditActivity{})

	// 权限过滤：学生只能看到自己创建或参与的活动，教师可以看到所有活动
	if userType == "student" {
		dbQuery = dbQuery.Where("owner_id = ? OR id IN (SELECT activity_id FROM activity_participants WHERE user_id = ?)", userID, userID)
	}

	if query != "" {
		// 关键词搜索：支持标题、描述、类别的模糊搜索
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

	var responses []models.ActivityResponse

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

	if userType == "student" {
		if activity.OwnerID != userID {
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

	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

	response := h.enrichActivityResponse(activity, authToken)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

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

	if activity.OwnerID != userID && userType != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限更新此活动",
			"data":    nil,
		})
		return
	}

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

	var compareStartDate, compareEndDate time.Time

	if req.StartDate != nil {
		compareStartDate = startDate
	} else {
		compareStartDate = activity.StartDate
	}

	if req.EndDate != nil {
		compareEndDate = endDate
	} else {
		compareEndDate = activity.EndDate
	}

	if !compareStartDate.IsZero() && !compareEndDate.IsZero() && compareStartDate.After(compareEndDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始日期不能晚于结束日期",
			"data":    nil,
		})
		return
	}

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

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "没有提供要更新的字段",
			"data":    nil,
		})
		return
	}

	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新活动失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	switch activity.Category {
	case "创新创业实践活动":
		if req.InnovationDetail != nil {
			var detail models.InnovationActivityDetail
			h.db.Where("activity_id = ?", activity.ID).First(&detail)
			if detail.ID != "" {
				h.db.Model(&detail).Updates(req.InnovationDetail)
			} else {
				detail = *req.InnovationDetail
				detail.ActivityID = activity.ID
				h.db.Create(&detail)
			}
		}
	case "学科竞赛":
		if req.CompetitionDetail != nil {
			var detail models.CompetitionActivityDetail
			h.db.Where("activity_id = ?", activity.ID).First(&detail)
			if detail.ID != "" {
				h.db.Model(&detail).Updates(req.CompetitionDetail)
			} else {
				detail = *req.CompetitionDetail
				detail.ActivityID = activity.ID
				h.db.Create(&detail)
			}
		}
	case "大学生创业项目":
		if req.EntrepreneurshipProjectDetail != nil {
			var detail models.EntrepreneurshipProjectDetail
			h.db.Where("activity_id = ?", activity.ID).First(&detail)
			if detail.ID != "" {
				h.db.Model(&detail).Updates(req.EntrepreneurshipProjectDetail)
			} else {
				detail = *req.EntrepreneurshipProjectDetail
				detail.ActivityID = activity.ID
				h.db.Create(&detail)
			}
		}
	case "创业实践项目":
		if req.EntrepreneurshipPracticeDetail != nil {
			var detail models.EntrepreneurshipPracticeDetail
			h.db.Where("activity_id = ?", activity.ID).First(&detail)
			if detail.ID != "" {
				h.db.Model(&detail).Updates(req.EntrepreneurshipPracticeDetail)
			} else {
				detail = *req.EntrepreneurshipPracticeDetail
				detail.ActivityID = activity.ID
				h.db.Create(&detail)
			}
		}
	case "论文专利":
		if req.PaperPatentDetail != nil {
			var detail models.PaperPatentDetail
			h.db.Where("activity_id = ?", activity.ID).First(&detail)
			if detail.ID != "" {
				h.db.Model(&detail).Updates(req.PaperPatentDetail)
			} else {
				detail = *req.PaperPatentDetail
				detail.ActivityID = activity.ID
				h.db.Create(&detail)
			}
		}
	}

	response := h.enrichActivityResponse(activity, "")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "活动更新成功",
		"data":    response,
	})
}

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

	if result != "活动删除成功" {
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
