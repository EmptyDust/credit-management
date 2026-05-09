package http

import (
	"errors"
	"mime/multipart"
	"net/http"
	"sort"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	activitydomain "credit-management/backend/internal/modules/activities/domain"
	fileservice "credit-management/backend/internal/modules/files/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *fileservice.Service
}

type keyCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
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
		"items":           items,
		"page":            1,
		"page_size":       len(items),
		"total":           len(items),
		"total_pages":     1,
		"total_size":      totalSize(items),
		"category_counts": countsBy(items, func(file fileservice.File) string { return file.FileCategory }),
		"type_counts":     countsBy(items, func(file fileservice.File) string { return file.ContentType }),
	})
}

func (h *Handler) Upload(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	fileHeaders := collectFileHeaders(c)
	if len(fileHeaders) == 0 {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "file is required", nil)
		return
	}

	files, err := h.service.UploadMany(c.Request.Context(), claims.UserID, claims.UserType, c.Param("id"), fileHeaders, fileservice.UploadOptions{
		FileCategory: c.PostForm("file_category"),
		Description:  c.PostForm("description"),
	})
	if err != nil {
		handleFileError(c, err, "upload file failed")
		return
	}

	payload := gin.H{"items": files}
	if len(files) > 0 {
		payload["file"] = files[0]
	}
	response.Success(c, http.StatusCreated, payload)
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
	if file.SignedURL != "" {
		c.Redirect(http.StatusFound, file.SignedURL)
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

func (h *Handler) Preview(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	file, err := h.service.Preview(c.Request.Context(), claims.UserID, claims.UserType, c.Param("file_id"))
	if err != nil {
		handleFileError(c, err, "preview file failed")
		return
	}
	if file.SignedURL != "" {
		c.Redirect(http.StatusFound, file.SignedURL)
		return
	}

	contentType := file.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", `inline; filename="`+file.File.Name+`"`)
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
	case errors.Is(err, fileservice.ErrFilePreviewInvalid):
		response.Error(c, http.StatusUnsupportedMediaType, "FILE_PREVIEW_UNSUPPORTED", "file preview unsupported", nil)
	case errors.Is(err, fileservice.ErrFileValidation):
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "file validation failed", nil)
	default:
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", fallback, nil)
	}
}

func collectFileHeaders(c *gin.Context) []*multipart.FileHeader {
	form, err := c.MultipartForm()
	if err != nil || form == nil || form.File == nil {
		return nil
	}

	files := make([]*multipart.FileHeader, 0)
	for _, fieldName := range []string{"files[]", "files", "file"} {
		files = append(files, form.File[fieldName]...)
	}
	return files
}

func totalSize(items []fileservice.File) int64 {
	var total int64
	for _, item := range items {
		total += item.Size
	}
	return total
}

func countsBy(items []fileservice.File, valueOf func(fileservice.File) string) []keyCount {
	counts := make(map[string]int)
	for _, item := range items {
		value := valueOf(item)
		if value == "" {
			value = "unknown"
		}
		counts[value]++
	}

	values := make([]string, 0, len(counts))
	for value := range counts {
		values = append(values, value)
	}
	sort.Strings(values)

	result := make([]keyCount, 0, len(values))
	for _, value := range values {
		result = append(result, keyCount{Value: value, Count: counts[value]})
	}
	return result
}
