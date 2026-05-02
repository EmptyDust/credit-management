package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	activityrepo "credit-management/backend/internal/modules/activities/repo"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/idgen"
	"credit-management/backend/internal/platform/storage"
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
	SignedURL   string
}

type Service struct {
	activities activityrepo.Repository
	repo       MetadataRepository
	store      storage.Store
}

type storedFile struct {
	meta           File
	content        []byte
	storageKey     string
	storageBackend string
}

func New(activities activityrepo.Repository) *Service {
	files := make(map[string]*storedFile)
	byActivity := make(map[string][]string)
	return &Service{
		activities: activities,
		repo:       newMemoryMetadataRepository(files, byActivity),
	}
}

func NewWithStore(activities activityrepo.Repository, repo MetadataRepository, store storage.Store) *Service {
	return &Service{
		activities: activities,
		repo:       repo,
		store:      store,
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

	return s.repo.ListByActivity(ctx, activityID)
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

	record := MetadataRecord{
		File:           file,
		StorageBackend: "memory",
	}
	if memoryRepo, ok := s.repo.(*memoryMetadataRepository); ok {
		memoryRepo.files[file.ID] = &storedFile{
			meta:           file,
			content:        append([]byte(nil), buf...),
			storageBackend: "memory",
		}
	}
	if s.store != nil {
		key := fmt.Sprintf("activities/%s/%s-%s", activityID, idgen.New(), fileHeader.Filename)
		object, err := s.store.Put(ctx, storage.PutRequest{
			Key:         key,
			Body:        bytes.NewReader(buf),
			Size:        int64(len(buf)),
			ContentType: file.ContentType,
		})
		if err != nil {
			return nil, err
		}
		record.StorageKey = object.Key
		record.StorageBackend = string(object.Backend)
		if memoryRepo, ok := s.repo.(*memoryMetadataRepository); ok {
			memoryRepo.files[file.ID].storageKey = object.Key
			memoryRepo.files[file.ID].storageBackend = string(object.Backend)
		}
	}
	if err := s.repo.SaveUpload(ctx, record); err != nil {
		if s.store != nil && record.StorageKey != "" {
			_ = s.store.Delete(ctx, record.StorageKey)
		}
		return nil, err
	}

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

	meta, err := s.repo.Get(ctx, fileID)
	if err != nil {
		return err
	}
	if err := s.repo.SoftDelete(ctx, activityID, fileID); err != nil {
		return err
	}
	if s.store != nil && meta.StorageKey != "" {
		return s.store.Delete(ctx, meta.StorageKey)
	}

	return nil
}

func (s *Service) Download(ctx context.Context, actorID string, actorRole string, fileID string) (*Download, error) {
	meta, err := s.repo.Get(ctx, fileID)
	if err != nil {
		return nil, err
	}

	activity, err := s.activities.Get(ctx, meta.ActivityID)
	if err != nil {
		return nil, err
	}
	if !activity.CanView(actorID, actorRole) {
		return nil, ErrFileAccessDenied
	}

	if s.store != nil {
		url, err := s.store.SignGetURL(ctx, meta.StorageKey, storage.SignedURLOptions{ExpiresIn: time.Hour})
		if err == nil {
			return &Download{
				File:        meta.File,
				ContentType: meta.ContentType,
				Content:     nil,
				SignedURL:   url,
			}, nil
		}
	}

	var content []byte
	if memoryRepo, ok := s.repo.(*memoryMetadataRepository); ok {
		if file, exists := memoryRepo.files[fileID]; exists {
			content = append([]byte(nil), file.content...)
		}
	}
	return &Download{
		File:        meta.File,
		Content:     content,
		ContentType: meta.ContentType,
	}, nil
}
