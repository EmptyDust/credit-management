package repo

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	usersdomain "credit-management/backend/internal/modules/users/domain"
	"credit-management/backend/internal/platform/apperrors"
	"credit-management/backend/internal/platform/authz"
	"credit-management/backend/internal/platform/idgen"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormUserRepository struct {
	db *gorm.DB
}

type userRow struct {
	ID                  string     `gorm:"column:id"`
	Username            string     `gorm:"column:username"`
	PasswordHash        string     `gorm:"column:password_hash"`
	Email               *string    `gorm:"column:email"`
	Phone               *string    `gorm:"column:phone"`
	RealName            string     `gorm:"column:real_name"`
	UserType            string     `gorm:"column:user_type"`
	Status              string     `gorm:"column:status"`
	AvatarFileID        *string    `gorm:"column:avatar_file_id"`
	DepartmentID        *string    `gorm:"column:department_id"`
	StudentID           *string    `gorm:"column:student_id"`
	TeacherID           *string    `gorm:"column:teacher_id"`
	Grade               *string    `gorm:"column:grade"`
	Title               *string    `gorm:"column:title"`
	RefreshTokenVersion int64      `gorm:"column:refresh_token_version"`
	LastLoginAt         *time.Time `gorm:"column:last_login_at"`
	DeletedAt           *time.Time `gorm:"column:deleted_at"`
}

type fileRow struct {
	ID             string     `gorm:"column:id"`
	StorageKey     string     `gorm:"column:storage_key"`
	OriginalName   string     `gorm:"column:original_name"`
	MimeType       string     `gorm:"column:mime_type"`
	FileExt        *string    `gorm:"column:file_ext"`
	FileSize       int64      `gorm:"column:file_size"`
	StorageBackend string     `gorm:"column:storage_backend"`
	UploadedBy     string     `gorm:"column:uploaded_by"`
	UploadedAt     time.Time  `gorm:"column:uploaded_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

type avatarFileRow struct {
	ID        string     `gorm:"column:id"`
	UserID    string     `gorm:"column:user_id"`
	FileID    string     `gorm:"column:file_id"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (userRow) TableName() string {
	return "users"
}

func (fileRow) TableName() string {
	return "files"
}

func (avatarFileRow) TableName() string {
	return "avatar_files"
}

func NewGorm(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByLoginID(ctx context.Context, loginID string) (*usersdomain.User, error) {
	var row userRow
	if err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("username = ? OR student_id = ? OR teacher_id = ?", loginID, loginID, loginID).
		First(&row).Error; err != nil {
		return nil, normalizeUserError(err)
	}

	return row.toDomain(), nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, userID string) (*usersdomain.User, error) {
	return r.getByID(ctx, userID)
}

func (r *GormUserRepository) CreateStudent(ctx context.Context, input usersdomain.RegisterStudentInput) (*usersdomain.User, error) {
	return r.createUser(ctx, CreateUserInput{
		Username:     input.Username,
		Password:     input.Password,
		RealName:     input.RealName,
		Email:        input.Email,
		Phone:        input.Phone,
		UserType:     authz.RoleStudent,
		DepartmentID: input.DepartmentID,
		StudentID:    input.StudentID,
		Grade:        input.Grade,
	})
}

func (r *GormUserRepository) UpdateProfile(ctx context.Context, userID string, patch usersdomain.ProfilePatch) (*usersdomain.User, error) {
	result := r.db.WithContext(ctx).
		Model(&userRow{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{
			"email":         nullableString(patch.Email),
			"phone":         nullableString(patch.Phone),
			"real_name":     patch.RealName,
			"grade":         nullableString(patch.Grade),
			"title":         nullableString(patch.Title),
			"department_id": nullableString(patch.DepartmentID),
		})
	if result.Error != nil {
		return nil, normalizeUserError(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrUserNotFound
	}

	return r.getByID(ctx, userID)
}

func (r *GormUserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string, refreshTokenVersion int64) error {
	result := r.db.WithContext(ctx).
		Model(&userRow{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{
			"password_hash":          passwordHash,
			"refresh_token_version":  refreshTokenVersion,
			"updated_at":             time.Now().UTC(),
		})
	if result.Error != nil {
		return normalizeUserError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

func (r *GormUserRepository) UpdateAvatar(ctx context.Context, userID string, fileID string) (*usersdomain.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureAvatarFileRow(tx, userID, fileID); err != nil {
			return err
		}

		now := time.Now().UTC()
		if err := tx.Model(&avatarFileRow{}).
			Where("user_id = ? AND deleted_at IS NULL", userID).
			Update("deleted_at", now).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&avatarFileRow{
			ID:        idgen.New(),
			UserID:    userID,
			FileID:    fileID,
			CreatedAt: now,
		}).Error; err != nil {
			return err
		}

		result := tx.Model(&userRow{}).
			Where("id = ? AND deleted_at IS NULL", userID).
			Update("avatar_file_id", nullableString(fileID))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return apperrors.ErrUserNotFound
		}
		return nil
	})
	if err != nil {
		return nil, normalizeUserError(err)
	}
	return r.getByID(ctx, userID)
}

func (r *GormUserRepository) DeleteAvatar(ctx context.Context, userID string) (*usersdomain.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&avatarFileRow{}).
			Where("user_id = ? AND deleted_at IS NULL", userID).
			Update("deleted_at", now).Error; err != nil {
			return err
		}

		result := tx.Model(&userRow{}).
			Where("id = ? AND deleted_at IS NULL", userID).
			Update("avatar_file_id", nil)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return apperrors.ErrUserNotFound
		}
		return nil
	})
	if err != nil {
		return nil, normalizeUserError(err)
	}
	return r.getByID(ctx, userID)
}

