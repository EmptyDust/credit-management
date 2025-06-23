package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// FileConfig 文件配置
type FileConfig struct {
	UploadDir       string   `json:"upload_dir"`
	MaxFileSize     int64    `json:"max_file_size"`    // 最大文件大小（字节）
	AllowedTypes    []string `json:"allowed_types"`    // 允许的文件类型
	ImageTypes      []string `json:"image_types"`      // 图片类型
	DocumentTypes   []string `json:"document_types"`   // 文档类型
	VideoTypes      []string `json:"video_types"`      // 视频类型
	AudioTypes      []string `json:"audio_types"`      // 音频类型
	ArchiveTypes    []string `json:"archive_types"`    // 压缩包类型
	ThumbnailDir    string   `json:"thumbnail_dir"`    // 缩略图目录
	PreviewDir      string   `json:"preview_dir"`      // 预览文件目录
	TempDir         string   `json:"temp_dir"`         // 临时文件目录
	PublicURL       string   `json:"public_url"`       // 公共访问URL
	EnablePreview   bool     `json:"enable_preview"`   // 是否启用预览
	EnableThumbnail bool     `json:"enable_thumbnail"` // 是否启用缩略图
}

// DefaultFileConfig 默认文件配置
func DefaultFileConfig() *FileConfig {
	return &FileConfig{
		UploadDir:       "./uploads",
		MaxFileSize:     50 * 1024 * 1024, // 50MB
		AllowedTypes:    []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".zip", ".rar", ".7z"},
		ImageTypes:      []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"},
		DocumentTypes:   []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"},
		VideoTypes:      []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"},
		AudioTypes:      []string{".mp3", ".wav", ".flac", ".aac", ".ogg"},
		ArchiveTypes:    []string{".zip", ".rar", ".7z", ".tar", ".gz"},
		ThumbnailDir:    "./thumbnails",
		PreviewDir:      "./previews",
		TempDir:         "./temp",
		PublicURL:       "http://localhost:8080/files",
		EnablePreview:   true,
		EnableThumbnail: true,
	}
}

// FileInfo 文件信息
type FileInfo struct {
	OriginalName string    `json:"original_name"`
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	FileType     string    `json:"file_type"`
	MimeType     string    `json:"mime_type"`
	MD5Hash      string    `json:"md5_hash"`
	UploadTime   time.Time `json:"upload_time"`
	IsImage      bool      `json:"is_image"`
	IsDocument   bool      `json:"is_document"`
	IsVideo      bool      `json:"is_video"`
	IsAudio      bool      `json:"is_audio"`
	IsArchive    bool      `json:"is_archive"`
}

// FileUploader 文件上传器
type FileUploader struct {
	config *FileConfig
}

// NewFileUploader 创建文件上传器
func NewFileUploader(config *FileConfig) *FileUploader {
	if config == nil {
		config = DefaultFileConfig()
	}

	// 创建必要的目录
	os.MkdirAll(config.UploadDir, 0755)
	os.MkdirAll(config.ThumbnailDir, 0755)
	os.MkdirAll(config.PreviewDir, 0755)
	os.MkdirAll(config.TempDir, 0755)

	return &FileUploader{config: config}
}

// UploadFile 上传文件
func (fu *FileUploader) UploadFile(file *multipart.FileHeader, category string) (*FileInfo, error) {
	// 检查文件大小
	if file.Size > fu.config.MaxFileSize {
		return nil, fmt.Errorf("文件大小超过限制: %d bytes", fu.config.MaxFileSize)
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !fu.isAllowedType(ext) {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 生成文件名
	fileName := fu.generateFileName(file.Filename)

	// 创建分类目录
	categoryDir := filepath.Join(fu.config.UploadDir, category)
	os.MkdirAll(categoryDir, 0755)

	// 文件路径
	filePath := filepath.Join(categoryDir, fileName)

	// 保存文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// 计算MD5
	hash := md5.New()
	tee := io.TeeReader(src, hash)

	_, err = io.Copy(dst, tee)
	if err != nil {
		return nil, err
	}

	md5Hash := fmt.Sprintf("%x", hash.Sum(nil))

	// 获取MIME类型
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = fu.getMimeType(ext)
	}

	fileInfo := &FileInfo{
		OriginalName: file.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     file.Size,
		FileType:     ext,
		MimeType:     mimeType,
		MD5Hash:      md5Hash,
		UploadTime:   time.Now(),
		IsImage:      fu.isImageType(ext),
		IsDocument:   fu.isDocumentType(ext),
		IsVideo:      fu.isVideoType(ext),
		IsAudio:      fu.isAudioType(ext),
		IsArchive:    fu.isArchiveType(ext),
	}

	return fileInfo, nil
}

// generateFileName 生成文件名
func (fu *FileUploader) generateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	hash := fmt.Sprintf("%x", md5.Sum([]byte(originalName+time.Now().String())))[:8]
	return fmt.Sprintf("%d_%s%s", timestamp, hash, ext)
}

