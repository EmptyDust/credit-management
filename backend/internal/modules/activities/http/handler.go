package http

import (
	"errors"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	activityservice "credit-management/backend/internal/modules/activities/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *activityservice.Service
}

type transferOwnerRequest struct {
	NewOwnerID string `json:"new_owner_id"`
}

func NewHandler(service *activityservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	items, err := h.service.List(c.Request.Context(), claims.UserID, claims.UserType, activityrepo.ListFilter{
		Keyword:      c.Query("keyword"),
		Status:       c.Query("status"),
		Category:     c.Query("category"),
		AcademicYear: c.Query("academic_year"),
		Semester:     c.Query("semester"),
		OwnerID:      c.Query("owner_id"),
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "list activities failed", nil)
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

func (h *Handler) Create(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var input activityservice.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid activity payload", nil)
		return
	}

	activity, err := h.service.Create(c.Request.Context(), claims.UserID, claims.UserType, input)
	if err != nil {
		handleActivityError(c, err, "create activity failed")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"activity": activity})
}

func (h *Handler) Get(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	activity, err := h.service.Get(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"))
	if err != nil {
		handleActivityError(c, err, "get activity failed")
		return
	}

	response.OK(c, gin.H{"activity": activity})
}

func (h *Handler) Update(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var input activityservice.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid activity payload", nil)
		return
	}

	activity, err := h.service.Update(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), input)
	if err != nil {
		handleActivityError(c, err, "update activity failed")
		return
	}

	response.OK(c, gin.H{"activity": activity})
}

func (h *Handler) Delete(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	if err := h.service.Delete(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id")); err != nil {
		handleActivityError(c, err, "delete activity failed")
		return
	}

	response.OK(c, gin.H{})
}

func (h *Handler) Submit(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	activity, err := h.service.Submit(c.Request.Context(), claims.UserID, c.Param("id"))
	if err != nil {
		handleActivityError(c, err, "submit activity failed")
		return
	}

	response.OK(c, gin.H{"activity": activity})
}

func (h *Handler) Withdraw(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	activity, err := h.service.Withdraw(c.Request.Context(), claims.UserID, c.Param("id"))
	if err != nil {
		handleActivityError(c, err, "withdraw activity failed")
		return
	}

	response.OK(c, gin.H{"activity": activity})
}

func (h *Handler) Review(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var input activityservice.ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid review payload", nil)
		return
	}

	activity, err := h.service.Review(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), input)
	if err != nil {
		handleActivityError(c, err, "review activity failed")
		return
	}

	response.OK(c, gin.H{"activity": activity})
}

func (h *Handler) Copy(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	activity, err := h.service.Copy(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"))
	if err != nil {
		handleActivityError(c, err, "copy activity failed")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"activity": activity})
}

func (h *Handler) TransferOwner(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var input transferOwnerRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid transfer payload", nil)
		return
	}

	activity, err := h.service.TransferOwner(c.Request.Context(), claims.UserType, c.Param("id"), input.NewOwnerID)
	if err != nil {
		handleActivityError(c, err, "transfer owner failed")
		return
	}

	response.OK(c, gin.H{"activity": activity})
}

func handleActivityError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, activitydomain.ErrActivityNotFound):
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "activity not found", nil)
	case errors.Is(err, activitydomain.ErrActivityForbidden):
		response.Error(c, http.StatusForbidden, "ACCESS_FORBIDDEN", "access forbidden", nil)
	case errors.Is(err, activitydomain.ErrActivityStatusInvalid):
		response.Error(c, http.StatusConflict, "ACTIVITY_STATUS_INVALID", "only draft activity can be edited", nil)
	default:
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", fallback, nil)
	}
}
