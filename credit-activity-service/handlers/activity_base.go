package handlers

import (
	"fmt"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

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

	// 获取 owner 信息
	if ownerInfo, err := h.getUserInfo(activity.OwnerID, authToken); err == nil {
		response.OwnerInfo = ownerInfo
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

		// 获取用户信息
		if userInfo, err := h.getUserInfo(application.UserID, authToken); err == nil {
			response.UserInfo = userInfo
		}

		applicationResponses = append(applicationResponses, response)
	}
	response.Applications = applicationResponses

	// 联查详情表
	switch activity.Category {
	case "创新创业实践活动":
		var detail models.InnovationActivityDetail
		h.db.Where("activity_id = ?", activity.ID).First(&detail)
		if detail.ID != "" {
			response.InnovationDetail = &detail
		}
	case "学科竞赛":
		var detail models.CompetitionActivityDetail
		h.db.Where("activity_id = ?", activity.ID).First(&detail)
		if detail.ID != "" {
			response.CompetitionDetail = &detail
		}
	case "大学生创业项目":
		var detail models.EntrepreneurshipProjectDetail
		h.db.Where("activity_id = ?", activity.ID).First(&detail)
		if detail.ID != "" {
			response.EntrepreneurshipProjectDetail = &detail
		}
	case "创业实践项目":
		var detail models.EntrepreneurshipPracticeDetail
		h.db.Where("activity_id = ?", activity.ID).First(&detail)
		if detail.ID != "" {
			response.EntrepreneurshipPracticeDetail = &detail
		}
	case "论文专利":
		var detail models.PaperPatentDetail
		h.db.Where("activity_id = ?", activity.ID).First(&detail)
		if detail.ID != "" {
			response.PaperPatentDetail = &detail
		}
	}

	return response
}

// getUserInfo 获取用户信息（使用真实用户服务）
func (h *ActivityHandler) getUserInfo(userID string, authToken string) (*models.UserInfo, error) {
	return utils.GetUserInfo(userID, authToken)
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
