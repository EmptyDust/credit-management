package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"credit-management/backend/internal/modules/users/repo"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/idgen"
)

var (
	ErrImportJobNotFound    = errors.New("import job not found")
	ErrImportJobExpired     = errors.New("import job expired")
	ErrImportFileHashMisfit = errors.New("import file hash mismatch")
	ErrImportHasInvalidRows = errors.New("import job has invalid rows")
)

type UserRow struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	RealName     string `json:"real_name"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	StudentID    string `json:"student_id,omitempty"`
	TeacherID    string `json:"teacher_id,omitempty"`
	Grade        string `json:"grade,omitempty"`
	Title        string `json:"title,omitempty"`
	DepartmentID string `json:"department_id,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RowResult struct {
	RowNo   int               `json:"row_no"`
	Status  string            `json:"status"`
	Payload UserRow           `json:"payload"`
	Errors  []ValidationError `json:"errors"`
}

type Summary struct {
	TotalRows   int `json:"total_rows"`
	ValidRows   int `json:"valid_rows"`
	InvalidRows int `json:"invalid_rows"`
}

type PrecheckResult struct {
	JobID     string      `json:"job_id"`
	JobType   string      `json:"job_type"`
	FileName  string      `json:"file_name"`
	FileHash  string      `json:"file_hash"`
	ExpiresAt time.Time   `json:"expires_at"`
	Summary   Summary     `json:"summary"`
	Rows      []RowResult `json:"rows"`
}

type ConfirmResult struct {
	JobID          string   `json:"job_id"`
	Stage          string   `json:"stage"`
	Summary        Summary  `json:"summary"`
	CreatedUserIDs []string `json:"created_user_ids"`
}

type job struct {
	ActorID       string
	JobType       string
	FileHash      string
	ExpiresAt     time.Time
	Rows          []UserRow
	Precheck      PrecheckResult
	ConfirmResult ConfirmResult
}

type UserImportService struct {
	users repo.UserRepository
	mu    sync.Mutex
	jobs  map[string]*job
}

func NewUserImportService(users repo.UserRepository) *UserImportService {
	return &UserImportService{
		users: users,
		jobs:  make(map[string]*job),
	}
}

func (s *UserImportService) PrecheckStudents(ctx context.Context, actorID string, fileName string, fileHash string, rows []UserRow) (PrecheckResult, error) {
	return s.precheck(ctx, actorID, "students", authz.RoleStudent, fileName, fileHash, rows)
}

func (s *UserImportService) PrecheckTeachers(ctx context.Context, actorID string, fileName string, fileHash string, rows []UserRow) (PrecheckResult, error) {
	return s.precheck(ctx, actorID, "teachers", authz.RoleTeacher, fileName, fileHash, rows)
}

func (s *UserImportService) ConfirmStudents(ctx context.Context, actorID string, jobID string, fileHash string) (ConfirmResult, error) {
	return s.confirm(ctx, actorID, jobID, fileHash, authz.RoleStudent)
}

func (s *UserImportService) ConfirmTeachers(ctx context.Context, actorID string, jobID string, fileHash string) (ConfirmResult, error) {
	return s.confirm(ctx, actorID, jobID, fileHash, authz.RoleTeacher)
}

func (s *UserImportService) precheck(ctx context.Context, actorID string, jobType string, userType string, fileName string, fileHash string, rows []UserRow) (PrecheckResult, error) {
	existing, err := s.users.ListUsers(ctx, repo.ListFilter{UserType: userType})
	if err != nil {
		return PrecheckResult{}, err
	}

	usernameSeen := make(map[string]struct{})
	idSeen := make(map[string]struct{})
	existingUsername := make(map[string]struct{})
	existingStudentID := make(map[string]struct{})
	existingTeacherID := make(map[string]struct{})
	for _, user := range existing {
		existingUsername[strings.ToLower(user.Username)] = struct{}{}
		if user.StudentID != "" {
			existingStudentID[user.StudentID] = struct{}{}
		}
		if user.TeacherID != "" {
			existingTeacherID[user.TeacherID] = struct{}{}
		}
	}

	rowResults := make([]RowResult, 0, len(rows))
	summary := Summary{TotalRows: len(rows)}

	for idx, row := range rows {
		errors := s.validateRow(row, userType, usernameSeen, idSeen, existingUsername, existingStudentID, existingTeacherID)
		status := "valid"
		if len(errors) > 0 {
			status = "invalid"
			summary.InvalidRows++
		} else {
			summary.ValidRows++
		}

		rowResults = append(rowResults, RowResult{
			RowNo:   idx + 1,
			Status:  status,
			Payload: row,
			Errors:  errors,
		})
	}

	jobID := idgen.New()
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	result := PrecheckResult{
		JobID:     jobID,
		JobType:   jobType,
		FileName:  fileName,
		FileHash:  fileHash,
		ExpiresAt: expiresAt,
		Summary:   summary,
		Rows:      rowResults,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[jobID] = &job{
		ActorID:   actorID,
		JobType:   jobType,
		FileHash:  fileHash,
		ExpiresAt: expiresAt,
		Rows:      rows,
		Precheck:  result,
	}

	return result, nil
}

func (s *UserImportService) confirm(ctx context.Context, actorID string, jobID string, fileHash string, userType string) (ConfirmResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return ConfirmResult{}, ErrImportJobNotFound
	}
	if time.Now().UTC().After(job.ExpiresAt) {
		return ConfirmResult{}, ErrImportJobExpired
	}
	if job.ActorID != actorID || job.FileHash != fileHash {
		return ConfirmResult{}, ErrImportFileHashMisfit
	}
	if job.ConfirmResult.JobID != "" {
		return job.ConfirmResult, nil
	}
	if job.Precheck.Summary.InvalidRows > 0 {
		return ConfirmResult{}, ErrImportHasInvalidRows
	}

	createdIDs := make([]string, 0, len(job.Rows))
	for _, row := range job.Rows {
		user, err := s.users.CreateUser(ctx, repo.CreateUserInput{
			Username:     row.Username,
			Password:     row.Password,
			RealName:     row.RealName,
			Email:        row.Email,
			Phone:        row.Phone,
			UserType:     userType,
			DepartmentID: row.DepartmentID,
			StudentID:    row.StudentID,
			TeacherID:    row.TeacherID,
			Grade:        row.Grade,
			Title:        row.Title,
		})
		if err != nil {
			return ConfirmResult{}, err
		}
		createdIDs = append(createdIDs, user.ID)
	}

	result := ConfirmResult{
		JobID:          jobID,
		Stage:          "completed",
		Summary:        job.Precheck.Summary,
		CreatedUserIDs: createdIDs,
	}
	job.ConfirmResult = result
	return result, nil
}

