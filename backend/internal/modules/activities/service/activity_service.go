package service

import (
	"context"
	"sync"
	"time"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/idgen"
)

type ApplicationHooks interface {
	SyncApprovedActivity(ctx context.Context, activityID string) error
	RevokeApprovedActivity(ctx context.Context, activityID string) error
}

type CopyHooks interface {
	CopyActivityRelations(ctx context.Context, sourceActivityID string, targetActivityID string, actorID string) error
}

type Service struct {
	repo           activityrepo.Repository
	hooks          ApplicationHooks
	copyHooks      CopyHooks
	schemaResolver activityCategorySchemaResolver
}

type Input struct {
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Category     string         `json:"category"`
	AcademicYear string         `json:"academic_year"`
	Semester     string         `json:"semester"`
	StartDate    string         `json:"start_date"`
	EndDate      string         `json:"end_date"`
	Details      map[string]any `json:"details"`
}

type ReviewInput struct {
	TargetStatus   string `json:"target_status"`
	ReviewComments string `json:"review_comments"`
}

type MemoryRepository struct {
	mu         sync.RWMutex
	activities map[string]*activitydomain.Activity
}

type noopHooks struct{}

func New(repo activityrepo.Repository, hooks ApplicationHooks) *Service {
	if hooks == nil {
		hooks = noopHooks{}
	}

	service := &Service{
		repo:  repo,
		hooks: hooks,
	}

	if resolver, ok := repo.(activityCategorySchemaResolver); ok {
		service.schemaResolver = resolver
	}

	return service
}

func (s *Service) UseCopyHooks(hooks CopyHooks) *Service {
	s.copyHooks = hooks
	return s
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		activities: make(map[string]*activitydomain.Activity),
	}
}

func (noopHooks) SyncApprovedActivity(_ context.Context, _ string) error {
	return nil
}

func (noopHooks) RevokeApprovedActivity(_ context.Context, _ string) error {
	return nil
}

func (s *Service) List(ctx context.Context, actorID string, actorRole string, filter activityrepo.ListFilter) ([]activitydomain.Activity, error) {
	if actorRole == authz.RoleStudent {
		filter.OwnerID = actorID
	}

	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, actorID string, actorRole string, activityID string) (*activitydomain.Activity, error) {
	activity, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if !activity.CanView(actorID, actorRole) {
		return nil, activitydomain.ErrActivityForbidden
	}

	return activity, nil
}

func (s *Service) Create(ctx context.Context, actorID string, _ string, input Input) (*activitydomain.Activity, error) {
	normalizedDetails, err := s.validateAndNormalizeDetails(ctx, "", input)
	if err != nil {
		return nil, err
	}

	activity := &activitydomain.Activity{
		ID:           idgen.New(),
		Title:        input.Title,
		Description:  input.Description,
		Category:     input.Category,
		AcademicYear: input.AcademicYear,
		Semester:     input.Semester,
		StartDate:    input.StartDate,
		EndDate:      input.EndDate,
		Status:       activitydomain.StatusDraft,
		OwnerID:      actorID,
		Details:      normalizedDetails,
	}

	if err := s.repo.Save(ctx, activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *Service) Update(ctx context.Context, actorID string, actorRole string, activityID string, input Input) (*activitydomain.Activity, error) {
	activity, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if actorRole != authz.RoleTeacher && actorRole != authz.RoleAdmin && activity.OwnerID != actorID {
		return nil, activitydomain.ErrActivityForbidden
	}
	if err := activity.AllowStructuralEdit(); err != nil {
		return nil, err
	}

	normalizedDetails, err := s.validateAndNormalizeDetails(ctx, activityID, input)
	if err != nil {
		return nil, err
	}

	activity.Title = input.Title
	activity.Description = input.Description
	activity.Category = input.Category
	activity.AcademicYear = input.AcademicYear
	activity.Semester = input.Semester
	activity.StartDate = input.StartDate
	activity.EndDate = input.EndDate
	activity.Details = normalizedDetails

	if err := s.repo.Save(ctx, activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *Service) Delete(ctx context.Context, actorID string, actorRole string, activityID string) error {
	activity, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return err
	}
	if !activity.CanDelete(actorID, actorRole) {
		return activitydomain.ErrActivityForbidden
	}

	return s.repo.Delete(ctx, activityID)
}

func (s *Service) Submit(ctx context.Context, actorID string, activityID string) (*activitydomain.Activity, error) {
	activity, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if err := activity.Submit(actorID); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *Service) Withdraw(ctx context.Context, actorID string, activityID string) (*activitydomain.Activity, error) {
	activity, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if err := activity.Withdraw(actorID); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *Service) Review(ctx context.Context, actorID string, actorRole string, activityID string, input ReviewInput) (*activitydomain.Activity, error) {
	activity, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	reviewedAt := time.Now().UTC()
	if err := activity.Review(actorRole, input.TargetStatus, input.ReviewComments, actorID, reviewedAt); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, activity); err != nil {
		return nil, err
	}
	if activity.Status == activitydomain.StatusApproved {
		if err := s.hooks.SyncApprovedActivity(ctx, activity.ID); err != nil {
			return nil, err
		}
	}
	if activity.Status == activitydomain.StatusRejected {
		if err := s.hooks.RevokeApprovedActivity(ctx, activity.ID); err != nil {
			return nil, err
		}
	}

	return activity, nil
}

