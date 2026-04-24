package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strconv"
	"sync"

	activityrepo "credit-management/backend/internal/modules/activities/repo"
	fileservice "credit-management/backend/internal/modules/files/service"
	participantservice "credit-management/backend/internal/modules/participants/service"
)

var ErrApplicationNotFound = errors.New("application not found")

type Application struct {
	ID             string  `json:"id"`
	ActivityID     string  `json:"activity_id"`
	UserID         string  `json:"user_id"`
	Status         string  `json:"status"`
	AppliedCredits float64 `json:"applied_credits"`
	AwardedCredits float64 `json:"awarded_credits"`
	Deleted        bool    `json:"deleted"`
}

type SyncEffect struct {
	Mode          string `json:"mode"`
	AffectedCount int    `json:"affected_count"`
}

type Service struct {
	mu           sync.Mutex
	applications map[string]*Application
	activities   activityrepo.Repository
	files        *fileservice.Service
	participants *participantservice.Service
}

func New() *Service {
	return &Service{
		applications: make(map[string]*Application),
	}
}

type Stats struct {
	Total         int     `json:"total"`
	ApprovedCount int     `json:"approved_count"`
	TotalCredits  float64 `json:"total_credits"`
}

type ApplicationView struct {
	ID             string             `json:"id"`
	Status         string             `json:"status"`
	AppliedCredits float64            `json:"applied_credits"`
	AwardedCredits float64            `json:"awarded_credits"`
	Activity       map[string]any     `json:"activity"`
	AcademicYear   string             `json:"academic_year,omitempty"`
	Semester       string             `json:"semester,omitempty"`
	RelatedFiles   []fileservice.File `json:"related_files,omitempty"`
}

func (s *Service) SetActivityRepository(repo activityrepo.Repository) {
	s.activities = repo
}

func (s *Service) SetFileService(files *fileservice.Service) {
	s.files = files
}

func (s *Service) SetParticipantService(participants *participantservice.Service) {
	s.participants = participants
}

func (s *Service) SyncApprovedActivity(activityID string, userIDs []string) SyncEffect {
	participants := make([]participantservice.Participant, 0, len(userIDs))
	for _, userID := range userIDs {
		participants = append(participants, participantservice.Participant{
			ActivityID: activityID,
			UserID:     userID,
		})
	}

	return s.syncApprovedParticipants(participants)
}

func (s *Service) syncApprovedParticipants(participants []participantservice.Participant) SyncEffect {
	s.mu.Lock()
	defer s.mu.Unlock()

	affected := 0
	for _, participant := range participants {
		key := participant.ActivityID + ":" + participant.UserID
		existing, ok := s.applications[key]
		if ok {
			if existing.Deleted {
				existing.Deleted = false
				existing.Status = "approved"
			}
			existing.AppliedCredits = participant.Credits
			existing.AwardedCredits = participant.Credits
			affected++
			continue
		}

		s.applications[key] = &Application{
			ID:             key,
			ActivityID:     participant.ActivityID,
			UserID:         participant.UserID,
			Status:         "approved",
			AppliedCredits: participant.Credits,
			AwardedCredits: participant.Credits,
		}
		affected++
	}

	return SyncEffect{
		Mode:          "created_or_restored",
		AffectedCount: affected,
	}
}

func (s *Service) SyncApprovedActivityContext(ctx context.Context, activityID string) error {
	if s.participants == nil {
		return nil
	}

	participants, err := s.participants.SnapshotByActivity(ctx, activityID)
	if err != nil {
		return err
	}

	s.syncApprovedParticipants(participants)
	return nil
}

type Hook struct {
	service *Service
}

func NewHook(service *Service) Hook {
	return Hook{service: service}
}

func (h Hook) SyncApprovedActivity(ctx context.Context, activityID string) error {
	return h.service.SyncApprovedActivityContext(ctx, activityID)
}

func (s *Service) List(actorID string, actorRole string) []ApplicationView {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]ApplicationView, 0, len(s.applications))
	for _, application := range s.applications {
		if application.Deleted {
			continue
		}
		if actorRole == "student" && application.UserID != actorID {
			continue
		}

		items = append(items, s.buildView(application))
	}

	return items
}

func (s *Service) Stats(actorID string, actorRole string) Stats {
	items := s.List(actorID, actorRole)
	stats := Stats{Total: len(items)}
	for _, item := range items {
		if item.Status == "approved" {
			stats.ApprovedCount++
		}
		stats.TotalCredits += item.AwardedCredits
	}

	return stats
}

func (s *Service) Get(actorID string, actorRole string, applicationID string) (ApplicationView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, application := range s.applications {
		if application.ID != applicationID || application.Deleted {
			continue
		}
		if actorRole == "student" && application.UserID != actorID {
			break
		}
		return s.buildView(application), nil
	}

	return ApplicationView{}, ErrApplicationNotFound
}

func (s *Service) Export(actorID string, actorRole string) ([]byte, error) {
	items := s.List(actorID, actorRole)
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	if err := writer.Write([]string{"id", "activity_id", "user_id", "status", "applied_credits", "awarded_credits", "academic_year", "semester"}); err != nil {
		return nil, err
	}

	for _, item := range items {
		if err := writer.Write([]string{
			item.ID,
			stringValue(item.Activity["id"]),
			stringValue(item.Activity["user_id"]),
			item.Status,
			floatString(item.AppliedCredits),
			floatString(item.AwardedCredits),
			item.AcademicYear,
			item.Semester,
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

func (s *Service) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, application := range s.applications {
		if !application.Deleted {
			count++
		}
	}
	return count
}

func (s *Service) buildView(application *Application) ApplicationView {
	view := ApplicationView{
		ID:             application.ID,
		Status:         application.Status,
		AppliedCredits: application.AppliedCredits,
		AwardedCredits: application.AwardedCredits,
		Activity: map[string]any{
			"id":      application.ActivityID,
			"user_id": application.UserID,
		},
	}

	if s.activities != nil {
		activity, err := s.activities.Get(context.Background(), application.ActivityID)
		if err == nil {
			view.Activity["title"] = activity.Title
			view.Activity["category"] = activity.Category
			view.Activity["owner_id"] = activity.OwnerID
			view.AcademicYear = activity.AcademicYear
			view.Semester = activity.Semester
		}
	}
	if s.files != nil {
		files, err := s.files.List(context.Background(), "", "admin", application.ActivityID)
		if err == nil {
			view.RelatedFiles = files
		}
	}

	return view
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func floatString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
