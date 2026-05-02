package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"credit-management/backend/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func TestDetailedHealthUsesStructuredDependencyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, DetailedHealthPayload{
		Database: DependencyHealth{Status: "ok"},
		Redis:    DependencyHealth{Status: "degraded", Message: "latency high"},
		Storage:  DependencyHealth{Status: "fail", Message: "root missing"},
		Version:  "test-build",
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.RequestIDKey, "req-test")
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/devtools/health", nil)

	handler.DetailedHealth(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Code string `json:"code"`
		Data struct {
			Database DependencyHealth `json:"database"`
			Redis    DependencyHealth `json:"redis"`
			Storage  DependencyHealth `json:"storage"`
			Version  string           `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload.Code != "OK" {
		t.Fatalf("expected code OK, got %s", payload.Code)
	}
	if payload.Data.Database.Status != "ok" {
		t.Fatalf("expected database status ok, got %s", payload.Data.Database.Status)
	}
	if payload.Data.Redis.Status != "degraded" || payload.Data.Redis.Message != "latency high" {
		t.Fatalf("unexpected redis payload: %+v", payload.Data.Redis)
	}
	if payload.Data.Storage.Status != "fail" || payload.Data.Storage.Message != "root missing" {
		t.Fatalf("unexpected storage payload: %+v", payload.Data.Storage)
	}
	if payload.Data.Version != "test-build" {
		t.Fatalf("expected version test-build, got %s", payload.Data.Version)
	}
}
