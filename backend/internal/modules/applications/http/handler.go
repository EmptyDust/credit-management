package http

import (
	"net/http"
	"strconv"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	appservice "credit-management/backend/internal/modules/applications/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *appservice.Service
}

func NewHandler(service *appservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	response.OK(c, h.service.ListFilteredPage(claims.UserID, claims.UserType, queryFilter(c)))
}

func (h *Handler) Stats(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	response.OK(c, h.service.StatsFiltered(claims.UserID, claims.UserType, queryFilter(c)))
}

func (h *Handler) Get(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	item, err := h.service.Get(claims.UserID, claims.UserType, c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "application not found", nil)
		return
	}

	response.OK(c, gin.H{"application": item})
}

func (h *Handler) Export(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	file, err := h.service.ExportFiltered(claims.UserID, claims.UserType, queryFilter(c), c.Query("format"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "export applications failed", nil)
		return
	}

	c.Header("Content-Type", file.ContentType)
	c.Header("Content-Disposition", `attachment; filename="`+file.Filename+`"`)
	c.Data(http.StatusOK, file.ContentType, file.Content)
}

func queryFilter(c *gin.Context) appservice.Filter {
	return appservice.Filter{
		Keyword:      c.Query("keyword"),
		ActivityID:   c.Query("activity_id"),
		UserID:       c.Query("user_id"),
		Status:       c.Query("status"),
		AcademicYear: c.Query("academic_year"),
		Semester:     c.Query("semester"),
		Page:         intQuery(c, "page"),
		PageSize:     intQuery(c, "page_size"),
	}
}

func intQuery(c *gin.Context, key string) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return 0
	}
	return value
}
