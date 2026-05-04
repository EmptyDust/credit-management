package router

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	activityhttp "credit-management/backend/internal/modules/activities/http"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	importhttp "credit-management/backend/internal/modules/imports/http"
	importservice "credit-management/backend/internal/modules/imports/service"
	activityservice "credit-management/backend/internal/modules/activities/service"
	"credit-management/backend/internal/platform/authn"

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

func TestActivityImportRoutesAllowAuthenticatedNonAdminUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		ImportHandler: importhttp.NewHandler(
			nil,
			importservice.NewActivityImportService(
				activityservice.NewMemoryRepository(),
				routerCategorySchemaRepo{
					schemas: map[string]*activityrepo.CategorySchemaEnvelope{
						"innovation": {
							Category: activityrepo.CategoryDescriptor{Key: "innovation", Label: "Innovation"},
							Schema: activityrepo.CategorySchema{
								Fields: []activityrepo.SchemaField{
									{
										Key:  "activity_mode",
										Type: "radio",
										Options: []activityrepo.SchemaOption{
											{Value: "online", Label: "Online"},
											{Value: "offline", Label: "Offline"},
										},
									},
									{
										Key:      "location",
										Type:     "text",
										Required: true,
										Rules: activityrepo.FieldRules{
											VisibleWhen: []activityrepo.FieldCondition{
												{
													Field: "activity_mode",
													Op:    "eq",
													Value: json.RawMessage(`"offline"`),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			),
		),
	})

	token, err := authn.IssueAccessToken(authn.ClaimsInput{
		UserID:   "teacher-01",
		Username: "teacher01",
		UserType: "teacher",
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "activities.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("title,description,category,academic_year,semester,start_date,end_date,details.activity_mode,details.location\nImported Activity,Desc,innovation,2025-2026,spring,2026-05-01,2026-05-03,offline,Room 101\n")); err != nil {
		t.Fatalf("write csv content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/imports/activities/precheck", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)

	engine.ServeHTTP(recorder, request)

	if recorder.Code == http.StatusForbidden {
		t.Fatalf("expected route to allow authenticated non-admin users, got 403: %s", recorder.Body.String())
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

type routerCategorySchemaRepo struct{
	schemas map[string]*activityrepo.CategorySchemaEnvelope
}

func (routerCategorySchemaRepo) ListCategories(context.Context) ([]activityrepo.CategorySummary, error) {
	return []activityrepo.CategorySummary{}, nil
}

func (r routerCategorySchemaRepo) GetCategorySchema(_ context.Context, key string) (*activityrepo.CategorySchemaEnvelope, error) {
	if schema, ok := r.schemas[key]; ok {
		clone := *schema
		clone.Schema = activityrepo.CategorySchema{
			Version: schema.Schema.Version,
			Fields:  append([]activityrepo.SchemaField(nil), schema.Schema.Fields...),
		}
		return &clone, nil
	}
	return &activityrepo.CategorySchemaEnvelope{}, nil
}
