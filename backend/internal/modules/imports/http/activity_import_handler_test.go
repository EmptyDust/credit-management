package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"credit-management/backend/internal/httpapi/middleware"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	activityservice "credit-management/backend/internal/modules/activities/service"
	importservice "credit-management/backend/internal/modules/imports/service"
	"credit-management/backend/internal/platform/authn"

	"github.com/gin-gonic/gin"
)

func TestActivityPrecheckReturnsRowValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := importservice.NewActivityImportService(
		activityservice.NewMemoryRepository(),
		handlerActivitySchemaRepo{
			schemas: map[string]*activityrepo.CategorySchemaEnvelope{
				"innovation": handlerInnovationSchema(t),
			},
		},
	)
	handler := NewActivityImportHandler(service)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "activities.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte(strings.Join([]string{
		"title,description,category,academic_year,semester,start_date,end_date,details.activity_mode,details.location",
		",Ignored,innovation,2025-2026,,2026-05-01,2026-05-03,offline,",
	}, "\n"))); err != nil {
		t.Fatalf("write csv content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "admin-01",
		Username: "admin01",
		UserType: "admin",
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/imports/activities/precheck", body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())

	handler.Precheck(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code string `json:"code"`
		Data struct {
			Summary struct {
				InvalidRows int `json:"invalid_rows"`
			} `json:"summary"`
			Rows []struct {
				Status string `json:"status"`
				Errors []struct {
					Field string `json:"field"`
					Code  string `json:"code"`
				} `json:"errors"`
			} `json:"rows"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "OK" {
		t.Fatalf("unexpected response code: %s", payload.Code)
	}
	if payload.Data.Summary.InvalidRows != 1 {
		t.Fatalf("expected 1 invalid row, got %d", payload.Data.Summary.InvalidRows)
	}
	if len(payload.Data.Rows) != 1 || payload.Data.Rows[0].Status != "invalid" {
		t.Fatalf("unexpected row payload: %+v", payload.Data.Rows)
	}
	if !hasHandlerActivityError(payload.Data.Rows[0].Errors, "title", "REQUIRED") {
		t.Fatalf("expected title required error, got %+v", payload.Data.Rows[0].Errors)
	}
	if !hasHandlerActivityError(payload.Data.Rows[0].Errors, "details", "ACTIVITY_DETAILS_INVALID") {
		t.Fatalf("expected details validation error, got %+v", payload.Data.Rows[0].Errors)
	}
}

func TestActivityConfirmReturnsConflictForInvalidPrecheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := importservice.NewActivityImportService(
		activityservice.NewMemoryRepository(),
		handlerActivitySchemaRepo{
			schemas: map[string]*activityrepo.CategorySchemaEnvelope{
				"innovation": handlerInnovationSchema(t),
			},
		},
	)
	precheck, err := service.Precheck(httptest.NewRequest(http.MethodPost, "/", nil).Context(), "admin-01", "activities.csv", "hash-1", []importservice.ActivityImportRow{
		{
			Title:        "",
			Category:     "innovation",
			AcademicYear: "2025-2026",
			Semester:     "spring",
		},
	})
	if err != nil {
		t.Fatalf("precheck: %v", err)
	}

	handler := NewActivityImportHandler(service)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "admin-01",
		Username: "admin01",
		UserType: "admin",
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/imports/activities/confirm", strings.NewReader(`{"job_id":"`+precheck.JobID+`","file_hash":"`+precheck.FileHash+`"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Confirm(ctx)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "IMPORT_CONFIRM_FORBIDDEN" {
		t.Fatalf("unexpected response code: %s", payload.Code)
	}
}

func TestActivityConfirmRejectsMissingJobFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := importservice.NewActivityImportService(
		activityservice.NewMemoryRepository(),
		handlerActivitySchemaRepo{
			schemas: map[string]*activityrepo.CategorySchemaEnvelope{
				"innovation": handlerInnovationSchema(t),
			},
		},
	)
	handler := NewActivityImportHandler(service)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(middleware.AuthClaimsKey, authn.Claims{
		UserID:   "admin-01",
		Username: "admin01",
		UserType: "admin",
	})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/imports/activities/confirm", strings.NewReader(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Confirm(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "VALIDATION_FAILED" {
		t.Fatalf("unexpected response code: %s", payload.Code)
	}
}

func hasHandlerActivityError(errs []struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}, field string, code string) bool {
	for _, err := range errs {
		if err.Field == field && err.Code == code {
			return true
		}
	}
	return false
}

func handlerInnovationSchema(t *testing.T) *activityrepo.CategorySchemaEnvelope {
	t.Helper()

	return &activityrepo.CategorySchemaEnvelope{
		Category: activityrepo.CategoryDescriptor{
			Key:   "innovation",
			Label: "Innovation",
		},
		Schema: activityrepo.CategorySchema{
			Version: 1,
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
	}
}

type handlerActivitySchemaRepo struct {
	schemas map[string]*activityrepo.CategorySchemaEnvelope
}

func (r handlerActivitySchemaRepo) ListCategories(context.Context) ([]activityrepo.CategorySummary, error) {
	return nil, nil
}

func (r handlerActivitySchemaRepo) GetCategorySchema(_ context.Context, key string) (*activityrepo.CategorySchemaEnvelope, error) {
	schema, ok := r.schemas[key]
	if !ok {
		return nil, activityrepo.ErrCategorySchemaNotFound
	}

	clone := *schema
	clone.Schema = activityrepo.CategorySchema{
		Version: schema.Schema.Version,
		Fields:  append([]activityrepo.SchemaField(nil), schema.Schema.Fields...),
	}
	return &clone, nil
}
