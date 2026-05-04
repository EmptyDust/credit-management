package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	activityrepo "credit-management/backend/internal/modules/activities/repo"
)

var (
	ErrActivitySchemaInvalid  = errors.New("activity schema invalid")
	ErrActivityDetailsInvalid = errors.New("activity details invalid")
)

type activityCategorySchemaResolver interface {
	GetActivityCategorySchema(ctx context.Context, activityID string, key string) (*activityrepo.CategorySchemaEnvelope, error)
}

type categorySchemaReaderAdapter struct {
	reader activityrepo.CategorySchemaReader
}

func (a categorySchemaReaderAdapter) GetActivityCategorySchema(ctx context.Context, _ string, key string) (*activityrepo.CategorySchemaEnvelope, error) {
	schema, err := a.reader.GetCategorySchema(ctx, key)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, activityrepo.ErrCategorySchemaNotFound
	}
	return schema, nil
}

func (s *Service) UseCategorySchemaReader(reader activityrepo.CategorySchemaReader) {
	if reader == nil {
		s.schemaResolver = nil
		return
	}

	s.schemaResolver = categorySchemaReaderAdapter{reader: reader}
}

func ValidateActivityDetails(schema activityrepo.CategorySchema, details map[string]any, baseFields map[string]any) (map[string]any, error) {
	fieldsByKey := make(map[string]activityrepo.SchemaField, len(schema.Fields))
	for _, field := range schema.Fields {
		if strings.TrimSpace(field.Key) == "" {
			return nil, fmt.Errorf("%w: field key is required", ErrActivitySchemaInvalid)
		}
		if _, exists := fieldsByKey[field.Key]; exists {
			return nil, fmt.Errorf("%w: duplicate field key %q", ErrActivitySchemaInvalid, field.Key)
		}
		if err := validateSchemaField(field); err != nil {
			return nil, err
		}
		fieldsByKey[field.Key] = field
	}

	rawDetails := cloneDetails(details)
	for key := range rawDetails {
		if _, ok := fieldsByKey[key]; !ok {
			return nil, fmt.Errorf("%w: unknown detail field %q", ErrActivityDetailsInvalid, key)
		}
	}

	state := cloneDetails(baseFields)
	for key, value := range rawDetails {
		state[key] = value
	}

	normalized := make(map[string]any, len(rawDetails))
	activeState := cloneDetails(baseFields)
	for _, field := range schema.Fields {
		active, err := isFieldActive(field, state)
		if err != nil {
			return nil, err
		}
		if !active {
			continue
		}

		value, present, err := normalizeFieldValue(field, rawDetails)
		if err != nil {
			return nil, err
		}
		if field.Required && !present {
			return nil, fmt.Errorf("%w: field %q is required", ErrActivityDetailsInvalid, field.Key)
		}
		if present {
			normalized[field.Key] = value
			activeState[field.Key] = value
		}
	}

	for _, field := range schema.Fields {
		active, err := isFieldActive(field, state)
		if err != nil {
			return nil, err
		}
		if !active {
			continue
		}
		if err := validateFieldRules(field, activeState); err != nil {
			return nil, err
		}
	}

	return normalized, nil
}

func (s *Service) validateAndNormalizeDetails(ctx context.Context, activityID string, input Input) (map[string]any, error) {
	if s.schemaResolver == nil {
		return cloneDetails(input.Details), nil
	}

	categoryKey := normalizeInputCategoryKey(input.Category)
	schema, err := s.schemaResolver.GetActivityCategorySchema(ctx, activityID, categoryKey)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, fmt.Errorf("%w: missing category schema for %q", ErrActivitySchemaInvalid, categoryKey)
	}

	return ValidateActivityDetails(schema.Schema, input.Details, activityBaseFields(input))
}

func activityBaseFields(input Input) map[string]any {
	return map[string]any{
		"title":         strings.TrimSpace(input.Title),
		"description":   strings.TrimSpace(input.Description),
		"category":      normalizeInputCategoryKey(input.Category),
		"academic_year": strings.TrimSpace(input.AcademicYear),
		"semester":      strings.TrimSpace(input.Semester),
		"start_date":    strings.TrimSpace(input.StartDate),
		"end_date":      strings.TrimSpace(input.EndDate),
	}
}

func normalizeInputCategoryKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "general"
	}
	return key
}

func validateSchemaField(field activityrepo.SchemaField) error {
	if !isSupportedFieldType(field.Type) {
		return fmt.Errorf("%w: unsupported field type %q", ErrActivitySchemaInvalid, field.Type)
	}
	for _, condition := range field.Rules.VisibleWhen {
		if err := validateFieldCondition(condition); err != nil {
			return err
		}
	}
	for _, condition := range field.Rules.EnabledWhen {
		if err := validateFieldCondition(condition); err != nil {
			return err
		}
	}
	for _, rule := range field.Rules.Validate {
		if err := validateRuleShape(rule); err != nil {
			return err
		}
	}
	return nil
}

func validateFieldCondition(condition activityrepo.FieldCondition) error {
	if strings.TrimSpace(condition.Field) == "" {
		return fmt.Errorf("%w: condition field is required", ErrActivitySchemaInvalid)
	}
	switch condition.Op {
	case "eq", "neq", "in", "not_in", "gt", "gte", "lt", "lte", "contains", "not_empty":
		return nil
	default:
		return fmt.Errorf("%w: unsupported condition operator %q", ErrActivitySchemaInvalid, condition.Op)
	}
}

func validateRuleShape(rule activityrepo.ValidateRule) error {
	if rule.Type != "compare" {
		return fmt.Errorf("%w: unsupported validate rule type %q", ErrActivitySchemaInvalid, rule.Type)
	}
	if strings.TrimSpace(rule.Left) == "" || strings.TrimSpace(rule.Right) == "" {
		return fmt.Errorf("%w: compare rule operands are required", ErrActivitySchemaInvalid)
	}
	switch rule.Op {
	case "eq", "neq", "gt", "gte", "lt", "lte":
		return nil
	default:
		return fmt.Errorf("%w: unsupported compare operator %q", ErrActivitySchemaInvalid, rule.Op)
	}
}

func isSupportedFieldType(fieldType string) bool {
	switch fieldType {
	case "text", "textarea", "number", "date", "select", "radio", "checkbox", "boolean":
		return true
	default:
		return false
	}
}

func isFieldActive(field activityrepo.SchemaField, values map[string]any) (bool, error) {
	visible, err := evaluateConditions(field.Rules.VisibleWhen, values)
	if err != nil {
		return false, err
	}
	if !visible {
		return false, nil
	}

	enabled, err := evaluateConditions(field.Rules.EnabledWhen, values)
	if err != nil {
		return false, err
	}
	return enabled, nil
}

