package dashboard

import (
	"context"
	"sort"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	participantservice "credit-management/backend/internal/modules/participants/service"
	"credit-management/backend/internal/platform/authz"
)

type activityListRepository interface {
	List(ctx context.Context, filter activityrepo.ListFilter) ([]activitydomain.Activity, error)
}

type activityCategoryReader interface {
	ListCategories(ctx context.Context) ([]activityrepo.CategorySummary, error)
}

type participantSnapshotter interface {
	SnapshotByActivity(ctx context.Context, activityID string) ([]participantservice.Participant, error)
}

type ActivityStatsService struct {
	activities   activityListRepository
	categories   activityCategoryReader
	participants participantSnapshotter
}

type ActivityStats struct {
	StatusCounts       map[string]int `json:"status_counts"`
	CategoryCounts     []KeyLabelCount `json:"category_counts"`
	AcademicYearCounts []ValueCount `json:"academic_year_counts"`
	SemesterCounts     []ValueCount `json:"semester_counts"`
	MineCount          int `json:"mine_count"`
	ParticipatedCount  int `json:"participated_count"`
}

type KeyLabelCount struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type ValueCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

func NewActivityStatsService(activities activityListRepository, categories activityCategoryReader, participants participantSnapshotter) *ActivityStatsService {
	return &ActivityStatsService{
		activities:   activities,
		categories:   categories,
		participants: participants,
	}
}

func (s *ActivityStatsService) Stats(ctx context.Context, actorID string, actorRole string, filter activityrepo.ListFilter) (ActivityStats, error) {
	items, err := s.activities.List(ctx, normalizeActivityFilter(actorID, actorRole, filter))
	if err != nil {
		return ActivityStats{}, err
	}

	categoryLabels := map[string]string{}
	categoryOrder := map[string]int{}
	orderedCategoryKeys := []string{}
	if s.categories != nil {
		categories, err := s.categories.ListCategories(ctx)
		if err != nil {
			return ActivityStats{}, err
		}
		for idx, category := range categories {
			categoryLabels[category.Key] = category.Label
			categoryOrder[category.Key] = idx
			orderedCategoryKeys = append(orderedCategoryKeys, category.Key)
		}
	}

	stats := ActivityStats{
		StatusCounts: map[string]int{
			activitydomain.StatusDraft:         0,
			activitydomain.StatusPendingReview: 0,
			activitydomain.StatusApproved:      0,
			activitydomain.StatusRejected:      0,
		},
		CategoryCounts:     []KeyLabelCount{},
		AcademicYearCounts: []ValueCount{},
		SemesterCounts:     []ValueCount{},
	}

	categoryCounts := map[string]int{}
	academicYearCounts := map[string]int{}
	semesterCounts := map[string]int{}

	for _, item := range items {
		stats.StatusCounts[item.Status]++
		categoryCounts[item.Category]++
		academicYearCounts[item.AcademicYear]++
		semesterCounts[item.Semester]++
		if item.OwnerID == actorID {
			stats.MineCount++
		}
	}

	if len(categoryCounts) > 0 {
		sort.SliceStable(orderedCategoryKeys, func(i, j int) bool {
			return categoryOrder[orderedCategoryKeys[i]] < categoryOrder[orderedCategoryKeys[j]]
		})
		for _, key := range orderedCategoryKeys {
			count := categoryCounts[key]
			if count == 0 {
				continue
			}
			stats.CategoryCounts = append(stats.CategoryCounts, KeyLabelCount{
				Key:   key,
				Label: categoryLabels[key],
				Count: count,
			})
			delete(categoryCounts, key)
		}

		if len(categoryCounts) > 0 {
			keys := make([]string, 0, len(categoryCounts))
			for key := range categoryCounts {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				stats.CategoryCounts = append(stats.CategoryCounts, KeyLabelCount{
					Key:   key,
					Label: categoryLabel(categoryLabels, key),
					Count: categoryCounts[key],
				})
			}
		}
	}

	stats.AcademicYearCounts = buildValueCounts(academicYearCounts)
	stats.SemesterCounts = buildValueCounts(semesterCounts)

	if s.participants != nil && actorID != "" {
		for _, item := range items {
			participants, err := s.participants.SnapshotByActivity(ctx, item.ID)
			if err != nil {
				return ActivityStats{}, err
			}
			for _, participant := range participants {
				if participant.UserID == actorID {
					stats.ParticipatedCount++
					break
				}
			}
		}
	}

	return stats, nil
}

func normalizeActivityFilter(actorID string, actorRole string, filter activityrepo.ListFilter) activityrepo.ListFilter {
	if actorRole == authz.RoleStudent {
		filter.OwnerID = actorID
	}
	return filter
}

func buildValueCounts(counts map[string]int) []ValueCount {
	values := make([]string, 0, len(counts))
	for value := range counts {
		values = append(values, value)
	}
	sort.Strings(values)

	items := make([]ValueCount, 0, len(values))
	for _, value := range values {
		items = append(items, ValueCount{
			Value: value,
			Count: counts[value],
		})
	}
	return items
}

func categoryLabel(labels map[string]string, key string) string {
	if labels[key] != "" {
		return labels[key]
	}
	return key
}
