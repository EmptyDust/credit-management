package database

import (
	"context"
	"os"
	"testing"

	authservice "credit-management/backend/internal/modules/auth/service"

	"gorm.io/gorm"
)

func TestSeederRejectsNilDatabase(t *testing.T) {
	err := NewSeeder().Run(context.Background(), nil)
	if err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestSeedUsersPreserveDefaultAdminCredentials(t *testing.T) {
	hash, err := authservice.HashPassword(defaultAdminPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	rows := seedUsers(hash)
	if len(rows) != 1 {
		t.Fatalf("expected exactly one seed user, got %d", len(rows))
	}

	admin := rows[0]
	if admin.ID != defaultAdminID || admin.Username != defaultAdminUsername || admin.UserType != "admin" {
		t.Fatalf("unexpected admin seed row: %#v", admin)
	}
	if err := authservice.ComparePassword(admin.PasswordHash, defaultAdminPassword); err != nil {
		t.Fatalf("default admin password hash mismatch: %v", err)
	}
}

func TestSeedOrganizationNodesUseCompatibilityIDs(t *testing.T) {
	rows := seedOrganizationNodes()
	if len(rows) != 4 {
		t.Fatalf("expected 4 organization nodes, got %d", len(rows))
	}

	expected := []string{"school-01", "college-01", "major-01", "class-01"}
	for i, id := range expected {
		if rows[i].ID != id {
			t.Fatalf("expected row %d id %s, got %s", i, id, rows[i].ID)
		}
	}
	if rows[1].ParentID == nil || *rows[1].ParentID != "school-01" {
		t.Fatalf("expected college parent school-01, got %#v", rows[1].ParentID)
	}
	if rows[3].Path != "school-01/college-01/major-01/class-01" {
		t.Fatalf("unexpected class path: %s", rows[3].Path)
	}
}

func TestSeedOptionDefinitionsCoverRequiredGroupsAndTitles(t *testing.T) {
	groups := seedOptionGroups()
	if len(groups) != 4 {
		t.Fatalf("expected 4 option groups, got %d", len(groups))
	}

	items := seedOptionItems(map[string]string{
		"grades":         "optgrp-grades-01",
		"academic_years": "optgrp-acyears-01",
		"semesters":      "optgrp-semesters01",
		"teacher_titles": "optgrp-ttitles-01",
	})
	titleCount := 0
	for _, item := range items {
		if item.GroupID == "optgrp-ttitles-01" {
			titleCount++
		}
	}
	if titleCount != 4 {
		t.Fatalf("expected 4 teacher title items, got %d", titleCount)
	}
}

func TestSeedOptionItemsUsesResolvedGroupIDs(t *testing.T) {
	items := seedOptionItems(map[string]string{
		"grades":         "db-grades-id",
		"academic_years": "db-academic-id",
		"semesters":      "db-semesters-id",
		"teacher_titles": "db-teacher-id",
	})

	for _, item := range items {
		switch item.Value {
		case "2024":
			if item.GroupID != "db-grades-id" {
				t.Fatalf("expected grades item to use resolved group id, got %s", item.GroupID)
			}
		case "associate_professor":
			if item.GroupID != "db-teacher-id" {
				t.Fatalf("expected teacher title item to use resolved group id, got %s", item.GroupID)
			}
		}
	}
}

func TestSeedActivityCategoryTemplatesProvideWorkspaceDefaults(t *testing.T) {
	rows := seedActivityCategoryTemplates()
	if len(rows) != 2 {
		t.Fatalf("expected 2 activity category seed rows, got %d", len(rows))
	}

	keys := map[string]activityCategoryTemplateSeedRow{}
	for _, row := range rows {
		keys[row.Key] = row
	}
	for _, key := range []string{"general", "innovation"} {
		row, ok := keys[key]
		if !ok || !row.Enabled {
			t.Fatalf("expected enabled %s category seed, got %#v", key, keys)
		}
	}

	fields := seedActivityFieldSchemas()
	if len(fields) != 1 {
		t.Fatalf("expected 1 activity field schema seed row, got %d", len(fields))
	}
	field := fields[0]
	if field.CategoryID != "actcat-innovation-01" || field.FieldKey != "location" || field.FieldType != "text" || !field.Required {
		t.Fatalf("unexpected innovation field schema seed: %#v", field)
	}
}

func TestSeederRunIsIdempotentAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx := context.Background()
	db, err := Open(ctx, Options{DSN: dsn})
	if err != nil {
		t.Skipf("test database unavailable: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := NewMigrator("../../../migrations").Apply(ctx, sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	seeder := NewSeeder()
	if err := seeder.Run(ctx, db); err != nil {
		t.Fatalf("first seed run: %v", err)
	}
	if err := seeder.Run(ctx, db); err != nil {
		t.Fatalf("second seed run: %v", err)
	}

	assertCount(t, db, `select count(*) from users where username = 'admin01' and user_type = 'admin'`, 1)
	assertCount(t, db, `select count(*) from option_groups where key in ('grades','academic_years','semesters','teacher_titles')`, 4)
	assertCount(t, db, `select count(*) from option_items oi join option_groups og on og.id = oi.group_id where og.key = 'teacher_titles'`, 4)
}

func assertCount(t *testing.T, db *gorm.DB, query string, expected int64) {
	t.Helper()

	var count int64
	if err := db.Raw(query).Scan(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != expected {
		t.Fatalf("expected count %d, got %d for query %q", expected, count, query)
	}
}
