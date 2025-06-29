package handlers

import (
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"gorm.io/gorm"
)

type ActivityHandler struct {
	db        *gorm.DB
	validator *utils.Validator
	base      *utils.BaseHandler
}

func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{
		db:        db,
		validator: utils.NewValidator(),
		base:      utils.NewBaseHandler(db),
	}
}

func (h *ActivityHandler) enrichActivityResponse(activity models.CreditActivity, authToken string) models.ActivityResponse {
	response := models.ActivityResponse{
		ID:             activity.ID,
		Title:          activity.Title,
		Description:    activity.Description,
		StartDate:      activity.StartDate,
		EndDate:        activity.EndDate,
		Status:         activity.Status,
		Category:       activity.Category,
		OwnerID:        activity.OwnerID,
		ReviewerID:     activity.ReviewerID,
		ReviewComments: activity.ReviewComments,
		ReviewedAt:     activity.ReviewedAt,
		CreatedAt:      activity.CreatedAt,
		UpdatedAt:      activity.UpdatedAt,
	}

	if ownerInfo, err := h.getUserInfo(activity.OwnerID, authToken); err == nil {
		response.OwnerInfo = ownerInfo
	}

	var participants []models.ActivityParticipant
	h.db.Where("activity_id = ?", activity.ID).Find(&participants)

	var participantResponses []models.ParticipantResponse
	for _, participant := range participants {
		userInfo, err := h.getUserInfo(participant.UserID, authToken)
		if err != nil {
			continue
		}

		response := models.ParticipantResponse{
			UserID:   participant.UserID,
			Credits:  participant.Credits,
			JoinedAt: participant.JoinedAt,
			UserInfo: userInfo,
		}

		participantResponses = append(participantResponses, response)
	}
	response.Participants = participantResponses

	var applications []models.Application
	h.db.Where("activity_id = ?", activity.ID).Find(&applications)

	var applicationResponses []models.ApplicationResponse
	for _, application := range applications {
		userInfo, err := h.getUserInfo(application.UserID, authToken)
		if err != nil {
			continue
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
			UserInfo:       userInfo,
		}

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

func (h *ActivityHandler) getUserInfo(userID string, authToken string) (*models.UserInfo, error) {
	return utils.GetUserInfo(userID, authToken)
}

func (h *ActivityHandler) validateActivityRequest(req models.ActivityRequest) error {
	return h.validator.ValidateActivityRequest(req)
}

func (h *ActivityHandler) validateUpdateRequest(req models.ActivityUpdateRequest) error {
	return h.validator.ValidateActivityUpdateRequest(req)
}

func (h *ActivityHandler) parseActivityDates(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	return utils.ParseDateRange(startDateStr, endDateStr)
}

func (h *ActivityHandler) parseSingleDate(dateStr string) (time.Time, error) {
	return utils.ParseDate(dateStr)
}
