package service

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"sync"
	"time"

	activityrepo "credit-management/backend/internal/modules/activities/repo"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/idgen"
)

var (
	ErrFileNotFound       = errors.New("file not found")
	ErrFileAccessDenied   = errors.New("file access denied")
	ErrFilePreviewInvalid = errors.New("file preview unsupported")
)

type File struct {
	ID          string    `json:"id"`
	ActivityID  string    `json:"activity_id"`
	Name        string    `json:"name"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Download struct {
	File        File
	Content     []byte
	ContentType string
}

type Service struct {
	activities activityrepo.Repository
	mu         sync.RWMutex
	files      map[string]*storedFile
	byActivity map[string][]string
}

type storedFile struct {
	meta    File
	content []byte
}

func New(activities activityrepo.Repository) *Service {
	return &Service{
		activities: activities,
		files:      make(map[string]*storedFile),
		byActivity: make(map[string][]string),
	}
}

func (s *Service) List(ctx context.Context, actorID string, actorRole string, activityID string) ([]File, error) {
	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if !activity.CanView(actorID, actorRole) {
		return nil, ErrFileAccessDenied
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.byActivity[activityID]
	items := make([]File, 0, len(ids))
	for _, id := range ids {
		if file, ok := s.files[id]; ok {
			items = append(items, file.meta)
		}
	}

	return items, nil
}

func (s *Service) Upload(ctx context.Context, actorID string, actorRole string, activityID string, fileHeader *multipart.FileHeader) (*File, error) {
	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if actorRole != authz.RoleTeacher && actorRole != authz.RoleAdmin && activity.OwnerID != actorID {
		return nil, ErrFileAccessDenied
	}
	if err := activity.AllowStructuralEdit(); err != nil {
		return nil, err
	}

	reader, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	file := File{
		ID:          idgen.New(),
		ActivityID:  activityID,
		Name:        fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		UploadedBy:  actorID,
		CreatedAt:   time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.files[file.ID] = &storedFile{
		meta:    file,
		content: buf,
	}
	s.byActivity[activityID] = append(s.byActivity[activityID], file.ID)

	return &file, nil
}

func (s *Service) Delete(ctx context.Context, actorID string, actorRole string, activityID string, fileID string) error {
	activity, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return err
	}
	if actorRole != authz.RoleTeacher && actorRole != authz.RoleAdmin && activity.OwnerID != actorID {
		return ErrFileAccessDenied
	}
	if err := activity.AllowStructuralEdit(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.files[fileID]; !ok {
		return ErrFileNotFound
	}
	delete(s.files, fileID)

	ids := s.byActivity[activityID]
	filtered := ids[:0]
	for _, id := range ids {
		if id != fileID {
			filtered = append(filtered, id)
		}
	}
	s.byActivity[activityID] = filtered

	return nil
}

func (s *Service) Download(ctx context.Context, actorID string, actorRole string, fileID string) (*Download, error) {
	s.mu.RLock()
	file, ok := s.files[fileID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrFileNotFound
	}

	activity, err := s.activities.Get(ctx, file.meta.ActivityID)
	if err != nil {
		return nil, err
	}
	if !activity.CanView(actorID, actorRole) {
		return nil, ErrFileAccessDenied
	}

	return &Download{
		File:        file.meta,
		Content:     append([]byte(nil), file.content...),
		ContentType: file.meta.ContentType,
	}, nil
}
