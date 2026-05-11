package repo_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	repo "credit-management/backend/internal/modules/activities/repo"
	"credit-management/backend/internal/platform/database"

	"gorm.io/gorm"
)

func TestCategorySchemaRepoGetsEnabledCategorySchema(t *testing.T) {
	categoryRepo, tx := newTestCategorySchemaRepo(t)
	defer func() { _ = tx.Rollback().Error }()

	const (
		innovationCategoryID = "actcat-innov-0000000000001"
		disabledCategoryID   = "actcat-hidden-0000000000002"
	)

	insertCategoryTemplate(t, tx, categoryTemplateSeed{
		ID:        innovationCategoryID,
		Key:       "innovation",
		Label:     "Innovation",
		SortOrder: 10,
		Enabled:   true,
	})
	insertCategoryTemplate(t, tx, categoryTemplateSeed{
		ID:        disabledCategoryID,
		Key:       "hidden",
		Label:     "Hidden",
		SortOrder: 99,
		Enabled:   false,
	})

	insertFieldSchema(t, tx, fieldSchemaSeed{
		ID:            "actfld-innov-000000000000001",
		CategoryID:    innovationCategoryID,
		FieldKey:      "outcome",
		FieldLabel:    "Outcome",
		FieldType:     "textarea",
		Required:      false,
		DefaultValue:  []byte(`"pending"`),
		Placeholder:   "Describe the expected outcome",
		OptionsSource: nil,
		ValidationRules: []byte(`{
			"visible_when": [{"field":"activity_mode","op":"eq","value":"offline"}],
			"enabled_when": [{"field":"approval_needed","op":"eq","value":true}],
			"validate": [{"type":"compare","left":"end_date","op":"gte","right":"start_date","message":"End date cannot be earlier than start date"}]
		}`),
		DisplayOrder: 20,
	})
	optionSource := "teacher_titles"
	insertFieldSchema(t, tx, fieldSchemaSeed{
		ID:              "actfld-innov-000000000000002",
		CategoryID:      innovationCategoryID,
		FieldKey:        "mentor_title",
		FieldLabel:      "Mentor Title",
		FieldType:       "select",
		Required:        true,
		DefaultValue:    []byte(`"assistant"`),
		Placeholder:     "Choose mentor title",
		OptionsSource:   &optionSource,
		ValidationRules: []byte(`{}`),
		DisplayOrder:    10,
	})
	insertFieldSchema(t, tx, fieldSchemaSeed{
		ID:              "actfld-innov-000000000000003",
		CategoryID:      innovationCategoryID,
		FieldKey:        "location",
		FieldLabel:      "Location",
		FieldType:       "text",
		Required:        true,
		DefaultValue:    []byte(`""`),
		ValidationRules: []byte(`{}`),
		DisplayOrder:    15,
	})

	categories, err := categoryRepo.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if containsCategoryKey(categories, "hidden") {
		t.Fatalf("expected disabled category to be excluded: %#v", categories)
	}
	if !containsCategoryKey(categories, "innovation") {
		t.Fatalf("expected innovation category in list: %#v", categories)
	}

	schema, err := categoryRepo.GetCategorySchema(context.Background(), "innovation")
	if err != nil {
		t.Fatalf("get category schema: %v", err)
	}

	if schema.Category.Key != "innovation" || schema.Category.Label != "Innovation" {
		t.Fatalf("unexpected category metadata: %#v", schema.Category)
	}
	if len(schema.Schema.Fields) != 3 {
		t.Fatalf("expected 3 schema fields, got %d", len(schema.Schema.Fields))
	}
	if schema.Schema.Fields[0].Key != "mentor_title" || schema.Schema.Fields[1].Key != "location" || schema.Schema.Fields[2].Key != "outcome" {
		t.Fatalf("expected fields sorted by display_order, got %#v", schema.Schema.Fields)
	}

	optionField := schema.Schema.Fields[0]
	if optionField.Placeholder != "Choose mentor title" {
		t.Fatalf("expected placeholder from schema storage, got %q", optionField.Placeholder)
	}
	if len(optionField.Options) != 4 {
		t.Fatalf("expected teacher title options from options_source, got %#v", optionField.Options)
	}
	if optionField.Options[0].Value != "assistant" || optionField.Options[0].Label != "Assistant" {
		t.Fatalf("unexpected first option: %#v", optionField.Options[0])
	}
	if optionField.Options[3].Value != "professor" || optionField.Options[3].Label != "Professor" {
		t.Fatalf("unexpected last option: %#v", optionField.Options[3])
	}

	field := schema.Schema.Fields[2]
	if got := string(field.DefaultValue); got != `"pending"` {
		t.Fatalf("expected default value to be preserved, got %s", got)
	}
	if field.Placeholder != "Describe the expected outcome" {
		t.Fatalf("expected placeholder from schema storage, got %q", field.Placeholder)
	}
	if len(field.Rules.VisibleWhen) != 1 || field.Rules.VisibleWhen[0].Field != "activity_mode" || string(field.Rules.VisibleWhen[0].Value) != `"offline"` {
		t.Fatalf("expected visible_when to decode, got %#v", field.Rules.VisibleWhen)
	}
	if len(field.Rules.EnabledWhen) != 1 || field.Rules.EnabledWhen[0].Field != "approval_needed" || string(field.Rules.EnabledWhen[0].Value) != "true" {
		t.Fatalf("expected enabled_when to decode, got %#v", field.Rules.EnabledWhen)
	}
	if len(field.Rules.Validate) != 1 || field.Rules.Validate[0].Type != "compare" || field.Rules.Validate[0].Right != "start_date" {
		t.Fatalf("expected validate rules to decode, got %#v", field.Rules.Validate)
	}
}

