package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"sort"
	"strconv"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	"credit-management/backend/internal/platform/authz"
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

func (s *ActivityExportService) Export(ctx context.Context, actorID string, actorRole string, filter activityrepo.ListFilter) ([]byte, error) {
	items, err := s.activities.List(ctx, normalizeExportFilter(actorID, actorRole, filter))
	if err != nil {
		return nil, err
	}

	detailKeys := collectDetailKeys(items)

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	header := append([]string{}, activityExportBaseHeader...)
	for _, key := range detailKeys {
		header = append(header, "details."+key)
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

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
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
