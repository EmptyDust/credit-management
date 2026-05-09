package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strconv"
	"time"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	usersrepo "credit-management/backend/internal/modules/users/repo"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/idgen"
)

var (
	ErrParticipantNotFound         = errors.New("participant not found")
	ErrParticipantDuplicate        = errors.New("participant duplicate")
	ErrParticipantOnlyStudent      = errors.New("participant only student allowed")
	ErrParticipantEditForbidden    = errors.New("participant edit forbidden")
	ErrParticipantVisibilityDenied = errors.New("participant visibility denied")
	ErrParticipantCreditsInvalid   = errors.New("participant credits invalid")
)

type Participant struct {
	ID         string    `json:"id"`
	ActivityID string    `json:"activity_id"`
	UserID     string    `json:"user_id"`
	Credits    float64   `json:"credits"`
	JoinedAt   time.Time `json:"joined_at"`
}

type Stats struct {
	TotalCount     int     `json:"total_count"`
	TotalCredits   float64 `json:"total_credits"`
	AverageCredits float64 `json:"average_credits"`
	RecentNewCount int     `json:"recent_new_count"`
}

type Service struct {
	activities activityrepo.Repository
	users      usersrepo.UserRepository
	repo       Repository
}

func New(activities activityrepo.Repository, users usersrepo.UserRepository) *Service {
	return NewWithRepository(activities, users, newMemoryRepository())
}

func NewWithRepository(activities activityrepo.Repository, users usersrepo.UserRepository, repo Repository) *Service {
	return &Service{
		activities: activities,
		users:      users,
		repo:       repo,
	}
}

func (s *Service) List(ctx context.Context, actorID string, actorRole string, activityID string) ([]Participant, error) {
	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if !activity.CanView(actorID, actorRole) {
		return nil, ErrParticipantVisibilityDenied
	}

	return s.repo.ListByActivity(ctx, activityID)
}

func (s *Service) Add(ctx context.Context, actorID string, actorRole string, activityID string, userIDs []string, credits float64) ([]Participant, error) {
	if credits < 0 {
		return nil, ErrParticipantCreditsInvalid
	}
	if _, err := s.editableActivity(ctx, actorID, actorRole, activityID); err != nil {
		return nil, err
	}

	added := make([]Participant, 0, len(userIDs))
	for _, userID := range userIDs {
		user, err := s.users.GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user.UserType != authz.RoleStudent {
			return nil, ErrParticipantOnlyStudent
		}
		if _, err := s.repo.GetActiveByActivityUser(ctx, activityID, userID); err == nil {
			return nil, ErrParticipantDuplicate
		} else if !errors.Is(err, ErrParticipantNotFound) {
			return nil, err
		}

		participant := Participant{
			ID:         idgen.New(),
			ActivityID: activityID,
			UserID:     userID,
			Credits:    credits,
			JoinedAt:   time.Now().UTC(),
		}
		added = append(added, participant)
	}

	if err := s.repo.AddBatch(ctx, activityID, added); err != nil {
		return nil, err
	}
	return added, nil
}

func (s *Service) UpdateCredits(ctx context.Context, actorID string, actorRole string, activityID string, participantID string, credits float64) (*Participant, error) {
	if credits < 0 {
		return nil, ErrParticipantCreditsInvalid
	}
	if _, err := s.editableActivity(ctx, actorID, actorRole, activityID); err != nil {
		return nil, err
	}
	return s.repo.UpdateCredits(ctx, activityID, participantID, credits)
}

func (s *Service) BatchSetCredits(ctx context.Context, actorID string, actorRole string, activityID string, participantIDs []string, credits float64) (int, error) {
	if credits < 0 {
		return 0, ErrParticipantCreditsInvalid
	}
	if _, err := s.editableActivity(ctx, actorID, actorRole, activityID); err != nil {
		return 0, err
	}
	return s.repo.BatchUpdateCredits(ctx, activityID, participantIDs, credits)
}

