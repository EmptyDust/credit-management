package database

import (
	"path/filepath"
	"testing"
)

func TestMigratorDiscoversSQLFiles(t *testing.T) {
	files, err := DiscoverMigrations("../../../migrations")
	if err != nil {
		t.Fatalf("expected no error discovering migrations, got %v", err)
	}

	if len(files) != 6 {
		t.Fatalf("expected 6 migration files, got %d", len(files))
	}

	if got := filepath.Base(files[0]); got != "0001_init_extensions.sql" {
		t.Fatalf("expected first migration 0001_init_extensions.sql, got %s", got)
	}

	if got := filepath.Base(files[len(files)-1]); got != "0006_add_activity_field_placeholders.sql" {
		t.Fatalf("expected last migration 0006_add_activity_field_placeholders.sql, got %s", got)
	}
}
