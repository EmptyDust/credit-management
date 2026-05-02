package repo_test

import (
	"context"
	"os"
	"testing"

	"credit-management/backend/internal/modules/users/domain"
	repo "credit-management/backend/internal/modules/users/repo"
	"credit-management/backend/internal/platform/database"
	"credit-management/backend/internal/platform/idgen"

	"gorm.io/gorm"
)

func TestGormUserRepositoryCreateStudentHashesPassword(t *testing.T) {
	repo, tx := newTestUserRepo(t)
	defer func() { _ = tx.Rollback().Error }()

	user, err := repo.CreateStudent(context.Background(), domain.RegisterStudentInput{
		Username:     "student_task3_repo",
		Password:     "Passw0rd!",
		RealName:     "Repo Student",
		Email:        "student_task3_repo@example.com",
		Phone:        "13800138123",
		StudentID:    "20249999",
		Grade:        "2024",
		DepartmentID: "class-01",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if user.PasswordHash == "Passw0rd!" || user.PasswordHash == "" {
		t.Fatalf("expected hashed password, got %q", user.PasswordHash)
	}

	found, err := repo.FindByLoginID(context.Background(), "student_task3_repo")
	if err != nil {
		t.Fatalf("find by login id: %v", err)
	}
	if found.ID != user.ID {
		t.Fatalf("expected same user id, got %s want %s", found.ID, user.ID)
	}
}

func TestGormUserRepositoryUpdateAvatarPersistsFileMetadata(t *testing.T) {
	repo, tx := newTestUserRepo(t)
	defer func() { _ = tx.Rollback().Error }()

	fileID := idgen.New()
	user, err := repo.UpdateAvatar(context.Background(), "01JSTUDENT0000000000000000", fileID)
	if err != nil {
		t.Fatalf("update avatar: %v", err)
	}
	if user.AvatarFileID != fileID {
		t.Fatalf("expected avatar file id %s, got %s", fileID, user.AvatarFileID)
	}

	var fileCount int64
	if err := tx.Table("files").Where("id = ?", fileID).Count(&fileCount).Error; err != nil {
		t.Fatalf("count files: %v", err)
	}
	if fileCount != 1 {
		t.Fatalf("expected one files row, got %d", fileCount)
	}

	var avatarLinkCount int64
	if err := tx.Table("avatar_files").Where("user_id = ? AND file_id = ? AND deleted_at IS NULL", "01JSTUDENT0000000000000000", fileID).Count(&avatarLinkCount).Error; err != nil {
		t.Fatalf("count avatar_files: %v", err)
	}
	if avatarLinkCount != 1 {
		t.Fatalf("expected one active avatar_files row, got %d", avatarLinkCount)
	}
}

func newTestUserRepo(t *testing.T) (*repo.GormUserRepository, *gorm.DB) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx := context.Background()
	db, err := database.Open(ctx, database.Options{DSN: dsn})
	if err != nil {
		t.Skipf("test database unavailable: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := database.NewMigrator("../../../../migrations").Apply(ctx, sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if err := database.NewSeeder().Run(ctx, db); err != nil {
		t.Fatalf("seed database: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}

	return repo.NewGorm(tx), tx
}
