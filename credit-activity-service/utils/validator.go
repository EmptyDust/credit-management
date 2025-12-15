package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"credit-management/credit-activity-service/models"
)

// Validator 验证器结构
type Validator struct{}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateActivityRequest 验证活动创建请求
func (v *Validator) ValidateActivityRequest(req models.ActivityRequest) error {
	if req.Title == "" {
		return fmt.Errorf("活动标题不能为空")
	}

	if len(req.Title) > 200 {
		return fmt.Errorf("活动标题长度不能超过200个字符")
	}

	if req.Category != "" {
		if err := v.ValidateCategory(req.Category); err != nil {
			return err
		}
	}

	// 验证日期范围
	if req.StartDate != "" || req.EndDate != "" {
		if _, _, err := ParseDateRange(req.StartDate, req.EndDate); err != nil {
			return err
		}
	}

	return nil
}

// ValidateActivityUpdateRequest 验证活动更新请求
func (v *Validator) ValidateActivityUpdateRequest(req models.ActivityUpdateRequest) error {
	if req.Title != nil && len(*req.Title) > 200 {
		return fmt.Errorf("活动标题长度不能超过200个字符")
	}

	if req.Category != nil {
		if err := v.ValidateCategory(*req.Category); err != nil {
			return err
		}
	}

	// 验证日期范围
	if req.StartDate != nil || req.EndDate != nil {
		startDateStr := ""
		endDateStr := ""
		if req.StartDate != nil {
			startDateStr = *req.StartDate
		}
		if req.EndDate != nil {
			endDateStr = *req.EndDate
		}
		if _, _, err := ParseDateRange(startDateStr, endDateStr); err != nil {
			return err
		}
	}

	return nil
}

// ValidateCategory 验证活动类别
func (v *Validator) ValidateCategory(category string) error {
	validCategories := models.GetActivityCategories()
	for _, validCategory := range validCategories {
		if category == validCategory {
			return nil
		}
	}
	return fmt.Errorf("无效的活动类别: %s", category)
}

// ValidatePagination 验证分页参数
func (v *Validator) ValidatePagination(pageStr, pageSizeStr string) (int, int, error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(pageSizeStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	return page, limit, nil
}

// ValidateUUID 验证UUID格式
func (v *Validator) ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("ID不能为空")
	}

	// 使用正则表达式验证UUID格式（标准UUID v4格式：8-4-4-4-12）
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(id) {
		return fmt.Errorf("无效的UUID格式")
	}

	return nil
}

// ValidateCredits 验证学分值
func (v *Validator) ValidateCredits(credits float64) error {
	if credits < 0 {
		return fmt.Errorf("学分不能为负数")
	}
	if credits > 100 {
		return fmt.Errorf("学分不能超过100")
	}
	return nil
}

// ValidateFileSize 验证文件大小
func (v *Validator) ValidateFileSize(size int64, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("文件大小超过限制，最大允许 %d 字节", maxSize)
	}
	return nil
}

// ValidateFileType 验证文件类型
func (v *Validator) ValidateFileType(filename string, allowedTypes []string) error {
	if filename == "" {
		return fmt.Errorf("文件名不能为空")
	}

	ext := strings.ToLower(getFileExtension(filename))
	for _, allowedType := range allowedTypes {
		if ext == strings.ToLower(allowedType) {
			return nil
		}
	}

	return fmt.Errorf("不支持的文件类型: %s，支持的类型: %v", ext, allowedTypes)
}

// getFileExtension 获取文件扩展名
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
