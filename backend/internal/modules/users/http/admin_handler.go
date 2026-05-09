package http

import (
	"errors"
	"net/http"
	"strconv"

	"credit-management/backend/internal/httpapi/response"
	authservice "credit-management/backend/internal/modules/auth/service"
	exportservice "credit-management/backend/internal/modules/exports/service"
	usersrepo "credit-management/backend/internal/modules/users/repo"
	userservice "credit-management/backend/internal/modules/users/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService  *userservice.AdminService
	exportService *exportservice.UserExportService
}

type userPayload struct {
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	RealName     string `json:"real_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	DepartmentID string `json:"department_id"`
	StudentID    string `json:"student_id"`
	TeacherID    string `json:"teacher_id"`
	Grade        string `json:"grade"`
	Title        string `json:"title"`
}

type statusPayload struct {
	Status string `json:"status"`
}

type resetPasswordPayload struct {
	NewPassword string `json:"new_password"`
}

func NewAdminHandler(adminService *userservice.AdminService, exportService *exportservice.UserExportService) *AdminHandler {
	return &AdminHandler{
		adminService:  adminService,
		exportService: exportService,
	}
}

func (h *AdminHandler) ListStudents(c *gin.Context) {
	result, err := h.adminService.ListStudents(c.Request.Context(), queryFilter(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "list students failed", nil)
		return
	}

	response.OK(c, gin.H{
		"items":       result.Items,
		"page":        result.Page,
		"page_size":   result.PageSize,
		"total":       result.Total,
		"total_pages": result.TotalPages,
	})
}

func (h *AdminHandler) StudentStats(c *gin.Context) {
	stats, err := h.adminService.StudentStats(c.Request.Context(), queryFilter(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "load student stats failed", nil)
		return
	}

	response.OK(c, stats)
}

func (h *AdminHandler) CreateStudent(c *gin.Context) {
	var payload userPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid student payload", nil)
		return
	}

	user, err := h.adminService.CreateStudent(c.Request.Context(), usersrepo.CreateUserInput{
		Username:     payload.Username,
		Password:     payload.Password,
		RealName:     payload.RealName,
		Email:        payload.Email,
		Phone:        payload.Phone,
		DepartmentID: payload.DepartmentID,
		StudentID:    payload.StudentID,
		Grade:        payload.Grade,
	})
	if err != nil {
		handleAdminError(c, err, "create student failed")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"user": user})
}

func (h *AdminHandler) GetStudent(c *gin.Context) {
	h.getUser(c, "student")
}

func (h *AdminHandler) UpdateStudent(c *gin.Context) {
	h.updateUser(c, "student")
}

func (h *AdminHandler) DeleteStudent(c *gin.Context) {
	if err := h.adminService.DeleteUser(c.Request.Context(), c.Param("id")); err != nil {
		handleAdminError(c, err, "delete student failed")
		return
	}

	response.OK(c, gin.H{})
}

func (h *AdminHandler) ExportStudents(c *gin.Context) {
	content, err := h.exportService.ExportStudents(c.Request.Context(), queryFilter(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "export students failed", nil)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="students.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", content)
}

func (h *AdminHandler) ListTeachers(c *gin.Context) {
	result, err := h.adminService.ListTeachers(c.Request.Context(), queryFilter(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "list teachers failed", nil)
		return
	}

	response.OK(c, gin.H{
		"items":       result.Items,
		"page":        result.Page,
		"page_size":   result.PageSize,
		"total":       result.Total,
		"total_pages": result.TotalPages,
	})
}

func (h *AdminHandler) TeacherStats(c *gin.Context) {
	stats, err := h.adminService.TeacherStats(c.Request.Context(), queryFilter(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "load teacher stats failed", nil)
		return
	}

	response.OK(c, stats)
}

func (h *AdminHandler) CreateTeacher(c *gin.Context) {
	var payload userPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid teacher payload", nil)
		return
	}

	user, err := h.adminService.CreateTeacher(c.Request.Context(), usersrepo.CreateUserInput{
		Username:     payload.Username,
		Password:     payload.Password,
		RealName:     payload.RealName,
		Email:        payload.Email,
		Phone:        payload.Phone,
		DepartmentID: payload.DepartmentID,
		TeacherID:    payload.TeacherID,
		Title:        payload.Title,
	})
	if err != nil {
		handleAdminError(c, err, "create teacher failed")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"user": user})
}

func (h *AdminHandler) GetTeacher(c *gin.Context) {
	h.getUser(c, "teacher")
}

func (h *AdminHandler) UpdateTeacher(c *gin.Context) {
	h.updateUser(c, "teacher")
}

func (h *AdminHandler) DeleteTeacher(c *gin.Context) {
	if err := h.adminService.DeleteUser(c.Request.Context(), c.Param("id")); err != nil {
		handleAdminError(c, err, "delete teacher failed")
		return
	}

	response.OK(c, gin.H{})
}

func (h *AdminHandler) ExportTeachers(c *gin.Context) {
	content, err := h.exportService.ExportTeachers(c.Request.Context(), queryFilter(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "export teachers failed", nil)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="teachers.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", content)
}

func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	var payload statusPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid status payload", nil)
		return
	}

	user, err := h.adminService.UpdateStatus(c.Request.Context(), c.Param("id"), payload.Status)
	if err != nil {
		handleAdminError(c, err, "update status failed")
		return
	}

	response.OK(c, gin.H{"user": user})
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	var payload resetPasswordPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid reset password payload", nil)
		return
	}

	if err := h.adminService.ResetPassword(c.Request.Context(), c.Param("id"), payload.NewPassword); err != nil {
		handleAdminError(c, err, "reset password failed")
		return
	}

	response.OK(c, gin.H{})
}

func (h *AdminHandler) getUser(c *gin.Context, userType string) {
	user, err := h.adminService.GetUser(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleAdminError(c, err, "get user failed")
		return
	}
	if user.UserType != userType {
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "user not found", nil)
		return
	}

	response.OK(c, gin.H{"user": user})
}

func (h *AdminHandler) updateUser(c *gin.Context, userType string) {
	var payload userPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid update payload", nil)
		return
	}

	user, err := h.adminService.GetUser(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleAdminError(c, err, "get user failed")
		return
	}
	if user.UserType != userType {
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "user not found", nil)
		return
	}

	updated, err := h.adminService.UpdateUser(c.Request.Context(), c.Param("id"), usersrepo.UpdateUserInput{
		Username:     payload.Username,
		RealName:     payload.RealName,
		Email:        payload.Email,
		Phone:        payload.Phone,
		DepartmentID: payload.DepartmentID,
		StudentID:    payload.StudentID,
		TeacherID:    payload.TeacherID,
		Grade:        payload.Grade,
		Title:        payload.Title,
	})
	if err != nil {
		handleAdminError(c, err, "update user failed")
		return
	}

	response.OK(c, gin.H{"user": updated})
}

func queryFilter(c *gin.Context) usersrepo.ListFilter {
	return usersrepo.ListFilter{
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
		Grade:    c.Query("grade"),
		Title:    c.Query("title"),
		Page:     queryInt(c, "page"),
		PageSize: queryInt(c, "page_size"),
	}
}

func queryInt(c *gin.Context, key string) int {
	value := c.Query(key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func handleAdminError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, authservice.ErrUserNotFound):
		response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "user not found", nil)
	case errors.Is(err, authservice.ErrConflict):
		response.Error(c, http.StatusConflict, "USER_CONFLICT", "user already exists", gin.H{"reason": err.Error()})
	default:
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", fallback, nil)
	}
}
