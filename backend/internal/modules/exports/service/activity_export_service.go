package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/exportfile"
)

var activityExportBaseHeader = []string{
	"title",
	"description",
	"category",
	"academic_year",
	"semester",
	"start_date",
	"end_date",
}

type activityExportRepository interface {
	List(ctx context.Context, filter activityrepo.ListFilter) ([]activitydomain.Activity, error)
}

type ActivityExportService struct {
	activities activityExportRepository
}

func NewActivityExportService(activities activityExportRepository) *ActivityExportService {
	return &ActivityExportService{activities: activities}
}

func (s *ActivityExportService) Export(ctx context.Context, actorID string, actorRole string, filter activityrepo.ListFilter, format string) (exportfile.File, error) {
	items, err := s.activities.List(ctx, normalizeExportFilter(actorID, actorRole, filter))
	if err != nil {
		return exportfile.File{}, err
	}

	rows := buildActivityExportRows(items)
	switch exportfile.NormalizeFormat(format) {
	case exportfile.FormatXLSX:
		content, err := exportfile.WriteXLSX("Activities", rows)
		if err != nil {
			return exportfile.File{}, err
		}
		return exportfile.File{
			Filename:    "activities.xlsx",
			ContentType: exportfile.ContentTypeXLSX,
			Content:     content,
		}, nil
	default:
		content, err := exportfile.WriteCSV(rows)
		if err != nil {
			return exportfile.File{}, err
		}
		return exportfile.File{
			Filename:    "activities.csv",
			ContentType: exportfile.ContentTypeCSV,
			Content:     content,
		}, nil
	}
}

func buildActivityExportRows(items []activitydomain.Activity) [][]string {
	detailKeys := collectDetailKeys(items)

	header := append([]string{}, activityExportBaseHeader...)
	for _, key := range detailKeys {
		header = append(header, "details."+key)
	}
	rows := [][]string{header}

	for _, item := range items {
		record := []string{
			item.Title,
			item.Description,
			item.Category,
			item.AcademicYear,
			item.Semester,
			item.StartDate,
			item.EndDate,
		}
		for _, key := range detailKeys {
			record = append(record, formatExportValue(item.Details[key]))
		}
		rows = append(rows, record)
	}

	return rows
}

func normalizeExportFilter(actorID string, actorRole string, filter activityrepo.ListFilter) activityrepo.ListFilter {
	if actorRole == authz.RoleStudent {
		filter.OwnerID = actorID
	}
	return filter
}

func collectDetailKeys(items []activitydomain.Activity) []string {
	keys := map[string]struct{}{}
	for _, item := range items {
		for key := range item.Details {
			keys[key] = struct{}{}
		}
	}

	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	return ordered
}

func formatExportValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int8:
		return strconv.FormatInt(int64(typed), 10)
	case int16:
		return strconv.FormatInt(int64(typed), 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint8:
		return strconv.FormatUint(uint64(typed), 10)
	case uint16:
		return strconv.FormatUint(uint64(typed), 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		payload, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(payload)
	}
}