func (r *GormUserRepository) ListUsers(ctx context.Context, filter ListFilter) ([]usersdomain.User, error) {
	query := r.db.WithContext(ctx).Model(&userRow{}).Where("deleted_at IS NULL")
	if filter.UserType != "" {
		query = query.Where("user_type = ?", filter.UserType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Grade != "" {
		query = query.Where("grade = ?", filter.Grade)
	}
	if filter.Title != "" {
		query = query.Where("title = ?", filter.Title)
	}
	if filter.Keyword != "" {
		like := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where(
			"username ILIKE ? OR real_name ILIKE ? OR COALESCE(student_id, '') ILIKE ? OR COALESCE(teacher_id, '') ILIKE ?",
			like, like, like, like,
		)
	}

	var rows []userRow
	if err := query.Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, normalizeUserError(err)
	}

	items := make([]usersdomain.User, 0, len(rows))
	for i := range rows {
		items = append(items, *rows[i].toDomain())
	}
	return items, nil
}

func (r *GormUserRepository) GetUser(ctx context.Context, userID string) (*usersdomain.User, error) {
	return r.getByID(ctx, userID)
}

func (r *GormUserRepository) CreateUser(ctx context.Context, input CreateUserInput) (*usersdomain.User, error) {
	return r.createUser(ctx, input)
}

func (r *GormUserRepository) UpdateUser(ctx context.Context, userID string, input UpdateUserInput) (*usersdomain.User, error) {
	updates := map[string]any{
		"real_name":     input.RealName,
		"email":         nullableString(input.Email),
		"phone":         nullableString(input.Phone),
		"department_id": nullableString(input.DepartmentID),
		"student_id":    nullableString(input.StudentID),
		"teacher_id":    nullableString(input.TeacherID),
		"grade":         nullableString(input.Grade),
		"title":         nullableString(input.Title),
	}
	if strings.TrimSpace(input.Username) != "" {
		updates["username"] = input.Username
	}

	result := r.db.WithContext(ctx).
		Model(&userRow{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(updates)
	if result.Error != nil {
		return nil, normalizeUserError(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrUserNotFound
	}

	return r.getByID(ctx, userID)
}

func (r *GormUserRepository) SoftDeleteUser(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&userRow{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("deleted_at", time.Now().UTC())
	if result.Error != nil {
		return normalizeUserError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

func (r *GormUserRepository) UpdateStatus(ctx context.Context, userID string, status string) (*usersdomain.User, error) {
	result := r.db.WithContext(ctx).
		Model(&userRow{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("status", status)
	if result.Error != nil {
		return nil, normalizeUserError(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrUserNotFound
	}

	return r.getByID(ctx, userID)
}

func (r *GormUserRepository) ResetPassword(ctx context.Context, userID string, passwordHash string, refreshTokenVersion int64) error {
	return r.UpdatePassword(ctx, userID, passwordHash, refreshTokenVersion)
}

func (r *GormUserRepository) createUser(ctx context.Context, input CreateUserInput) (*usersdomain.User, error) {
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	row := userRow{
		ID:                  idgen.New(),
		Username:            input.Username,
		PasswordHash:        passwordHash,
		Email:               nullableString(input.Email),
		Phone:               nullableString(input.Phone),
		RealName:            input.RealName,
		UserType:            input.UserType,
		Status:              "active",
		DepartmentID:        nullableString(input.DepartmentID),
		StudentID:           nullableString(input.StudentID),
		TeacherID:           nullableString(input.TeacherID),
		Grade:               nullableString(input.Grade),
		Title:               nullableString(input.Title),
		RefreshTokenVersion: 1,
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, normalizeUserError(err)
	}
	return row.toDomain(), nil
}

func (r *GormUserRepository) getByID(ctx context.Context, userID string) (*usersdomain.User, error) {
	var row userRow
	if err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		First(&row, "id = ?", userID).Error; err != nil {
		return nil, normalizeUserError(err)
	}
	return row.toDomain(), nil
}

func normalizeUserError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "uq_users_username":
			return fmt.Errorf("%w: username", apperrors.ErrConflict)
		case "uq_users_student_id":
			return fmt.Errorf("%w: student_id", apperrors.ErrConflict)
		case "uq_users_teacher_id":
			return fmt.Errorf("%w: teacher_id", apperrors.ErrConflict)
		case "uq_users_email_active":
			return fmt.Errorf("%w: email", apperrors.ErrConflict)
		case "uq_users_phone_active":
			return fmt.Errorf("%w: phone", apperrors.ErrConflict)
		}
		return fmt.Errorf("%w: %s", apperrors.ErrConflict, pgErr.ConstraintName)
	}

	return err
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func (r userRow) toDomain() *usersdomain.User {
	return &usersdomain.User{
		ID:                  r.ID,
		Username:            r.Username,
		PasswordHash:        r.PasswordHash,
		Email:               derefString(r.Email),
		Phone:               derefString(r.Phone),
		RealName:            r.RealName,
		UserType:            r.UserType,
		Status:              r.Status,
		AvatarFileID:        derefString(r.AvatarFileID),
		DepartmentID:        derefString(r.DepartmentID),
		StudentID:           derefString(r.StudentID),
		TeacherID:           derefString(r.TeacherID),
		Grade:               derefString(r.Grade),
		Title:               derefString(r.Title),
		RefreshTokenVersion: r.RefreshTokenVersion,
		LastLoginAt:         cloneTimePtr(r.LastLoginAt),
		DeletedAt:           cloneTimePtr(r.DeletedAt),
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	t := value.UTC()
	return &t
}

func ensureAvatarFileRow(tx *gorm.DB, userID string, fileID string) error {
	now := time.Now().UTC()
	fileExt := strings.TrimPrefix(filepath.Ext(fileID), ".")

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&fileRow{
		ID:             fileID,
		StorageKey:     "avatars/" + fileID,
		OriginalName:   fileID,
		MimeType:       "application/octet-stream",
		FileExt:        nullableString(fileExt),
		FileSize:       0,
		StorageBackend: "local",
		UploadedBy:     userID,
		UploadedAt:     now,
	}).Error
}
