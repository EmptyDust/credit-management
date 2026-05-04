package service

import (
	"context"
	"encoding/csv"
	"reflect"
	"strings"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
)

func TestActivityExportServiceFlattensDetailsAndRespectsFilters(t *testing.T) {
	repo := &fakeActivityExportRepository{
		items: []activitydomain.Activity{
			{
				ID:           "activity-1",
				Title:        "Innovation Lab",
				Description:  "Build prototypes",
				Category:     "innovation",
				AcademicYear: "2025-2026",
				Semester:     "autumn",
				StartDate:    "2025-09-01",
				EndDate:      "2025-09-15",
				Status:       activitydomain.StatusApproved,
				OwnerID:      "teacher-1",
				Details: map[string]any{
					"location":         "Room 101",
					"activity_mode":    "offline",
					"max_participants": float64(30),
				},
			},
			{
				ID:           "activity-2",
				Title:        "General Seminar",
				Description:  "Weekly discussion",
				Category:     "general",
				AcademicYear: "2025-2026",
				Semester:     "spring",
				StartDate:    "2026-03-02",
				EndDate:      "2026-03-02",
				Status:       activitydomain.StatusDraft,
				OwnerID:      "teacher-2",
				Details: map[string]any{
					"location": "Room 202",
				},
			},
		},
	}
	service := NewActivityExportService(repo)

	content, err := service.Export(context.Background(), "teacher-1", "teacher", activityrepo.ListFilter{
		Category: "innovation",
	})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	if repo.lastFilter.Category != "innovation" {
		t.Fatalf("expected export filter category innovation, got %#v", repo.lastFilter)
	}

	records, err := csv.NewReader(strings.NewReader(string(content))).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v\ncsv=%s", err, string(content))
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 csv records, got %d", len(records))
	}

	expectedHeader := []string{
		"title",
		"description",
		"category",
		"academic_year",
		"semester",
		"start_date",
		"end_date",
		"details.activity_mode",
		"details.location",
		"details.max_participants",
	}
	if !reflect.DeepEqual(records[0], expectedHeader) {
		t.Fatalf("unexpected header: %#v", records[0])
	}

	expectedRow := []string{
		"Innovation Lab",
		"Build prototypes",
		"innovation",
		"2025-2026",
		"autumn",
		"2025-09-01",
		"2025-09-15",
		"offline",
		"Room 101",
		"30",
	}
	if !reflect.DeepEqual(records[1], expectedRow) {
		t.Fatalf("unexpected csv row: %#v", records[1])
	}
}

func TestActivityExportServiceScopesStudentExportsToOwnedActivities(t *testing.T) {
	repo := &fakeActivityExportRepository{}
	service := NewActivityExportService(repo)

	if _, err := service.Export(context.Background(), "student-1", "student", activityrepo.ListFilter{
		Status: activitydomain.StatusApproved,
	}); err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	expectedFilter := activityrepo.ListFilter{
		Status:  activitydomain.StatusApproved,
		OwnerID: "student-1",
	}
	if !reflect.DeepEqual(repo.lastFilter, expectedFilter) {
		t.Fatalf("unexpected export filter: %#v", repo.lastFilter)
	}
}

type fakeActivityExportRepository struct {
	items      []activitydomain.Activity
	lastFilter activityrepo.ListFilter
}

func (f *fakeActivityExportRepository) List(_ context.Context, filter activityrepo.ListFilter) ([]activitydomain.Activity, error) {
	f.lastFilter = filter

	items := make([]activitydomain.Activity, 0, len(f.items))
	for _, item := range f.items {
		if filter.OwnerID != "" && item.OwnerID != filter.OwnerID {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.Category != "" && item.Category != filter.Category {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}
