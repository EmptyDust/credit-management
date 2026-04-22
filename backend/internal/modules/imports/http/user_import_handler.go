package http

import (
	"encoding/csv"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"credit-management/backend/internal/httpapi/middleware"
	"credit-management/backend/internal/httpapi/response"
	authservice "credit-management/backend/internal/modules/auth/service"
	importservice "credit-management/backend/internal/modules/imports/service"

	"github.com/gin-gonic/gin"
)

type UserImportHandler struct {
	service *importservice.UserImportService
}

type confirmRequest struct {
	JobID    string `json:"job_id"`
	FileHash string `json:"file_hash"`
}

func NewUserImportHandler(service *importservice.UserImportService) *UserImportHandler {
	return &UserImportHandler{service: service}
}

func (h *UserImportHandler) StudentTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="students-template.csv"`)
	c.String(http.StatusOK, "username,password,real_name,email,phone,student_id,grade,department_id\n")
}

func (h *UserImportHandler) TeacherTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="teachers-template.csv"`)
	c.String(http.StatusOK, "username,password,real_name,email,phone,teacher_id,title,department_id\n")
}

func (h *UserImportHandler) PrecheckStudents(c *gin.Context) {
	actorID, ok := actorIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	file, rows, err := readRowsFromUpload(c, "student")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
		return
	}

	result, err := h.service.PrecheckStudents(c.Request.Context(), actorID, file.Filename, file.Filename, rows)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "student import precheck failed", nil)
		return
	}

	response.OK(c, result)
}

func (h *UserImportHandler) ConfirmStudents(c *gin.Context) {
	h.confirm(c, "student")
}

func (h *UserImportHandler) PrecheckTeachers(c *gin.Context) {
	actorID, ok := actorIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	file, rows, err := readRowsFromUpload(c, "teacher")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
		return
	}

	result, err := h.service.PrecheckTeachers(c.Request.Context(), actorID, file.Filename, file.Filename, rows)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "teacher import precheck failed", nil)
		return
	}

	response.OK(c, result)
}

func (h *UserImportHandler) ConfirmTeachers(c *gin.Context) {
	h.confirm(c, "teacher")
}

func (h *UserImportHandler) confirm(c *gin.Context, userType string) {
	actorID, ok := actorIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	var payload confirmRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid confirm payload", nil)
		return
	}

	var (
		result importservice.ConfirmResult
		err    error
	)
	if userType == "student" {
		result, err = h.service.ConfirmStudents(c.Request.Context(), actorID, payload.JobID, payload.FileHash)
	} else {
		result, err = h.service.ConfirmTeachers(c.Request.Context(), actorID, payload.JobID, payload.FileHash)
	}
	if err != nil {
		switch {
		case errors.Is(err, importservice.ErrImportJobNotFound):
			response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "import job not found", nil)
		case errors.Is(err, importservice.ErrImportJobExpired):
			response.Error(c, http.StatusConflict, "IMPORT_JOB_EXPIRED", "import job expired", nil)
		case errors.Is(err, importservice.ErrImportFileHashMisfit):
			response.Error(c, http.StatusConflict, "IMPORT_FILE_HASH_MISMATCH", "file hash mismatch", nil)
		case errors.Is(err, importservice.ErrImportHasInvalidRows):
			response.Error(c, http.StatusConflict, "IMPORT_CONFIRM_FORBIDDEN", "import job contains invalid rows", nil)
		case errors.Is(err, authservice.ErrConflict):
			response.Error(c, http.StatusConflict, "USER_CONFLICT", "import conflicts with existing data", nil)
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "confirm import failed", nil)
		}
		return
	}

	response.OK(c, result)
}

func actorIDFromContext(c *gin.Context) (string, bool) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return "", false
	}

	return claims.UserID, true
}

func readRowsFromUpload(c *gin.Context, userType string) (*multipart.FileHeader, []importservice.UserRow, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, nil, errors.New("file is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, errors.New("failed to open upload file")
	}
	defer file.Close()

	rows, err := parseCSV(file, userType)
	if err != nil {
		return nil, nil, err
	}

	return nil, rows, nil
}

func parseCSV(reader io.Reader, userType string) ([]importservice.UserRow, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.New("failed to parse csv file")
	}
	if len(records) <= 1 {
		return nil, errors.New("csv file must contain at least one data row")
	}

	rows := make([]importservice.UserRow, 0, len(records)-1)
	for _, record := range records[1:] {
		for len(record) < 8 {
			record = append(record, "")
		}

		row := importservice.UserRow{
			Username:     strings.TrimSpace(record[0]),
			Password:     strings.TrimSpace(record[1]),
			RealName:     strings.TrimSpace(record[2]),
			Email:        strings.TrimSpace(record[3]),
			Phone:        strings.TrimSpace(record[4]),
			DepartmentID: strings.TrimSpace(record[7]),
		}
		if userType == "student" {
			row.StudentID = strings.TrimSpace(record[5])
			row.Grade = strings.TrimSpace(record[6])
		} else {
			row.TeacherID = strings.TrimSpace(record[5])
			row.Title = strings.TrimSpace(record[6])
		}

		rows = append(rows, row)
	}

	return rows, nil
}
