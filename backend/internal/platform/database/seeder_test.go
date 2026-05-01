package database

import (
	"context"
	"testing"

	authservice "credit-management/backend/internal/modules/auth/service"
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

	items := seedOptionItems()
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

func TestSeedActivityCategoryTemplatesProvideMinimalDefault(t *testing.T) {
	rows := seedActivityCategoryTemplates()
	if len(rows) != 1 {
		t.Fatalf("expected 1 activity category seed row, got %d", len(rows))
	}
	if rows[0].Key != "general" || !rows[0].Enabled {
		t.Fatalf("unexpected activity category seed: %#v", rows[0])
	}
}
