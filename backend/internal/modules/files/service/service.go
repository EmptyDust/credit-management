package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
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
	ErrFileValidation     = errors.New("file validation failed")
)

type File struct {
	ID            string    `json:"id"`
	ActivityID    string    `json:"activity_id"`
	Name          string    `json:"name"`
	OriginalName  string    `json:"original_name"`
	ContentType   string    `json:"content_type"`
	FileCategory  string    `json:"file_category"`
	Description   string    `json:"description"`
	Size          int64     `json:"size"`
	UploadedBy    string    `json:"uploaded_by"`
	CreatedAt     time.Time `json:"created_at"`
	DownloadCount int64     `json:"download_count"`
	MD5Hash       string    `json:"md5_hash"`
}

type Download struct {
	File        File
	Content     []byte
	ContentType string
	SignedURL   string
}

type UploadOptions struct {
	FileCategory string
	Description  string
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
	return s.UploadWithOptions(ctx, actorID, actorRole, activityID, fileHeader, UploadOptions{})
}

func (s *Service) UploadMany(ctx context.Context, actorID string, actorRole string, activityID string, fileHeaders []*multipart.FileHeader, options UploadOptions) ([]File, error) {
	if len(fileHeaders) == 0 {
		return nil, ErrFileValidation
	}
	if len(fileHeaders) > 10 {
		return nil, ErrFileValidation
	}

	items := make([]File, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		file, err := s.UploadWithOptions(ctx, actorID, actorRole, activityID, fileHeader, options)
		if err != nil {
			return nil, err
		}
		items = append(items, *file)
	}

	return items, nil
}

func (s *Service) UploadWithOptions(ctx context.Context, actorID string, actorRole string, activityID string, fileHeader *multipart.FileHeader, options UploadOptions) (*File, error) {
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

	contentType := normalizeContentType(fileHeader.Header.Get("Content-Type"), fileHeader.Filename, buf)
	if err := validateFile(fileHeader.Filename, contentType, int64(len(buf))); err != nil {
		return nil, err
	}

	hash := md5.Sum(buf)
	file := File{
		ID:           idgen.New(),
		ActivityID:   activityID,
		Name:         fileHeader.Filename,
		OriginalName: fileHeader.Filename,
		ContentType:  contentType,
		FileCategory: strings.TrimSpace(options.FileCategory),
		Description:  strings.TrimSpace(options.Description),
		Size:         int64(len(buf)),
		UploadedBy:   actorID,
		CreatedAt:    time.Now().UTC(),
		MD5Hash:      hex.EncodeToString(hash[:]),
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

func (s *Service) CopyActivityFiles(ctx context.Context, sourceActivityID string, targetActivityID string, actorID string) error {
	sourceFiles, err := s.repo.ListRecordsByActivity(ctx, sourceActivityID)
	if err != nil {
		return err
	}
	if len(sourceFiles) == 0 {
		return nil
	}

	for _, source := range sourceFiles {
		copied := File{
			ID:           idgen.New(),
			ActivityID:   targetActivityID,
			Name:         source.Name,
			OriginalName: source.OriginalName,
			ContentType:  source.ContentType,
			FileCategory: source.FileCategory,
			Description:  source.Description,
			Size:         source.Size,
			UploadedBy:   actorID,
			CreatedAt:    time.Now().UTC(),
			MD5Hash:      source.MD5Hash,
		}
		record := MetadataRecord{
			File:           copied,
			StorageBackend: source.StorageBackend,
		}

		if memoryRepo, ok := s.repo.(*memoryMetadataRepository); ok {
			sourceFile, exists := memoryRepo.files[source.ID]
			if !exists {
				return ErrFileNotFound
			}
			memoryRepo.files[copied.ID] = &storedFile{
				meta:           copied,
				content:        append([]byte(nil), sourceFile.content...),
				storageBackend: sourceFile.storageBackend,
			}
		}

		if s.store != nil {
			reader, err := s.store.Get(ctx, source.StorageKey)
			if err != nil {
				return err
			}
			key := fmt.Sprintf("activities/%s/%s-%s", targetActivityID, idgen.New(), source.Name)
			object, putErr := s.store.Put(ctx, storage.PutRequest{
				Key:         key,
				Body:        reader,
				Size:        source.Size,
				ContentType: source.ContentType,
			})
			closeErr := reader.Close()
			if putErr != nil {
				return putErr
			}
			if closeErr != nil {
				return closeErr
			}
			record.StorageKey = object.Key
			record.StorageBackend = string(object.Backend)
			if memoryRepo, ok := s.repo.(*memoryMetadataRepository); ok {
				memoryRepo.files[copied.ID].storageKey = object.Key
				memoryRepo.files[copied.ID].storageBackend = string(object.Backend)
			}
		}

		if err := s.repo.SaveUpload(ctx, record); err != nil {
			if s.store != nil && record.StorageKey != "" {
				_ = s.store.Delete(ctx, record.StorageKey)
			}
			return err
		}
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
	if err := s.repo.IncrementDownloadCount(ctx, meta.ActivityID, fileID); err != nil {
		return nil, err
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

func (s *Service) Preview(ctx context.Context, actorID string, actorRole string, fileID string) (*Download, error) {
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
	if !isPreviewable(meta.Name, meta.ContentType) {
		return nil, ErrFilePreviewInvalid
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

func normalizeContentType(contentType string, filename string, content []byte) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(content)
	}
	if strings.HasPrefix(contentType, "text/plain") && strings.EqualFold(filepath.Ext(filename), ".txt") {
		return "text/plain"
	}
	return contentType
}

func validateFile(filename string, contentType string, size int64) error {
	if size <= 0 || size > 50*1024*1024 {
		return ErrFileValidation
	}

	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string][]string{
		".txt":  {"text/plain"},
		".pdf":  {"application/pdf"},
		".jpg":  {"image/jpeg"},
		".jpeg": {"image/jpeg"},
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".csv":  {"text/csv", "text/plain", "application/vnd.ms-excel"},
		".xls":  {"application/vnd.ms-excel", "application/octet-stream"},
		".xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/octet-stream"},
		".ppt":  {"application/vnd.ms-powerpoint", "application/octet-stream"},
		".pptx": {"application/vnd.openxmlformats-officedocument.presentationml.presentation", "application/octet-stream"},
		".doc":  {"application/msword", "application/octet-stream"},
		".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/octet-stream"},
		".zip":  {"application/zip", "application/x-zip-compressed", "application/octet-stream"},
		".rar":  {"application/vnd.rar", "application/x-rar-compressed", "application/octet-stream"},
		".7z":   {"application/x-7z-compressed", "application/octet-stream"},
		".mp4":  {"video/mp4", "application/octet-stream"},
		".mov":  {"video/quicktime", "application/octet-stream"},
		".mp3":  {"audio/mpeg", "application/octet-stream"},
		".wav":  {"audio/wav", "audio/x-wav", "application/octet-stream"},
	}

	allowedTypes, ok := allowed[ext]
	if !ok {
		return ErrFileValidation
	}
	for _, allowedType := range allowedTypes {
		if strings.HasPrefix(contentType, allowedType) {
			return nil
		}
	}

	return ErrFileValidation
}

func isPreviewable(filename string, contentType string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".txt", ".pdf", ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return strings.HasPrefix(contentType, "text/plain") ||
		strings.HasPrefix(contentType, "application/pdf") ||
		strings.HasPrefix(contentType, "image/")
}
