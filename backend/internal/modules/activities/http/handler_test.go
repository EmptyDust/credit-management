package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"credit-management/backend/internal/httpapi/middleware"
	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	activityservice "credit-management/backend/internal/modules/activities/service"
	"credit-management/backend/internal/modules/dashboard"
	"credit-management/backend/internal/platform/authn"

	"github.com/gin-gonic/gin"
)

func TestListCategoriesReturnsCategoryItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, fakeCategorySchemaRepo{
		categories: []activityrepo.CategorySummary{
			{Key: "innovation", Label: "Innovation", SortOrder: 10},
			{Key: "research", Label: "Research", SortOrder: 20},
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/activity-categories", nil)

	handler.ListCategories(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code string `json:"code"`
		Data struct {
			Items []activityrepo.CategorySummary `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "OK" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if len(payload.Data.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(payload.Data.Items))
	}
	if payload.Data.Items[0].Key != "innovation" || payload.Data.Items[0].SortOrder != 10 {
		t.Fatalf("unexpected first item: %#v", payload.Data.Items[0])
	}
}

func TestGetCategorySchemaReturnsSchemaPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, fakeCategorySchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": {
				Category: activityrepo.CategoryDescriptor{
					Key:   "innovation",
					Label: "Innovation",
				},
				Schema: activityrepo.CategorySchema{
					Version: 1,
					Fields: []activityrepo.SchemaField{
						{
							Key:          "mentor_title",
							Label:        "Mentor Title",
							Type:         "select",
							Required:     true,
							DefaultValue: json.RawMessage(`"assistant"`),
							Placeholder:  "",
							Options: []activityrepo.SchemaOption{
								{Value: "assistant", Label: "Assistant"},
								{Value: "lecturer", Label: "Lecturer"},
							},
							Rules:        activityrepo.FieldRules{},
							DisplayOrder: 10,
						},
					},
				},
			},
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "category", Value: "innovation"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/activity-categories/innovation/schema", nil)

	handler.GetCategorySchema(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code string `json:"code"`
		Data struct {
			Category activityrepo.CategoryDescriptor `json:"category"`
			Schema   activityrepo.CategorySchema     `json:"schema"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "OK" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Data.Category.Key != "innovation" || payload.Data.Category.Label != "Innovation" {
		t.Fatalf("unexpected category payload: %#v", payload.Data.Category)
	}
	if len(payload.Data.Schema.Fields) != 1 {
		t.Fatalf("expected 1 schema field, got %d", len(payload.Data.Schema.Fields))
	}
	field := payload.Data.Schema.Fields[0]
	if field.Key != "mentor_title" || len(field.Options) != 2 {
		t.Fatalf("unexpected schema field payload: %#v", field)
	}
}

func TestGetCategorySchemaReturnsNotFoundForMissingCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, fakeCategorySchemaRepo{
		schemaErr: activityrepo.ErrCategorySchemaNotFound,
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "category", Value: "missing"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/activity-categories/missing/schema", nil)

	handler.GetCategorySchema(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "ACTIVITY_CATEGORY_NOT_FOUND" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Message != "activity category not found" {
		t.Fatalf("unexpected message: %s", payload.Message)
	}
}

func TestCreateReturnsCategoryNotFoundWhenWriteCategoryIsUnknown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := activityservice.New(fakeActivityRepository{
		saveErr: activityrepo.ErrCategorySchemaNotFound,
	}, nil)
	handler := NewHandler(service, fakeCategorySchemaRepo{})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "student-01",
		Username: "student01",
		UserType: "student",
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/activities", strings.NewReader(`{
		"title":"Test Activity",
		"category":"missing-category",
		"academic_year":"2024-2025",
		"semester":"spring",
		"details":{}
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Create(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "ACTIVITY_CATEGORY_NOT_FOUND" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Message != "activity category not found" {
		t.Fatalf("unexpected message: %s", payload.Message)
	}
}

func TestCreateReturnsDetailsInvalidForSchemaValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	schemaRepo := fakeCategorySchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": {
				Category: activityrepo.CategoryDescriptor{Key: "innovation", Label: "Innovation"},
				Schema: activityrepo.CategorySchema{
					Fields: []activityrepo.SchemaField{
						{
							Key:      "location",
							Type:     "text",
							Required: true,
						},
					},
				},
			},
		},
	}
	service := activityservice.New(fakeActivityRepository{}, nil)
	service.UseCategorySchemaReader(schemaRepo)
	handler := NewHandler(service, schemaRepo)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "student-01",
		Username: "student01",
		UserType: "student",
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/activities", strings.NewReader(`{
		"title":"Test Activity",
		"category":"innovation",
		"academic_year":"2024-2025",
		"semester":"spring",
		"details":{}
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Create(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "ACTIVITY_DETAILS_INVALID" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Message != "activity details are invalid" {
		t.Fatalf("unexpected message: %s", payload.Message)
	}
}

func TestCreateReturnsSchemaInvalidForUnsupportedFieldType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	schemaRepo := fakeCategorySchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": {
				Category: activityrepo.CategoryDescriptor{Key: "innovation", Label: "Innovation"},
				Schema: activityrepo.CategorySchema{
					Fields: []activityrepo.SchemaField{
						{
							Key:  "attachment",
							Type: "file",
						},
					},
				},
			},
		},
	}
	service := activityservice.New(fakeActivityRepository{}, nil)
	service.UseCategorySchemaReader(schemaRepo)
	handler := NewHandler(service, schemaRepo)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "student-01",
		Username: "student01",
		UserType: "student",
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/activities", strings.NewReader(`{
		"title":"Test Activity",
		"category":"innovation",
		"academic_year":"2024-2025",
		"semester":"spring",
		"details":{"attachment":"proof.pdf"}
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Create(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "ACTIVITY_SCHEMA_INVALID" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Message != "activity schema is invalid" {
		t.Fatalf("unexpected message: %s", payload.Message)
	}
}

