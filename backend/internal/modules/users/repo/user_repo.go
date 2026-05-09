package repo

import (
	"context"

	usersdomain "credit-management/backend/internal/modules/users/domain"
)

type ListFilter struct {
	UserType string
	Keyword  string
	Status   string
	Grade    string
	Title    string
	Page     int
	PageSize int
}

type CreateUserInput struct {
	Username     string
	Password     string
	RealName     string
	Email        string
	Phone        string
	UserType     string
	DepartmentID string
	StudentID    string
	TeacherID    string
	Grade        string
	Title        string
}

type UpdateUserInput struct {
	Username     string
	RealName     string
	Email        string
	Phone        string
	DepartmentID string
	StudentID    string
	TeacherID    string
	Grade        string
	Title        string
}

type UserRepository interface {
	ListUsers(ctx context.Context, filter ListFilter) ([]usersdomain.User, error)
	GetUser(ctx context.Context, userID string) (*usersdomain.User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (*usersdomain.User, error)
	UpdateUser(ctx context.Context, userID string, input UpdateUserInput) (*usersdomain.User, error)
	SoftDeleteUser(ctx context.Context, userID string) error
	UpdateStatus(ctx context.Context, userID string, status string) (*usersdomain.User, error)
	ResetPassword(ctx context.Context, userID string, passwordHash string, refreshTokenVersion int64) error
}
