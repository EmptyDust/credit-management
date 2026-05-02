package repo_test

import (
	"context"
	"os"
	"testing"
	"time"

	authdomain "credit-management/backend/internal/modules/auth/domain"
	repo "credit-management/backend/internal/modules/auth/repo"
	"credit-management/backend/internal/platform/database"

	"gorm.io/gorm"
)

func TestGormSessionRepoSaveAndRevokeAllUserSessions(t *testing.T) {
	repo, tx := newTestSessionRepo(t)
	defer func() { _ = tx.Rollback().Error }()

	session := authdomain.Session{
		ID:               "01JSESSIONTASK3000000000000",
		UserID:           "01JSTUDENT0000000000000000",
		RefreshTokenHash: "hash-value",
		IssuedAt:         time.Now().UTC().Add(-time.Minute),
		ExpiresAt:        time.Now().UTC().Add(time.Hour),
	}
	if err := repo.SaveSession(context.Background(), session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	if err := repo.RevokeAllUserSessions(context.Background(), session.UserID, "password_changed"); err != nil {
		t.Fatalf("revoke all user sessions: %v", err)
	}

	stored, err := repo.GetSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if stored.RevokedAt == nil {
		t.Fatal("expected revoked session")
	}
	if stored.RevokeReason != "password_changed" {
		t.Fatalf("expected revoke reason password_changed, got %s", stored.RevokeReason)
	}
}

func newTestSessionRepo(t *testing.T) (*repo.GormSessionRepo, *gorm.DB) {
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

	return repo.NewGorm(tx, nil), tx
}
