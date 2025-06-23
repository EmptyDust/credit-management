package handlers

import (
	"net/http"
	"os"
	"strconv"

	"credit-management/user-management-service/models"
	"credit-management/user-management-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileHandler struct {
	db                  *gorm.DB
	fileUploader        *utils.FileUploader
	notificationManager *utils.NotificationManager
}

func NewFileHandler(db *gorm.DB) *FileHandler {
	fileUploader := utils.NewFileUploader(nil)
	notificationManager := utils.NewNotificationManager(db)

	return &FileHandler{
		db:                  db,
		fileUploader:        fileUploader,
		notificationManager: notificationManager,
	}
}

// UploadFile 上传文件
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}

	var req models.FileUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 上传文件
	fileInfo, err := h.fileUploader.UploadFile(file, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}

	// 保存文件记录到数据库
	userFile := models.UserFile{
		UserID:       userID.(uint),
		FileName:     fileInfo.FileName,
		OriginalName: fileInfo.OriginalName,
		FilePath:     fileInfo.FilePath,
		FileSize:     fileInfo.FileSize,
		FileType:     fileInfo.FileType,
		MimeType:     fileInfo.MimeType,
		Category:     req.Category,
		Description:  req.Description,
		IsPublic:     req.IsPublic,
	}

	if err := h.db.Create(&userFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件记录失败"})
		return
	}

	// 发送通知
	h.notificationManager.SendTemplateNotification(userID.(uint), utils.FileUploadedTemplate, map[string]interface{}{
		"filename": fileInfo.OriginalName,
		"filesize": fileInfo.FileSize,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "文件上传成功",
		"file": models.FileResponse{
			ID:            userFile.ID,
			FileName:      userFile.FileName,
			OriginalName:  userFile.OriginalName,
			FileSize:      userFile.FileSize,
			FileType:      userFile.FileType,
			MimeType:      userFile.MimeType,
			Category:      userFile.Category,
			Description:   userFile.Description,
			IsPublic:      userFile.IsPublic,
			DownloadCount: userFile.DownloadCount,
			DownloadURL:   h.fileUploader.GetFileURL(userFile.FilePath),
			PreviewURL:    h.fileUploader.GetPreviewURL(userFile.FilePath),
			CreatedAt:     userFile.CreatedAt,
		},
	})
}

// DownloadFile 下载文件
func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var userFile models.UserFile
	if err := h.db.First(&userFile, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		}
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(userFile.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 增加下载次数
	h.db.Model(&userFile).Update("download_count", userFile.DownloadCount+1)

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+userFile.OriginalName)
	c.Header("Content-Type", userFile.MimeType)
	c.File(userFile.FilePath)
}

// GetFile 获取文件信息
func (h *FileHandler) GetFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var userFile models.UserFile
	if err := h.db.Preload("User").First(&userFile, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		}
		return
	}

	c.JSON(http.StatusOK, models.FileResponse{
		ID:            userFile.ID,
		FileName:      userFile.FileName,
		OriginalName:  userFile.OriginalName,
		FileSize:      userFile.FileSize,
		FileType:      userFile.FileType,
		MimeType:      userFile.MimeType,
		Category:      userFile.Category,
		Description:   userFile.Description,
		IsPublic:      userFile.IsPublic,
		DownloadCount: userFile.DownloadCount,
		DownloadURL:   h.fileUploader.GetFileURL(userFile.FilePath),
		PreviewURL:    h.fileUploader.GetPreviewURL(userFile.FilePath),
		CreatedAt:     userFile.CreatedAt,
	})
}

// GetUserFiles 获取用户文件列表
func (h *FileHandler) GetUserFiles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	isPublic := c.Query("is_public")

	query := h.db.Where("user_id = ?", userID.(uint))

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isPublic != "" {
		if isPublic == "true" {
			query = query.Where("is_public = ?", true)
		} else {
			query = query.Where("is_public = ?", false)
		}
	}

	var total int64
	query.Model(&models.UserFile{}).Count(&total)

	var userFiles []models.UserFile
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&userFiles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		return
	}

	var fileResponses []models.FileResponse
	for _, userFile := range userFiles {
		fileResponse := models.FileResponse{
			ID:            userFile.ID,
			FileName:      userFile.FileName,
			OriginalName:  userFile.OriginalName,
			FileSize:      userFile.FileSize,
			FileType:      userFile.FileType,
			MimeType:      userFile.MimeType,
			Category:      userFile.Category,
			Description:   userFile.Description,
			IsPublic:      userFile.IsPublic,
			DownloadCount: userFile.DownloadCount,
			DownloadURL:   h.fileUploader.GetFileURL(userFile.FilePath),
			PreviewURL:    h.fileUploader.GetPreviewURL(userFile.FilePath),
			CreatedAt:     userFile.CreatedAt,
		}
		fileResponses = append(fileResponses, fileResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"files":       fileResponses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (int(total) + pageSize - 1) / pageSize,
	})
}

