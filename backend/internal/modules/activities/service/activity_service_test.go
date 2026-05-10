package service

import (
	"context"
	"errors"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
)

func TestUpdateDraftOnly(t *testing.T) {
	activity := activitydomain.Activity{Status: activitydomain.StatusApproved}

	err := activity.AllowStructuralEdit()
	if !errors.Is(err, activitydomain.ErrActivityStatusInvalid) {
		t.Fatalf("expected ErrActivityStatusInvalid, got %v", err)
	}
}

func TestListPagePaginatesAfterFiltering(t *testing.T) {
	svc := New(fakeActivityRepository{
		items: []activitydomain.Activity{
			{ID: "activity-1", Title: "A", Status: activitydomain.StatusApproved, OwnerID: "teacher-1"},
			{ID: "activity-2", Title: "B", Status: activitydomain.StatusRejected, OwnerID: "teacher-1"},
			{ID: "activity-3", Title: "C", Status: activitydomain.StatusApproved, OwnerID: "teacher-2"},
			{ID: "activity-4", Title: "D", Status: activitydomain.StatusApproved, OwnerID: "teacher-3"},
		},
	}, nil)

	result, err := svc.ListPage(context.Background(), "teacher-1", "teacher", activityrepo.ListFilter{
		Status:   activitydomain.StatusApproved,
		Page:     2,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("ListPage returned error: %v", err)
	}

	if result.Total != 3 || result.Page != 2 || result.PageSize != 2 || result.TotalPages != 2 {
		t.Fatalf("unexpected pagination metadata: %#v", result)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "activity-4" {
		t.Fatalf("unexpected page items: %#v", result.Items)
	}
}

type fakeActivityRepository struct {
	items []activitydomain.Activity
}

func (f fakeActivityRepository) List(_ context.Context, filter activityrepo.ListFilter) ([]activitydomain.Activity, error) {
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

func (f fakeActivityRepository) Get(_ context.Context, activityID string) (*activitydomain.Activity, error) {
	for _, item := range f.items {
		if item.ID == activityID {
			copy := item
			return &copy, nil
		}
	}
	return nil, activitydomain.ErrActivityNotFound
}

func (f fakeActivityRepository) Save(context.Context, *activitydomain.Activity) error {
	return nil
}

func (f fakeActivityRepository) Delete(context.Context, string) error {
	return nil
}
