package service

import (
	"context"
	"testing"

	authservice "credit-management/backend/internal/modules/auth/service"
)

func TestStudentImportPrecheckRejectsDuplicateStudentID(t *testing.T) {
	store, err := authservice.NewMemoryStore()
	if err != nil {
		t.Fatalf("expected in-memory store setup to succeed, got %v", err)
	}

	svc := NewUserImportService(store)

	result, err := svc.PrecheckStudents(context.Background(), "01JADMIN000000000000000000", "students.csv", "hash-1", []UserRow{
		{
			Username:     "student-a",
			Password:     "Passw0rd!",
			RealName:     "Student A",
			StudentID:    "20240088",
			Grade:        "2024",
			DepartmentID: "class-01",
		},
		{
			Username:     "student-b",
			Password:     "Passw0rd!",
			RealName:     "Student B",
			StudentID:    "20240088",
			Grade:        "2024",
			DepartmentID: "class-01",
		},
	})
	if err != nil {
		t.Fatalf("expected precheck to succeed, got %v", err)
	}

	if result.Summary.InvalidRows != 1 {
		t.Fatalf("expected 1 invalid row, got %d", result.Summary.InvalidRows)
	}

	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 row results, got %d", len(result.Rows))
	}

	if result.Rows[1].Status != "invalid" {
		t.Fatalf("expected second row invalid, got %s", result.Rows[1].Status)
	}

	if len(result.Rows[1].Errors) == 0 || result.Rows[1].Errors[0].Code != "DUPLICATE_STUDENT_ID" {
		t.Fatalf("expected duplicate student id error, got %+v", result.Rows[1].Errors)
	}
}
