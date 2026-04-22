package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	authdomain "credit-management/backend/internal/modules/auth/domain"
	usersdomain "credit-management/backend/internal/modules/users/domain"
	"credit-management/backend/internal/platform/authn"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/idgen"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInactive    = errors.New("account inactive")
	ErrRefreshInvalid     = errors.New("refresh token invalid")
	ErrUserNotFound       = errors.New("user not found")
	ErrConflict           = errors.New("resource conflict")
)

type UserStore interface {
	FindByLoginID(ctx context.Context, loginID string) (*usersdomain.User, error)
	FindByID(ctx context.Context, userID string) (*usersdomain.User, error)
	CreateStudent(ctx context.Context, input RegisterStudentInput) (*usersdomain.User, error)
	UpdateProfile(ctx context.Context, userID string, patch usersdomain.ProfilePatch) (*usersdomain.User, error)
	UpdatePassword(ctx context.Context, userID string, passwordHash string, refreshTokenVersion int64) error
	UpdateAvatar(ctx context.Context, userID string, fileID string) (*usersdomain.User, error)
	DeleteAvatar(ctx context.Context, userID string) (*usersdomain.User, error)
}

type SessionStore interface {
	SaveSession(ctx context.Context, session authdomain.Session) error
	GetSession(ctx context.Context, sessionID string) (*authdomain.Session, error)
	RevokeSession(ctx context.Context, sessionID string, reason string) error
	RevokeAllUserSessions(ctx context.Context, userID string, reason string) error
}

type Service struct {
	users    UserStore
	sessions SessionStore
}

type LoginInput struct {
	LoginID  string
	Password string
}

type RegisterStudentInput struct {
	Username     string
	Password     string
	RealName     string
	Email        string
	Phone        string
	StudentID    string
	Grade        string
	DepartmentID string
}

type UserSummary struct {
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
	AvatarURL    string `json:"avatar_url,omitempty"`
}

type AuthResult struct {
	AccessToken      string      `json:"access_token"`
	RefreshToken     string      `json:"refresh_token"`
	ExpiresIn        int64       `json:"expires_in"`
	RefreshExpiresIn int64       `json:"refresh_expires_in"`
	User             UserSummary `json:"user"`
}

type MemoryStore struct {
	mu       sync.RWMutex
	users    map[string]*usersdomain.User
	sessions map[string]*authdomain.Session
}

func New(users UserStore, sessions SessionStore) *Service {
	return &Service{
		users:    users,
		sessions: sessions,
	}
}

