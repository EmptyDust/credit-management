package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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

// GetAttachments 获取活动附件列表
func (h *AttachmentHandler) GetAttachments(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 检查活动是否存在
	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：学生只能查看自己创建或参与的活动
	if userType == "student" {
		if activity.OwnerID != userID {
			// 检查是否为参与者
			var participant models.ActivityParticipant
			if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&participant).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "无权限查看此活动的附件",
					"data":    nil,
				})
				return
			}
		}
	}
	// 教师和管理员可以查看所有活动的附件，无需额外权限检查

	// 获取查询参数
	category := c.Query("category")
	fileType := c.Query("file_type")
	uploadedBy := c.Query("uploaded_by")

	// 构建查询
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

	// 获取附件列表
	var attachments []models.Attachment
	if err := query.Order("uploaded_at DESC").Find(&attachments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取附件列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应
	var responses []models.AttachmentResponse
	var totalSize int64
	categoryCount := make(map[string]int64)
	fileTypeCount := make(map[string]int64)

	// 获取认证令牌用于调用用户服务
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

		// 获取上传者信息
		if userInfo, err := utils.GetUserInfo(attachment.UploadedBy, authToken); err == nil {
			response.Uploader = *userInfo
		}

		responses = append(responses, response)

		// 统计信息
		totalSize += attachment.FileSize
		categoryCount[attachment.FileCategory]++
		fileTypeCount[attachment.FileType]++
	}

	// 构建统计信息
	stats := models.AttachmentStats{
		TotalCount:    int64(len(attachments)),
		TotalSize:     totalSize,
		CategoryCount: categoryCount,
		FileTypeCount: fileTypeCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "获取附件列表成功",
		"data": gin.H{
			"attachments": responses,
			"stats":       stats,
		},
	})
}

// UploadAttachment 上传单个附件
func (h *AttachmentHandler) UploadAttachment(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 检查活动是否存在
	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：只有活动创建者和管理员可以上传附件
	if userType != "admin" && activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限上传附件到此活动",
			"data":    nil,
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未找到上传的文件",
			"data":    nil,
		})
		return
	}
	defer file.Close()

	// 获取文件描述
	description := c.PostForm("description")

	// 验证文件
	if err := h.validateFile(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 读取文件内容并计算MD5
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "读取文件失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	md5Hash := fmt.Sprintf("%x", md5.Sum(fileBytes))

	// 生成存储文件名
	fileExt := filepath.Ext(header.Filename)
	// fileName := fmt.Sprintf("%s_%d%s", md5Hash[:8], time.Now().Unix(), fileExt)

	// 创建存储目录
	uploadDir := "uploads/attachments"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建存储目录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 保存文件（如果物理文件已存在则跳过写入）
	// filePath := filepath.Join(uploadDir, fileName)
	md5FilePath := filepath.Join(uploadDir, md5Hash+fileExt)
	if _, err := os.Stat(md5FilePath); os.IsNotExist(err) {
		if err := os.WriteFile(md5FilePath, fileBytes, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "保存文件失败: " + err.Error(),
				"data":    nil,
			})
			return
		}
	}

	// 创建附件记录（每次都插入新记录，哪怕md5_hash一样）
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建附件记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应
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

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "附件上传成功",
		"data":    response,
	})
}

// BatchUploadAttachments 批量上传附件
func (h *AttachmentHandler) BatchUploadAttachments(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 检查活动是否存在
	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：只有活动创建者和管理员可以上传附件
	if userType != "admin" && activity.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限上传附件到此活动",
			"data":    nil,
		})
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取上传文件失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "未找到上传的文件",
			"data":    nil,
		})
		return
	}

	// 检查文件数量限制
	if len(files) > models.MaxBatchUploadCount {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("批量上传文件数量不能超过%d个", models.MaxBatchUploadCount),
			"data":    nil,
		})
		return
	}

	// 创建存储目录
	uploadDir := "uploads/attachments"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建存储目录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	var results []models.BatchUploadResult
	successCount := 0
	failCount := 0

	// 开始事务
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量上传失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	response := models.BatchUploadResponse{
		TotalFiles:   len(files),
		SuccessCount: successCount,
		FailCount:    failCount,
		Results:      results,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "批量上传完成",
		"data":    response,
	})
}

