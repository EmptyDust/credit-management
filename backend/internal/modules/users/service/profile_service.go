package service

import (
	"context"
	"time"

	authservice "credit-management/backend/internal/modules/auth/service"
	usersdomain "credit-management/backend/internal/modules/users/domain"
	"credit-management/backend/internal/platform/idgen"
)

type UserStore interface {
	FindByID(ctx context.Context, userID string) (*usersdomain.User, error)
	UpdateProfile(ctx context.Context, userID string, patch usersdomain.ProfilePatch) (*usersdomain.User, error)
	UpdatePassword(ctx context.Context, userID string, passwordHash string, refreshTokenVersion int64) error
	UpdateAvatar(ctx context.Context, userID string, fileID string) (*usersdomain.User, error)
	DeleteAvatar(ctx context.Context, userID string) (*usersdomain.User, error)
}

type ProfileSessionStore interface {
	RevokeAllUserSessions(ctx context.Context, userID string, reason string) error
}

type Service struct {
	users    UserStore
	sessions ProfileSessionStore
}

type Actions struct {
	CanEdit           bool `json:"can_edit"`
	CanChangePassword bool `json:"can_change_password"`
	CanUploadAvatar   bool `json:"can_upload_avatar"`
	CanDeleteAvatar   bool `json:"can_delete_avatar"`
}

type Profile struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	Email        string  `json:"email,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	RealName     string  `json:"real_name"`
	UserType     string  `json:"user_type"`
	Status       string  `json:"status"`
	DepartmentID string  `json:"department_id,omitempty"`
	StudentID    string  `json:"student_id,omitempty"`
	TeacherID    string  `json:"teacher_id,omitempty"`
	Grade        string  `json:"grade,omitempty"`
	Title        string  `json:"title,omitempty"`
	AvatarURL    string  `json:"avatar_url,omitempty"`
	Actions      Actions `json:"actions"`
}

type AvatarResult struct {
	FileID    string    `json:"file_id"`
	AvatarURL string    `json:"avatar_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewProfileService(users UserStore, sessions ProfileSessionStore) *Service {
	return &Service{
		users:    users,
		sessions: sessions,
	}
}

func (s *Service) GetProfile(ctx context.Context, userID string) (Profile, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return Profile{}, err
	}

	return toProfile(user), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, patch usersdomain.ProfilePatch) (Profile, error) {
	user, err := s.users.UpdateProfile(ctx, userID, patch)
	if err != nil {
		return Profile{}, err
	}

	return toProfile(user), nil
}

func (s *Service) ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := authservice.ComparePassword(user.PasswordHash, oldPassword); err != nil {
		return authservice.ErrInvalidCredentials
	}

	hash, err := authservice.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePassword(ctx, userID, hash, user.RefreshTokenVersion+1); err != nil {
		return err
	}

	return s.sessions.RevokeAllUserSessions(ctx, userID, "password_changed")
}

func (s *Service) UploadAvatar(ctx context.Context, userID string, _ string) (AvatarResult, error) {
	fileID := idgen.New()
	user, err := s.users.UpdateAvatar(ctx, userID, fileID)
	if err != nil {
		return AvatarResult{}, err
	}

	expiresAt := time.Now().UTC().Add(time.Hour)
	return AvatarResult{
		FileID:    fileID,
		AvatarURL: avatarURL(user.AvatarFileID),
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) DeleteAvatar(ctx context.Context, userID string) error {
	_, err := s.users.DeleteAvatar(ctx, userID)
	return err
}

func toProfile(user *usersdomain.User) Profile {
	return Profile{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		RealName:     user.RealName,
		UserType:     user.UserType,
		Status:       user.Status,
		DepartmentID: user.DepartmentID,
		StudentID:    user.StudentID,
		TeacherID:    user.TeacherID,
		Grade:        user.Grade,
		Title:        user.Title,
		AvatarURL:    avatarURL(user.AvatarFileID),
		Actions: Actions{
			CanEdit:           true,
			CanChangePassword: true,
			CanUploadAvatar:   true,
			CanDeleteAvatar:   user.AvatarFileID != "",
		},
	}
}

func avatarURL(fileID string) string {
	if fileID == "" {
		return ""
	}

	return "https://example.invalid/avatars/" + fileID
}
