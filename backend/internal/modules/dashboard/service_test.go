package dashboard

import (
	"context"
	"reflect"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	appservice "credit-management/backend/internal/modules/applications/service"
	participantservice "credit-management/backend/internal/modules/participants/service"
	usersdomain "credit-management/backend/internal/modules/users/domain"
	usersrepo "credit-management/backend/internal/modules/users/repo"
)

type testCounterRepo struct {
	count int
}

func (r testCounterRepo) Count() int {
	return r.count
}

func TestOverviewCounts(t *testing.T) {
	service := NewService(
		func() int { return 3 },
		func() int { return 5 },
		func() int { return 7 },
	)

	overview := service.Overview()
	if overview.UserCount != 3 {
		t.Fatalf("expected user count 3, got %d", overview.UserCount)
	}
	if overview.ActivityCount != 5 {
		t.Fatalf("expected activity count 5, got %d", overview.ActivityCount)
	}
	if overview.ApplicationCount != 7 {
		t.Fatalf("expected application count 7, got %d", overview.ApplicationCount)
	}
}

func TestOverviewCountsFromCounterRepositories(t *testing.T) {
	service := NewService(
		testCounterRepo{count: 11},
		testCounterRepo{count: 13},
		testCounterRepo{count: 17},
	)

	overview := service.Overview()
	if overview.UserCount != 11 {
		t.Fatalf("expected user count 11, got %d", overview.UserCount)
	}
	if overview.ActivityCount != 13 {
		t.Fatalf("expected activity count 13, got %d", overview.ActivityCount)
	}
	if overview.ApplicationCount != 17 {
		t.Fatalf("expected application count 17, got %d", overview.ApplicationCount)
	}
}

func TestOverviewForReturnsPermissionAwareDashboardMetrics(t *testing.T) {
	activities := fakeDashboardActivityRepository{
		items: []activitydomain.Activity{
			{
				ID:           "activity-1",
				Title:        "Newest visible",
				Category:     "general",
				AcademicYear: "2025-2026",
				Semester:     "autumn",
				Status:       activitydomain.StatusApproved,
				OwnerID:      "teacher-1",
			},
			{
				ID:           "activity-2",
				Title:        "Popular visible",
				Category:     "innovation",
				AcademicYear: "2025-2026",
				Semester:     "autumn",
				Status:       activitydomain.StatusDraft,
				OwnerID:      "teacher-2",
			},
			{
				ID:           "activity-3",
				Title:        "Older visible",
				Category:     "innovation",
				AcademicYear: "2024-2025",
				Semester:     "spring",
				Status:       activitydomain.StatusRejected,
				OwnerID:      "teacher-1",
			},
		},
	}
	participants := fakeDashboardParticipantSnapshotter{
		byActivity: map[string][]participantservice.Participant{
			"activity-1": {{UserID: "student-1"}},
			"activity-2": {{UserID: "student-1"}, {UserID: "student-2"}, {UserID: "student-3"}},
			"activity-3": {{UserID: "student-1"}, {UserID: "student-2"}},
		},
	}
	service := NewService(
		func() int { return 4 },
		func() int { return 3 },
		func() int { return 8 },
		WithUserStats(fakeDashboardUserRepository{
			items: []usersdomain.User{
				{ID: "student-1", UserType: "student", Status: "active"},
				{ID: "student-2", UserType: "student", Status: "disabled"},
				{ID: "teacher-1", UserType: "teacher", Status: "active"},
				{ID: "admin-1", UserType: "admin", Status: "active"},
			},
		}),
		WithActivityOverview(activities, fakeActivityStatsCategoryReader{
			items: []activityrepo.CategorySummary{
				{Key: "general", Label: "General"},
				{Key: "innovation", Label: "Innovation"},
			},
		}, participants),
		WithApplicationStats(func(actorID string, actorRole string) appservice.Stats {
			if actorID != "teacher-1" || actorRole != "teacher" {
				t.Fatalf("unexpected application stats actor: %s/%s", actorID, actorRole)
			}
			return appservice.Stats{Total: 8, ApprovedCount: 6, TotalCredits: 12.5}
		}),
	)

	overview, err := service.OverviewFor(context.Background(), "teacher-1", "teacher")
	if err != nil {
		t.Fatalf("OverviewFor returned error: %v", err)
	}

	if overview.UserStats.StudentCount != 2 || overview.UserStats.TeacherCount != 1 || overview.UserStats.AdminCount != 1 {
		t.Fatalf("unexpected user stats: %#v", overview.UserStats)
	}
	if overview.UserStats.ActiveCount != 3 || overview.UserStats.DisabledCount != 1 {
		t.Fatalf("unexpected user status stats: %#v", overview.UserStats)
	}

	expectedStatusCounts := map[string]int{
		activitydomain.StatusDraft:         1,
		activitydomain.StatusPendingReview: 0,
		activitydomain.StatusApproved:      1,
		activitydomain.StatusRejected:      1,
	}
	if !reflect.DeepEqual(overview.ActivityStats.StatusCounts, expectedStatusCounts) {
		t.Fatalf("unexpected activity status counts: %#v", overview.ActivityStats.StatusCounts)
	}

	expectedCategoryDistribution := []KeyLabelCount{
		{Key: "general", Label: "General", Count: 1},
		{Key: "innovation", Label: "Innovation", Count: 2},
	}
	if !reflect.DeepEqual(overview.CategoryDistribution, expectedCategoryDistribution) {
		t.Fatalf("unexpected category distribution: %#v", overview.CategoryDistribution)
	}

	if len(overview.RecentActivities) != 3 || overview.RecentActivities[0].ID != "activity-1" {
		t.Fatalf("unexpected recent activities: %#v", overview.RecentActivities)
	}
	if len(overview.PopularActivities) != 3 || overview.PopularActivities[0].ID != "activity-2" || overview.PopularActivities[0].ParticipantCount != 3 {
		t.Fatalf("unexpected popular activities: %#v", overview.PopularActivities)
	}
	if overview.ApplicationStats.Total != 8 || overview.ApplicationStats.ApprovedCount != 6 || overview.ApplicationStats.TotalCredits != 12.5 {
		t.Fatalf("unexpected application stats: %#v", overview.ApplicationStats)
	}
}

