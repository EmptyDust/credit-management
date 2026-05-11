package repo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var ErrCategorySchemaNotFound = errors.New("activity category schema not found")

type CategorySchemaReader interface {
	ListCategories(ctx context.Context) ([]CategorySummary, error)
	GetCategorySchema(ctx context.Context, key string) (*CategorySchemaEnvelope, error)
}

type CategorySummary struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

type CategoryDescriptor struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type CategorySchemaEnvelope struct {
	Category CategoryDescriptor `json:"category"`
	Schema   CategorySchema     `json:"schema"`
}

type CategorySchema struct {
	Version int           `json:"version"`
	Fields  []SchemaField `json:"fields"`
}

type SchemaField struct {
	Key          string          `json:"key"`
	Label        string          `json:"label"`
	Type         string          `json:"type"`
	Required     bool            `json:"required"`
	DefaultValue json.RawMessage `json:"default_value"`
	Placeholder  string          `json:"placeholder"`
	Options      []SchemaOption  `json:"options"`
	Rules        FieldRules      `json:"rules"`
	DisplayOrder int             `json:"display_order"`
}

type SchemaOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type FieldCondition struct {
	Field string          `json:"field"`
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value"`
}

type ValidateRule struct {
	Type    string `json:"type"`
	Left    string `json:"left"`
	Op      string `json:"op"`
	Right   string `json:"right"`
	Message string `json:"message"`
}

type FieldRules struct {
	VisibleWhen []FieldCondition `json:"visible_when"`
	EnabledWhen []FieldCondition `json:"enabled_when"`
	Validate    []ValidateRule   `json:"validate"`
}

type CategorySchemaRepo struct {
	db *gorm.DB
}

type categorySummaryRow struct {
	ID        string `gorm:"column:id"`
	Key       string `gorm:"column:key"`
	Label     string `gorm:"column:label"`
	SortOrder int    `gorm:"column:sort_order"`
}

type fieldSchemaRow struct {
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

type optionItemRow struct {
	Value string `gorm:"column:value"`
	Label string `gorm:"column:label"`
}

func (categorySummaryRow) TableName() string {
	return "activity_category_templates"
}

func (fieldSchemaRow) TableName() string {
	return "activity_field_schemas"
}

func NewCategorySchemaGorm(db *gorm.DB) *CategorySchemaRepo {
	return &CategorySchemaRepo{db: db}
}

func (r *CategorySchemaRepo) ListCategories(ctx context.Context) ([]CategorySummary, error) {
	var rows []categorySummaryRow
	if err := r.db.WithContext(ctx).
		Model(&categorySummaryRow{}).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Order("sort_order ASC, key ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]CategorySummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, CategorySummary{
			Key:       row.Key,
			Label:     row.Label,
			SortOrder: row.SortOrder,
		})
	}

	return items, nil
}

func (r *CategorySchemaRepo) GetCategorySchema(ctx context.Context, key string) (*CategorySchemaEnvelope, error) {
	var category categorySummaryRow
	if err := r.db.WithContext(ctx).
		Model(&categorySummaryRow{}).
		Where("key = ? AND enabled = ? AND deleted_at IS NULL", key, true).
		First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategorySchemaNotFound
		}
		return nil, err
	}

	var rows []fieldSchemaRow
	if err := r.db.WithContext(ctx).
		Model(&fieldSchemaRow{}).
		Where("category_id = ? AND deleted_at IS NULL", category.ID).
		Order("display_order ASC, field_key ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	fields := make([]SchemaField, 0, len(rows))
	for _, row := range rows {
		rules, err := decodeFieldRules(row.ValidationRules)
		if err != nil {
			return nil, err
		}
		options, err := r.loadSchemaOptions(ctx, row.OptionsSource)
		if err != nil {
			return nil, err
		}

		fields = append(fields, SchemaField{
			Key:          row.FieldKey,
			Label:        row.FieldLabel,
			Type:         row.FieldType,
			Required:     row.Required,
			DefaultValue: cloneRawMessage(row.DefaultValue),
			Placeholder:  placeholderForField(row),
			Options:      options,
			Rules:        rules,
			DisplayOrder: row.DisplayOrder,
		})
	}

	return &CategorySchemaEnvelope{
		Category: CategoryDescriptor{
			Key:   category.Key,
			Label: category.Label,
		},
		Schema: CategorySchema{
			Version: 1,
			Fields:  fields,
		},
	}, nil
}

func decodeFieldRules(payload json.RawMessage) (FieldRules, error) {
	rules := FieldRules{
		VisibleWhen: []FieldCondition{},
		EnabledWhen: []FieldCondition{},
		Validate:    []ValidateRule{},
	}
	if len(payload) == 0 {
		return rules, nil
	}
	if err := json.Unmarshal(payload, &rules); err != nil {
		return FieldRules{}, err
	}
	if rules.VisibleWhen == nil {
		rules.VisibleWhen = []FieldCondition{}
	}
	if rules.EnabledWhen == nil {
		rules.EnabledWhen = []FieldCondition{}
	}
	if rules.Validate == nil {
		rules.Validate = []ValidateRule{}
	}
	return rules, nil
}

func (r *CategorySchemaRepo) loadSchemaOptions(ctx context.Context, optionsSource *string) ([]SchemaOption, error) {
	if optionsSource == nil {
		return []SchemaOption{}, nil
	}

	groupKey := strings.TrimSpace(*optionsSource)
	if groupKey == "" {
		return []SchemaOption{}, nil
	}

	rows := make([]optionItemRow, 0)
	err := r.db.WithContext(ctx).
		Table("option_items").
		Select("option_items.value", "option_items.label").
		Joins("JOIN option_groups ON option_groups.id = option_items.group_id").
		Where("option_groups.key = ? AND option_groups.scope = ? AND option_groups.deleted_at IS NULL AND option_items.deleted_at IS NULL AND option_items.enabled = ?", groupKey, "system", true).
		Order("option_items.sort_order ASC, option_items.value ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	options := make([]SchemaOption, 0, len(rows))
	for _, row := range rows {
		options = append(options, SchemaOption{
			Value: row.Value,
			Label: row.Label,
		})
	}

	return options, nil
}

func placeholderForField(row fieldSchemaRow) string {
	return row.Placeholder
}

func cloneRawMessage(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return nil
	}
	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}
