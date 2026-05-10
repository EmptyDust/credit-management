package http

import (
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
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
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	overview, err := h.service.OverviewFor(c.Request.Context(), claims.UserID, claims.UserType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "load dashboard overview failed", nil)
		return
	}

	response.Success(c, http.StatusOK, overview)
}
