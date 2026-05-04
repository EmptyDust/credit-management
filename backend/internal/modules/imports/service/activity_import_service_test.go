package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	activityservice "credit-management/backend/internal/modules/activities/service"
)

func TestActivityImportPrecheckReportsFixedAndDynamicValidationErrors(t *testing.T) {
	activityRepo := activityservice.NewMemoryRepository()
	schemaRepo := activityImportSchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": activityImportInnovationSchema(t),
		},
	}

	svc := NewActivityImportService(activityRepo, schemaRepo)

	result, err := svc.Precheck(context.Background(), "admin-01", "activities.csv", "hash-1", []ActivityImportRow{
		{
			Title:        "",
			Category:     "innovation",
			AcademicYear: "2025-2026",
			Semester:     "",
			Details: map[string]any{
				"activity_mode": "offline",
			},
		},
		{
			Title:        "Unknown Category",
			Category:     "missing",
			AcademicYear: "2025-2026",
			Semester:     "spring",
		},
		{
			Title:        "Innovation Week",
			Category:     "innovation",
			AcademicYear: "2025-2026",
			Semester:     "spring",
			Details: map[string]any{
				"activity_mode": "onsite",
			},
		},
		{
			Title:        "Innovation Week",
			Category:     "innovation",
			AcademicYear: "2025-2026",
			Semester:     "spring",
			Details: map[string]any{
				"activity_mode": "offline",
				"location":      "Room 101",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected precheck to succeed, got %v", err)
	}

	if result.JobType != "activities" {
		t.Fatalf("expected job type activities, got %s", result.JobType)
	}
	if result.Summary.TotalRows != 4 {
		t.Fatalf("expected 4 total rows, got %d", result.Summary.TotalRows)
	}
	if result.Summary.ValidRows != 1 {
		t.Fatalf("expected 1 valid row, got %d", result.Summary.ValidRows)
	}
	if result.Summary.InvalidRows != 3 {
		t.Fatalf("expected 3 invalid rows, got %d", result.Summary.InvalidRows)
	}

	if result.Rows[0].Status != "invalid" {
		t.Fatalf("expected first row invalid, got %s", result.Rows[0].Status)
	}
	assertActivityImportError(t, result.Rows[0].Errors, "title", "REQUIRED")
	assertActivityImportError(t, result.Rows[0].Errors, "semester", "REQUIRED")
	assertActivityImportError(t, result.Rows[0].Errors, "details", "ACTIVITY_DETAILS_INVALID")

	if result.Rows[1].Status != "invalid" {
		t.Fatalf("expected second row invalid, got %s", result.Rows[1].Status)
	}
	assertActivityImportError(t, result.Rows[1].Errors, "category", "ACTIVITY_CATEGORY_NOT_FOUND")

	if result.Rows[2].Status != "invalid" {
		t.Fatalf("expected third row invalid, got %s", result.Rows[2].Status)
	}
	assertActivityImportError(t, result.Rows[2].Errors, "details", "ACTIVITY_DETAILS_INVALID")

	if result.Rows[3].Status != "valid" {
		t.Fatalf("expected fourth row valid, got %s", result.Rows[3].Status)
	}
}

func TestActivityImportConfirmIsIdempotent(t *testing.T) {
	activityRepo := activityservice.NewMemoryRepository()
	schemaRepo := activityImportSchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": activityImportInnovationSchema(t),
		},
	}

	svc := NewActivityImportService(activityRepo, schemaRepo)

	precheck, err := svc.Precheck(context.Background(), "admin-01", "activities.csv", "hash-1", []ActivityImportRow{
		{
			Title:        "Innovation Week",
			Description:  "A schema-aware import row",
			Category:     "innovation",
			AcademicYear: "2025-2026",
			Semester:     "spring",
			StartDate:    "2026-05-01",
			EndDate:      "2026-05-03",
			Details: map[string]any{
				"activity_mode": "offline",
				"location":      "Room 101",
			},
		},
	})
	if err != nil {
		t.Fatalf("precheck: %v", err)
	}

	first, err := svc.Confirm(context.Background(), "admin-01", precheck.JobID, precheck.FileHash)
	if err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	second, err := svc.Confirm(context.Background(), "admin-01", precheck.JobID, precheck.FileHash)
	if err != nil {
		t.Fatalf("second confirm: %v", err)
	}

	if first.JobID != second.JobID || first.Stage != second.Stage {
		t.Fatalf("expected stable confirm result, got first=%+v second=%+v", first, second)
	}
	if len(first.CreatedActivityIDs) != 1 {
		t.Fatalf("expected 1 created activity id, got %+v", first.CreatedActivityIDs)
	}
	if strings.Join(first.CreatedActivityIDs, ",") != strings.Join(second.CreatedActivityIDs, ",") {
		t.Fatalf("expected stable created ids, got first=%v second=%v", first.CreatedActivityIDs, second.CreatedActivityIDs)
	}

	items, err := activityRepo.List(context.Background(), activityrepo.ListFilter{})
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 persisted activity, got %d", len(items))
	}
	if items[0].Status != activitydomain.StatusDraft {
		t.Fatalf("expected draft activity, got %s", items[0].Status)
	}
	if items[0].OwnerID != "admin-01" {
		t.Fatalf("expected owner admin-01, got %s", items[0].OwnerID)
	}
}

