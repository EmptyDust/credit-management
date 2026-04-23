package http

import (
	"errors"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	activitydomain "credit-management/backend/internal/modules/activities/domain"
	fileservice "credit-management/backend/internal/modules/files/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *fileservice.Service
}

func NewHandler(service *fileservice.Service) *Handler {
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
		handleFileError(c, err, "list files failed")
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

func (h *Handler) Upload(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "file is required", nil)
		return
	}

	file, err := h.service.Upload(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), fileHeader)
	if err != nil {
		handleFileError(c, err, "upload file failed")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"file": file})
}

func (h *Handler) Delete(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	if err := h.service.Delete(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), c.Param("file_id")); err != nil {
		handleFileError(c, err, "delete file failed")
		return
	}

	response.OK(c, gin.H{})
}

func (h *Handler) Download(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	file, err := h.service.Download(c.Request.Context(), claims.UserID, claims.UserType, c.Param("file_id"))
	if err != nil {
		handleFileError(c, err, "download file failed")
		return
	}

	contentType := file.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", `attachment; filename="`+file.File.Name+`"`)
	c.Data(http.StatusOK, contentType, file.Content)
}

func handleFileError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, activitydomain.ErrActivityNotFound), errors.Is(err, fileservice.ErrFileNotFound):
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "file resource not found", nil)
	case errors.Is(err, activitydomain.ErrActivityStatusInvalid):
		response.Error(c, http.StatusConflict, "ACTIVITY_STATUS_INVALID", "only draft activity can be edited", nil)
	case errors.Is(err, activitydomain.ErrActivityForbidden), errors.Is(err, fileservice.ErrFileAccessDenied):
		response.Error(c, http.StatusForbidden, "FILE_ACCESS_FORBIDDEN", "file access forbidden", nil)
	default:
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", fallback, nil)
	}
}
