package http

import (
	"net/http"
	"log/slog"

	"credit-management/backend/internal/httpapi/response"
	auditservice "credit-management/backend/internal/modules/audit/service"
	"credit-management/backend/internal/platform/auditlog"
	"credit-management/backend/internal/platform/health"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	audit   *auditservice.Service
	health  DetailedHealthPayload
	checker *health.Checker
	logger  *slog.Logger
}

type DependencyHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type DetailedHealthPayload struct {
	Database DependencyHealth `json:"database"`
	Redis    DependencyHealth `json:"redis"`
	Storage  DependencyHealth `json:"storage"`
	Version  string           `json:"version"`
}

func NewHandler(audit *auditservice.Service, payload ...DetailedHealthPayload) *Handler {
	handler := &Handler{
		audit: audit,
		health: DetailedHealthPayload{
			Database: DependencyHealth{Status: "unknown"},
			Redis:    DependencyHealth{Status: "unknown"},
			Storage:  DependencyHealth{Status: "unknown"},
			Version:  "dev",
		},
	}
	if len(payload) > 0 {
		handler.health = payload[0]
	}

	return handler
}

func NewHandlerWithHealth(audit *auditservice.Service, logger *slog.Logger, checker *health.Checker) *Handler {
	payload := DetailedHealthPayload{
		Database: DependencyHealth{Status: "unknown"},
		Redis:    DependencyHealth{Status: "unknown"},
		Storage:  DependencyHealth{Status: "unknown"},
		Version:  "dev",
	}
	return &Handler{audit: audit, health: payload, checker: checker, logger: logger}
}

func (h *Handler) DetailedHealth(c *gin.Context) {
	if h.checker != nil {
		result := h.checker.Detailed(c.Request.Context())
		response.OK(c, DetailedHealthPayload{
			Database: mapDependency(result.Database),
			Redis:    mapDependency(result.Redis),
			Storage:  mapDependency(result.Storage),
			Version:  result.Version,
		})
		return
	}
	response.OK(c, h.health)
}

func (h *Handler) Logs(c *gin.Context) {
	response.OK(c, gin.H{
		"items": []gin.H{
			{"level": "info", "message": "runtime log streaming not yet enabled", "source": "backend"},
		},
		"page":        1,
		"page_size":   1,
		"total":       1,
		"total_pages": 1,
	})
}

func (h *Handler) AuditLogs(c *gin.Context) {
	items := h.audit.List(auditlog.Filter{
		ActorID:    c.Query("actor_id"),
		TargetType: c.Query("target_type"),
		TargetID:   c.Query("target_id"),
		Action:     c.Query("action"),
		Result:     c.Query("result"),
	})
	response.OK(c, gin.H{
		"items":       items,
		"page":        1,
		"page_size":   len(items),
		"total":       len(items),
		"total_pages": 1,
	})
}

func (h *Handler) ExportAuditLogs(c *gin.Context) {
	content, err := h.audit.Export(auditlog.Filter{
		ActorID:    c.Query("actor_id"),
		TargetType: c.Query("target_type"),
		TargetID:   c.Query("target_id"),
		Action:     c.Query("action"),
		Result:     c.Query("result"),
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "export audit logs failed", nil)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="audit-logs.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", content)
}

func mapDependency(result health.DependencyResult) DependencyHealth {
	return DependencyHealth{
		Status:  string(result.Status),
		Message: result.Error,
	}
}
