package service

import (
	"context"

	authservice "credit-management/backend/internal/modules/auth/service"
	usersdomain "credit-management/backend/internal/modules/users/domain"
	usersrepo "credit-management/backend/internal/modules/users/repo"
	"credit-management/backend/internal/platform/authz"
)

type AdminSessionStore interface {
	RevokeAllUserSessions(ctx context.Context, userID string, reason string) error
}

type AdminService struct {
	users    usersrepo.UserRepository
	sessions AdminSessionStore
}

type AdminUser struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	RealName     string `json:"real_name"`
	UserType     string `json:"user_type"`
	Status       string `json:"status"`
	DepartmentID string `json:"department_id,omitempty"`
	StudentID    string `json:"student_id,omitempty"`
	TeacherID    string `json:"teacher_id,omitempty"`
	Grade        string `json:"grade,omitempty"`
	Title        string `json:"title,omitempty"`
}

type ListResult struct {
	Items []AdminUser `json:"items"`
	Total int         `json:"total"`
}

func NewAdminService(users usersrepo.UserRepository, sessions AdminSessionStore) *AdminService {
	return &AdminService{
		users:    users,
		sessions: sessions,
	}
}

func (s *AdminService) ListStudents(ctx context.Context, filter usersrepo.ListFilter) (ListResult, error) {
	filter.UserType = authz.RoleStudent
	return s.listUsers(ctx, filter)
}

func (s *AdminService) ListTeachers(ctx context.Context, filter usersrepo.ListFilter) (ListResult, error) {
	filter.UserType = authz.RoleTeacher
	return s.listUsers(ctx, filter)
}

func (s *AdminService) GetUser(ctx context.Context, userID string) (AdminUser, error) {
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return AdminUser{}, err
	}

	return toAdminUser(user), nil
}

func (s *AdminService) CreateStudent(ctx context.Context, input usersrepo.CreateUserInput) (AdminUser, error) {
	input.UserType = authz.RoleStudent
	user, err := s.users.CreateUser(ctx, input)
	if err != nil {
		return AdminUser{}, err
	}

	return toAdminUser(user), nil
}

func (s *AdminService) CreateTeacher(ctx context.Context, input usersrepo.CreateUserInput) (AdminUser, error) {
	input.UserType = authz.RoleTeacher
	user, err := s.users.CreateUser(ctx, input)
	if err != nil {
		return AdminUser{}, err
	}

	return toAdminUser(user), nil
}

func (s *AdminService) UpdateUser(ctx context.Context, userID string, input usersrepo.UpdateUserInput) (AdminUser, error) {
	user, err := s.users.UpdateUser(ctx, userID, input)
	if err != nil {
		return AdminUser{}, err
	}

	return toAdminUser(user), nil
}

func (s *AdminService) DeleteUser(ctx context.Context, userID string) error {
	return s.users.SoftDeleteUser(ctx, userID)
}

func (s *AdminService) UpdateStatus(ctx context.Context, userID string, status string) (AdminUser, error) {
	user, err := s.users.UpdateStatus(ctx, userID, status)
	if err != nil {
		return AdminUser{}, err
	}

	if status == "inactive" || status == "suspended" {
		if err := s.sessions.RevokeAllUserSessions(ctx, userID, "status_changed"); err != nil {
			return AdminUser{}, err
		}
	}

	return toAdminUser(user), nil
}

func (s *AdminService) ResetPassword(ctx context.Context, userID string, newPassword string) error {
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	hash, err := authservice.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.users.ResetPassword(ctx, userID, hash, user.RefreshTokenVersion+1); err != nil {
		return err
	}

	return s.sessions.RevokeAllUserSessions(ctx, userID, "password_reset")
}

func (s *AdminService) listUsers(ctx context.Context, filter usersrepo.ListFilter) (ListResult, error) {
	users, err := s.users.ListUsers(ctx, filter)
	if err != nil {
		return ListResult{}, err
	}

	items := make([]AdminUser, 0, len(users))
	for i := range users {
		user := users[i]
		items = append(items, toAdminUser(&user))
	}

	return ListResult{
		Items: items,
		Total: len(items),
	}, nil
}

func toAdminUser(user *usersdomain.User) AdminUser {
	return AdminUser{
		ID:           user.ID,
		Username:     user.Username,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		DepartmentID: user.DepartmentID,
		StudentID:    user.StudentID,
		TeacherID:    user.TeacherID,
		Grade:        user.Grade,
		Title:        user.Title,
	}
}
