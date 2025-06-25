package handlers

import (
	"credit-management/user-service/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// getCurrentUserRole 获取当前用户角色
func getCurrentUserRole(c *gin.Context) string {
	claims, exists := c.Get("claims")
	if !exists {
		return ""
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	userType, exists := claimsMap["user_type"]
	if !exists {
		return ""
	}
	return userType.(string)
}

// getCurrentUserID 获取当前用户ID
func getCurrentUserID(c *gin.Context) string {
	claims, exists := c.Get("claims")
	if !exists {
		return ""
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	userID, exists := claimsMap["user_id"]
	if !exists {
		return ""
	}
	return userID.(string)
}

// convertToStudentBasicResponse 转换为学生基本信息响应
func (h *UserHandler) convertToStudentBasicResponse(user models.User) models.StudentBasicResponse {
	return models.StudentBasicResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		RealName:     user.RealName,
		StudentID:    user.StudentID,
		College:      user.College,
		Major:        user.Major,
		Class:        user.Class,
		Grade:        user.Grade,
		Status:       user.Status,
		Avatar:       user.Avatar,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// convertToTeacherBasicResponse 转换为教师基本信息响应
func (h *UserHandler) convertToTeacherBasicResponse(user models.User) models.TeacherBasicResponse {
	return models.TeacherBasicResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		RealName:     user.RealName,
		Department:   user.Department,
		Title:        user.Title,
		Status:       user.Status,
		Avatar:       user.Avatar,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// convertToStudentDetailResponse 转换为学生详细信息响应
func (h *UserHandler) convertToStudentDetailResponse(user models.User) models.StudentDetailResponse {
	return models.StudentDetailResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		StudentID:    user.StudentID,
		College:      user.College,
		Major:        user.Major,
		Class:        user.Class,
		Grade:        user.Grade,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// convertToTeacherDetailResponse 转换为教师详细信息响应
func (h *UserHandler) convertToTeacherDetailResponse(user models.User) models.TeacherDetailResponse {
	return models.TeacherDetailResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		Department:   user.Department,
		Title:        user.Title,
		Specialty:    user.Specialty,
		Status:       user.Status,
		Avatar:       user.Avatar,
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// convertToRoleBasedResponse 根据角色转换响应
func (h *UserHandler) convertToRoleBasedResponse(user models.User, currentUserRole string, isOwnProfile bool) interface{} {
	if isOwnProfile {
		// 查看自己的信息，返回详细信息
		return h.convertToUserResponse(user)
	}

	// 查看他人信息，根据角色返回不同详细程度
	switch currentUserRole {
	case "admin":
		return h.convertToUserResponse(user) // 管理员看到所有信息
	case "teacher":
		if user.UserType == "student" {
			return h.convertToStudentDetailResponse(user) // 教师看到学生详细信息
		} else {
			return h.convertToTeacherBasicResponse(user) // 教师看到其他教师基本信息
		}
	case "student":
		if user.UserType == "student" {
			return h.convertToStudentBasicResponse(user) // 学生看到其他学生基本信息
		} else {
			return h.convertToTeacherBasicResponse(user) // 学生看到教师基本信息
		}
	default:
		// 默认返回基本信息
		if user.UserType == "student" {
			return h.convertToStudentBasicResponse(user)
		} else {
			return h.convertToTeacherBasicResponse(user)
		}
	}
}

// canViewUserDetails 检查是否可以查看用户详细信息
func canViewUserDetails(currentUserRole, targetUserType string) bool {
	switch currentUserRole {
	case "admin":
		return true // 管理员可以查看所有用户详细信息
	case "teacher":
		return targetUserType == "student" // 教师只能查看学生详细信息
	case "student":
		return false // 学生不能查看任何用户的详细信息
	default:
		return false
	}
}
