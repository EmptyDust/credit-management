package service

import (
	"context"
	"errors"
	"time"

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
)

type Participant struct {
	ID         string    `json:"id"`
	ActivityID string    `json:"activity_id"`
	UserID     string    `json:"user_id"`
	Credits    float64   `json:"credits"`
	JoinedAt   time.Time `json:"joined_at"`
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
	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if actorRole != authz.RoleTeacher && actorRole != authz.RoleAdmin && activity.OwnerID != actorID {
		return nil, ErrParticipantEditForbidden
	}
	if err := activity.AllowStructuralEdit(); err != nil {
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
