package service

import (
	"context"
	"errors"
	"time"

	"credit-management/backend/internal/platform/idgen"

	"gorm.io/gorm"
)

type MetadataRecord struct {
	File
	StorageKey     string
	StorageBackend string
}

type MetadataRepository interface {
	SaveUpload(ctx context.Context, meta MetadataRecord) error
	Get(ctx context.Context, fileID string) (*MetadataRecord, error)
	ListByActivity(ctx context.Context, activityID string) ([]File, error)
	SoftDelete(ctx context.Context, activityID string, fileID string) error
}

type memoryMetadataRepository struct {
	files      map[string]*storedFile
	byActivity map[string][]string
}

type gormMetadataRepository struct {
	db *gorm.DB
}

type fileRow struct {
	ID             string     `gorm:"column:id"`
	StorageKey     string     `gorm:"column:storage_key"`
	OriginalName   string     `gorm:"column:original_name"`
	MimeType       string     `gorm:"column:mime_type"`
	FileSize       int64      `gorm:"column:file_size"`
	StorageBackend string     `gorm:"column:storage_backend"`
	UploadedBy     string     `gorm:"column:uploaded_by"`
	UploadedAt     time.Time  `gorm:"column:uploaded_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

type activityFileRow struct {
	ID          string     `gorm:"column:id"`
	ActivityID  string     `gorm:"column:activity_id"`
	FileID      string     `gorm:"column:file_id"`
	Description *string    `gorm:"column:description"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

type joinedFileRow struct {
	ID             string    `gorm:"column:id"`
	ActivityID     string    `gorm:"column:activity_id"`
	StorageKey     string    `gorm:"column:storage_key"`
	OriginalName   string    `gorm:"column:original_name"`
	MimeType       string    `gorm:"column:mime_type"`
	FileSize       int64     `gorm:"column:file_size"`
	StorageBackend string    `gorm:"column:storage_backend"`
	UploadedBy     string    `gorm:"column:uploaded_by"`
	UploadedAt     time.Time `gorm:"column:uploaded_at"`
}

func (fileRow) TableName() string {
	return "files"
}

func (activityFileRow) TableName() string {
	return "activity_files"
}

func newMemoryMetadataRepository(files map[string]*storedFile, byActivity map[string][]string) MetadataRepository {
	return &memoryMetadataRepository{
		files:      files,
		byActivity: byActivity,
	}
}

func NewGormMetadataRepository(db *gorm.DB) MetadataRepository {
	return &gormMetadataRepository{db: db}
}

func (r *memoryMetadataRepository) SaveUpload(_ context.Context, meta MetadataRecord) error {
	file, ok := r.files[meta.ID]
	if !ok {
		file = &storedFile{}
		r.files[meta.ID] = file
	}
	file.meta = File{
		ID:          meta.ID,
		ActivityID:  meta.ActivityID,
		Name:        meta.Name,
		ContentType: meta.ContentType,
		Size:        meta.Size,
		UploadedBy:  meta.UploadedBy,
		CreatedAt:   meta.CreatedAt,
	}
	file.storageKey = meta.StorageKey
	r.byActivity[meta.ActivityID] = append(r.byActivity[meta.ActivityID], meta.ID)
	return nil
}

func (r *memoryMetadataRepository) Get(_ context.Context, fileID string) (*MetadataRecord, error) {
	file, ok := r.files[fileID]
	if !ok {
		return nil, ErrFileNotFound
	}
	return &MetadataRecord{File: file.meta, StorageKey: file.storageKey, StorageBackend: file.storageBackend}, nil
}

func (r *memoryMetadataRepository) ListByActivity(_ context.Context, activityID string) ([]File, error) {
	ids := r.byActivity[activityID]
	items := make([]File, 0, len(ids))
	for _, id := range ids {
		if file, ok := r.files[id]; ok {
			items = append(items, file.meta)
		}
	}
	return items, nil
}

func (r *memoryMetadataRepository) SoftDelete(_ context.Context, activityID string, fileID string) error {
	if _, ok := r.files[fileID]; !ok {
		return ErrFileNotFound
	}
	delete(r.files, fileID)

	ids := r.byActivity[activityID]
	filtered := ids[:0]
	for _, id := range ids {
		if id != fileID {
			filtered = append(filtered, id)
		}
	}
	r.byActivity[activityID] = filtered
	return nil
}

func (r *gormMetadataRepository) SaveUpload(ctx context.Context, meta MetadataRecord) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&fileRow{
			ID:             meta.ID,
			StorageKey:     meta.StorageKey,
			OriginalName:   meta.Name,
			MimeType:       meta.ContentType,
			FileSize:       meta.Size,
			StorageBackend: meta.StorageBackend,
			UploadedBy:     meta.UploadedBy,
			UploadedAt:     meta.CreatedAt,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&activityFileRow{
			ID:         idgen.New(),
			ActivityID: meta.ActivityID,
			FileID:     meta.ID,
		}).Error
	})
}

func (r *gormMetadataRepository) Get(ctx context.Context, fileID string) (*MetadataRecord, error) {
	var row joinedFileRow
	if err := r.db.WithContext(ctx).
		Table("files").
		Select("files.id, activity_files.activity_id, files.storage_key, files.original_name, files.mime_type, files.file_size, files.storage_backend, files.uploaded_by, files.uploaded_at").
		Joins("JOIN activity_files ON activity_files.file_id = files.id AND activity_files.deleted_at IS NULL").
		Where("files.id = ? AND files.deleted_at IS NULL", fileID).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return &MetadataRecord{
		File: File{
			ID:          row.ID,
			ActivityID:  row.ActivityID,
			Name:        row.OriginalName,
			ContentType: row.MimeType,
			Size:        row.FileSize,
			UploadedBy:  row.UploadedBy,
			CreatedAt:   row.UploadedAt.UTC(),
		},
		StorageKey:     row.StorageKey,
		StorageBackend: row.StorageBackend,
	}, nil
}

func (r *gormMetadataRepository) ListByActivity(ctx context.Context, activityID string) ([]File, error) {
	var rows []joinedFileRow
	if err := r.db.WithContext(ctx).
		Table("files").
		Select("files.id, activity_files.activity_id, files.storage_key, files.original_name, files.mime_type, files.file_size, files.storage_backend, files.uploaded_by, files.uploaded_at").
		Joins("JOIN activity_files ON activity_files.file_id = files.id AND activity_files.deleted_at IS NULL").
		Where("activity_files.activity_id = ? AND files.deleted_at IS NULL", activityID).
		Order("files.uploaded_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]File, 0, len(rows))
	for i := range rows {
		items = append(items, File{
			ID:          rows[i].ID,
			ActivityID:  rows[i].ActivityID,
			Name:        rows[i].OriginalName,
			ContentType: rows[i].MimeType,
			Size:        rows[i].FileSize,
			UploadedBy:  rows[i].UploadedBy,
			CreatedAt:   rows[i].UploadedAt.UTC(),
		})
	}
	return items, nil
}

func (r *gormMetadataRepository) SoftDelete(ctx context.Context, activityID string, fileID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		linkResult := tx.Model(&activityFileRow{}).
			Where("activity_id = ? AND file_id = ? AND deleted_at IS NULL", activityID, fileID).
			Update("deleted_at", now)
		if linkResult.Error != nil {
			return linkResult.Error
		}
		if linkResult.RowsAffected == 0 {
			return ErrFileNotFound
		}

		return tx.Model(&fileRow{}).
			Where("id = ? AND deleted_at IS NULL", fileID).
			Update("deleted_at", now).Error
	})
}
