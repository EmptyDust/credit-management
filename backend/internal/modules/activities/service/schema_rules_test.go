package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
)

func TestValidateDetailsStripsHiddenFields(t *testing.T) {
	schema := activityrepo.CategorySchema{
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
				Key:  "location",
				Type: "text",
				Rules: activityrepo.FieldRules{
					VisibleWhen: []activityrepo.FieldCondition{
						rawCondition(t, "activity_mode", "eq", `"offline"`),
					},
				},
			},
		},
	}

	details, err := ValidateActivityDetails(schema, map[string]any{
		"activity_mode": "online",
		"location":      "Room 101",
	}, map[string]any{})
	if err != nil {
		t.Fatalf("validate details: %v", err)
	}

	if _, ok := details["location"]; ok {
		t.Fatalf("expected hidden field to be stripped, got %#v", details)
	}
}

func TestValidateDetailsRequiresFieldOnlyWhenVisible(t *testing.T) {
	schema := activityrepo.CategorySchema{
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
						rawCondition(t, "activity_mode", "eq", `"offline"`),
					},
				},
			},
		},
	}

	if _, err := ValidateActivityDetails(schema, map[string]any{
		"activity_mode": "online",
	}, map[string]any{}); err != nil {
		t.Fatalf("expected hidden required field to be ignored, got %v", err)
	}

	_, err := ValidateActivityDetails(schema, map[string]any{
		"activity_mode": "offline",
	}, map[string]any{})
	if !errors.Is(err, ErrActivityDetailsInvalid) {
		t.Fatalf("expected ErrActivityDetailsInvalid when visible required field is missing, got %v", err)
	}
}

func TestValidateDetailsRespectsEnabledWhen(t *testing.T) {
	schema := activityrepo.CategorySchema{
		Fields: []activityrepo.SchemaField{
			{
				Key:  "approval_needed",
				Type: "boolean",
			},
			{
				Key:      "approver",
				Type:     "text",
				Required: true,
				Rules: activityrepo.FieldRules{
					EnabledWhen: []activityrepo.FieldCondition{
						rawCondition(t, "approval_needed", "eq", `true`),
					},
				},
			},
		},
	}

	details, err := ValidateActivityDetails(schema, map[string]any{
		"approval_needed": false,
		"approver":        "teacher-01",
	}, map[string]any{})
	if err != nil {
		t.Fatalf("validate details: %v", err)
	}

	if _, ok := details["approver"]; ok {
		t.Fatalf("expected disabled field to be ignored, got %#v", details)
	}
}

func TestValidateDetailsRejectsInvalidCrossFieldRange(t *testing.T) {
	schema := activityrepo.CategorySchema{
		Fields: []activityrepo.SchemaField{
			{
				Key:  "schedule_guard",
				Type: "boolean",
				Rules: activityrepo.FieldRules{
					Validate: []activityrepo.ValidateRule{
						{
							Type:    "compare",
							Left:    "end_date",
							Op:      "gte",
							Right:   "start_date",
							Message: "end date cannot be earlier than start date",
						},
					},
				},
			},
		},
	}

	_, err := ValidateActivityDetails(schema, map[string]any{}, map[string]any{
		"start_date": "2026-06-01",
		"end_date":   "2026-05-01",
	})
	if !errors.Is(err, ErrActivityDetailsInvalid) {
		t.Fatalf("expected ErrActivityDetailsInvalid, got %v", err)
	}
}

func TestValidateDetailsRejectsUnsupportedFieldType(t *testing.T) {
	schema := activityrepo.CategorySchema{
		Fields: []activityrepo.SchemaField{
			{
				Key:  "attachment",
				Type: "file",
			},
		},
	}

	_, err := ValidateActivityDetails(schema, map[string]any{
		"attachment": "proof.pdf",
	}, map[string]any{})
	if !errors.Is(err, ErrActivitySchemaInvalid) {
		t.Fatalf("expected ErrActivitySchemaInvalid, got %v", err)
	}
}