// DeleteFile 删除文件
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var userFile models.UserFile
	if err := h.db.Where("id = ? AND user_id = ?", fileID, userID.(uint)).First(&userFile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在或无权限删除"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		}
		return
	}

	// 删除物理文件
	if err := h.fileUploader.DeleteFile(userFile.FilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件失败"})
		return
	}

	// 删除数据库记录
	if err := h.db.Delete(&userFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件删除成功"})
}

// UpdateFile 更新文件信息
func (h *FileHandler) UpdateFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var req struct {
		Description string `json:"description"`
		IsPublic    *bool  `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var userFile models.UserFile
	if err := h.db.Where("id = ? AND user_id = ?", fileID, userID.(uint)).First(&userFile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在或无权限修改"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		}
		return
	}

	updates := make(map[string]interface{})
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if err := h.db.Model(&userFile).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件更新成功"})
}

// GetPublicFiles 获取公开文件列表
func (h *FileHandler) GetPublicFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")

	query := h.db.Where("is_public = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	var total int64
	query.Model(&models.UserFile{}).Count(&total)

	var userFiles []models.UserFile
	offset := (page - 1) * pageSize
	err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&userFiles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		return
	}

	var fileResponses []models.FileResponse
	for _, userFile := range userFiles {
		fileResponse := models.FileResponse{
			ID:            userFile.ID,
			FileName:      userFile.FileName,
			OriginalName:  userFile.OriginalName,
			FileSize:      userFile.FileSize,
			FileType:      userFile.FileType,
			MimeType:      userFile.MimeType,
			Category:      userFile.Category,
			Description:   userFile.Description,
			IsPublic:      userFile.IsPublic,
			DownloadCount: userFile.DownloadCount,
			DownloadURL:   h.fileUploader.GetFileURL(userFile.FilePath),
			PreviewURL:    h.fileUploader.GetPreviewURL(userFile.FilePath),
			CreatedAt:     userFile.CreatedAt,
		}
		fileResponses = append(fileResponses, fileResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"files":       fileResponses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (int(total) + pageSize - 1) / pageSize,
	})
}

// GetFileStats 获取文件统计信息
func (h *FileHandler) GetFileStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var stats struct {
		TotalFiles    int64 `json:"total_files"`
		TotalSize     int64 `json:"total_size"`
		PublicFiles   int64 `json:"public_files"`
		PrivateFiles  int64 `json:"private_files"`
		ImageFiles    int64 `json:"image_files"`
		DocumentFiles int64 `json:"document_files"`
		VideoFiles    int64 `json:"video_files"`
		AudioFiles    int64 `json:"audio_files"`
		ArchiveFiles  int64 `json:"archive_files"`
	}

	// 总文件数和大小
	h.db.Model(&models.UserFile{}).Where("user_id = ?", userID.(uint)).Count(&stats.TotalFiles)
	h.db.Model(&models.UserFile{}).Where("user_id = ?", userID.(uint)).Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalSize)

	// 公开/私有文件数
	h.db.Model(&models.UserFile{}).Where("user_id = ? AND is_public = ?", userID.(uint), true).Count(&stats.PublicFiles)
	h.db.Model(&models.UserFile{}).Where("user_id = ? AND is_public = ?", userID.(uint), false).Count(&stats.PrivateFiles)

	// 各类型文件数
	h.db.Model(&models.UserFile{}).Where("user_id = ? AND file_type IN ('.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp')", userID.(uint)).Count(&stats.ImageFiles)
	h.db.Model(&models.UserFile{}).Where("user_id = ? AND file_type IN ('.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.txt')", userID.(uint)).Count(&stats.DocumentFiles)
	h.db.Model(&models.UserFile{}).Where("user_id = ? AND file_type IN ('.mp4', '.avi', '.mov', '.wmv', '.flv', '.mkv')", userID.(uint)).Count(&stats.VideoFiles)
	h.db.Model(&models.UserFile{}).Where("user_id = ? AND file_type IN ('.mp3', '.wav', '.flac', '.aac', '.ogg')", userID.(uint)).Count(&stats.AudioFiles)
	h.db.Model(&models.UserFile{}).Where("user_id = ? AND file_type IN ('.zip', '.rar', '.7z', '.tar', '.gz')", userID.(uint)).Count(&stats.ArchiveFiles)

	c.JSON(http.StatusOK, stats)
}
