package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validator 验证器结构
type Validator struct{}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{}
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

	// 简单的UUID格式验证
	if len(id) != 36 || !strings.Contains(id, "-") {
		return fmt.Errorf("无效的UUID格式")
	}

	return nil
}

// ValidateEmail 验证邮箱格式
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("邮箱格式不正确")
	}

	return nil
}

// ValidatePhone 验证手机号格式
func (v *Validator) ValidatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("手机号不能为空")
	}

	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("手机号格式不正确")
	}

	return nil
}

// ValidateUsername 验证用户名格式
func (v *Validator) ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	if len(username) < 3 || len(username) > 20 {
		return fmt.Errorf("用户名长度必须在3-20个字符之间")
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("用户名只能包含字母、数字和下划线")
	}

	return nil
}

// ValidatePassword 验证密码复杂度
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	if len(password) < 8 {
		return fmt.Errorf("密码长度至少8位")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("密码必须包含大小写字母和数字")
	}

	return nil
}

// ValidateStudentID 验证学号格式
func (v *Validator) ValidateStudentID(studentID string) error {
	if studentID == "" {
		return fmt.Errorf("学号不能为空")
	}

	// 学号格式：年份(4位) + 序号(4位)
	studentIDRegex := regexp.MustCompile(`^\d{8}$`)
	if !studentIDRegex.MatchString(studentID) {
		return fmt.Errorf("学号格式不正确，应为8位数字")
	}

	year := studentID[:4]
	yearNum, _ := strconv.Atoi(year)
	currentYear := 2024 // 可以根据实际情况调整

	if yearNum < 2000 || yearNum > currentYear+10 {
		return fmt.Errorf("学号年份不合理")
	}

	return nil
}

// ValidateGrade 验证年级格式
func (v *Validator) ValidateGrade(grade string) error {
	if grade == "" {
		return fmt.Errorf("年级不能为空")
	}

	gradeRegex := regexp.MustCompile(`^\d{4}$`)
	if !gradeRegex.MatchString(grade) {
		return fmt.Errorf("年级格式不正确，应为4位数字")
	}

	year, _ := strconv.Atoi(grade)
	currentYear := 2024 // 可以根据实际情况调整

	if year < 2000 || year > currentYear+10 {
		return fmt.Errorf("年级年份不合理")
	}

	return nil
}

// ValidateUserType 验证用户类型
func (v *Validator) ValidateUserType(userType string) error {
	validTypes := []string{"student", "teacher", "admin"}
	for _, validType := range validTypes {
		if userType == validType {
			return nil
		}
	}
	return fmt.Errorf("无效的用户类型: %s", userType)
}

// ValidateStatus 验证用户状态
func (v *Validator) ValidateStatus(status string) error {
	validStatuses := []string{"active", "inactive", "suspended", "graduated"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return fmt.Errorf("无效的用户状态: %s", status)
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
