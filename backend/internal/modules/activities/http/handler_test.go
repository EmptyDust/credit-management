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
	saveErr error
}

func (f fakeActivityRepository) List(context.Context, activityrepo.ListFilter) ([]activitydomain.Activity, error) {
	return nil, errors.New("not implemented")
}

func (f fakeActivityRepository) Get(context.Context, string) (*activitydomain.Activity, error) {
	return nil, errors.New("not implemented")
}

func (f fakeActivityRepository) Save(context.Context, *activitydomain.Activity) error {
	return f.saveErr
}

func (f fakeActivityRepository) Delete(context.Context, string) error {
	return errors.New("not implemented")
}
