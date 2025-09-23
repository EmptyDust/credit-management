package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AttachmentHandler 附件处理器
type AttachmentHandler struct {
	db *gorm.DB
}

// NewAttachmentHandler 创建附件处理器
func NewAttachmentHandler(db *gorm.DB) *AttachmentHandler {
	return &AttachmentHandler{db: db}
}

func (h *AttachmentHandler) GetAttachments(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		utils.SendBadRequest(c, "活动ID不能为空")
		return
	}

	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	category := c.Query("category")
	fileType := c.Query("file_type")
	uploadedBy := c.Query("uploaded_by")

	query := h.db.Model(&models.Attachment{}).Where("activity_id = ? AND deleted_at IS NULL", activityID)

	if category != "" {
		query = query.Where("file_category = ?", category)
	}
	if fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	if uploadedBy != "" {
		query = query.Where("uploaded_by = ?", uploadedBy)
	}

	var attachments []models.Attachment
	if err := query.Order("uploaded_at DESC").Find(&attachments).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var responses []models.AttachmentResponse
	var totalSize int64
	categoryCount := make(map[string]int64)
	fileTypeCount := make(map[string]int64)

	authToken := c.GetHeader("Authorization")

	for _, attachment := range attachments {
		response := models.AttachmentResponse{
			ID:            attachment.ID,
			ActivityID:    attachment.ActivityID,
			FileName:      attachment.FileName,
			OriginalName:  attachment.OriginalName,
			FileSize:      attachment.FileSize,
			FileType:      attachment.FileType,
			FileCategory:  attachment.FileCategory,
			Description:   attachment.Description,
			UploadedBy:    attachment.UploadedBy,
			UploadedAt:    attachment.UploadedAt,
			DownloadCount: attachment.DownloadCount,
			DownloadURL:   fmt.Sprintf("/api/activities/%s/attachments/%s/download", activityID, attachment.ID),
		}

		if userInfo, err := utils.GetUserInfo(attachment.UploadedBy, authToken); err == nil {
			response.Uploader = *userInfo
		}

		responses = append(responses, response)

		totalSize += attachment.FileSize
		categoryCount[attachment.FileCategory]++
		fileTypeCount[attachment.FileType]++
	}

	stats := models.AttachmentStats{
		TotalCount:    int64(len(attachments)),
		TotalSize:     totalSize,
		CategoryCount: categoryCount,
		FileTypeCount: fileTypeCount,
	}

	utils.SendSuccessResponse(c, gin.H{
		"attachments": responses,
		"stats":       stats,
	})
}

func (h *AttachmentHandler) UploadAttachment(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		utils.SendBadRequest(c, "活动ID不能为空")
		return
	}

	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.SendBadRequest(c, "未找到上传的文件")
		return
	}
	defer file.Close()

	description := c.PostForm("description")

	if err := h.validateFile(header); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	// 读取文件内容并计算MD5
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	md5Hash := fmt.Sprintf("%x", md5.Sum(fileBytes))

	fileExt := filepath.Ext(header.Filename)

	// 创建存储目录
	uploadDir := "uploads/attachments"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	md5FilePath := filepath.Join(uploadDir, md5Hash+fileExt)
	if _, err := os.Stat(md5FilePath); os.IsNotExist(err) {
		if err := os.WriteFile(md5FilePath, fileBytes, 0644); err != nil {
			utils.SendInternalServerError(c, err)
			return
		}
	}
	userID, _ := c.Get("id")

	attachment := models.Attachment{
		ActivityID:   activityID,
		FileName:     md5Hash + fileExt, // 物理文件名统一用md5+ext
		OriginalName: header.Filename,
		FileSize:     header.Size,
		FileType:     fileExt,
		FileCategory: models.GetFileCategory(fileExt),
		MD5Hash:      md5Hash,
		Description:  description,
		UploadedBy:   userID.(string),
		UploadedAt:   time.Now(),
	}

	if err := h.db.Create(&attachment).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	response := models.AttachmentResponse{
		ID:            attachment.ID,
		ActivityID:    attachment.ActivityID,
		FileName:      attachment.FileName,
		OriginalName:  attachment.OriginalName,
		FileSize:      attachment.FileSize,
		FileType:      attachment.FileType,
		FileCategory:  attachment.FileCategory,
		Description:   attachment.Description,
		UploadedBy:    attachment.UploadedBy,
		UploadedAt:    attachment.UploadedAt,
		DownloadCount: attachment.DownloadCount,
		DownloadURL:   fmt.Sprintf("/api/activities/%s/attachments/%s/download", activityID, attachment.ID),
	}

	utils.SendCreatedResponse(c, "附件上传成功", response)
}