// isAllowedType 检查是否为允许的文件类型
func (fu *FileUploader) isAllowedType(ext string) bool {
	for _, allowedType := range fu.config.AllowedTypes {
		if ext == allowedType {
			return true
		}
	}
	return false
}

// isImageType 检查是否为图片类型
func (fu *FileUploader) isImageType(ext string) bool {
	for _, imageType := range fu.config.ImageTypes {
		if ext == imageType {
			return true
		}
	}
	return false
}

// isDocumentType 检查是否为文档类型
func (fu *FileUploader) isDocumentType(ext string) bool {
	for _, docType := range fu.config.DocumentTypes {
		if ext == docType {
			return true
		}
	}
	return false
}

// isVideoType 检查是否为视频类型
func (fu *FileUploader) isVideoType(ext string) bool {
	for _, videoType := range fu.config.VideoTypes {
		if ext == videoType {
			return true
		}
	}
	return false
}

// isAudioType 检查是否为音频类型
func (fu *FileUploader) isAudioType(ext string) bool {
	for _, audioType := range fu.config.AudioTypes {
		if ext == audioType {
			return true
		}
	}
	return false
}

// isArchiveType 检查是否为压缩包类型
func (fu *FileUploader) isArchiveType(ext string) bool {
	for _, archiveType := range fu.config.ArchiveTypes {
		if ext == archiveType {
			return true
		}
	}
	return false
}

// getMimeType 获取MIME类型
func (fu *FileUploader) getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".7z":   "application/x-7z-compressed",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".mkv":  "video/x-matroska",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".ogg":  "audio/ogg",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream"
}

// GetFileURL 获取文件URL
func (fu *FileUploader) GetFileURL(filePath string) string {
	// 将文件路径转换为相对路径
	relPath, err := filepath.Rel(fu.config.UploadDir, filePath)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", fu.config.PublicURL, relPath)
}

// GetPreviewURL 获取预览URL
func (fu *FileUploader) GetPreviewURL(filePath string) string {
	if !fu.config.EnablePreview {
		return ""
	}

	relPath, err := filepath.Rel(fu.config.UploadDir, filePath)
	if err != nil {
		return ""
	}

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)

	// 对于图片，直接返回原图URL
	if fu.isImageType(ext) {
		return fu.GetFileURL(filePath)
	}

	// 对于文档，返回预览URL
	if fu.isDocumentType(ext) {
		previewPath := filepath.Join(fu.config.PreviewDir, relPath)
		previewPath = strings.TrimSuffix(previewPath, ext) + ".html"
		return fmt.Sprintf("%s/preview/%s", fu.config.PublicURL, strings.TrimPrefix(previewPath, "./"))
	}

	return ""
}

// GetThumbnailURL 获取缩略图URL
func (fu *FileUploader) GetThumbnailURL(filePath string) string {
	if !fu.config.EnableThumbnail {
		return ""
	}

	relPath, err := filepath.Rel(fu.config.UploadDir, filePath)
	if err != nil {
		return ""
	}

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)

	// 只对图片生成缩略图
	if !fu.isImageType(ext) {
		return ""
	}

	thumbnailPath := filepath.Join(fu.config.ThumbnailDir, relPath)
	thumbnailPath = strings.TrimSuffix(thumbnailPath, ext) + "_thumb" + ext

	return fmt.Sprintf("%s/thumbnail/%s", fu.config.PublicURL, strings.TrimPrefix(thumbnailPath, "./"))
}

// DeleteFile 删除文件
func (fu *FileUploader) DeleteFile(filePath string) error {
	// 删除原文件
	if err := os.Remove(filePath); err != nil {
		return err
	}

	// 删除缩略图
	thumbnailPath := fu.getThumbnailPath(filePath)
	if _, err := os.Stat(thumbnailPath); err == nil {
		os.Remove(thumbnailPath)
	}

	// 删除预览文件
	previewPath := fu.getPreviewPath(filePath)
	if _, err := os.Stat(previewPath); err == nil {
		os.Remove(previewPath)
	}

	return nil
}

// getThumbnailPath 获取缩略图路径
func (fu *FileUploader) getThumbnailPath(filePath string) string {
	relPath, _ := filepath.Rel(fu.config.UploadDir, filePath)
	return filepath.Join(fu.config.ThumbnailDir, relPath)
}

// getPreviewPath 获取预览文件路径
func (fu *FileUploader) getPreviewPath(filePath string) string {
	relPath, _ := filepath.Rel(fu.config.UploadDir, filePath)
	ext := filepath.Ext(filePath)
	return filepath.Join(fu.config.PreviewDir, strings.TrimSuffix(relPath, ext)+".html")
}

// FileMiddleware 文件上传中间件
func FileMiddleware(uploader *FileUploader) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("fileUploader", uploader)
		c.Next()
	}
}
