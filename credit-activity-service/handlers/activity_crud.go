package handlers

import (
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req models.ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	// 使用统一的验证器
	if err := h.validateActivityRequest(req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 使用统一的日期解析工具
	startDate, endDate, err := utils.ParseDateRange(req.StartDate, req.EndDate)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
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
		utils.SendInternalServerError(c, err)
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

	utils.SendCreatedResponse(c, "活动创建成功", response)
}

func (h *ActivityHandler) GetActivities(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	query := c.Query("query")
	status := c.Query("status")
	category := c.Query("category")
	ownerID := c.Query("owner_id")

	// 使用统一的验证器处理分页参数
	page, limit, _ := h.validator.ValidatePagination(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("page_size", c.DefaultQuery("limit", "10")),
	)

	// 使用数据库基类进行查询
	activities, total, err := h.base.SearchActivities(query, status, category, ownerID, userID.(string), userType.(string), page, limit)
	if err != nil {
		utils.SendInternalServerError(c, err)
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

	utils.SendPaginatedResponse(c, responses, total, page, limit)
}

func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id := c.Param("id")
	if err := h.validator.ValidateUUID(id); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	userType, _ := c.Get("user_type")

	// 使用数据库基类获取活动
	activity, err := h.base.GetActivityByIDWithParticipants(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	// 权限检查
	if userType == "student" {
		if activity.OwnerID != userID {
			if err := h.base.CheckUserParticipant(id, userID.(string)); err != nil {
				utils.SendForbidden(c, "无权限查看此活动")
				return
			}
		}
	}

	authToken := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		authToken = authHeader
	}

	response := h.enrichActivityResponse(*activity, authToken)
	utils.SendSuccessResponse(c, response)
}

func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	if err := h.validator.ValidateUUID(id); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	userType, _ := c.Get("user_type")

	// 使用数据库基类获取活动
	activity, err := h.base.GetActivityByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	// 权限检查
	if userType == "student" && activity.OwnerID != userID {
		utils.SendForbidden(c, "无权限修改此活动")
		return
	}

	var req models.ActivityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 使用统一的验证器
	if err := h.validateUpdateRequest(req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}

	// 处理日期字段
	if req.StartDate != nil || req.EndDate != nil {
		startDateStr := ""
		endDateStr := ""
		if req.StartDate != nil {
			startDateStr = *req.StartDate
		}
		if req.EndDate != nil {
			endDateStr = *req.EndDate
		}

		startDate, endDate, err := utils.ParseDateRange(startDateStr, endDateStr)
		if err != nil {
			utils.SendBadRequest(c, err.Error())
			return
		}

		if !startDate.IsZero() {
			updates["start_date"] = startDate
		}
		if !endDate.IsZero() {
			updates["end_date"] = endDate
		}
	}

	if err := h.db.Model(&models.CreditActivity{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	// 获取更新后的活动
	updatedActivity, err := h.base.GetActivityByID(id)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	response := h.enrichActivityResponse(*updatedActivity, "")
	utils.SendSuccessResponse(c, response)
}

func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	if err := h.validator.ValidateUUID(id); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 使用数据库基类检查活动是否存在
	if err := h.base.CheckActivityExists(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	if err := h.db.Delete(&models.CreditActivity{}, id).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "活动删除成功"})
}
