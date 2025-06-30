package utils

import (
	"os"
	"strings"
	"time"
)

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDatabaseURL 获取数据库连接URL
func GetDatabaseURL() string {
	return GetEnv("DATABASE_URL", "host=localhost user=postgres password=password dbname=credit_management port=5432 sslmode=disable")
}

// GetServerPort 获取服务器端口
func GetServerPort() string {
	return GetEnv("SERVER_PORT", "8084")
}

// GetJWTSecret 获取JWT密钥
func GetJWTSecret() string {
	return GetEnv("JWT_SECRET", "your-secret-key")
}

// GetJWTExpiration 获取JWT过期时间（小时）
func GetJWTExpiration() int {
	expStr := GetEnv("JWT_EXPIRATION", "24")
	if exp, err := time.ParseDuration(expStr + "h"); err == nil {
		return int(exp.Hours())
	}
	return 24
}

// GetMaxFileSize 获取最大文件大小（字节）
func GetMaxFileSize() int64 {
	sizeStr := GetEnv("MAX_FILE_SIZE", "10485760") // 10MB
	if size, err := time.ParseDuration(sizeStr); err == nil {
		return int64(size)
	}
	return 10 * 1024 * 1024 // 10MB
}

// GetAllowedFileTypes 获取允许的文件类型
func GetAllowedFileTypes() []string {
	typesStr := GetEnv("ALLOWED_FILE_TYPES", "csv,xlsx,xls")
	return strings.Split(typesStr, ",")
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}

// IsEmptyString 检查字符串是否为空
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// TruncateString 截断字符串
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// GenerateDefaultPassword 生成默认密码
func GenerateDefaultPassword() string {
	return "Password123"
}

// ValidatePasswordComplexity 验证密码复杂度（兼容现有代码）
func ValidatePasswordComplexity(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(password, "0123456789")

	return hasUpper && hasLower && hasDigit
}

// ValidatePhoneFormat 验证手机号格式（兼容现有代码）
func ValidatePhoneFormat(phone string) bool {
	if len(phone) != 11 {
		return false
	}

	if !strings.HasPrefix(phone, "1") {
		return false
	}

	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// ValidateStudentIDFormat 验证学号格式（兼容现有代码）
func ValidateStudentIDFormat(studentID string) bool {
	if len(studentID) != 8 {
		return false
	}

	for _, char := range studentID {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// ValidateGradeFormat 验证年级格式（兼容现有代码）
func ValidateGradeFormat(grade string) bool {
	if len(grade) != 4 {
		return false
	}

	for _, char := range grade {
		if char < '0' || char > '9' {
			return false
		}
	}

	year, _ := time.Parse("2006", grade)
	currentYear := time.Now().Year()

	return year.Year() >= 2000 && year.Year() <= currentYear+10
}

// DerefString returns the value of a *string or "" if nil
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
