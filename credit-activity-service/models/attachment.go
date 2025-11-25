package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Attachment 附件模型
type Attachment struct {
	ID            string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityID    string         `json:"activity_id" gorm:"type:uuid;not null;index"`
	FileName      string         `json:"file_name" gorm:"not null"`     // 存储的文件名
	OriginalName  string         `json:"original_name" gorm:"not null"` // 原始文件名
	FileSize      int64          `json:"file_size" gorm:"not null"`     // 文件大小（字节）
	FileType      string         `json:"file_type" gorm:"not null"`     // 文件扩展名
	FileCategory  string         `json:"file_category" gorm:"not null"` // 文件类别
	Description   string         `json:"description"`                   // 文件描述
	UploadedBy    string         `json:"uploaded_by" gorm:"type:uuid;not null;index"`
	UploadedAt    time.Time      `json:"uploaded_at" gorm:"default:CURRENT_TIMESTAMP"`
	DownloadCount int64          `json:"download_count" gorm:"default:0"` // 下载次数
	MD5Hash       string         `json:"md5_hash" gorm:"uniqueIndex"`     // 文件MD5哈希（可选）
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联关系
	Activity CreditActivity `json:"activity" gorm:"foreignKey:ActivityID"`
	// Uploader field is populated manually in handlers, not via GORM foreign key
}

func (a *Attachment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func (Attachment) TableName() string {
	return "attachments"
}

// AttachmentUploadRequest 附件上传请求
type AttachmentUploadRequest struct {
	Description string `json:"description" form:"description"`
}

// AttachmentUpdateRequest 附件更新请求
type AttachmentUpdateRequest struct {
	Description string `json:"description" binding:"required"`
}

// AttachmentResponse 附件响应
type AttachmentResponse struct {
	ID            string    `json:"id"`
	ActivityID    string    `json:"activity_id"`
	FileName      string    `json:"file_name"`
	OriginalName  string    `json:"original_name"`
	FileSize      int64     `json:"file_size"`
	FileType      string    `json:"file_type"`
	FileCategory  string    `json:"file_category"`
	Description   string    `json:"description"`
	UploadedBy    string    `json:"uploaded_by"`
	UploadedAt    time.Time `json:"uploaded_at"`
	DownloadCount int64     `json:"download_count"`
	DownloadURL   string    `json:"download_url"`
	Uploader      UserInfo  `json:"uploader"`
}

// AttachmentStats 附件统计
type AttachmentStats struct {
	TotalCount    int64            `json:"total_count"`
	TotalSize     int64            `json:"total_size"`
	CategoryCount map[string]int64 `json:"category_count"`
	FileTypeCount map[string]int64 `json:"file_type_count"`
}

// BatchUploadResult 批量上传结果
type BatchUploadResult struct {
	FileName string `json:"file_name"`
	Status   string `json:"status"` // success, failed
	FileID   string `json:"file_id,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	Message  string `json:"message,omitempty"`
}

// BatchUploadResponse 批量上传响应
type BatchUploadResponse struct {
	TotalFiles   int                 `json:"total_files"`
	SuccessCount int                 `json:"success_count"`
	FailCount    int                 `json:"fail_count"`
	Results      []BatchUploadResult `json:"results"`
}

// 文件类别常量
const (
	CategoryDocument     = "document"     // 文档
	CategoryImage        = "image"        // 图片
	CategoryVideo        = "video"        // 视频
	CategoryAudio        = "audio"        // 音频
	CategoryArchive      = "archive"      // 压缩包
	CategorySpreadsheet  = "spreadsheet"  // 表格
	CategoryPresentation = "presentation" // 演示文稿
	CategoryOther        = "other"        // 其他
)

// 支持的文件类型映射
var SupportedFileTypes = map[string]string{
	// 文档
	".pdf":  CategoryDocument,
	".doc":  CategoryDocument,
	".docx": CategoryDocument,
	".txt":  CategoryDocument,
	".rtf":  CategoryDocument,
	".odt":  CategoryDocument,

	// 图片
	".jpg":  CategoryImage,
	".jpeg": CategoryImage,
	".png":  CategoryImage,
	".gif":  CategoryImage,
	".bmp":  CategoryImage,
	".webp": CategoryImage,

	// 视频
	".mp4": CategoryVideo,
	".avi": CategoryVideo,
	".mov": CategoryVideo,
	".wmv": CategoryVideo,
	".flv": CategoryVideo,

	// 音频
	".mp3": CategoryAudio,
	".wav": CategoryAudio,
	".ogg": CategoryAudio,
	".aac": CategoryAudio,

	// 压缩包
	".zip": CategoryArchive,
	".rar": CategoryArchive,
	".7z":  CategoryArchive,
	".tar": CategoryArchive,
	".gz":  CategoryArchive,

	// 表格
	".xls":  CategorySpreadsheet,
	".xlsx": CategorySpreadsheet,
	".csv":  CategorySpreadsheet,

	// 演示文稿
	".ppt":  CategoryPresentation,
	".pptx": CategoryPresentation,
}

// 文件大小限制（20MB）
const MaxFileSize = 20 * 1024 * 1024

// 批量上传文件数量限制
const MaxBatchUploadCount = 10

func GetFileCategory(fileType string) string {
	if category, exists := SupportedFileTypes[fileType]; exists {
		return category
	}
	return CategoryOther
}

func IsSupportedFileType(fileType string) bool {
	_, exists := SupportedFileTypes[fileType]
	return exists
}

func GetSupportedFileTypes() []string {
	var types []string
	for fileType := range SupportedFileTypes {
		types = append(types, fileType)
	}
	return types
}

func GetFileCategories() []string {
	return []string{
		CategoryDocument,
		CategoryImage,
		CategoryVideo,
		CategoryAudio,
		CategoryArchive,
		CategorySpreadsheet,
		CategoryPresentation,
		CategoryOther,
	}
}