func TestActivityImportConfirmRollsBackPartialCreatesBeforeRetry(t *testing.T) {
	activityRepo := &flakyActivityImportRepo{
		MemoryRepository: activityservice.NewMemoryRepository(),
		failAtCall:       2,
	}
	schemaRepo := activityImportSchemaRepo{
		schemas: map[string]*activityrepo.CategorySchemaEnvelope{
			"innovation": activityImportInnovationSchema(t),
		},
	}

	svc := NewActivityImportService(activityRepo, schemaRepo)
	precheck, err := svc.Precheck(context.Background(), "admin-01", "activities.csv", "hash-1", []ActivityImportRow{
		{
			Title:        "Innovation Week 1",
			Category:     "innovation",
			AcademicYear: "2025-2026",
			Semester:     "spring",
			Details: map[string]any{
				"activity_mode": "offline",
				"location":      "Room 101",
			},
		},
		{
			Title:        "Innovation Week 2",
			Category:     "innovation",
			AcademicYear: "2025-2026",
			Semester:     "spring",
			Details: map[string]any{
				"activity_mode": "offline",
				"location":      "Room 102",
			},
		},
	})
	if err != nil {
		t.Fatalf("precheck: %v", err)
	}

	if _, err := svc.Confirm(context.Background(), "admin-01", precheck.JobID, precheck.FileHash); !errors.Is(err, errFlakyActivityImportSave) {
		t.Fatalf("expected flaky save error, got %v", err)
	}

	items, err := activityRepo.List(context.Background(), activityrepo.ListFilter{})
	if err != nil {
		t.Fatalf("list after rollback: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected rollback to leave no persisted activities, got %d", len(items))
	}

	activityRepo.failAtCall = 0
	second, err := svc.Confirm(context.Background(), "admin-01", precheck.JobID, precheck.FileHash)
	if err != nil {
		t.Fatalf("second confirm: %v", err)
	}
	if len(second.CreatedActivityIDs) != 2 {
		t.Fatalf("expected 2 created activity ids after retry, got %+v", second.CreatedActivityIDs)
	}

	items, err = activityRepo.List(context.Background(), activityrepo.ListFilter{})
	if err != nil {
		t.Fatalf("list after retry: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 persisted activities after retry, got %d", len(items))
	}
}

func assertActivityImportError(t *testing.T, errs []ValidationError, field string, code string) {
	t.Helper()

	for _, err := range errs {
		if err.Field == field && err.Code == code {
			return
		}
	}

	t.Fatalf("expected error %s/%s in %+v", field, code, errs)
}

func activityImportInnovationSchema(t *testing.T) *activityrepo.CategorySchemaEnvelope {
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
								Value: mustActivityImportRawJSON(t, `"offline"`),
							},
						},
					},
				},
			},
		},
	}
}

func mustActivityImportRawJSON(t *testing.T, value string) json.RawMessage {
	t.Helper()

	var payload any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		t.Fatalf("decode raw json %s: %v", value, err)
	}
	return json.RawMessage(value)
}

type activityImportSchemaRepo struct {
	schemas map[string]*activityrepo.CategorySchemaEnvelope
}

var errFlakyActivityImportSave = errors.New("flaky activity import save")

type flakyActivityImportRepo struct {
	*activityservice.MemoryRepository
	saveCalls  int
	failAtCall int
}

func (r activityImportSchemaRepo) ListCategories(context.Context) ([]activityrepo.CategorySummary, error) {
	return nil, nil
}

func (r activityImportSchemaRepo) GetCategorySchema(_ context.Context, key string) (*activityrepo.CategorySchemaEnvelope, error) {
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

func (r *flakyActivityImportRepo) Save(ctx context.Context, activity *activitydomain.Activity) error {
	r.saveCalls++
	if r.failAtCall > 0 && r.saveCalls == r.failAtCall {
		return errFlakyActivityImportSave
	}
	return r.MemoryRepository.Save(ctx, activity)
}
