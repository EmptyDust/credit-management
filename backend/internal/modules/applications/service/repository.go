package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"credit-management/backend/internal/platform/idgen"

	"gorm.io/gorm"
)

type Repository interface {
	UpsertApproved(ctx context.Context, application Application) error
	SoftDeleteByActivity(ctx context.Context, activityID string) error
	List(ctx context.Context, actorID string, actorRole string) ([]Application, error)
	Get(ctx context.Context, actorID string, actorRole string, applicationID string) (*Application, error)
	Count(ctx context.Context) (int, error)
}

type memoryRepository struct {
	mu           sync.Mutex
	applications map[string]*Application
}

type gormRepository struct {
	db *gorm.DB
}

type applicationRow struct {
	ID             string     `gorm:"column:id"`
	ActivityID     string     `gorm:"column:activity_id"`
	UserID         string     `gorm:"column:user_id"`
	Status         string     `gorm:"column:status"`
	AppliedCredits float64    `gorm:"column:applied_credits"`
	AwardedCredits float64    `gorm:"column:awarded_credits"`
	SubmittedAt    time.Time  `gorm:"column:submitted_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (applicationRow) TableName() string {
	return "applications"
}

func newMemoryRepository() Repository {
	return &memoryRepository{
		applications: make(map[string]*Application),
	}
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *memoryRepository) UpsertApproved(_ context.Context, application Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := application.ActivityID + ":" + application.UserID
	existing, ok := r.applications[key]
	if ok {
		if existing.Deleted {
			existing.Deleted = false
			existing.Status = "approved"
		}
		existing.AppliedCredits = application.AppliedCredits
		existing.AwardedCredits = application.AwardedCredits
		return nil
	}

	copy := application
	if copy.ID == "" {
		copy.ID = key
	}
	r.applications[key] = &copy
	return nil
}

func (r *memoryRepository) SoftDeleteByActivity(_ context.Context, activityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, application := range r.applications {
		if application.ActivityID == activityID {
			application.Deleted = true
		}
	}
	return nil
}

func (r *memoryRepository) List(_ context.Context, actorID string, actorRole string) ([]Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Application, 0, len(r.applications))
	for _, application := range r.applications {
		if application.Deleted {
			continue
		}
		if actorRole == "student" && application.UserID != actorID {
			continue
		}
		copy := *application
		items = append(items, copy)
	}
	return items, nil
}

func (r *memoryRepository) Get(_ context.Context, actorID string, actorRole string, applicationID string) (*Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, application := range r.applications {
		if application.ID != applicationID || application.Deleted {
			continue
		}
		if actorRole == "student" && application.UserID != actorID {
			break
		}
		copy := *application
		return &copy, nil
	}
	return nil, ErrApplicationNotFound
}

func (r *memoryRepository) Count(_ context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for _, application := range r.applications {
		if !application.Deleted {
			count++
		}
	}
	return count, nil
}

func (r *gormRepository) UpsertApproved(ctx context.Context, application Application) error {
	var row applicationRow
	err := r.db.WithContext(ctx).
		Where("activity_id = ? AND user_id = ?", application.ActivityID, application.UserID).
		Order("created_at DESC").
		First(&row).Error
	if err == nil {
		result := r.db.WithContext(ctx).Model(&applicationRow{}).
			Where("id = ?", row.ID).
			Updates(map[string]any{
				"status":          "approved",
				"applied_credits": application.AppliedCredits,
				"awarded_credits": application.AwardedCredits,
				"deleted_at":      nil,
				"updated_at":      time.Now().UTC(),
			})
		return result.Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	id := application.ID
	if id == "" || len(id) != 26 {
		id = idgen.New()
	}
	return r.db.WithContext(ctx).Create(&applicationRow{
		ID:             id,
		ActivityID:     application.ActivityID,
		UserID:         application.UserID,
		Status:         "approved",
		AppliedCredits: application.AppliedCredits,
		AwardedCredits: application.AwardedCredits,
		SubmittedAt:    time.Now().UTC(),
	}).Error
}

func (r *gormRepository) SoftDeleteByActivity(ctx context.Context, activityID string) error {
	return r.db.WithContext(ctx).
		Model(&applicationRow{}).
		Where("activity_id = ? AND deleted_at IS NULL", activityID).
		Update("deleted_at", time.Now().UTC()).Error
}

func (r *gormRepository) List(ctx context.Context, actorID string, actorRole string) ([]Application, error) {
	query := r.db.WithContext(ctx).Model(&applicationRow{}).Where("deleted_at IS NULL")
	if actorRole == "student" {
		query = query.Where("user_id = ?", actorID)
	}

	var rows []applicationRow
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]Application, 0, len(rows))
	for i := range rows {
		items = append(items, rows[i].toDomain())
	}
	return items, nil
}

func (r *gormRepository) Get(ctx context.Context, actorID string, actorRole string, applicationID string) (*Application, error) {
	query := r.db.WithContext(ctx).Model(&applicationRow{}).Where("id = ? AND deleted_at IS NULL", applicationID)
	if actorRole == "student" {
		query = query.Where("user_id = ?", actorID)
	}

	var row applicationRow
	if err := query.First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	application := row.toDomain()
	return &application, nil
}

func (r *gormRepository) Count(ctx context.Context) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&applicationRow{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r applicationRow) toDomain() Application {
	return Application{
		ID:             r.ID,
		ActivityID:     r.ActivityID,
		UserID:         r.UserID,
		Status:         r.Status,
		AppliedCredits: r.AppliedCredits,
		AwardedCredits: r.AwardedCredits,
		Deleted:        r.DeletedAt != nil,
	}
}
