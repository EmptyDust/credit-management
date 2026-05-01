package http

import (
	"net/http"

	"credit-management/backend/internal/httpapi/response"
	optionsservice "credit-management/backend/internal/modules/options/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *optionsservice.Service
}

func NewHandler(service *optionsservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetOrganizationTree(c *gin.Context) {
	items, err := h.service.OrganizationTree(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "load organization tree failed", nil)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetGrades(c *gin.Context) {
	h.getOptionList(c, "grades")
}

func (h *Handler) GetAcademicYears(c *gin.Context) {
	h.getOptionList(c, "academic_years")
}

func (h *Handler) GetSemesters(c *gin.Context) {
	h.getOptionList(c, "semesters")
}

func (h *Handler) GetTeacherTitles(c *gin.Context) {
	h.getOptionList(c, "teacher_titles")
}

func (h *Handler) getOptionList(c *gin.Context, groupKey string) {
	items, err := h.service.OptionList(c.Request.Context(), groupKey)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "load options failed", nil)
		return
	}

	response.OK(c, gin.H{"items": items})
}
