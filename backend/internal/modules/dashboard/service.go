package dashboard

import (
	"context"
	"sort"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	appservice "credit-management/backend/internal/modules/applications/service"
	usersdomain "credit-management/backend/internal/modules/users/domain"
	usersrepo "credit-management/backend/internal/modules/users/repo"
)

type Counter func() int
type CounterRepository interface {
	Count() int
}

type userStatsRepository interface {
	ListUsers(ctx context.Context, filter usersrepo.ListFilter) ([]usersdomain.User, error)
}

type applicationStatsProvider func(actorID string, actorRole string) appservice.Stats

type Service struct {
	usersCounter        Counter
	activitiesCounter   Counter
	applicationsCounter Counter
	userStats           userStatsRepository
	activityStats       *ActivityStatsService
	activityList        activityListRepository
	participants        participantSnapshotter
	applicationStats    applicationStatsProvider
}

type ServiceOption func(*Service)

type UserStats struct {
	Total         int            `json:"total"`
	StudentCount  int            `json:"student_count"`
	TeacherCount  int            `json:"teacher_count"`
	AdminCount    int            `json:"admin_count"`
	ActiveCount   int            `json:"active_count"`
	DisabledCount int            `json:"disabled_count"`
	RoleCounts    map[string]int `json:"role_counts"`
	StatusCounts  map[string]int `json:"status_counts"`
}

type ActivitySummary struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Category         string `json:"category"`
	AcademicYear     string `json:"academic_year"`
	Semester         string `json:"semester"`
	Status           string `json:"status"`
	OwnerID          string `json:"owner_id"`
	ParticipantCount int    `json:"participant_count"`
}

type Overview struct {
	UserCount            int               `json:"user_count"`
	ActivityCount        int               `json:"activity_count"`
	ApplicationCount     int               `json:"application_count"`
	UserStats            UserStats         `json:"user_stats"`
	ActivityStats        ActivityStats     `json:"activity_stats"`
	ApplicationStats     appservice.Stats  `json:"application_stats"`
	CategoryDistribution []KeyLabelCount   `json:"category_distribution"`
	RecentActivities     []ActivitySummary `json:"recent_activities"`
	PopularActivities    []ActivitySummary `json:"popular_activities"`
}

func NewService(users any, activities any, applications any, options ...ServiceOption) *Service {
	service := &Service{
		usersCounter:        normalizeCounter(users),
		activitiesCounter:   normalizeCounter(activities),
		applicationsCounter: normalizeCounter(applications),
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithUserStats(users userStatsRepository) ServiceOption {
	return func(service *Service) {
		service.userStats = users
	}
}

func WithActivityOverview(activities activityListRepository, categories activityCategoryReader, participants participantSnapshotter) ServiceOption {
	return func(service *Service) {
		service.activityList = activities
		service.participants = participants
		service.activityStats = NewActivityStatsService(activities, categories, participants)
	}
}

func WithApplicationStats(provider func(actorID string, actorRole string) appservice.Stats) ServiceOption {
	return func(service *Service) {
		service.applicationStats = provider
	}
}

func normalizeCounter(source any) Counter {
	switch v := source.(type) {
	case nil:
		return nil
	case Counter:
		return v
	case func() int:
		return Counter(v)
	case CounterRepository:
		return v.Count
	default:
		return nil
	}
}

func (s *Service) Overview() Overview {
	overview, _ := s.OverviewFor(context.Background(), "", "")
	return overview
}

func (s *Service) OverviewFor(ctx context.Context, actorID string, actorRole string) (Overview, error) {
	overview := Overview{}
	if s.usersCounter != nil {
		overview.UserCount = s.usersCounter()
	}
	if s.activitiesCounter != nil {
		overview.ActivityCount = s.activitiesCounter()
	}
	if s.applicationsCounter != nil {
		overview.ApplicationCount = s.applicationsCounter()
	}

	if s.userStats != nil {
		stats, err := s.buildUserStats(ctx)
		if err != nil {
			return Overview{}, err
		}
		overview.UserStats = stats
	}

	if s.activityStats != nil {
		stats, err := s.activityStats.Stats(ctx, actorID, actorRole, activityrepo.ListFilter{})
		if err != nil {
			return Overview{}, err
		}
		overview.ActivityStats = stats
		overview.CategoryDistribution = stats.CategoryCounts
	}

	if s.activityList != nil {
		activities, err := s.activityList.List(ctx, normalizeActivityFilter(actorID, actorRole, activityrepo.ListFilter{}))
		if err != nil {
			return Overview{}, err
		}
		summaries, err := s.buildActivitySummaries(ctx, activities)
		if err != nil {
			return Overview{}, err
		}
		overview.RecentActivities = limitActivitySummaries(summaries, 10)
		overview.PopularActivities = popularActivitySummaries(summaries, 10)
	}

	if s.applicationStats != nil {
		overview.ApplicationStats = s.applicationStats(actorID, actorRole)
	} else {
		overview.ApplicationStats.Total = overview.ApplicationCount
	}

	return overview, nil
}

func (s *Service) buildUserStats(ctx context.Context) (UserStats, error) {
	users, err := s.userStats.ListUsers(ctx, usersrepo.ListFilter{})
	if err != nil {
		return UserStats{}, err
	}

	stats := UserStats{
		Total:        len(users),
		RoleCounts:   map[string]int{},
		StatusCounts: map[string]int{},
	}
	for _, user := range users {
		stats.RoleCounts[user.UserType]++
		stats.StatusCounts[user.Status]++

		switch user.UserType {
		case "student":
			stats.StudentCount++
		case "teacher":
			stats.TeacherCount++
		case "admin":
			stats.AdminCount++
		}

		switch user.Status {
		case "active":
			stats.ActiveCount++
		case "disabled", "inactive", "suspended":
			stats.DisabledCount++
		}
	}

	return stats, nil
}

func (s *Service) buildActivitySummaries(ctx context.Context, activities []activitydomain.Activity) ([]ActivitySummary, error) {
	summaries := make([]ActivitySummary, 0, len(activities))
	for _, activity := range activities {
		summary := ActivitySummary{
			ID:           activity.ID,
			Title:        activity.Title,
			Category:     activity.Category,
			AcademicYear: activity.AcademicYear,
			Semester:     activity.Semester,
			Status:       activity.Status,
			OwnerID:      activity.OwnerID,
		}
		if s.participants != nil {
			participants, err := s.participants.SnapshotByActivity(ctx, activity.ID)
			if err != nil {
				return nil, err
			}
			summary.ParticipantCount = len(participants)
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func limitActivitySummaries(items []ActivitySummary, limit int) []ActivitySummary {
	if limit < 0 {
		limit = 0
	}
	if len(items) <= limit {
		return cloneActivitySummaries(items)
	}
	return cloneActivitySummaries(items[:limit])
}

func popularActivitySummaries(items []ActivitySummary, limit int) []ActivitySummary {
	popular := cloneActivitySummaries(items)
	sort.SliceStable(popular, func(i, j int) bool {
		return popular[i].ParticipantCount > popular[j].ParticipantCount
	})
	return limitActivitySummaries(popular, limit)
}

func cloneActivitySummaries(items []ActivitySummary) []ActivitySummary {
	copied := make([]ActivitySummary, len(items))
	copy(copied, items)
	return copied
}
