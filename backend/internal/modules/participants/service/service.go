package service

import (
	"context"
	"errors"
	"sync"
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
	mu         sync.RWMutex
	byActivity map[string]map[string]*Participant
}

func New(activities activityrepo.Repository, users usersrepo.UserRepository) *Service {
	return &Service{
		activities: activities,
		users:      users,
		byActivity: make(map[string]map[string]*Participant),
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

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := s.byActivity[activityID]
	items := make([]Participant, 0, len(rows))
	for _, participant := range rows {
		items = append(items, *participant)
	}

	return items, nil
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

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.byActivity[activityID] == nil {
		s.byActivity[activityID] = make(map[string]*Participant)
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
		if _, exists := s.byActivity[activityID][userID]; exists {
			return nil, ErrParticipantDuplicate
		}

		participant := &Participant{
			ID:         idgen.New(),
			ActivityID: activityID,
			UserID:     userID,
			Credits:    credits,
			JoinedAt:   time.Now().UTC(),
		}
		s.byActivity[activityID][userID] = participant
		added = append(added, *participant)
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

	s.mu.Lock()
	defer s.mu.Unlock()

	rows := s.byActivity[activityID]
	if rows == nil {
		return ErrParticipantNotFound
	}
	if _, ok := rows[actorID]; !ok {
		return ErrParticipantNotFound
	}

	delete(rows, actorID)
	return nil
}
