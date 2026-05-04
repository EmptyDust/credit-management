package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	activityrepo "credit-management/backend/internal/modules/activities/repo"
	activityservice "credit-management/backend/internal/modules/activities/service"
	"credit-management/backend/internal/platform/idgen"
)

type ActivityImportRow struct {
	Title        string         `json:"title"`
	Description  string         `json:"description,omitempty"`
	Category     string         `json:"category"`
	AcademicYear string         `json:"academic_year"`
	Semester     string         `json:"semester"`
	StartDate    string         `json:"start_date,omitempty"`
	EndDate      string         `json:"end_date,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
}

type ActivityRowResult struct {
	RowNo   int               `json:"row_no"`
	Status  string            `json:"status"`
	Payload ActivityImportRow `json:"payload"`
	Errors  []ValidationError `json:"errors"`
}

type ActivityPrecheckResult struct {
	JobID     string              `json:"job_id"`
	JobType   string              `json:"job_type"`
	FileName  string              `json:"file_name"`
	FileHash  string              `json:"file_hash"`
	ExpiresAt time.Time           `json:"expires_at"`
	Summary   Summary             `json:"summary"`
	Rows      []ActivityRowResult `json:"rows"`
}

type ActivityConfirmResult struct {
	JobID              string   `json:"job_id"`
	Stage              string   `json:"stage"`
	Summary            Summary  `json:"summary"`
	CreatedActivityIDs []string `json:"created_activity_ids"`
}

type activityImportJob struct {
	ActorID       string
	FileHash      string
	ExpiresAt     time.Time
	Rows          []ActivityImportRow
	Precheck      ActivityPrecheckResult
	ConfirmResult ActivityConfirmResult
}

type ActivityImportService struct {
	activities activityrepo.Repository
	schemas activityrepo.CategorySchemaReader
	creator    *activityservice.Service
	mu         sync.Mutex
	jobs       map[string]*activityImportJob
}

func NewActivityImportService(activities activityrepo.Repository, schemas activityrepo.CategorySchemaReader) *ActivityImportService {
	creator := activityservice.New(activities, nil)
	creator.UseCategorySchemaReader(schemas)

	return &ActivityImportService{
		activities: activities,
		schemas:    schemas,
		creator:    creator,
		jobs:       make(map[string]*activityImportJob),
	}
}

func (s *ActivityImportService) Precheck(ctx context.Context, actorID string, fileName string, fileHash string, rows []ActivityImportRow) (ActivityPrecheckResult, error) {
	results := make([]ActivityRowResult, 0, len(rows))
	summary := Summary{TotalRows: len(rows)}

	for idx, row := range rows {
		_, errs, err := s.validateRow(ctx, row)
		if err != nil {
			return ActivityPrecheckResult{}, err
		}

		status := "valid"
		if len(errs) > 0 {
			status = "invalid"
			summary.InvalidRows++
		} else {
			summary.ValidRows++
		}

		results = append(results, ActivityRowResult{
			RowNo:   idx + 1,
			Status:  status,
			Payload: cloneActivityImportRow(row),
			Errors:  errs,
		})
	}

	jobID := idgen.New()
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	result := ActivityPrecheckResult{
		JobID:     jobID,
		JobType:   "activities",
		FileName:  fileName,
		FileHash:  fileHash,
		ExpiresAt: expiresAt,
		Summary:   summary,
		Rows:      results,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[jobID] = &activityImportJob{
		ActorID:   actorID,
		FileHash:  fileHash,
		ExpiresAt: expiresAt,
		Rows:      cloneActivityImportRows(rows),
		Precheck:  result,
	}

	return result, nil
}

func (s *ActivityImportService) Confirm(ctx context.Context, actorID string, jobID string, fileHash string) (ActivityConfirmResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return ActivityConfirmResult{}, ErrImportJobNotFound
	}
	if time.Now().UTC().After(job.ExpiresAt) {
		return ActivityConfirmResult{}, ErrImportJobExpired
	}
	if job.ActorID != actorID || job.FileHash != fileHash {
		return ActivityConfirmResult{}, ErrImportFileHashMisfit
	}
	if job.ConfirmResult.JobID != "" {
		return job.ConfirmResult, nil
	}
	if job.Precheck.Summary.InvalidRows > 0 {
		return ActivityConfirmResult{}, ErrImportHasInvalidRows
	}

	inputs := make([]activityservice.Input, 0, len(job.Rows))
	for _, row := range job.Rows {
		input, errs, err := s.validateRow(ctx, row)
		if err != nil {
			return ActivityConfirmResult{}, err
		}
		if len(errs) > 0 {
			return ActivityConfirmResult{}, ErrImportHasInvalidRows
		}
		inputs = append(inputs, input)
	}

	createdIDs := make([]string, 0, len(inputs))
	for _, input := range inputs {
		activity, err := s.creator.Create(ctx, actorID, "", input)
		if err != nil {
			if rollbackErr := s.rollbackCreatedActivities(ctx, createdIDs); rollbackErr != nil {
				return ActivityConfirmResult{}, rollbackErr
			}
			if errors.Is(err, activityrepo.ErrCategorySchemaNotFound) ||
				errors.Is(err, activityservice.ErrActivitySchemaInvalid) ||
				errors.Is(err, activityservice.ErrActivityDetailsInvalid) {
				return ActivityConfirmResult{}, ErrImportHasInvalidRows
			}
			return ActivityConfirmResult{}, err
		}
		createdIDs = append(createdIDs, activity.ID)
	}

	result := ActivityConfirmResult{
		JobID:              jobID,
		Stage:              "completed",
		Summary:            job.Precheck.Summary,
		CreatedActivityIDs: createdIDs,
	}
	job.ConfirmResult = result
	return result, nil
}

func (s *ActivityImportService) rollbackCreatedActivities(ctx context.Context, activityIDs []string) error {
	for i := len(activityIDs) - 1; i >= 0; i-- {
		if err := s.activities.Delete(ctx, activityIDs[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *ActivityImportService) validateRow(ctx context.Context, row ActivityImportRow) (activityservice.Input, []ValidationError, error) {
	input := activityservice.Input{
		Title:        strings.TrimSpace(row.Title),
		Description:  strings.TrimSpace(row.Description),
		Category:     strings.TrimSpace(row.Category),
		AcademicYear: strings.TrimSpace(row.AcademicYear),
		Semester:     strings.TrimSpace(row.Semester),
		StartDate:    strings.TrimSpace(row.StartDate),
		EndDate:      strings.TrimSpace(row.EndDate),
		Details:      map[string]any{},
	}

	errs := make([]ValidationError, 0)
	if input.Title == "" {
		errs = append(errs, ValidationError{Field: "title", Code: "REQUIRED", Message: "title is required"})
	}
	if input.Category == "" {
		errs = append(errs, ValidationError{Field: "category", Code: "REQUIRED", Message: "category is required"})
	}
	if input.AcademicYear == "" {
		errs = append(errs, ValidationError{Field: "academic_year", Code: "REQUIRED", Message: "academic_year is required"})
	}
	if input.Semester == "" {
		errs = append(errs, ValidationError{Field: "semester", Code: "REQUIRED", Message: "semester is required"})
	}
	if input.StartDate != "" {
		if _, err := time.Parse("2006-01-02", input.StartDate); err != nil {
			errs = append(errs, ValidationError{Field: "start_date", Code: "INVALID_DATE", Message: "start_date must use YYYY-MM-DD"})
		}
	}
	if input.EndDate != "" {
		if _, err := time.Parse("2006-01-02", input.EndDate); err != nil {
			errs = append(errs, ValidationError{Field: "end_date", Code: "INVALID_DATE", Message: "end_date must use YYYY-MM-DD"})
		}
	}

	if input.Category == "" {
		return input, errs, nil
	}

	schema, err := s.schemas.GetCategorySchema(ctx, input.Category)
	if err != nil {
		if errors.Is(err, activityrepo.ErrCategorySchemaNotFound) {
			errs = append(errs, ValidationError{
				Field:   "category",
				Code:    "ACTIVITY_CATEGORY_NOT_FOUND",
				Message: "activity category not found",
			})
			return input, errs, nil
		}
		return activityservice.Input{}, nil, err
	}

	normalized, err := activityservice.ValidateActivityDetails(
		schema.Schema,
		coerceImportedDetails(schema.Schema.Fields, row.Details),
		map[string]any{
			"title":         input.Title,
			"description":   input.Description,
			"category":      input.Category,
			"academic_year": input.AcademicYear,
			"semester":      input.Semester,
			"start_date":    input.StartDate,
			"end_date":      input.EndDate,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, activityservice.ErrActivitySchemaInvalid):
			errs = append(errs, ValidationError{
				Field:   "details",
				Code:    "ACTIVITY_SCHEMA_INVALID",
				Message: trimValidationPrefix(err.Error()),
			})
		case errors.Is(err, activityservice.ErrActivityDetailsInvalid):
			errs = append(errs, ValidationError{
				Field:   "details",
				Code:    "ACTIVITY_DETAILS_INVALID",
				Message: trimValidationPrefix(err.Error()),
			})
		default:
			return activityservice.Input{}, nil, err
		}
		return input, errs, nil
	}

	input.Details = normalized
	return input, errs, nil
}

func coerceImportedDetails(fields []activityrepo.SchemaField, details map[string]any) map[string]any {
	if details == nil {
		return map[string]any{}
	}

	fieldsByKey := make(map[string]activityrepo.SchemaField, len(fields))
	for _, field := range fields {
		fieldsByKey[field.Key] = field
	}

	coerced := make(map[string]any, len(details))
	for key, value := range details {
		field, ok := fieldsByKey[key]
		if !ok {
			coerced[key] = value
			continue
		}
		coerced[key] = coerceImportedDetailValue(field.Type, value)
	}
	return coerced
}

func coerceImportedDetailValue(fieldType string, value any) any {
	raw, ok := value.(string)
	if !ok {
		return value
	}
	raw = strings.TrimSpace(raw)

	switch fieldType {
	case "number":
		if raw == "" {
			return ""
		}
		number, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return raw
		}
		return number
	case "boolean":
		if raw == "" {
			return ""
		}
		boolean, ok := parseImportedBool(raw)
		if !ok {
			return raw
		}
		return boolean
	case "checkbox":
		if raw == "" {
			return []string{}
		}
		if strings.HasPrefix(raw, "[") {
			var values []string
			if err := json.Unmarshal([]byte(raw), &values); err == nil {
				return values
			}
		}

		separator := "|"
		if strings.Contains(raw, "|") {
			separator = "|"
		} else if strings.Contains(raw, ",") {
			separator = ","
		}

		parts := strings.Split(raw, separator)
		values := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			values = append(values, part)
		}
		if len(values) == 0 {
			return []string{}
		}
		return values
	default:
		return raw
	}
}

func parseImportedBool(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "y":
		return true, true
	case "false", "0", "no", "n":
		return false, true
	default:
		return false, false
	}
}

func trimValidationPrefix(message string) string {
	message = strings.TrimSpace(message)
	prefixes := []string{
		activityservice.ErrActivitySchemaInvalid.Error() + ":",
		activityservice.ErrActivityDetailsInvalid.Error() + ":",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(message, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(message, prefix))
		}
	}
	return message
}

func cloneActivityImportRows(rows []ActivityImportRow) []ActivityImportRow {
	cloned := make([]ActivityImportRow, 0, len(rows))
	for _, row := range rows {
		cloned = append(cloned, cloneActivityImportRow(row))
	}
	return cloned
}

func cloneActivityImportRow(row ActivityImportRow) ActivityImportRow {
	return ActivityImportRow{
		Title:        row.Title,
		Description:  row.Description,
		Category:     row.Category,
		AcademicYear: row.AcademicYear,
		Semester:     row.Semester,
		StartDate:    row.StartDate,
		EndDate:      row.EndDate,
		Details:      cloneActivityImportDetails(row.Details),
	}
}

func cloneActivityImportDetails(details map[string]any) map[string]any {
	if details == nil {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(details))
	for key, value := range details {
		cloned[key] = value
	}
	return cloned
}
