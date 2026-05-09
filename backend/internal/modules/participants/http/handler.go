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

type updateCreditsRequest struct {
	Credits float64 `json:"credits"`
}

type batchParticipantIDsRequest struct {
	ParticipantIDs []string `json:"participant_ids"`
}

type batchSetCreditsRequest struct {
	ParticipantIDs []string `json:"participant_ids"`
	Credits        float64  `json:"credits"`
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

func (h *Handler) Stats(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	stats, err := h.service.Stats(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"))
	if err != nil {
		handleParticipantError(c, err, "load participant stats failed")
		return
	}

	response.OK(c, stats)
}

func (h *Handler) Export(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	content, err := h.service.Export(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"))
	if err != nil {
		handleParticipantError(c, err, "export participants failed")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="participants.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", content)
}

func (h *Handler) UpdateCredits(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var payload updateCreditsRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid participant payload", nil)
		return
	}

	participant, err := h.service.UpdateCredits(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), c.Param("participant_id"), payload.Credits)
	if err != nil {
		handleParticipantError(c, err, "update participant failed")
		return
	}

	response.OK(c, gin.H{"participant": participant})
}

func (h *Handler) BatchSetCredits(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var payload batchSetCreditsRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid participant payload", nil)
		return
	}

	affected, err := h.service.BatchSetCredits(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), payload.ParticipantIDs, payload.Credits)
	if err != nil {
		handleParticipantError(c, err, "batch set participant credits failed")
		return
	}

	response.OK(c, gin.H{"affected_count": affected})
}

func (h *Handler) Delete(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	if err := h.service.Delete(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), c.Param("participant_id")); err != nil {
		handleParticipantError(c, err, "delete participant failed")
		return
	}

	response.OK(c, gin.H{})
}

func (h *Handler) BatchDelete(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var payload batchParticipantIDsRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid participant payload", nil)
		return
	}

	affected, err := h.service.BatchDelete(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), payload.ParticipantIDs)
	if err != nil {
		handleParticipantError(c, err, "batch delete participants failed")
		return
	}

	response.OK(c, gin.H{"affected_count": affected})
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
	case errors.Is(err, participantservice.ErrParticipantCreditsInvalid):
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "participant credits must be greater than or equal to 0", nil)
	default:
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", fallback, nil)
	}
}