func NewMemoryStore() (*MemoryStore, error) {
	passwordHash, err := HashPassword("Passw0rd!")
	if err != nil {
		return nil, err
	}

	return &MemoryStore{
		users: map[string]*usersdomain.User{
			"01JSTUDENT00000000000000000": {
				ID:                  "01JSTUDENT00000000000000000",
				Username:            "student01",
				PasswordHash:        passwordHash,
				Email:               "student01@example.com",
				Phone:               "13800138000",
				RealName:            "Student One",
				UserType:            authz.RoleStudent,
				Status:              "active",
				DepartmentID:        "class-01",
				StudentID:           "20240001",
				Grade:               "2024",
				RefreshTokenVersion: 1,
			},
			"01JTEACHER00000000000000000": {
				ID:                  "01JTEACHER00000000000000000",
				Username:            "teacher01",
				PasswordHash:        passwordHash,
				Email:               "teacher01@example.com",
				Phone:               "13800138001",
				RealName:            "Teacher One",
				UserType:            authz.RoleTeacher,
				Status:              "active",
				DepartmentID:        "college-01",
				TeacherID:           "T2024001",
				Title:               "lecturer",
				RefreshTokenVersion: 1,
			},
			"01JADMIN0000000000000000000": {
				ID:                  "01JADMIN0000000000000000000",
				Username:            "admin01",
				PasswordHash:        passwordHash,
				Email:               "admin01@example.com",
				Phone:               "13800138002",
				RealName:            "Admin One",
				UserType:            authz.RoleAdmin,
				Status:              "active",
				RefreshTokenVersion: 1,
			},
		},
		sessions: make(map[string]*authdomain.Session),
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	user, err := s.users.FindByLoginID(ctx, strings.TrimSpace(input.LoginID))
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if user.Status != "active" {
		return AuthResult{}, ErrAccountInactive
	}
	if err := ComparePassword(user.PasswordHash, input.Password); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) RegisterStudent(ctx context.Context, input RegisterStudentInput) (*usersdomain.User, error) {
	user, err := s.users.CreateStudent(ctx, input)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (AuthResult, error) {
	claims, err := authn.ParseRefreshToken(refreshToken)
	if err != nil {
		return AuthResult{}, ErrRefreshInvalid
	}

	session, err := s.sessions.GetSession(ctx, claims.SessionID)
	if err != nil {
		return AuthResult{}, ErrRefreshInvalid
	}
	if session.IsRevoked() || time.Now().UTC().After(session.ExpiresAt) {
		return AuthResult{}, ErrRefreshInvalid
	}
	if session.RefreshTokenHash != hashToken(refreshToken) {
		return AuthResult{}, ErrRefreshInvalid
	}

	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return AuthResult{}, ErrRefreshInvalid
	}
	if user.Status != "active" {
		return AuthResult{}, ErrAccountInactive
	}
	if user.RefreshTokenVersion != claims.TokenVersion {
		return AuthResult{}, ErrRefreshInvalid
	}

	if err := s.sessions.RevokeSession(ctx, session.ID, "refresh_rotated"); err != nil {
		return AuthResult{}, err
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Logout(ctx context.Context, accessToken string) error {
	claims, err := authn.ParseAccessToken(accessToken)
	if err != nil {
		return err
	}
	if claims.SessionID == "" {
		return nil
	}

	return s.sessions.RevokeSession(ctx, claims.SessionID, "logout")
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ComparePassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *Service) issueTokens(ctx context.Context, user *usersdomain.User) (AuthResult, error) {
	sessionID := idgen.New()

	accessToken, err := authn.IssueAccessToken(authn.ClaimsInput{
		UserID:    user.ID,
		Username:  user.Username,
		UserType:  user.UserType,
		SessionID: sessionID,
	})
	if err != nil {
		return AuthResult{}, err
	}

	refreshToken, err := authn.IssueRefreshToken(authn.ClaimsInput{
		UserID:       user.ID,
		Username:     user.Username,
		UserType:     user.UserType,
		SessionID:    sessionID,
		TokenVersion: user.RefreshTokenVersion,
	})
	if err != nil {
		return AuthResult{}, err
	}

	now := time.Now().UTC()
	if err := s.sessions.SaveSession(ctx, authdomain.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshToken),
		IssuedAt:         now,
		ExpiresAt:        now.Add(authn.RefreshTTL()),
	}); err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(authn.AccessTTL().Seconds()),
		RefreshExpiresIn: int64(authn.RefreshTTL().Seconds()),
		User:             toUserSummary(user),
	}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func toUserSummary(user *usersdomain.User) UserSummary {
	return UserSummary{
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
		AvatarURL:    avatarURL(user.AvatarFileID),
	}
}

func avatarURL(fileID string) string {
	if fileID == "" {
		return ""
	}

	return fmt.Sprintf("https://example.invalid/avatars/%s", fileID)
}

func (s *MemoryStore) FindByLoginID(_ context.Context, loginID string) (*usersdomain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == loginID || user.StudentID == loginID || user.TeacherID == loginID {
			return cloneUser(user), nil
		}
	}

	return nil, ErrUserNotFound
}

func (s *MemoryStore) FindByID(_ context.Context, userID string) (*usersdomain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return cloneUser(user), nil
}