func TestGormRepositorySaveReturnsCategoryNotFoundForUnknownCategory(t *testing.T) {
	activityRepo, tx := newTestActivityRepo(t)
	defer func() { _ = tx.Rollback().Error }()

	const missingCategoryKey = "not-a-real-category"
	err := activityRepo.Save(context.Background(), &activitydomain.Activity{
		ID:           "act-missing-category-000001",
		Title:        "Test Activity",
		Category:     missingCategoryKey,
		AcademicYear: "2024-2025",
		Semester:     "spring",
		Status:       activitydomain.StatusDraft,
		OwnerID:      "01JADMIN000000000000000000",
		Details:      map[string]any{},
	})
	if !errors.Is(err, repo.ErrCategorySchemaNotFound) {
		t.Fatalf("expected ErrCategorySchemaNotFound, got %v", err)
	}

	var count int64
	if err := tx.Table("activity_category_templates").Where("key = ?", missingCategoryKey).Count(&count).Error; err != nil {
		t.Fatalf("count category templates: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no category template to be created, got %d", count)
	}
}

func TestGormRepositorySavePreservesExistingDisabledCategoryBinding(t *testing.T) {
	activityRepo, tx := newTestActivityRepo(t)
	defer func() { _ = tx.Rollback().Error }()

	const (
		legacyCategoryID  = "actcat-legacy-0000000000001"
		legacyCategoryKey = "legacy"
		disabledSwitchID  = "actcat-switch-0000000000002"
		disabledSwitchKey = "disabled-switch"
		activityID        = "act-existing-disabled-000001"
	)

	insertCategoryTemplate(t, tx, categoryTemplateSeed{
		ID:        legacyCategoryID,
		Key:       legacyCategoryKey,
		Label:     "Legacy",
		SortOrder: 30,
		Enabled:   true,
	})
	insertCategoryTemplate(t, tx, categoryTemplateSeed{
		ID:        disabledSwitchID,
		Key:       disabledSwitchKey,
		Label:     "Disabled Switch",
		SortOrder: 40,
		Enabled:   false,
	})

	if err := activityRepo.Save(context.Background(), &activitydomain.Activity{
		ID:           activityID,
		Title:        "Existing Activity",
		Category:     legacyCategoryKey,
		AcademicYear: "2024-2025",
		Semester:     "spring",
		Status:       activitydomain.StatusDraft,
		OwnerID:      "01JADMIN000000000000000000",
		Details:      map[string]any{"source": "initial"},
	}); err != nil {
		t.Fatalf("save initial activity: %v", err)
	}

	if err := tx.Model(&categoryTemplateSeed{}).Where("id = ?", legacyCategoryID).Update("enabled", false).Error; err != nil {
		t.Fatalf("disable legacy category: %v", err)
	}

	if err := activityRepo.Save(context.Background(), &activitydomain.Activity{
		ID:           activityID,
		Title:        "Updated Activity",
		Category:     legacyCategoryKey,
		AcademicYear: "2024-2025",
		Semester:     "spring",
		Status:       activitydomain.StatusDraft,
		OwnerID:      "01JADMIN000000000000000000",
		Details:      map[string]any{"source": "updated"},
	}); err != nil {
		t.Fatalf("save with existing disabled binding: %v", err)
	}

	stored := lookupStoredActivity(t, tx, activityID)
	if stored.CategoryID != legacyCategoryID {
		t.Fatalf("expected category_id %s to be preserved, got %s", legacyCategoryID, stored.CategoryID)
	}
	if stored.Title != "Updated Activity" {
		t.Fatalf("expected title update to persist, got %s", stored.Title)
	}

	err := activityRepo.Save(context.Background(), &activitydomain.Activity{
		ID:           activityID,
		Title:        "Updated Activity",
		Category:     disabledSwitchKey,
		AcademicYear: "2024-2025",
		Semester:     "spring",
		Status:       activitydomain.StatusDraft,
		OwnerID:      "01JADMIN000000000000000000",
		Details:      map[string]any{"source": "switch"},
	})
	if !errors.Is(err, repo.ErrCategorySchemaNotFound) {
		t.Fatalf("expected ErrCategorySchemaNotFound when switching to disabled category, got %v", err)
	}

	stored = lookupStoredActivity(t, tx, activityID)
	if stored.CategoryID != legacyCategoryID {
		t.Fatalf("expected failed switch to keep category_id %s, got %s", legacyCategoryID, stored.CategoryID)
	}
}

