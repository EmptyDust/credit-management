package database

import (
	"context"
	"fmt"

	authservice "credit-management/backend/internal/modules/auth/service"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultAdminID       = "01JADMIN0000000000000000000"
	defaultAdminUsername = "admin01"
	defaultAdminPassword = "Passw0rd!"
)

type Seeder struct{}

func NewSeeder() Seeder {
	return Seeder{}
}

func (Seeder) Run(ctx context.Context, db *gorm.DB) error {
	passwordHash, err := authservice.HashPassword(defaultAdminPassword)
	if err != nil {
		return fmt.Errorf("hash default admin password: %w", err)
	}

	return WithTransaction(ctx, db, func(tx *gorm.DB) error {
		if err := insertSeedRows(tx, seedUsers(passwordHash)); err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
		if err := insertSeedRows(tx, seedOrganizationNodes()); err != nil {
			return fmt.Errorf("seed organization nodes: %w", err)
		}
		if err := insertSeedRows(tx, seedOptionGroups()); err != nil {
			return fmt.Errorf("seed option groups: %w", err)
		}
		if err := insertSeedRows(tx, seedOptionItems()); err != nil {
			return fmt.Errorf("seed option items: %w", err)
		}
		if err := insertSeedRows(tx, seedActivityCategoryTemplates()); err != nil {
			return fmt.Errorf("seed activity category templates: %w", err)
		}

		return nil
	})
}

func insertSeedRows[T any](tx *gorm.DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}

	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

type userSeedRow struct {
	ID                  string `gorm:"column:id"`
	Username            string `gorm:"column:username"`
	PasswordHash        string `gorm:"column:password_hash"`
	Email               string `gorm:"column:email"`
	Phone               string `gorm:"column:phone"`
	RealName            string `gorm:"column:real_name"`
	UserType            string `gorm:"column:user_type"`
	Status              string `gorm:"column:status"`
	RefreshTokenVersion int64  `gorm:"column:refresh_token_version"`
}

func (userSeedRow) TableName() string {
	return "users"
}

type organizationNodeSeedRow struct {
	ID        string  `gorm:"column:id"`
	ParentID  *string `gorm:"column:parent_id"`
	NodeType  string  `gorm:"column:node_type"`
	Name      string  `gorm:"column:name"`
	Code      string  `gorm:"column:code"`
	Path      string  `gorm:"column:path"`
	Level     int     `gorm:"column:level"`
	SortOrder int     `gorm:"column:sort_order"`
}

func (organizationNodeSeedRow) TableName() string {
	return "organization_nodes"
}

type optionGroupSeedRow struct {
	ID      string `gorm:"column:id"`
	Key     string `gorm:"column:key"`
	Label   string `gorm:"column:label"`
	Scope   string `gorm:"column:scope"`
	Version int    `gorm:"column:version"`
}

func (optionGroupSeedRow) TableName() string {
	return "option_groups"
}

type optionItemSeedRow struct {
	ID        string `gorm:"column:id"`
	GroupID   string `gorm:"column:group_id"`
	Value     string `gorm:"column:value"`
	Label     string `gorm:"column:label"`
	SortOrder int    `gorm:"column:sort_order"`
	Enabled   bool   `gorm:"column:enabled"`
}

func (optionItemSeedRow) TableName() string {
	return "option_items"
}

type activityCategoryTemplateSeedRow struct {
	ID        string `gorm:"column:id"`
	Key       string `gorm:"column:key"`
	Label     string `gorm:"column:label"`
	SortOrder int    `gorm:"column:sort_order"`
	Enabled   bool   `gorm:"column:enabled"`
}

func (activityCategoryTemplateSeedRow) TableName() string {
	return "activity_category_templates"
}

func seedUsers(passwordHash string) []userSeedRow {
	return []userSeedRow{
		{
			ID:                  defaultAdminID,
			Username:            defaultAdminUsername,
			PasswordHash:        passwordHash,
			Email:               "admin01@example.com",
			Phone:               "13800138002",
			RealName:            "Admin One",
			UserType:            "admin",
			Status:              "active",
			RefreshTokenVersion: 1,
		},
	}
}

