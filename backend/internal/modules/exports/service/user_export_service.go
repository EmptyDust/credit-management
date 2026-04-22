package service

import (
	"bytes"
	"context"
	"encoding/csv"

	usersrepo "credit-management/backend/internal/modules/users/repo"
	"credit-management/backend/internal/platform/authz"
)

type UserExportService struct {
	users usersrepo.UserRepository
}

func NewUserExportService(users usersrepo.UserRepository) *UserExportService {
	return &UserExportService{users: users}
}

func (s *UserExportService) ExportStudents(ctx context.Context, filter usersrepo.ListFilter) ([]byte, error) {
	filter.UserType = authz.RoleStudent
	return s.exportUsers(ctx, filter, []string{"id", "username", "real_name", "student_id", "grade", "department_id", "status"})
}

func (s *UserExportService) ExportTeachers(ctx context.Context, filter usersrepo.ListFilter) ([]byte, error) {
	filter.UserType = authz.RoleTeacher
	return s.exportUsers(ctx, filter, []string{"id", "username", "real_name", "teacher_id", "title", "department_id", "status"})
}

func (s *UserExportService) exportUsers(ctx context.Context, filter usersrepo.ListFilter, header []string) ([]byte, error) {
	users, err := s.users.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, user := range users {
		record := []string{
			user.ID,
			user.Username,
			user.RealName,
		}
		if filter.UserType == authz.RoleStudent {
			record = append(record, user.StudentID, user.Grade, user.DepartmentID, user.Status)
		} else {
			record = append(record, user.TeacherID, user.Title, user.DepartmentID, user.Status)
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
