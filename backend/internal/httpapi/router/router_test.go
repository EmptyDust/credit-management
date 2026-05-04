package router

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	activityhttp "credit-management/backend/internal/modules/activities/http"
	activityrepo "credit-management/backend/internal/modules/activities/repo"

	"github.com/gin-gonic/gin"
)

func TestActivityCategoryEndpointsRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := New(Dependencies{
		Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		ActivityHandler: activityhttp.NewHandler(nil, routerCategorySchemaRepo{}),
	})

	testCases := []struct {
		name string
		path string
	}{
		{name: "list categories", path: "/api/v1/activity-categories"},
		{name: "category schema", path: "/api/v1/activity-categories/innovation/schema"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)

			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", recorder.Code, recorder.Body.String())
			}

			var payload struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
			}
			if payload.Code != "AUTH_TOKEN_REQUIRED" {
				t.Fatalf("unexpected code: %s", payload.Code)
			}
			if payload.Message != "authorization header is required" {
				t.Fatalf("unexpected message: %s", payload.Message)
			}
		})
	}
}

type routerCategorySchemaRepo struct{}

func (routerCategorySchemaRepo) ListCategories(context.Context) ([]activityrepo.CategorySummary, error) {
	return []activityrepo.CategorySummary{}, nil
}

func (routerCategorySchemaRepo) GetCategorySchema(context.Context, string) (*activityrepo.CategorySchemaEnvelope, error) {
	return &activityrepo.CategorySchemaEnvelope{}, nil
}
