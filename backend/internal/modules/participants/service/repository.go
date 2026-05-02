package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"credit-management/backend/internal/platform/idgen"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository interface {
	ListByActivity(ctx context.Context, activityID string) ([]Participant, error)
	GetActiveByActivityUser(ctx context.Context, activityID string, userID string) (*Participant, error)
	AddBatch(ctx context.Context, activityID string, rows []Participant) error
	SoftDeleteByActivityUser(ctx context.Context, activityID string, userID string) error
}

type memoryRepository struct {
	mu         sync.RWMutex
	byActivity map[string]map[string]*Participant
}

type gormRepository struct {
	db *gorm.DB
}

type participantRow struct {
	ID         string     `gorm:"column:id"`
	ActivityID string     `gorm:"column:activity_id"`
	UserID     string     `gorm:"column:user_id"`
	Credits    float64    `gorm:"column:credits"`
	JoinedAt   time.Time  `gorm:"column:joined_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

func (participantRow) TableName() string {
	return "activity_participants"
}

func newMemoryRepository() Repository {
	return &memoryRepository{
		byActivity: make(map[string]map[string]*Participant),
	}
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *memoryRepository) ListByActivity(_ context.Context, activityID string) ([]Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows := r.byActivity[activityID]
	items := make([]Participant, 0, len(rows))
	for _, participant := range rows {
		items = append(items, *participant)
	}

	return items, nil
}

func (r *memoryRepository) GetActiveByActivityUser(_ context.Context, activityID string, userID string) (*Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows := r.byActivity[activityID]
	if rows == nil {
		return nil, ErrParticipantNotFound
	}
	participant, ok := rows[userID]
	if !ok {
		return nil, ErrParticipantNotFound
	}
	copy := *participant
	return &copy, nil
}

func (r *memoryRepository) AddBatch(_ context.Context, activityID string, rows []Participant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.byActivity[activityID] == nil {
		r.byActivity[activityID] = make(map[string]*Participant)
	}

	for i := range rows {
		participant := rows[i]
		if _, exists := r.byActivity[activityID][participant.UserID]; exists {
			return ErrParticipantDuplicate
		}
		copy := participant
		if copy.ID == "" {
			copy.ID = idgen.New()
		}
		if copy.JoinedAt.IsZero() {
			copy.JoinedAt = time.Now().UTC()
		}
		r.byActivity[activityID][copy.UserID] = &copy
	}
	return nil
}

func (r *memoryRepository) SoftDeleteByActivityUser(_ context.Context, activityID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := r.byActivity[activityID]
	if rows == nil {
		return ErrParticipantNotFound
	}
	if _, ok := rows[userID]; !ok {
		return ErrParticipantNotFound
	}
	delete(rows, userID)
	return nil
}

func (r *gormRepository) ListByActivity(ctx context.Context, activityID string) ([]Participant, error) {
	var rows []participantRow
	if err := r.db.WithContext(ctx).
		Where("activity_id = ? AND deleted_at IS NULL", activityID).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]Participant, 0, len(rows))
	for i := range rows {
		items = append(items, rows[i].toDomain())
	}
	return items, nil
}

func (r *gormRepository) GetActiveByActivityUser(ctx context.Context, activityID string, userID string) (*Participant, error) {
	var row participantRow
	if err := r.db.WithContext(ctx).
		Where("activity_id = ? AND user_id = ? AND deleted_at IS NULL", activityID, userID).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}

	participant := row.toDomain()
	return &participant, nil
}

func (r *gormRepository) AddBatch(ctx context.Context, activityID string, rows []Participant) error {
	items := make([]participantRow, 0, len(rows))
	now := time.Now().UTC()
	for i := range rows {
		row := rows[i]
		if row.ID == "" {
			row.ID = idgen.New()
		}
		if row.JoinedAt.IsZero() {
			row.JoinedAt = now
		}
		items = append(items, participantRow{
			ID:         row.ID,
			ActivityID: activityID,
			UserID:     row.UserID,
			Credits:    row.Credits,
			JoinedAt:   row.JoinedAt.UTC(),
		})
	}

	if err := r.db.WithContext(ctx).Create(&items).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrParticipantDuplicate
		}
		return err
	}
	return nil
}

func (r *gormRepository) SoftDeleteByActivityUser(ctx context.Context, activityID string, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&participantRow{}).
		Where("activity_id = ? AND user_id = ? AND deleted_at IS NULL", activityID, userID).
		Update("deleted_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrParticipantNotFound
	}
	return nil
}

func (r participantRow) toDomain() Participant {
	return Participant{
		ID:         r.ID,
		ActivityID: r.ActivityID,
		UserID:     r.UserID,
		Credits:    r.Credits,
		JoinedAt:   r.JoinedAt.UTC(),
	}
}
