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
