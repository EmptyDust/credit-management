package http

import (
	"errors"
	"net/http"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	authservice "credit-management/backend/internal/modules/auth/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *authservice.Service
}

type loginRequest struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type registerStudentRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	RealName     string `json:"real_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	StudentID    string `json:"student_id"`
	Grade        string `json:"grade"`
	DepartmentID string `json:"department_id"`
}

func NewHandler(service *authservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid login payload", nil)
		return
	}

	result, err := h.service.Login(c.Request.Context(), authservice.LoginInput{
		LoginID:  req.LoginID,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidCredentials):
			response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "invalid credentials", nil)
		case errors.Is(err, authservice.ErrAccountInactive):
			response.Error(c, http.StatusForbidden, "AUTH_ACCOUNT_INACTIVE", "account is inactive", nil)
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "login failed", nil)
		}
		return
	}

	response.OK(c, result)
}

func (h *Handler) RegisterStudent(c *gin.Context) {
	var req registerStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid register payload", nil)
		return
	}

	user, err := h.service.RegisterStudent(c.Request.Context(), authservice.RegisterStudentInput{
		Username:     req.Username,
		Password:     req.Password,
		RealName:     req.RealName,
		Email:        req.Email,
		Phone:        req.Phone,
		StudentID:    req.StudentID,
		Grade:        req.Grade,
		DepartmentID: req.DepartmentID,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrConflict):
			response.Error(c, http.StatusConflict, "USER_CONFLICT", "user already exists", gin.H{"reason": err.Error()})
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "register failed", nil)
		}
		return
	}

	response.OK(c, gin.H{
		"user": gin.H{
			"id":            user.ID,
			"username":      user.Username,
			"real_name":     user.RealName,
			"user_type":     user.UserType,
			"status":        user.Status,
			"student_id":    user.StudentID,
			"grade":         user.Grade,
			"department_id": user.DepartmentID,
		},
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid refresh payload", nil)
		return
	}

	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrRefreshInvalid):
			response.Error(c, http.StatusUnauthorized, "AUTH_REFRESH_INVALID", "refresh token is invalid", nil)
		case errors.Is(err, authservice.ErrAccountInactive):
			response.Error(c, http.StatusForbidden, "AUTH_ACCOUNT_INACTIVE", "account is inactive", nil)
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "refresh failed", nil)
		}
		return
	}

	response.OK(c, result)
}

func (h *Handler) Logout(c *gin.Context) {
	accessToken, _ := c.Get(middleware.AccessTokenKey)
	token, _ := accessToken.(string)
	if token == "" {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_REQUIRED", "access token is required", nil)
		return
	}

	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	response.OK(c, gin.H{})
}