func (h *AttachmentHandler) BatchUploadAttachments(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		utils.SendBadRequest(c, "活动ID不能为空")
		return
	}

	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.SendBadRequest(c, "获取上传文件失败: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.SendBadRequest(c, "未找到上传的文件")
		return
	}

	if len(files) > models.MaxBatchUploadCount {
		utils.SendBadRequest(c, fmt.Sprintf("批量上传文件数量不能超过%d个", models.MaxBatchUploadCount))
		return
	}

	uploadDir := "uploads/attachments"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	var results []models.BatchUploadResult
	successCount := 0
	failCount := 0

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, fileHeader := range files {
		result := models.BatchUploadResult{
			FileName: fileHeader.Filename,
		}

		// 验证文件
		if err := h.validateFile(fileHeader); err != nil {
			result.Status = "failed"
			result.Message = err.Error()
			results = append(results, result)
			failCount++
			continue
		}

		// 打开文件
		file, err := fileHeader.Open()
		if err != nil {
			result.Status = "failed"
			result.Message = "打开文件失败: " + err.Error()
			results = append(results, result)
			failCount++
			continue
		}

		// 读取文件内容
		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			result.Status = "failed"
			result.Message = "读取文件失败: " + err.Error()
			results = append(results, result)
			failCount++
			continue
		}

		// 计算MD5
		md5Hash := fmt.Sprintf("%x", md5.Sum(fileBytes))

		// 检查文件是否已存在
		var existingAttachment models.Attachment
		if err := tx.Where("md5_hash = ? AND activity_id = ? AND deleted_at IS NULL", md5Hash, activityID).First(&existingAttachment).Error; err == nil {
			result.Status = "failed"
			result.Message = "文件已存在"
			results = append(results, result)
			failCount++
			continue
		}

		// 生成存储文件名
		fileExt := filepath.Ext(fileHeader.Filename)
		fileName := fmt.Sprintf("%s_%d%s", md5Hash[:8], time.Now().UnixNano(), fileExt)

		// 保存文件
		filePath := filepath.Join(uploadDir, fileName)
		if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
			result.Status = "failed"
			result.Message = "保存文件失败: " + err.Error()
			results = append(results, result)
			failCount++
			continue
		}

		userID, _ := c.Get("id")
		// 创建附件记录
		attachment := models.Attachment{
			ActivityID:   activityID,
			FileName:     fileName,
			OriginalName: fileHeader.Filename,
			FileSize:     fileHeader.Size,
			FileType:     fileExt,
			FileCategory: models.GetFileCategory(fileExt),
			MD5Hash:      md5Hash,
			Description:  "", // 批量上传时默认为空描述
			UploadedBy:   userID.(string),
			UploadedAt:   time.Now(),
		}

		if err := tx.Create(&attachment).Error; err != nil {
			// 删除已保存的文件
			os.Remove(filePath)
			result.Status = "failed"
			result.Message = "创建附件记录失败: " + err.Error()
			results = append(results, result)
			failCount++
			continue
		}

		result.Status = "success"
		result.FileID = attachment.ID
		result.FileSize = attachment.FileSize
		results = append(results, result)
		successCount++
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	response := models.BatchUploadResponse{
		TotalFiles:   len(files),
		SuccessCount: successCount,
		FailCount:    failCount,
		Results:      results,
	}

	utils.SendSuccessResponse(c, response)
}

func (h *AttachmentHandler) DownloadAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		utils.SendBadRequest(c, "活动ID或附件ID不能为空")
		return
	}

	// 获取当前用户信息
	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "附件不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	filePath := filepath.Join("uploads/attachments", attachment.FileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.SendNotFound(c, "文件不存在")
		return
	}

	h.db.Model(&attachment).Update("download_count", attachment.DownloadCount+1)

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.OriginalName))

	contentType := "application/octet-stream"
	if attachment.FileType != "" {
		switch strings.ToLower(attachment.FileType) {
		case ".pdf":
			contentType = "application/pdf"
		case ".doc", ".docx":
			contentType = "application/msword"
		case ".xls", ".xlsx":
			contentType = "application/vnd.ms-excel"
		case ".ppt", ".pptx":
			contentType = "application/vnd.ms-powerpoint"
		case ".txt":
			contentType = "text/plain"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".zip":
			contentType = "application/zip"
		case ".rar":
			contentType = "application/x-rar-compressed"
		}
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(attachment.FileSize, 10))

	c.File(filePath)
}

