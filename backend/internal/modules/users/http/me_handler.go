package http

import (
	"errors"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	authservice "credit-management/backend/internal/modules/auth/service"
	usersdomain "credit-management/backend/internal/modules/users/domain"
	userservice "credit-management/backend/internal/modules/users/service"

	"github.com/gin-gonic/gin"
)

type MeHandler struct {
	service *userservice.Service
}

type updateProfileRequest struct {
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	RealName     string `json:"real_name"`
	Grade        string `json:"grade"`
	Title        string `json:"title"`
	DepartmentID string `json:"department_id"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func NewMeHandler(service *userservice.Service) *MeHandler {
	return &MeHandler{service: service}
}

func (h *MeHandler) GetProfile(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "user not found", nil)
		return
	}

	response.OK(c, profile)
}

func (h *MeHandler) UpdateProfile(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid profile payload", nil)
		return
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), claims.UserID, usersdomain.ProfilePatch{
		Email:        req.Email,
		Phone:        req.Phone,
		RealName:     req.RealName,
		Grade:        req.Grade,
		Title:        req.Title,
		DepartmentID: req.DepartmentID,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrConflict):
			response.Error(c, http.StatusConflict, "USER_CONFLICT", "user profile conflicts with existing data", gin.H{"reason": err.Error()})
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "profile update failed", nil)
		}
		return
	}

	response.OK(c, profile)
}

func (h *MeHandler) ChangePassword(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid password payload", nil)
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidCredentials):
			response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "old password is incorrect", nil)
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "password change failed", nil)
		}
		return
	}

	response.OK(c, gin.H{})
}

func (h *MeHandler) UploadAvatar(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "avatar file is required", nil)
		return
	}

	result, err := h.service.UploadAvatar(c.Request.Context(), claims.UserID, file.Filename)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "avatar upload failed", nil)
		return
	}

	response.OK(c, result)
}

func (h *MeHandler) DeleteAvatar(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	if err := h.service.DeleteAvatar(c.Request.Context(), claims.UserID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "avatar delete failed", nil)
		return
	}

	response.OK(c, gin.H{})
}