// DownloadAttachment 下载附件
func (h *AttachmentHandler) DownloadAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID或附件ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 检查活动是否存在
	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：学生只能下载自己创建或参与的活动的附件
	if userType == "student" {
		if activity.OwnerID != userID {
			// 检查是否为参与者
			var participant models.ActivityParticipant
			if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&participant).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "无权限下载此活动的附件",
					"data":    nil,
				})
				return
			}
		}
	}
	// 教师和管理员可以下载所有活动的附件，无需额外权限检查

	// 获取附件信息
	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "附件不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取附件失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 检查文件是否存在
	filePath := filepath.Join("uploads/attachments", attachment.FileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
			"data":    nil,
		})
		return
	}

	// 更新下载次数
	h.db.Model(&attachment).Update("download_count", attachment.DownloadCount+1)

	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.OriginalName))

	// 根据文件类型设置正确的Content-Type
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

	// 发送文件
	c.File(filePath)
}

// PreviewAttachment 预览附件
func (h *AttachmentHandler) PreviewAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID或附件ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 检查活动是否存在
	var activity models.CreditActivity
	if err := h.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "活动不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取活动失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：学生只能预览自己创建或参与的活动的附件
	if userType == "student" {
		if activity.OwnerID != userID {
			// 检查是否为参与者
			var participant models.ActivityParticipant
			if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&participant).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "无权限预览此活动的附件",
					"data":    nil,
				})
				return
			}
		}
	}
	// 教师和管理员可以预览所有活动的附件，无需额外权限检查

	// 获取附件信息
	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "附件不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取附件失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 检查文件是否存在
	filePath := filepath.Join("uploads/attachments", attachment.FileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
			"data":    nil,
		})
		return
	}

	// 根据文件类型设置正确的Content-Type用于预览
	contentType := "application/octet-stream"
	if attachment.FileType != "" {
		switch strings.ToLower(attachment.FileType) {
		case ".pdf":
			contentType = "application/pdf"
		case ".txt":
			contentType = "text/plain; charset=utf-8"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".bmp":
			contentType = "image/bmp"
		case ".webp":
			contentType = "image/webp"
		case ".mp4":
			contentType = "video/mp4"
		case ".avi":
			contentType = "video/x-msvideo"
		case ".mov":
			contentType = "video/quicktime"
		case ".wmv":
			contentType = "video/x-ms-wmv"
		case ".flv":
			contentType = "video/x-flv"
		case ".mp3":
			contentType = "audio/mpeg"
		case ".wav":
			contentType = "audio/wav"
		case ".ogg":
			contentType = "audio/ogg"
		case ".aac":
			contentType = "audio/aac"
		}
	}

	// 设置预览响应头（不设置Content-Disposition为attachment，允许浏览器直接显示）
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(attachment.FileSize, 10))
	c.Header("Cache-Control", "public, max-age=3600") // 缓存1小时

	// 发送文件
	c.File(filePath)
}

// UpdateAttachment 更新附件信息
func (h *AttachmentHandler) UpdateAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID或附件ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 获取附件信息
	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "附件不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取附件失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：只有上传者和管理员可以更新附件信息
	if userType != "admin" && attachment.UploadedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限更新此附件",
			"data":    nil,
		})
		return
	}

	var req models.AttachmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 更新附件信息
	attachment.Description = req.Description
	if err := h.db.Save(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新附件失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 构建响应
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "附件信息更新成功",
		"data":    response,
	})
}

// DeleteAttachment 删除附件
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	activityID := c.Param("id")
	attachmentID := c.Param("attachment_id")

	if activityID == "" || attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "活动ID或附件ID不能为空",
			"data":    nil,
		})
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
			"data":    nil,
		})
		return
	}

	userType, _ := c.Get("user_type")

	// 获取附件信息
	var attachment models.Attachment
	if err := h.db.Where("id = ? AND activity_id = ? AND deleted_at IS NULL", attachmentID, activityID).First(&attachment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "附件不存在",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取附件失败: " + err.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 权限检查：只有上传者和管理员可以删除附件
	if userType != "admin" && attachment.UploadedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限删除此附件",
			"data":    nil,
		})
		return
	}

	// 检查是否有其他活动使用相同的文件
	var otherAttachmentsCount int64
	h.db.Model(&models.Attachment{}).
		Where("md5_hash = ? AND activity_id != ? AND deleted_at IS NULL", attachment.MD5Hash, activityID).
		Count(&otherAttachmentsCount)

	// 软删除附件记录
	if err := h.db.Delete(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除附件失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 如果没有其他活动使用该文件，则删除物理文件
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "附件删除成功",
		"data": gin.H{
			"attachment_id": attachmentID,
			"deleted_at":    time.Now(),
			"file_removed":  otherAttachmentsCount == 0,
		},
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
