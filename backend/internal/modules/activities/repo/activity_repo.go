package repo

import (
	"context"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
)

type ListFilter struct {
	Keyword      string
	Status       string
	Category     string
	AcademicYear string
	Semester     string
	OwnerID      string
	Page         int
	PageSize     int
}

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]activitydomain.Activity, error)
	Get(ctx context.Context, activityID string) (*activitydomain.Activity, error)
	Save(ctx context.Context, activity *activitydomain.Activity) error
	Delete(ctx context.Context, activityID string) error
}