func seedOrganizationNodes() []organizationNodeSeedRow {
	schoolID := "school-01"
	collegeID := "college-01"
	majorID := "major-01"

	return []organizationNodeSeedRow{
		{
			ID:        schoolID,
			NodeType:  "school",
			Name:      "Demo University",
			Code:      "DEMO",
			Path:      schoolID,
			Level:     1,
			SortOrder: 1,
		},
		{
			ID:        collegeID,
			ParentID:  stringPtr(schoolID),
			NodeType:  "college",
			Name:      "Engineering College",
			Code:      "ENG",
			Path:      schoolID + "/" + collegeID,
			Level:     2,
			SortOrder: 1,
		},
		{
			ID:        majorID,
			ParentID:  stringPtr(collegeID),
			NodeType:  "major",
			Name:      "Software Engineering",
			Code:      "SE",
			Path:      schoolID + "/" + collegeID + "/" + majorID,
			Level:     3,
			SortOrder: 1,
		},
		{
			ID:        "class-01",
			ParentID:  stringPtr(majorID),
			NodeType:  "class",
			Name:      "2024 Class 1",
			Code:      "SE2401",
			Path:      schoolID + "/" + collegeID + "/" + majorID + "/class-01",
			Level:     4,
			SortOrder: 1,
		},
	}
}

func seedOptionGroups() []optionGroupSeedRow {
	return []optionGroupSeedRow{
		{ID: "optgrp-grades-01", Key: "grades", Label: "Grades", Scope: "system", Version: 1},
		{ID: "optgrp-acyears-01", Key: "academic_years", Label: "Academic Years", Scope: "system", Version: 1},
		{ID: "optgrp-semesters01", Key: "semesters", Label: "Semesters", Scope: "system", Version: 1},
		{ID: "optgrp-ttitles-01", Key: "teacher_titles", Label: "Teacher Titles", Scope: "system", Version: 1},
	}
}

func seedOptionItems() []optionItemSeedRow {
	return []optionItemSeedRow{
		{ID: "grade-2024", GroupID: "optgrp-grades-01", Value: "2024", Label: "2024", SortOrder: 1, Enabled: true},
		{ID: "grade-2025", GroupID: "optgrp-grades-01", Value: "2025", Label: "2025", SortOrder: 2, Enabled: true},
		{ID: "grade-2026", GroupID: "optgrp-grades-01", Value: "2026", Label: "2026", SortOrder: 3, Enabled: true},
		{ID: "acyear-2024-2025", GroupID: "optgrp-acyears-01", Value: "2024-2025", Label: "2024-2025", SortOrder: 1, Enabled: true},
		{ID: "acyear-2025-2026", GroupID: "optgrp-acyears-01", Value: "2025-2026", Label: "2025-2026", SortOrder: 2, Enabled: true},
		{ID: "acyear-2026-2027", GroupID: "optgrp-acyears-01", Value: "2026-2027", Label: "2026-2027", SortOrder: 3, Enabled: true},
		{ID: "semester-spring", GroupID: "optgrp-semesters01", Value: "spring", Label: "Spring", SortOrder: 1, Enabled: true},
		{ID: "semester-autumn", GroupID: "optgrp-semesters01", Value: "autumn", Label: "Autumn", SortOrder: 2, Enabled: true},
		{ID: "ttitle-assistant", GroupID: "optgrp-ttitles-01", Value: "assistant", Label: "Assistant", SortOrder: 1, Enabled: true},
		{ID: "ttitle-lecturer", GroupID: "optgrp-ttitles-01", Value: "lecturer", Label: "Lecturer", SortOrder: 2, Enabled: true},
		{ID: "ttitle-assoc-prof", GroupID: "optgrp-ttitles-01", Value: "associate_professor", Label: "Associate Professor", SortOrder: 3, Enabled: true},
		{ID: "ttitle-professor", GroupID: "optgrp-ttitles-01", Value: "professor", Label: "Professor", SortOrder: 4, Enabled: true},
	}
}

func seedActivityCategoryTemplates() []activityCategoryTemplateSeedRow {
	return []activityCategoryTemplateSeedRow{
		{
			ID:        "actcat-general-01",
			Key:       "general",
			Label:     "General",
			SortOrder: 1,
			Enabled:   true,
		},
	}
}

func stringPtr(value string) *string {
	return &value
}