func TestUpdateReturnsDetailsInvalidForSchemaValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	schemaRepo := fakeCategorySchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": {
				Category: activityrepo.CategoryDescriptor{Key: "innovation", Label: "Innovation"},
				Schema: activityrepo.CategorySchema{
					Fields: []activityrepo.SchemaField{
						{
							Key:      "location",
							Type:     "text",
							Required: true,
						},
					},
				},
			},
		},
	}
	service := activityservice.New(fakeActivityRepository{
		activity: &activitydomain.Activity{
			ID:           "activity-01",
			Title:        "Innovation Week",
			Category:     "innovation",
			AcademicYear: "2024-2025",
			Semester:     "spring",
			Status:       activitydomain.StatusDraft,
			OwnerID:      "student-01",
			Details:      map[string]any{},
		},
	}, nil)
	service.UseCategorySchemaReader(schemaRepo)
	handler := NewHandler(service, schemaRepo)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "activity-01"}}
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "student-01",
		Username: "student01",
		UserType: "student",
	})
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/activities/activity-01", strings.NewReader(`{
		"title":"Test Activity",
		"category":"innovation",
		"academic_year":"2024-2025",
		"semester":"spring",
		"details":{}
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Update(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "ACTIVITY_DETAILS_INVALID" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Message != "activity details are invalid" {
		t.Fatalf("unexpected message: %s", payload.Message)
	}
}

func TestUpdateReturnsSchemaInvalidForUnsupportedFieldType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	schemaRepo := fakeCategorySchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": {
				Category: activityrepo.CategoryDescriptor{Key: "innovation", Label: "Innovation"},
				Schema: activityrepo.CategorySchema{
					Fields: []activityrepo.SchemaField{
						{
							Key:  "attachment",
							Type: "file",
						},
					},
				},
			},
		},
	}
	service := activityservice.New(fakeActivityRepository{
		activity: &activitydomain.Activity{
			ID:           "activity-01",
			Title:        "Innovation Week",
			Category:     "innovation",
			AcademicYear: "2024-2025",
			Semester:     "spring",
			Status:       activitydomain.StatusDraft,
			OwnerID:      "student-01",
			Details:      map[string]any{},
		},
	}, nil)
	service.UseCategorySchemaReader(schemaRepo)
	handler := NewHandler(service, schemaRepo)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "activity-01"}}
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "student-01",
		Username: "student01",
		UserType: "student",
	})
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/activities/activity-01", strings.NewReader(`{
		"title":"Test Activity",
		"category":"innovation",
		"academic_year":"2024-2025",
		"semester":"spring",
		"details":{"attachment":"proof.pdf"}
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Update(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "ACTIVITY_SCHEMA_INVALID" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Message != "activity schema is invalid" {
		t.Fatalf("unexpected message: %s", payload.Message)
	}
}

func TestStatsReturnsActivityStatsPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, fakeCategorySchemaRepo{}).
		UseStatsService(fakeActivityStatsService{
			stats: dashboard.ActivityStats{
				StatusCounts: map[string]int{
					activitydomain.StatusDraft:    2,
					activitydomain.StatusApproved: 3,
				},
				CategoryCounts: []dashboard.KeyLabelCount{
					{Key: "innovation", Label: "Innovation", Count: 3},
				},
				AcademicYearCounts: []dashboard.ValueCount{
					{Value: "2025-2026", Count: 3},
				},
				SemesterCounts: []dashboard.ValueCount{
					{Value: "autumn", Count: 3},
				},
				MineCount:         1,
				ParticipatedCount: 2,
			},
		})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "teacher-01",
		Username: "teacher01",
		UserType: "teacher",
	})
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/activities/stats?category=innovation", nil)

	handler.Stats(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code string                  `json:"code"`
		Data dashboard.ActivityStats `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "OK" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Data.MineCount != 1 || payload.Data.ParticipatedCount != 2 {
		t.Fatalf("unexpected stats payload: %#v", payload.Data)
	}
	if payload.Data.StatusCounts[activitydomain.StatusApproved] != 3 {
		t.Fatalf("unexpected status counts: %#v", payload.Data.StatusCounts)
	}
}

func TestExportReturnsCSVAttachment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, fakeCategorySchemaRepo{}).
		UseExportService(fakeActivityExportService{
			content: []byte("title,category\nExpo,innovation\n"),
		})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "teacher-01",
		Username: "teacher01",
		UserType: "teacher",
	})
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/activities/export?category=innovation", nil)

	handler.Export(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", got)
	}
	if got := recorder.Header().Get("Content-Disposition"); got != `attachment; filename="activities.csv"` {
		t.Fatalf("unexpected content disposition: %s", got)
	}
	if recorder.Body.String() != "title,category\nExpo,innovation\n" {
		t.Fatalf("unexpected csv body: %q", recorder.Body.String())
	}
}

type fakeCategorySchemaRepo struct {
	categories []activityrepo.CategorySummary
	schemas    map[string]*activityrepo.CategorySchemaEnvelope
	listErr    error
	schemaErr  error
}

func (f fakeCategorySchemaRepo) ListCategories(context.Context) ([]activityrepo.CategorySummary, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.categories, nil
}

func (f fakeCategorySchemaRepo) GetCategorySchema(_ context.Context, key string) (*activityrepo.CategorySchemaEnvelope, error) {
	if f.schemaErr != nil {
		return nil, f.schemaErr
	}
	return f.schemas[key], nil
}

type fakeActivityRepository struct {
	activity *activitydomain.Activity
	saveErr  error
}

type fakeActivityStatsService struct {
	stats dashboard.ActivityStats
	err   error
}

func (f fakeActivityStatsService) Stats(context.Context, string, string, activityrepo.ListFilter) (dashboard.ActivityStats, error) {
	return f.stats, f.err
}

type fakeActivityExportService struct {
	content []byte
	err     error
}

func (f fakeActivityExportService) Export(context.Context, string, string, activityrepo.ListFilter) ([]byte, error) {
	return f.content, f.err
}

func (f fakeActivityRepository) List(context.Context, activityrepo.ListFilter) ([]activitydomain.Activity, error) {
	return nil, errors.New("not implemented")
}

func (f fakeActivityRepository) Get(context.Context, string) (*activitydomain.Activity, error) {
	if f.activity == nil {
		return nil, errors.New("not implemented")
	}
	clone := *f.activity
	clone.Details = cloneTestDetails(f.activity.Details)
	return &clone, nil
}

func (f fakeActivityRepository) Save(context.Context, *activitydomain.Activity) error {
	return f.saveErr
}

func (f fakeActivityRepository) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func cloneTestDetails(source map[string]any) map[string]any {
	if source == nil {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
