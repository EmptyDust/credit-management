package http

import (
	"net/http"

	"credit-management/backend/internal/httpapi/response"
	dashboard "credit-management/backend/internal/modules/dashboard"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *dashboard.Service
}

func NewHandler(service *dashboard.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Overview(c *gin.Context) {
	response.Success(c, http.StatusOK, h.service.Overview())
}
