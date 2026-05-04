package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"credit-management/backend/internal/httpapi/response"
	importservice "credit-management/backend/internal/modules/imports/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	users      *UserImportHandler
	activities *ActivityImportHandler
}

type ActivityImportHandler struct {
	service *importservice.ActivityImportService
}

func NewHandler(userService *importservice.UserImportService, activityService *importservice.ActivityImportService) *Handler {
	return &Handler{
		users:      NewUserImportHandler(userService),
		activities: NewActivityImportHandler(activityService),
	}
}

func NewActivityImportHandler(service *importservice.ActivityImportService) *ActivityImportHandler {
	return &ActivityImportHandler{service: service}
}

func (h *Handler) StudentTemplate(c *gin.Context) {
	h.users.StudentTemplate(c)
}

func (h *Handler) TeacherTemplate(c *gin.Context) {
	h.users.TeacherTemplate(c)
}

func (h *Handler) PrecheckStudents(c *gin.Context) {
	h.users.PrecheckStudents(c)
}

func (h *Handler) ConfirmStudents(c *gin.Context) {
	h.users.ConfirmStudents(c)
}

func (h *Handler) PrecheckTeachers(c *gin.Context) {
	h.users.PrecheckTeachers(c)
}

func (h *Handler) ConfirmTeachers(c *gin.Context) {
	h.users.ConfirmTeachers(c)
}

func (h *Handler) ActivityPrecheck(c *gin.Context) {
	h.activities.Precheck(c)
}

func (h *Handler) ActivityConfirm(c *gin.Context) {
	h.activities.Confirm(c)
}

func (h *ActivityImportHandler) Precheck(c *gin.Context) {
	actorID, ok := actorIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "access token is invalid", nil)
		return
	}

	file, fileHash, rows, err := readActivityRowsFromUpload(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
		return
	}

	result, err := h.service.Precheck(c.Request.Context(), actorID, file.Filename, fileHash, rows)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "activity import precheck failed", nil)
		return
	}

	response.OK(c, result)
}

func (h *ActivityImportHandler) Confirm(c *gin.Context) {
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
	if strings.TrimSpace(payload.JobID) == "" || strings.TrimSpace(payload.FileHash) == "" {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid confirm payload", nil)
		return
	}

	result, err := h.service.Confirm(c.Request.Context(), actorID, payload.JobID, payload.FileHash)
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
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "confirm activity import failed", nil)
		}
		return
	}

	response.OK(c, result)
}

func readActivityRowsFromUpload(c *gin.Context) (*multipart.FileHeader, string, []importservice.ActivityImportRow, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, "", nil, errors.New("file is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", nil, errors.New("failed to open upload file")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, "", nil, errors.New("failed to read upload file")
	}

	rows, err := parseActivityCSV(bytes.NewReader(content))
	if err != nil {
		return nil, "", nil, err
	}

	sum := sha256.Sum256(content)
	return fileHeader, hex.EncodeToString(sum[:]), rows, nil
}

func parseActivityCSV(reader io.Reader) ([]importservice.ActivityImportRow, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.New("failed to parse csv file")
	}
	if len(records) <= 1 {
		return nil, errors.New("csv file must contain at least one data row")
	}

	headers := records[0]
	rows := make([]importservice.ActivityImportRow, 0, len(records)-1)
	for _, record := range records[1:] {
		if len(record) < len(headers) {
			padded := append([]string{}, record...)
			for len(padded) < len(headers) {
				padded = append(padded, "")
			}
			record = padded
		}

		row := importservice.ActivityImportRow{
			Details: map[string]any{},
		}
		for idx, header := range headers {
			value := strings.TrimSpace(record[idx])
			switch strings.TrimSpace(header) {
			case "title":
				row.Title = value
			case "description":
				row.Description = value
			case "category":
				row.Category = value
			case "academic_year":
				row.AcademicYear = value
			case "semester":
				row.Semester = value
			case "start_date":
				row.StartDate = value
			case "end_date":
				row.EndDate = value
			default:
				if strings.HasPrefix(strings.TrimSpace(header), "details.") {
					key := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(header), "details."))
					if key != "" {
						row.Details[key] = value
					}
				}
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}