func (h *AttachmentHandler) PreviewAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		utils.SendBadRequest(c, "活动ID或附件ID不能为空")
		return
	}

	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "活动不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "附件不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	filePath := filepath.Join("uploads/attachments", attachment.FileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.SendNotFound(c, "文件不存在")
		return
	}

	// 检查文件类型是否支持预览
	previewableTypes := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".txt":  true,
	}

	if !previewableTypes[strings.ToLower(attachment.FileType)] {
		utils.SendBadRequest(c, "不支持预览此类型的文件")
		return
	}

	// 设置适当的Content-Type
	contentType := "application/octet-stream"
	switch strings.ToLower(attachment.FileType) {
	case ".pdf":
		contentType = "application/pdf"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".txt":
		contentType = "text/plain"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(attachment.FileSize, 10))
	c.Header("Cache-Control", "public, max-age=3600") // 缓存1小时

	c.File(filePath)
}

func (h *AttachmentHandler) UpdateAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		utils.SendBadRequest(c, "活动ID或附件ID不能为空")
		return
	}

	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "附件不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}

	var req models.AttachmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	attachment.Description = req.Description
	if err := h.db.Save(&attachment).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	response := models.AttachmentResponse{
		ID:            attachment.ID,
		ActivityID:    attachment.ActivityID,
		FileName:      attachment.FileName,
		OriginalName:  attachment.OriginalName,
		FileSize:      attachment.FileSize,
		FileType:      attachment.FileType,
		FileCategory:  attachment.FileCategory,
		Description:   attachment.Description,
		UploadedBy:    attachment.UploadedBy,
		UploadedAt:    attachment.UploadedAt,
		DownloadCount: attachment.DownloadCount,
		DownloadURL:   fmt.Sprintf("/api/activities/%s/attachments/%s/download", activityID, attachment.ID),
	}

	utils.SendSuccessResponse(c, response)
}

func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		utils.SendBadRequest(c, "活动ID或附件ID不能为空")
		return
	}

	_, exists := c.Get("id")
	if !exists {
		utils.SendUnauthorized(c)
		return
	}

	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFound(c, "附件不存在")
		} else {
			utils.SendInternalServerError(c, err)
		}
		return
	}
	var otherAttachmentsCount int64
	h.db.Model(&models.Attachment{}).
		Where("md5_hash = ? AND activity_id != ? AND deleted_at IS NULL", attachment.MD5Hash, activityID).
		Count(&otherAttachmentsCount)

	if err := h.db.Delete(&attachment).Error; err != nil {
		utils.SendInternalServerError(c, err)
		return
	}

	if otherAttachmentsCount == 0 {
		filePath := filepath.Join("uploads/attachments", attachment.FileName)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			// 记录错误但不影响响应
			fmt.Printf("删除物理文件失败: %v\n", err)
		} else {
			fmt.Printf("彻底删除物理文件: %s\n", filePath)
		}
	} else {
		fmt.Printf("文件被其他活动使用，保留物理文件: %s (其他活动数量: %d)\n", attachment.FileName, otherAttachmentsCount)
	}

	utils.SendSuccessResponse(c, gin.H{
		"attachment_id": attachmentID,
		"deleted_at":    time.Now(),
		"file_removed":  otherAttachmentsCount == 0,
	})
}

// validateFile 验证文件
func (h *AttachmentHandler) validateFile(header *multipart.FileHeader) error {
	// 检查文件大小
	if header.Size > models.MaxFileSize {
		return fmt.Errorf("文件大小不能超过%dMB", models.MaxFileSize/1024/1024)
	}

	// 检查文件类型
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	if !models.IsSupportedFileType(fileExt) {
		return fmt.Errorf("不支持的文件类型: %s", fileExt)
	}

	return nil
}