func evaluateConditions(conditions []activityrepo.FieldCondition, values map[string]any) (bool, error) {
	for _, condition := range conditions {
		expected, err := decodeConditionValue(condition.Value)
		if err != nil {
			return false, fmt.Errorf("%w: decode condition value: %v", ErrActivitySchemaInvalid, err)
		}
		matched, err := evaluateCondition(values[condition.Field], condition.Op, expected)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

func decodeConditionValue(payload json.RawMessage) (any, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func evaluateCondition(actual any, op string, expected any) (bool, error) {
	switch op {
	case "eq":
		return valuesEqual(actual, expected), nil
	case "neq":
		return !valuesEqual(actual, expected), nil
	case "in":
		items, ok := toAnySlice(expected)
		if !ok {
			return false, fmt.Errorf("%w: in operator requires an array value", ErrActivitySchemaInvalid)
		}
		return anySliceContains(items, actual), nil
	case "not_in":
		items, ok := toAnySlice(expected)
		if !ok {
			return false, fmt.Errorf("%w: not_in operator requires an array value", ErrActivitySchemaInvalid)
		}
		return !anySliceContains(items, actual), nil
	case "contains":
		return evaluateContains(actual, expected)
	case "not_empty":
		return !isEmptyInputValue(actual), nil
	case "gt", "gte", "lt", "lte":
		return compareOrderedValues(actual, expected, op)
	default:
		return false, fmt.Errorf("%w: unsupported condition operator %q", ErrActivitySchemaInvalid, op)
	}
}

func normalizeFieldValue(field activityrepo.SchemaField, rawDetails map[string]any) (any, bool, error) {
	raw, exists := rawDetails[field.Key]
	if !exists || isEmptyInputValue(raw) {
		return nil, false, nil
	}

	switch field.Type {
	case "text", "textarea":
		value, ok := raw.(string)
		if !ok {
			return nil, false, fmt.Errorf("%w: field %q must be a string", ErrActivityDetailsInvalid, field.Key)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false, nil
		}
		return value, true, nil
	case "date":
		value, ok := raw.(string)
		if !ok {
			return nil, false, fmt.Errorf("%w: field %q must be a date string", ErrActivityDetailsInvalid, field.Key)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false, nil
		}
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return nil, false, fmt.Errorf("%w: field %q must use YYYY-MM-DD", ErrActivityDetailsInvalid, field.Key)
		}
		return parsed.Format("2006-01-02"), true, nil
	case "number":
		number, ok := toFloat64(raw)
		if !ok {
			return nil, false, fmt.Errorf("%w: field %q must be numeric", ErrActivityDetailsInvalid, field.Key)
		}
		return number, true, nil
	case "boolean":
		value, ok := raw.(bool)
		if !ok {
			return nil, false, fmt.Errorf("%w: field %q must be a boolean", ErrActivityDetailsInvalid, field.Key)
		}
		return value, true, nil
	case "select", "radio":
		value, ok := raw.(string)
		if !ok {
			return nil, false, fmt.Errorf("%w: field %q must be a string option", ErrActivityDetailsInvalid, field.Key)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false, nil
		}
		if !allowedOption(field.Options, value) {
			return nil, false, fmt.Errorf("%w: field %q contains an unsupported option", ErrActivityDetailsInvalid, field.Key)
		}
		return value, true, nil
	case "checkbox":
		values, ok := toStringSlice(raw)
		if !ok {
			return nil, false, fmt.Errorf("%w: field %q must be an array of strings", ErrActivityDetailsInvalid, field.Key)
		}
		filtered := make([]string, 0, len(values))
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if !allowedOption(field.Options, value) {
				return nil, false, fmt.Errorf("%w: field %q contains an unsupported option", ErrActivityDetailsInvalid, field.Key)
			}
			filtered = append(filtered, value)
		}
		if len(filtered) == 0 {
			return nil, false, nil
		}
		return filtered, true, nil
	default:
		return nil, false, fmt.Errorf("%w: unsupported field type %q", ErrActivitySchemaInvalid, field.Type)
	}
}

func validateFieldRules(field activityrepo.SchemaField, values map[string]any) error {
	for _, rule := range field.Rules.Validate {
		switch rule.Type {
		case "compare":
			ok, err := evaluateCompareRule(values[rule.Left], values[rule.Right], rule.Op)
			if err != nil {
				return err
			}
			if !ok {
				message := strings.TrimSpace(rule.Message)
				if message == "" {
					message = fmt.Sprintf("%s must be %s %s", rule.Left, rule.Op, rule.Right)
				}
				return fmt.Errorf("%w: %s", ErrActivityDetailsInvalid, message)
			}
		default:
			return fmt.Errorf("%w: unsupported validate rule type %q", ErrActivitySchemaInvalid, rule.Type)
		}
	}
	return nil
}

func evaluateCompareRule(left any, right any, op string) (bool, error) {
	if isEmptyInputValue(left) || isEmptyInputValue(right) {
		return true, nil
	}

	if leftDate, ok := parseDateValue(left); ok {
		rightDate, ok := parseDateValue(right)
		if !ok {
			return false, fmt.Errorf("%w: compare rule operands must both be dates", ErrActivitySchemaInvalid)
		}
		return compareInts(leftDate.Unix(), rightDate.Unix(), op), nil
	}

	leftNumber, ok := toFloat64(left)
	if ok {
		rightNumber, ok := toFloat64(right)
		if !ok {
			return false, fmt.Errorf("%w: compare rule operands must both be numeric", ErrActivitySchemaInvalid)
		}
		return compareFloats(leftNumber, rightNumber, op), nil
	}

	switch op {
	case "eq":
		return valuesEqual(left, right), nil
	case "neq":
		return !valuesEqual(left, right), nil
	default:
		return false, fmt.Errorf("%w: compare rule operands must be numbers or dates for %q", ErrActivitySchemaInvalid, op)
	}
}

func compareOrderedValues(left any, right any, op string) (bool, error) {
	if isEmptyInputValue(left) || isEmptyInputValue(right) {
		return false, nil
	}

	if leftDate, ok := parseDateValue(left); ok {
		rightDate, ok := parseDateValue(right)
		if !ok {
			return false, fmt.Errorf("%w: ordered condition requires comparable operands", ErrActivitySchemaInvalid)
		}
		return compareInts(leftDate.Unix(), rightDate.Unix(), op), nil
	}

	leftNumber, ok := toFloat64(left)
	if ok {
		rightNumber, ok := toFloat64(right)
		if !ok {
			return false, fmt.Errorf("%w: ordered condition requires comparable operands", ErrActivitySchemaInvalid)
		}
		return compareFloats(leftNumber, rightNumber, op), nil
	}

	return false, fmt.Errorf("%w: ordered condition requires numeric or date operands", ErrActivitySchemaInvalid)
}

func compareFloats(left float64, right float64, op string) bool {
	switch op {
	case "gt":
		return left > right
	case "gte":
		return left >= right
	case "lt":
		return left < right
	case "lte":
		return left <= right
	case "eq":
		return left == right
	case "neq":
		return left != right
	default:
		return false
	}
}

func compareInts(left int64, right int64, op string) bool {
	switch op {
	case "gt":
		return left > right
	case "gte":
		return left >= right
	case "lt":
		return left < right
	case "lte":
		return left <= right
	case "eq":
		return left == right
	case "neq":
		return left != right
	default:
		return false
	}
}

func parseDateValue(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(text))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func allowedOption(options []activityrepo.SchemaOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func valuesEqual(left any, right any) bool {
	if leftNumber, ok := toFloat64(left); ok {
		rightNumber, ok := toFloat64(right)
		return ok && leftNumber == rightNumber
	}

	if leftDate, ok := parseDateValue(left); ok {
		rightDate, ok := parseDateValue(right)
		return ok && leftDate.Equal(rightDate)
	}

	switch leftTyped := left.(type) {
	case string:
		rightTyped, ok := right.(string)
		return ok && strings.TrimSpace(leftTyped) == strings.TrimSpace(rightTyped)
	case bool:
		rightTyped, ok := right.(bool)
		return ok && leftTyped == rightTyped
	}

	leftSlice, leftOK := toStringSlice(left)
	rightSlice, rightOK := toStringSlice(right)
	if leftOK && rightOK {
		if len(leftSlice) != len(rightSlice) {
			return false
		}
		for index := range leftSlice {
			if leftSlice[index] != rightSlice[index] {
				return false
			}
		}
		return true
	}

	return left == right
}

func evaluateContains(actual any, expected any) (bool, error) {
	switch actualTyped := actual.(type) {
	case string:
		expectedTyped, ok := expected.(string)
		if !ok {
			return false, fmt.Errorf("%w: contains operator requires a string match value", ErrActivitySchemaInvalid)
		}
		return strings.Contains(actualTyped, expectedTyped), nil
	}

	values, ok := toStringSlice(actual)
	if !ok {
		return false, fmt.Errorf("%w: contains operator requires a string or string array field", ErrActivitySchemaInvalid)
	}
	expectedText, ok := expected.(string)
	if !ok {
		return false, fmt.Errorf("%w: contains operator requires a string match value", ErrActivitySchemaInvalid)
	}
	for _, value := range values {
		if value == expectedText {
			return true, nil
		}
	}
	return false, nil
}

func toFloat64(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func toStringSlice(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...), true
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, false
			}
			items = append(items, text)
		}
		return items, true
	default:
		return nil, false
	}
}

func toAnySlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []string:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, item)
		}
		return items, true
	default:
		return nil, false
	}
}

func anySliceContains(items []any, needle any) bool {
	for _, item := range items {
		if valuesEqual(item, needle) {
			return true
		}
	}
	return false
}

func isEmptyInputValue(value any) bool {
	if value == nil {
		return true
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	case []string:
		return len(typed) == 0
	case []any:
		return len(typed) == 0
	default:
		return false
	}
}