func (s *UserImportService) validateRow(
	row UserRow,
	userType string,
	usernameSeen map[string]struct{},
	idSeen map[string]struct{},
	existingUsername map[string]struct{},
	existingStudentID map[string]struct{},
	existingTeacherID map[string]struct{},
) []ValidationError {
	var errs []ValidationError

	username := strings.TrimSpace(row.Username)
	if username == "" {
		errs = append(errs, ValidationError{Field: "username", Code: "REQUIRED", Message: "username is required"})
	} else {
		key := strings.ToLower(username)
		if _, ok := existingUsername[key]; ok {
			errs = append(errs, ValidationError{Field: "username", Code: "DUPLICATE_USERNAME", Message: "username already exists"})
		}
		if _, ok := usernameSeen[key]; ok {
			errs = append(errs, ValidationError{Field: "username", Code: "DUPLICATE_USERNAME", Message: "username is duplicated in file"})
		}
		usernameSeen[key] = struct{}{}
	}

	if strings.TrimSpace(row.Password) == "" {
		errs = append(errs, ValidationError{Field: "password", Code: "REQUIRED", Message: "password is required"})
	}
	if strings.TrimSpace(row.RealName) == "" {
		errs = append(errs, ValidationError{Field: "real_name", Code: "REQUIRED", Message: "real_name is required"})
	}

	switch userType {
	case authz.RoleStudent:
		if strings.TrimSpace(row.StudentID) == "" {
			errs = append(errs, ValidationError{Field: "student_id", Code: "REQUIRED", Message: "student_id is required"})
		} else {
			if _, ok := existingStudentID[row.StudentID]; ok {
				errs = append(errs, ValidationError{Field: "student_id", Code: "DUPLICATE_STUDENT_ID", Message: "student_id already exists"})
			}
			if _, ok := idSeen[row.StudentID]; ok {
				errs = append(errs, ValidationError{Field: "student_id", Code: "DUPLICATE_STUDENT_ID", Message: "student_id is duplicated in file"})
			}
			idSeen[row.StudentID] = struct{}{}
		}
		if strings.TrimSpace(row.Grade) == "" {
			errs = append(errs, ValidationError{Field: "grade", Code: "REQUIRED", Message: "grade is required"})
		}
	case authz.RoleTeacher:
		if strings.TrimSpace(row.TeacherID) == "" {
			errs = append(errs, ValidationError{Field: "teacher_id", Code: "REQUIRED", Message: "teacher_id is required"})
		} else {
			if _, ok := existingTeacherID[row.TeacherID]; ok {
				errs = append(errs, ValidationError{Field: "teacher_id", Code: "DUPLICATE_TEACHER_ID", Message: "teacher_id already exists"})
			}
			if _, ok := idSeen[row.TeacherID]; ok {
				errs = append(errs, ValidationError{Field: "teacher_id", Code: "DUPLICATE_TEACHER_ID", Message: "teacher_id is duplicated in file"})
			}
			idSeen[row.TeacherID] = struct{}{}
		}
		if strings.TrimSpace(row.Title) == "" {
			errs = append(errs, ValidationError{Field: "title", Code: "REQUIRED", Message: "title is required"})
		}
	}

	if strings.TrimSpace(row.DepartmentID) == "" {
		errs = append(errs, ValidationError{Field: "department_id", Code: "REQUIRED", Message: "department_id is required"})
	}

	return errs
}

func fileHashMismatchDetails(expected string, got string) string {
	return fmt.Sprintf("expected %s, got %s", expected, got)
}
