package http

import (
	"errors"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	activitydomain "credit-management/backend/internal/modules/activities/domain"
	participantservice "credit-management/backend/internal/modules/participants/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *participantservice.Service
}

type addParticipantsRequest struct {
	UserIDs []string `json:"user_ids"`
	Credits float64  `json:"credits"`
}

func NewHandler(service *participantservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	items, err := h.service.List(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"))
	if err != nil {
		handleParticipantError(c, err, "list participants failed")
		return
	}

	response.OK(c, gin.H{
		"items":       items,
		"page":        1,
		"page_size":   len(items),
		"total":       len(items),
		"total_pages": 1,
	})
}

func (h *Handler) Add(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var payload addParticipantsRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid participant payload", nil)
		return
	}

	items, err := h.service.Add(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), payload.UserIDs, payload.Credits)
	if err != nil {
		handleParticipantError(c, err, "add participants failed")
		return
	}

	response.OK(c, gin.H{"items": items})
}

func (h *Handler) Leave(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	if err := h.service.Leave(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id")); err != nil {
		handleParticipantError(c, err, "leave activity failed")
		return
	}

	response.OK(c, gin.H{})
}

func handleParticipantError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, activitydomain.ErrActivityNotFound), errors.Is(err, participantservice.ErrParticipantNotFound):
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "participant resource not found", nil)
	case errors.Is(err, activitydomain.ErrActivityForbidden), errors.Is(err, participantservice.ErrParticipantVisibilityDenied), errors.Is(err, participantservice.ErrParticipantEditForbidden):
		response.Error(c, http.StatusForbidden, "ACCESS_FORBIDDEN", "access forbidden", nil)
	case errors.Is(err, activitydomain.ErrActivityStatusInvalid):
		response.Error(c, http.StatusConflict, "ACTIVITY_STATUS_INVALID", "only draft activity can be edited", nil)
	case errors.Is(err, participantservice.ErrParticipantDuplicate):
		response.Error(c, http.StatusConflict, "PARTICIPANT_DUPLICATE", "participant already exists", nil)
	case errors.Is(err, participantservice.ErrParticipantOnlyStudent):
		response.Error(c, http.StatusConflict, "PARTICIPANT_ONLY_STUDENT_ALLOWED", "only students can join activity", nil)
	default:
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", fallback, nil)
	}
}