func (s *MemoryStore) CreateStudent(_ context.Context, input RegisterStudentInput) (*usersdomain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureUnique("", input.Username, input.StudentID, input.Email, input.Phone); err != nil {
		return nil, err
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &usersdomain.User{
		ID:                  idgen.New(),
		Username:            input.Username,
		PasswordHash:        passwordHash,
		Email:               input.Email,
		Phone:               input.Phone,
		RealName:            input.RealName,
		UserType:            authz.RoleStudent,
		Status:              "active",
		DepartmentID:        input.DepartmentID,
		StudentID:           input.StudentID,
		Grade:               input.Grade,
		RefreshTokenVersion: 1,
	}
	s.users[user.ID] = user

	return cloneUser(user), nil
}

func (s *MemoryStore) UpdateProfile(_ context.Context, userID string, patch usersdomain.ProfilePatch) (*usersdomain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	if err := s.ensureUnique(userID, user.Username, user.StudentID, patch.Email, patch.Phone); err != nil {
		return nil, err
	}

	user.Email = patch.Email
	user.Phone = patch.Phone
	user.RealName = patch.RealName
	user.Grade = patch.Grade
	user.Title = patch.Title
	user.DepartmentID = patch.DepartmentID

	return cloneUser(user), nil
}

func (s *MemoryStore) UpdatePassword(_ context.Context, userID string, passwordHash string, refreshTokenVersion int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return ErrUserNotFound
	}

	user.PasswordHash = passwordHash
	user.RefreshTokenVersion = refreshTokenVersion
	return nil
}

func (s *MemoryStore) UpdateAvatar(_ context.Context, userID string, fileID string) (*usersdomain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	user.AvatarFileID = fileID
	return cloneUser(user), nil
}

func (s *MemoryStore) DeleteAvatar(_ context.Context, userID string) (*usersdomain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	user.AvatarFileID = ""
	return cloneUser(user), nil
}

func (s *MemoryStore) SaveSession(_ context.Context, session authdomain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	copy := session
	s.sessions[session.ID] = &copy
	return nil
}

func (s *MemoryStore) GetSession(_ context.Context, sessionID string) (*authdomain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrRefreshInvalid
	}

	return cloneSession(session), nil
}

func (s *MemoryStore) RevokeSession(_ context.Context, sessionID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}

	now := time.Now().UTC()
	session.RevokedAt = &now
	session.RevokeReason = reason
	return nil
}

func (s *MemoryStore) RevokeAllUserSessions(_ context.Context, userID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for _, session := range s.sessions {
		if session.UserID != userID {
			continue
		}

		session.RevokedAt = &now
		session.RevokeReason = reason
	}

	return nil
}

func (s *MemoryStore) ensureUnique(userID string, username string, studentID string, email string, phone string) error {
	for id, user := range s.users {
		if id == userID {
			continue
		}
		if user.Username == username {
			return fmt.Errorf("%w: username", ErrConflict)
		}
		if studentID != "" && user.StudentID == studentID {
			return fmt.Errorf("%w: student_id", ErrConflict)
		}
		if email != "" && user.Status == "active" && strings.EqualFold(user.Email, email) {
			return fmt.Errorf("%w: email", ErrConflict)
		}
		if phone != "" && user.Status == "active" && user.Phone == phone {
			return fmt.Errorf("%w: phone", ErrConflict)
		}
	}

	return nil
}

func cloneUser(user *usersdomain.User) *usersdomain.User {
	if user == nil {
		return nil
	}

	copy := *user
	if user.LastLoginAt != nil {
		t := *user.LastLoginAt
		copy.LastLoginAt = &t
	}
	return &copy
}

func cloneSession(session *authdomain.Session) *authdomain.Session {
	if session == nil {
		return nil
	}

	copy := *session
	if session.RevokedAt != nil {
		t := *session.RevokedAt
		copy.RevokedAt = &t
	}
	return &copy
}
