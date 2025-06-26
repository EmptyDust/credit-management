package handlers

import (
	"fmt"
	"regexp"
	"strings"
)

// validatePassword 验证密码强度
func (h *UserHandler) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度至少8位")
	}

	// 检查是否包含至少一个大写字母
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("密码必须包含至少一个大写字母")
	}

	// 检查是否包含至少一个小写字母
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return fmt.Errorf("密码必须包含至少一个小写字母")
	}

	// 检查是否包含至少一个数字
	if !strings.ContainsAny(password, "0123456789") {
		return fmt.Errorf("密码必须包含至少一个数字")
	}

	return nil
}

// validatePhone 验证手机号格式
func (h *UserHandler) validatePhone(phone string) error {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("手机号格式不正确，请输入11位有效的手机号")
	}
	return nil
}

// validateStudentID 验证学号格式
func (h *UserHandler) validateStudentID(studentID string) error {
	// 学号应该是8位数字
	studentIDRegex := regexp.MustCompile(`^\d{8}$`)
	if !studentIDRegex.MatchString(studentID) {
		return fmt.Errorf("学号格式不正确，请输入8位数字学号")
	}
	return nil
}
