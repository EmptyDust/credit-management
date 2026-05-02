package auditlog

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"credit-management/backend/internal/platform/database"

	"gorm.io/gorm"
)

func TestWriterRecordAppendsEntries(t *testing.T) {
	writer := NewWriter()

	if err := writer.Record(context.Background(), Entry{Action: "create"}); err != nil {
		t.Fatalf("expected first record to succeed, got %v", err)
	}
	if err := writer.Record(context.Background(), Entry{Action: "delete"}); err != nil {
		t.Fatalf("expected second record to succeed, got %v", err)
	}

	entries := writer.List(Filter{})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Action != "create" || entries[1].Action != "delete" {
		t.Fatalf("expected append-only order, got %+v", entries)
	}
}

func TestGormWriterRecordListAndExport(t *testing.T) {
	writer, tx := newTestGormWriter(t)
	defer func() { _ = tx.Rollback().Error }()

	firstTime := time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC)
	secondTime := firstTime.Add(time.Minute)

	if err := writer.Record(context.Background(), Entry{
		ID:         "01JAUDIT000000000000000001",
		ActorID:    "01JADMIN000000000000000000",
		ActorRole:  "admin",
		TargetType: "activity",
		TargetID:   "01JACTIVITY000000000000000",
		Action:     "approve",
		Result:     "success",
		RequestID:  "req-1",
		CreatedAt:  firstTime,
	}); err != nil {
		t.Fatalf("record first audit log: %v", err)
	}
	if err := writer.Record(context.Background(), Entry{
		ID:         "01JAUDIT000000000000000002",
		ActorRole:  "teacher",
		TargetType: "activity",
		TargetID:   "01JACTIVITY000000000000001",
		Action:     "reject",
		Result:     "failed",
		RequestID:  "req-2",
		CreatedAt:  secondTime,
	}); err != nil {
		t.Fatalf("record second audit log: %v", err)
	}

	filtered := writer.List(Filter{Action: "approve"})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered entry, got %d", len(filtered))
	}
	if filtered[0].ID != "01JAUDIT000000000000000001" {
		t.Fatalf("unexpected filtered entry id %s", filtered[0].ID)
	}

	all := writer.List(Filter{})
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if !all[0].CreatedAt.Equal(firstTime) || !all[1].CreatedAt.Equal(secondTime) {
		t.Fatalf("expected created_at ordering, got %+v", all)
	}

	content, err := writer.ExportCSV(Filter{TargetType: "activity"})
	if err != nil {
		t.Fatalf("export csv: %v", err)
	}
	if string(content) == "" {
		t.Fatal("expected csv content")
	}
	if got := string(content); !strings.Contains(got, "approve") || !strings.Contains(got, "reject") {
		t.Fatalf("expected csv to contain both actions, got %q", got)
	}
}

func newTestGormWriter(t *testing.T) (*Writer, *gorm.DB) {
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

	return NewGormWriter(tx), tx
}
