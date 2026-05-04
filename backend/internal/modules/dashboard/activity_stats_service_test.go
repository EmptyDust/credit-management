package dashboard

import (
	"context"
	"reflect"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	participantservice "credit-management/backend/internal/modules/participants/service"
)

func TestActivityStatsServiceAggregatesFilteredCounts(t *testing.T) {
	service := NewActivityStatsService(
		fakeActivityListRepository{
			items: []activitydomain.Activity{
				{
					ID:           "activity-1",
					Title:        "Hackathon",
					Category:     "general",
					AcademicYear: "2025-2026",
					Semester:     "autumn",
					Status:       activitydomain.StatusDraft,
					OwnerID:      "teacher-1",
				},
				{
					ID:           "activity-2",
					Title:        "Innovation Lab",
					Category:     "innovation",
					AcademicYear: "2025-2026",
					Semester:     "autumn",
					Status:       activitydomain.StatusApproved,
					OwnerID:      "teacher-2",
				},
				{
					ID:           "activity-3",
					Title:        "Archived Event",
					Category:     "innovation",
					AcademicYear: "2024-2025",
					Semester:     "spring",
					Status:       activitydomain.StatusPendingReview,
					OwnerID:      "teacher-1",
				},
				{
					ID:           "activity-4",
					Title:        "Seminar",
					Category:     "general",
					AcademicYear: "2025-2026",
					Semester:     "spring",
					Status:       activitydomain.StatusRejected,
					OwnerID:      "teacher-3",
				},
			},
		},
		fakeActivityStatsCategoryReader{
			items: []activityrepo.CategorySummary{
				{Key: "general", Label: "General", SortOrder: 10},
				{Key: "innovation", Label: "Innovation", SortOrder: 20},
			},
		},
		fakeParticipantSnapshotter{
			byActivity: map[string][]participantservice.Participant{
				"activity-1": {{UserID: "teacher-2"}},
				"activity-2": {{UserID: "teacher-1"}},
				"activity-4": {{UserID: "teacher-1"}},
			},
		},
	)

	stats, err := service.Stats(context.Background(), "teacher-1", "teacher", activityrepo.ListFilter{
		AcademicYear: "2025-2026",
	})
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}

	expectedStatusCounts := map[string]int{
		activitydomain.StatusDraft:         1,
		activitydomain.StatusPendingReview: 0,
		activitydomain.StatusApproved:      1,
		activitydomain.StatusRejected:      1,
	}
	if !reflect.DeepEqual(stats.StatusCounts, expectedStatusCounts) {
		t.Fatalf("unexpected status counts: %#v", stats.StatusCounts)
	}

	expectedCategoryCounts := []KeyLabelCount{
		{Key: "general", Label: "General", Count: 2},
		{Key: "innovation", Label: "Innovation", Count: 1},
	}
	if !reflect.DeepEqual(stats.CategoryCounts, expectedCategoryCounts) {
		t.Fatalf("unexpected category counts: %#v", stats.CategoryCounts)
	}

	expectedAcademicYearCounts := []ValueCount{
		{Value: "2025-2026", Count: 3},
	}
	if !reflect.DeepEqual(stats.AcademicYearCounts, expectedAcademicYearCounts) {
		t.Fatalf("unexpected academic year counts: %#v", stats.AcademicYearCounts)
	}

	expectedSemesterCounts := []ValueCount{
		{Value: "autumn", Count: 2},
		{Value: "spring", Count: 1},
	}
	if !reflect.DeepEqual(stats.SemesterCounts, expectedSemesterCounts) {
		t.Fatalf("unexpected semester counts: %#v", stats.SemesterCounts)
	}

	if stats.MineCount != 1 {
		t.Fatalf("expected mine count 1, got %d", stats.MineCount)
	}
	if stats.ParticipatedCount != 2 {
		t.Fatalf("expected participated count 2, got %d", stats.ParticipatedCount)
	}
}

type fakeActivityListRepository struct {
	items       []activitydomain.Activity
	lastFilter  activityrepo.ListFilter
	listInvoked bool
}

func (f fakeActivityListRepository) List(_ context.Context, filter activityrepo.ListFilter) ([]activitydomain.Activity, error) {
	f.lastFilter = filter
	f.listInvoked = true

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
		if filter.AcademicYear != "" && item.AcademicYear != filter.AcademicYear {
			continue
		}
		if filter.Semester != "" && item.Semester != filter.Semester {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

type fakeActivityStatsCategoryReader struct {
	items []activityrepo.CategorySummary
	err   error
}

func (f fakeActivityStatsCategoryReader) ListCategories(context.Context) ([]activityrepo.CategorySummary, error) {
	return f.items, f.err
}

type fakeParticipantSnapshotter struct {
	byActivity map[string][]participantservice.Participant
}

func (f fakeParticipantSnapshotter) SnapshotByActivity(_ context.Context, activityID string) ([]participantservice.Participant, error) {
	return f.byActivity[activityID], nil
}