type fakeDashboardUserRepository struct {
	items []usersdomain.User
}

func (f fakeDashboardUserRepository) ListUsers(_ context.Context, filter usersrepo.ListFilter) ([]usersdomain.User, error) {
	items := make([]usersdomain.User, 0, len(f.items))
	for _, item := range f.items {
		if filter.UserType != "" && item.UserType != filter.UserType {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (f fakeDashboardUserRepository) GetUser(context.Context, string) (*usersdomain.User, error) {
	return nil, nil
}

func (f fakeDashboardUserRepository) CreateUser(context.Context, usersrepo.CreateUserInput) (*usersdomain.User, error) {
	return nil, nil
}

func (f fakeDashboardUserRepository) UpdateUser(context.Context, string, usersrepo.UpdateUserInput) (*usersdomain.User, error) {
	return nil, nil
}

func (f fakeDashboardUserRepository) SoftDeleteUser(context.Context, string) error {
	return nil
}

func (f fakeDashboardUserRepository) UpdateStatus(context.Context, string, string) (*usersdomain.User, error) {
	return nil, nil
}

func (f fakeDashboardUserRepository) ResetPassword(context.Context, string, string, int64) error {
	return nil
}

type fakeDashboardActivityRepository struct {
	items []activitydomain.Activity
}

func (f fakeDashboardActivityRepository) List(_ context.Context, filter activityrepo.ListFilter) ([]activitydomain.Activity, error) {
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

func (f fakeDashboardActivityRepository) Get(_ context.Context, activityID string) (*activitydomain.Activity, error) {
	for _, item := range f.items {
		if item.ID == activityID {
			copy := item
			return &copy, nil
		}
	}
	return nil, activitydomain.ErrActivityNotFound
}

func (f fakeDashboardActivityRepository) Save(context.Context, *activitydomain.Activity) error {
	return nil
}

func (f fakeDashboardActivityRepository) Delete(context.Context, string) error {
	return nil
}

type fakeDashboardParticipantSnapshotter struct {
	byActivity map[string][]participantservice.Participant
}

func (f fakeDashboardParticipantSnapshotter) SnapshotByActivity(_ context.Context, activityID string) ([]participantservice.Participant, error) {
	return f.byActivity[activityID], nil
}
