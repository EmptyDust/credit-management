package utils

import (
	"fmt"
	"time"
)

// DateFormats 支持的日期格式
var DateFormats = []string{
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// ParseDate 解析日期字符串，支持多种格式
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("日期字符串不能为空")
	}

	for _, format := range DateFormats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("日期格式错误，支持格式：YYYY-MM-DD、YYYY-MM-DD HH:mm:ss、YYYY-MM-DDTHH:mm:ss")
}

// ParseDateOptional 解析可选日期字符串，空字符串返回零值
func ParseDateOptional(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return ParseDate(dateStr)
}

// ParseDateRange 解析日期范围
func ParseDateRange(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	startDate, err := ParseDateOptional(startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期格式错误: %v", err)
	}

	endDate, err := ParseDateOptional(endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("结束日期格式错误: %v", err)
	}

	// 验证日期范围
	if !startDate.IsZero() && !endDate.IsZero() && startDate.After(endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期不能晚于结束日期")
	}

	return startDate, endDate, nil
}

// FormatDate 格式化日期为标准格式
func FormatDate(date time.Time) string {
	return date.Format("2006-01-02")
}

// FormatDateTime 格式化日期时间为标准格式
func FormatDateTime(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}

// IsValidDateRange 验证日期范围是否有效
func IsValidDateRange(startDate, endDate time.Time) bool {
	if startDate.IsZero() || endDate.IsZero() {
		return true // 允许空日期
	}
	return !startDate.After(endDate)
} 