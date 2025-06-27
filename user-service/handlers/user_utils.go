package handlers

import (
	"credit-management/user-service/models"
	"credit-management/user-service/utils"
)

// convertToUserResponse 将User模型转换为UserResponse
func (h *UserHandler) convertToUserResponse(user models.User) models.UserResponse {
	return models.UserResponse{
		UserID:       user.UserID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        utils.DerefString(user.Phone),
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		Avatar:       utils.DerefString(user.Avatar),
		LastLoginAt:  user.LastLoginAt,
		RegisterTime: user.RegisterTime,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		StudentID:    user.StudentID,
		College:      user.College,
		Major:        user.Major,
		Class:        user.Class,
		Grade:        user.Grade,
		Department:   user.Department,
		Title:        user.Title,
	}
}