type categoryTemplateSeed struct {
	ID        string `gorm:"column:id"`
	Key       string `gorm:"column:key"`
	Label     string `gorm:"column:label"`
	SortOrder int    `gorm:"column:sort_order"`
	Enabled   bool   `gorm:"column:enabled"`
}

func (categoryTemplateSeed) TableName() string {
	return "activity_category_templates"
}

type fieldSchemaSeed struct {
	ID              string          `gorm:"column:id"`
	CategoryID      string          `gorm:"column:category_id"`
	FieldKey        string          `gorm:"column:field_key"`
	FieldLabel      string          `gorm:"column:field_label"`
	FieldType       string          `gorm:"column:field_type"`
	Required        bool            `gorm:"column:required"`
	DefaultValue    json.RawMessage `gorm:"column:default_value"`
	Placeholder     string          `gorm:"column:placeholder"`
	OptionsSource   *string         `gorm:"column:options_source"`
	ValidationRules json.RawMessage `gorm:"column:validation_rules"`
	DisplayOrder    int             `gorm:"column:display_order"`
}

func (fieldSchemaSeed) TableName() string {
	return "activity_field_schemas"
}

func newTestCategorySchemaRepo(t *testing.T) (*repo.CategorySchemaRepo, *gorm.DB) {
	t.Helper()

	tx := newTestActivityDB(t)
	return repo.NewCategorySchemaGorm(tx), tx
}

func newTestActivityRepo(t *testing.T) (*repo.GormRepository, *gorm.DB) {
	t.Helper()

	tx := newTestActivityDB(t)
	return repo.NewGorm(tx), tx
}

func newTestActivityDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx := context.Background()
	db, err := database.Open(ctx, database.Options{DSN: dsn})
	if err != nil {
		t.Skipf("test database unavailable: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := database.NewMigrator("../../../../migrations").Apply(ctx, sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if err := database.NewSeeder().Run(ctx, db); err != nil {
		t.Fatalf("seed database: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}

	return tx
}

type storedActivityRow struct {
	ID         string    `gorm:"column:id"`
	Title      string    `gorm:"column:title"`
	CategoryID string    `gorm:"column:category_id"`
	StartDate  time.Time `gorm:"column:start_date"`
	EndDate    time.Time `gorm:"column:end_date"`
}

func (storedActivityRow) TableName() string {
	return "activities"
}

func insertCategoryTemplate(t *testing.T, db *gorm.DB, row categoryTemplateSeed) {
	t.Helper()

	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("insert category template: %v", err)
	}
}

func insertFieldSchema(t *testing.T, db *gorm.DB, row fieldSchemaSeed) {
	t.Helper()

	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("insert field schema: %v", err)
	}
}

func lookupStoredActivity(t *testing.T, db *gorm.DB, activityID string) storedActivityRow {
	t.Helper()

	var row storedActivityRow
	if err := db.First(&row, "id = ?", activityID).Error; err != nil {
		t.Fatalf("load stored activity: %v", err)
	}
	return row
}

func containsCategoryKey(items []repo.CategorySummary, key string) bool {
	for _, item := range items {
		if item.Key == key {
			return true
		}
	}
	return false
}
