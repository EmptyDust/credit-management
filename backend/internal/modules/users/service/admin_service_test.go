package service

import (
	"context"
	"testing"

	authservice "credit-management/backend/internal/modules/auth/service"
	usersrepo "credit-management/backend/internal/modules/users/repo"
)

func TestListStudentsReturnsOnlyStudents(t *testing.T) {
	store, err := authservice.NewMemoryStore()
	if err != nil {
		t.Fatalf("expected in-memory store setup to succeed, got %v", err)
	}

	svc := NewAdminService(store, store)

	result, err := svc.ListStudents(context.Background(), usersrepo.ListFilter{})
	if err != nil {
		t.Fatalf("expected list students to succeed, got %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected exactly one seeded student, got %d", len(result.Items))
	}

	if result.Items[0].UserType != "student" {
		t.Fatalf("expected seeded user type student, got %s", result.Items[0].UserType)
	}
}

func TestListStudentsPaginatesResults(t *testing.T) {
	store, err := authservice.NewMemoryStore()
	if err != nil {
		t.Fatalf("expected in-memory store setup to succeed, got %v", err)
	}

	svc := NewAdminService(store, store)
	for _, username := range []string{"student02", "student03", "student04"} {
		if _, err := svc.CreateStudent(context.Background(), usersrepo.CreateUserInput{
			Username:  username,
			Password:  "Passw0rd!",
			RealName:  username,
			StudentID: username,
			Grade:     "2024",
		}); err != nil {
			t.Fatalf("create student %s: %v", username, err)
		}
	}

	result, err := svc.ListStudents(context.Background(), usersrepo.ListFilter{
		Page:     2,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("expected list students to succeed, got %v", err)
	}

	if result.Total != 4 {
		t.Fatalf("expected total 4, got %d", result.Total)
	}
	if result.Page != 2 || result.PageSize != 2 || result.TotalPages != 2 {
		t.Fatalf("unexpected pagination metadata: %#v", result)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected second page to contain 2 students, got %d", len(result.Items))
	}
}

func TestUserStatsAggregatesStatusAndRoleSpecificFields(t *testing.T) {
	store, err := authservice.NewMemoryStore()
	if err != nil {
		t.Fatalf("expected in-memory store setup to succeed, got %v", err)
	}

	svc := NewAdminService(store, store)
	student, err := svc.CreateStudent(context.Background(), usersrepo.CreateUserInput{
		Username:  "student_disabled",
		Password:  "Passw0rd!",
		RealName:  "Student Disabled",
		StudentID: "20249998",
		Grade:     "2025",
	})
	if err != nil {
		t.Fatalf("create disabled student: %v", err)
	}
	if _, err := svc.UpdateStatus(context.Background(), student.ID, "disabled"); err != nil {
		t.Fatalf("disable student: %v", err)
	}
	if _, err := svc.CreateTeacher(context.Background(), usersrepo.CreateUserInput{
		Username:  "teacher_professor",
		Password:  "Passw0rd!",
		RealName:  "Teacher Professor",
		TeacherID: "T2024999",
		Title:     "professor",
	}); err != nil {
		t.Fatalf("create teacher: %v", err)
	}

	studentStats, err := svc.StudentStats(context.Background(), usersrepo.ListFilter{})
	if err != nil {
		t.Fatalf("student stats: %v", err)
	}
	if studentStats.Total != 2 || studentStats.StatusCounts["active"] != 1 || studentStats.StatusCounts["disabled"] != 1 {
		t.Fatalf("unexpected student status stats: %#v", studentStats)
	}
	if countFor(studentStats.GradeCounts, "2024") != 1 || countFor(studentStats.GradeCounts, "2025") != 1 {
		t.Fatalf("unexpected student grade stats: %#v", studentStats.GradeCounts)
	}

	teacherStats, err := svc.TeacherStats(context.Background(), usersrepo.ListFilter{})
	if err != nil {
		t.Fatalf("teacher stats: %v", err)
	}
	if teacherStats.Total != 2 || teacherStats.StatusCounts["active"] != 2 {
		t.Fatalf("unexpected teacher status stats: %#v", teacherStats)
	}
	if countFor(teacherStats.TitleCounts, "lecturer") != 1 || countFor(teacherStats.TitleCounts, "professor") != 1 {
		t.Fatalf("unexpected teacher title stats: %#v", teacherStats.TitleCounts)
	}
}

func countFor(items []KeyCount, key string) int {
	for _, item := range items {
		if item.Key == key {
			return item.Count
		}
	}
	return 0
}
