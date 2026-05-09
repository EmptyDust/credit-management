package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"

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
	repo         Repository
	activities   activityrepo.Repository
	files        *fileservice.Service
	participants *participantservice.Service
}

type Filter struct {
	Keyword      string
	ActivityID   string
	UserID       string
	Status       string
	AcademicYear string
	Semester     string
}

func New() *Service {
	return &Service{
		repo: newMemoryRepository(),
	}
}

func NewWithRepository(repo Repository) *Service {
	return &Service{repo: repo}
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
	affected := 0
	for _, participant := range participants {
		if err := s.repo.UpsertApproved(context.Background(), Application{
			ActivityID:     participant.ActivityID,
			UserID:         participant.UserID,
			Status:         "approved",
			AppliedCredits: participant.Credits,
			AwardedCredits: participant.Credits,
		}); err == nil {
			affected++
		}
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

	for _, participant := range participants {
		if err := s.repo.UpsertApproved(ctx, Application{
			ActivityID:     participant.ActivityID,
			UserID:         participant.UserID,
			Status:         "approved",
			AppliedCredits: participant.Credits,
			AwardedCredits: participant.Credits,
		}); err != nil {
			return err
		}
	}
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

func (h Hook) RevokeApprovedActivity(ctx context.Context, activityID string) error {
	return h.service.RevokeApprovedActivityContext(ctx, activityID)
}

func (s *Service) RevokeApprovedActivityContext(ctx context.Context, activityID string) error {
	return s.repo.SoftDeleteByActivity(ctx, activityID)
}

func (s *Service) List(actorID string, actorRole string) []ApplicationView {
	return s.ListFiltered(actorID, actorRole, Filter{})
}

func (s *Service) ListFiltered(actorID string, actorRole string, filter Filter) []ApplicationView {
	items, err := s.repo.List(context.Background(), actorID, actorRole)
	if err != nil {
		return []ApplicationView{}
	}

	views := make([]ApplicationView, 0, len(items))
	for i := range items {
		application := items[i]
		view := s.buildView(&application)
		if !matchesFilter(view, filter) {
			continue
		}
		views = append(views, view)
	}
	return views
}

func (s *Service) Stats(actorID string, actorRole string) Stats {
	return s.StatsFiltered(actorID, actorRole, Filter{})
}

func (s *Service) StatsFiltered(actorID string, actorRole string, filter Filter) Stats {
	items := s.ListFiltered(actorID, actorRole, filter)
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
	application, err := s.repo.Get(context.Background(), actorID, actorRole, applicationID)
	if err != nil {
		return ApplicationView{}, err
	}
	return s.buildView(application), nil
}

func (s *Service) Export(actorID string, actorRole string) ([]byte, error) {
	return s.ExportFiltered(actorID, actorRole, Filter{})
}

func (s *Service) ExportFiltered(actorID string, actorRole string, filter Filter) ([]byte, error) {
	items := s.ListFiltered(actorID, actorRole, filter)
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
	count, err := s.repo.Count(context.Background())
	if err != nil {
		return 0
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

func matchesFilter(item ApplicationView, filter Filter) bool {
	if filter.Status != "" && item.Status != filter.Status {
		return false
	}
	if filter.ActivityID != "" && stringValue(item.Activity["id"]) != filter.ActivityID {
		return false
	}
	if filter.UserID != "" && stringValue(item.Activity["user_id"]) != filter.UserID {
		return false
	}
	if filter.AcademicYear != "" && item.AcademicYear != filter.AcademicYear {
		return false
	}
	if filter.Semester != "" && item.Semester != filter.Semester {
		return false
	}
	if strings.TrimSpace(filter.Keyword) != "" {
		keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))
		haystack := strings.ToLower(strings.Join([]string{
			item.ID,
			stringValue(item.Activity["id"]),
			stringValue(item.Activity["user_id"]),
			stringValue(item.Activity["title"]),
			stringValue(item.Activity["category"]),
			item.Status,
			item.AcademicYear,
			item.Semester,
		}, " "))
		if !strings.Contains(haystack, keyword) {
			return false
		}
	}
	return true
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func floatString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
