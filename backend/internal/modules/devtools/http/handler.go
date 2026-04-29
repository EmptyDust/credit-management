package http

import (
	"net/http"

	"credit-management/backend/internal/httpapi/response"
	auditservice "credit-management/backend/internal/modules/audit/service"
	"credit-management/backend/internal/platform/auditlog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	audit *auditservice.Service
}

func NewHandler(audit *auditservice.Service) *Handler {
	return &Handler{audit: audit}
}

func (h *Handler) DetailedHealth(c *gin.Context) {
	response.OK(c, gin.H{
		"database": gin.H{"status": "unknown"},
		"redis":    gin.H{"status": "unknown"},
		"storage":  gin.H{"status": "unknown"},
		"version":  "dev",
	})
}

func (h *Handler) Logs(c *gin.Context) {
	response.OK(c, gin.H{
		"items":       []gin.H{},
		"page":        1,
		"page_size":   0,
		"total":       0,
		"total_pages": 0,
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