func TestCreateValidatesAndNormalizesSchemaDetails(t *testing.T) {
	repo := newSchemaAwareTestRepo(t)
	repo.schemas["innovation"] = &activityrepo.CategorySchemaEnvelope{
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
							rawCondition(t, "activity_mode", "eq", `"offline"`),
						},
					},
				},
			},
		},
	}

	svc := New(repo, nil)
	activity, err := svc.Create(context.Background(), "student-01", "student", Input{
		Title:        "Innovation Week",
		Category:     "innovation",
		AcademicYear: "2025-2026",
		Semester:     "spring",
		Details: map[string]any{
			"activity_mode": "online",
			"location":      "Room 101",
		},
	})
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}

	if _, ok := activity.Details["location"]; ok {
		t.Fatalf("expected hidden field removed from returned activity, got %#v", activity.Details)
	}

	saved, err := repo.Get(context.Background(), activity.ID)
	if err != nil {
		t.Fatalf("load saved activity: %v", err)
	}
	if _, ok := saved.Details["location"]; ok {
		t.Fatalf("expected hidden field removed before save, got %#v", saved.Details)
	}
}

func TestUpdateRejectsInvalidCrossFieldRange(t *testing.T) {
	repo := newSchemaAwareTestRepo(t)
	repo.schemas["innovation"] = &activityrepo.CategorySchemaEnvelope{
		Category: activityrepo.CategoryDescriptor{Key: "innovation", Label: "Innovation"},
		Schema: activityrepo.CategorySchema{
			Fields: []activityrepo.SchemaField{
				{
					Key:  "schedule_guard",
					Type: "boolean",
					Rules: activityrepo.FieldRules{
						Validate: []activityrepo.ValidateRule{
							{
								Type:    "compare",
								Left:    "end_date",
								Op:      "gte",
								Right:   "start_date",
								Message: "end date cannot be earlier than start date",
							},
						},
					},
				},
			},
		},
	}
	repo.activities["activity-01"] = &activitydomain.Activity{
		ID:           "activity-01",
		Title:        "Innovation Week",
		Category:     "innovation",
		AcademicYear: "2025-2026",
		Semester:     "spring",
		Status:       activitydomain.StatusDraft,
		OwnerID:      "student-01",
		Details:      map[string]any{"schedule_guard": true},
	}

	svc := New(repo, nil)
	_, err := svc.Update(context.Background(), "student-01", "student", "activity-01", Input{
		Title:        "Innovation Week",
		Category:     "innovation",
		AcademicYear: "2025-2026",
		Semester:     "spring",
		StartDate:    "2026-06-01",
		EndDate:      "2026-05-01",
		Details: map[string]any{
			"schedule_guard": true,
		},
	})
	if !errors.Is(err, ErrActivityDetailsInvalid) {
		t.Fatalf("expected ErrActivityDetailsInvalid, got %v", err)
	}
}

type schemaAwareTestRepo struct {
	*MemoryRepository
	schemas map[string]*activityrepo.CategorySchemaEnvelope
}

func newSchemaAwareTestRepo(t *testing.T) *schemaAwareTestRepo {
	t.Helper()

	return &schemaAwareTestRepo{
		MemoryRepository: NewMemoryRepository(),
		schemas:          map[string]*activityrepo.CategorySchemaEnvelope{},
	}
}

func (r *schemaAwareTestRepo) GetActivityCategorySchema(_ context.Context, activityID string, key string) (*activityrepo.CategorySchemaEnvelope, error) {
	if schema, ok := r.schemas[key]; ok {
		return cloneCategorySchemaEnvelope(schema), nil
	}
	return nil, activityrepo.ErrCategorySchemaNotFound
}

func cloneCategorySchemaEnvelope(schema *activityrepo.CategorySchemaEnvelope) *activityrepo.CategorySchemaEnvelope {
	if schema == nil {
		return nil
	}

	clone := *schema
	clone.Schema = activityrepo.CategorySchema{
		Version: schema.Schema.Version,
		Fields:  append([]activityrepo.SchemaField(nil), schema.Schema.Fields...),
	}
	return &clone
}

func rawCondition(t *testing.T, field string, op string, payload string) activityrepo.FieldCondition {
	t.Helper()

	return activityrepo.FieldCondition{
		Field: field,
		Op:    op,
		Value: mustRawJSON(t, payload),
	}
}

func mustRawJSON(t *testing.T, value string) json.RawMessage {
	t.Helper()

	var payload any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		t.Fatalf("decode raw json %s: %v", value, err)
	}
	return json.RawMessage(value)
}