func (s *Service) Delete(ctx context.Context, actorID string, actorRole string, activityID string, participantID string) error {
	if _, err := s.editableActivity(ctx, actorID, actorRole, activityID); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, activityID, participantID)
}

func (s *Service) BatchDelete(ctx context.Context, actorID string, actorRole string, activityID string, participantIDs []string) (int, error) {
	if _, err := s.editableActivity(ctx, actorID, actorRole, activityID); err != nil {
		return 0, err
	}
	return s.repo.BatchSoftDelete(ctx, activityID, participantIDs)
}

func (s *Service) Stats(ctx context.Context, actorID string, actorRole string, activityID string) (Stats, error) {
	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return Stats{}, err
	}
	if !activity.CanView(actorID, actorRole) {
		return Stats{}, ErrParticipantVisibilityDenied
	}

	items, err := s.repo.ListByActivity(ctx, activityID)
	if err != nil {
		return Stats{}, err
	}

	stats := Stats{TotalCount: len(items)}
	recentThreshold := time.Now().UTC().AddDate(0, 0, -7)
	for _, item := range items {
		stats.TotalCredits += item.Credits
		if !item.JoinedAt.Before(recentThreshold) {
			stats.RecentNewCount++
		}
	}
	if stats.TotalCount > 0 {
		stats.AverageCredits = stats.TotalCredits / float64(stats.TotalCount)
	}
	return stats, nil
}

func (s *Service) Export(ctx context.Context, actorID string, actorRole string, activityID string) ([]byte, error) {
	if _, err := s.manageableActivity(ctx, actorID, actorRole, activityID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByActivity(ctx, activityID)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	if err := writer.Write([]string{"participant_id", "user_id", "credits", "joined_at"}); err != nil {
		return nil, err
	}
	for _, item := range items {
		if err := writer.Write([]string{
			item.ID,
			item.UserID,
			strconv.FormatFloat(item.Credits, 'f', -1, 64),
			item.JoinedAt.Format(time.RFC3339),
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Service) Leave(ctx context.Context, actorID string, actorRole string, activityID string) error {
	if actorRole != authz.RoleStudent {
		return ErrParticipantEditForbidden
	}

	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return err
	}
	if err := activity.AllowStructuralEdit(); err != nil {
		return err
	}

	return s.repo.SoftDeleteByActivityUser(ctx, activityID, actorID)
}

func (s *Service) SnapshotByActivity(ctx context.Context, activityID string) ([]Participant, error) {
	return s.repo.ListByActivity(ctx, activityID)
}

func (s *Service) CopyActivityParticipants(ctx context.Context, sourceActivityID string, targetActivityID string) error {
	source, err := s.repo.ListByActivity(ctx, sourceActivityID)
	if err != nil {
		return err
	}
	if len(source) == 0 {
		return nil
	}

	now := time.Now().UTC()
	copied := make([]Participant, 0, len(source))
	for _, participant := range source {
		copied = append(copied, Participant{
			ID:         idgen.New(),
			ActivityID: targetActivityID,
			UserID:     participant.UserID,
			Credits:    participant.Credits,
			JoinedAt:   now,
		})
	}

	return s.repo.AddBatch(ctx, targetActivityID, copied)
}

func (s *Service) editableActivity(ctx context.Context, actorID string, actorRole string, activityID string) (*activitydomain.Activity, error) {
	activity, err := s.manageableActivity(ctx, actorID, actorRole, activityID)
	if err != nil {
		return nil, err
	}
	if err := activity.AllowStructuralEdit(); err != nil {
		return nil, err
	}
	return activity, nil
}

func (s *Service) manageableActivity(ctx context.Context, actorID string, actorRole string, activityID string) (*activitydomain.Activity, error) {
	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if actorRole != authz.RoleTeacher && actorRole != authz.RoleAdmin && activity.OwnerID != actorID {
		return nil, ErrParticipantEditForbidden
	}
	return activity, nil
}
