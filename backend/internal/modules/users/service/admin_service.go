package service

import (
	"context"
	"sort"

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
	Items      []AdminUser `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

type KeyCount struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type UserStats struct {
	Total        int            `json:"total"`
	StatusCounts map[string]int `json:"status_counts"`
	GradeCounts  []KeyCount     `json:"grade_counts,omitempty"`
	TitleCounts  []KeyCount     `json:"title_counts,omitempty"`
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

func (s *AdminService) StudentStats(ctx context.Context, filter usersrepo.ListFilter) (UserStats, error) {
	filter.UserType = authz.RoleStudent
	users, err := s.users.ListUsers(ctx, withoutPagination(filter))
	if err != nil {
		return UserStats{}, err
	}

	return buildUserStats(users, "grade"), nil
}

func (s *AdminService) TeacherStats(ctx context.Context, filter usersrepo.ListFilter) (UserStats, error) {
	filter.UserType = authz.RoleTeacher
	users, err := s.users.ListUsers(ctx, withoutPagination(filter))
	if err != nil {
		return UserStats{}, err
	}

	return buildUserStats(users, "title"), nil
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

	total := len(items)
	page, pageSize := normalizePagination(filter.Page, filter.PageSize)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return ListResult{
		Items:      items[start:end],
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func normalizePagination(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func withoutPagination(filter usersrepo.ListFilter) usersrepo.ListFilter {
	filter.Page = 0
	filter.PageSize = 0
	return filter
}

func buildUserStats(users []usersdomain.User, dimension string) UserStats {
	stats := UserStats{
		Total:        len(users),
		StatusCounts: map[string]int{},
	}
	dimensionCounts := map[string]int{}

	for _, user := range users {
		stats.StatusCounts[user.Status]++
		switch dimension {
		case "grade":
			if user.Grade != "" {
				dimensionCounts[user.Grade]++
			}
		case "title":
			if user.Title != "" {
				dimensionCounts[user.Title]++
			}
		}
	}

	switch dimension {
	case "grade":
		stats.GradeCounts = toKeyCounts(dimensionCounts)
	case "title":
		stats.TitleCounts = toKeyCounts(dimensionCounts)
	}

	return stats
}

func toKeyCounts(counts map[string]int) []KeyCount {
	items := make([]KeyCount, 0, len(counts))
	for key, count := range counts {
		items = append(items, KeyCount{
			Key:   key,
			Label: key,
			Count: count,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})
	return items
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