func (s *Service) Copy(ctx context.Context, actorID string, actorRole string, activityID string) (*activitydomain.Activity, error) {
	source, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if actorRole != authz.RoleTeacher && actorRole != authz.RoleAdmin && source.OwnerID != actorID {
		return nil, activitydomain.ErrActivityForbidden
	}

	copyActivity := &activitydomain.Activity{
		ID:           idgen.New(),
		Title:        source.Title,
		Description:  source.Description,
		Category:     source.Category,
		AcademicYear: source.AcademicYear,
		Semester:     source.Semester,
		StartDate:    source.StartDate,
		EndDate:      source.EndDate,
		Status:       activitydomain.StatusDraft,
		OwnerID:      actorID,
		Details:      cloneDetails(source.Details),
	}
	if err := s.repo.Save(ctx, copyActivity); err != nil {
		return nil, err
	}
	if s.copyHooks != nil {
		if err := s.copyHooks.CopyActivityRelations(ctx, source.ID, copyActivity.ID, actorID); err != nil {
			_ = s.repo.Delete(ctx, copyActivity.ID)
			return nil, err
		}
	}

	return copyActivity, nil
}

func (s *Service) TransferOwner(ctx context.Context, actorRole string, activityID string, newOwnerID string) (*activitydomain.Activity, error) {
	activity, err := s.repo.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if err := activity.TransferOwner(actorRole, newOwnerID); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (r *MemoryRepository) List(_ context.Context, filter activityrepo.ListFilter) ([]activitydomain.Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]activitydomain.Activity, 0, len(r.activities))
	for _, activity := range r.activities {
		if activity.DeletedAt != nil {
			continue
		}
		if filter.OwnerID != "" && activity.OwnerID != filter.OwnerID {
			continue
		}
		if filter.Status != "" && activity.Status != filter.Status {
			continue
		}
		if filter.Category != "" && activity.Category != filter.Category {
			continue
		}
		if filter.AcademicYear != "" && activity.AcademicYear != filter.AcademicYear {
			continue
		}
		if filter.Semester != "" && activity.Semester != filter.Semester {
			continue
		}
		if filter.Keyword != "" && activity.Title != filter.Keyword {
			continue
		}

		items = append(items, cloneActivity(*activity))
	}

	return items, nil
}

func (r *MemoryRepository) Get(_ context.Context, activityID string) (*activitydomain.Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activity, ok := r.activities[activityID]
	if !ok || activity.DeletedAt != nil {
		return nil, activitydomain.ErrActivityNotFound
	}

	copy := cloneActivity(*activity)
	return &copy, nil
}

func (r *MemoryRepository) Save(_ context.Context, activity *activitydomain.Activity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := cloneActivity(*activity)
	r.activities[activity.ID] = &copy
	return nil
}

func (r *MemoryRepository) Delete(_ context.Context, activityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	activity, ok := r.activities[activityID]
	if !ok || activity.DeletedAt != nil {
		return activitydomain.ErrActivityNotFound
	}
	now := time.Now().UTC()
	activity.DeletedAt = &now
	return nil
}

func cloneActivity(activity activitydomain.Activity) activitydomain.Activity {
	copy := activity
	copy.Details = cloneDetails(activity.Details)
	if activity.ReviewedAt != nil {
		t := *activity.ReviewedAt
		copy.ReviewedAt = &t
	}
	if activity.DeletedAt != nil {
		t := *activity.DeletedAt
		copy.DeletedAt = &t
	}
	return copy
}

func cloneDetails(details map[string]any) map[string]any {
	if details == nil {
		return map[string]any{}
	}

	copy := make(map[string]any, len(details))
	for key, value := range details {
		copy[key] = value
	}
	return copy
}
