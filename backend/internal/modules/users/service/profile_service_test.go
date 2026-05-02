package service

import (
	"context"
	"testing"

	authservice "credit-management/backend/internal/modules/auth/service"
)

func TestUploadAvatarReturnsDownloadURL(t *testing.T) {
	store, err := authservice.NewMemoryStore()
	if err != nil {
		t.Fatalf("expected in-memory store setup to succeed, got %v", err)
	}

	svc := NewProfileService(store, store)
	result, err := svc.UploadAvatar(context.Background(), "01JSTUDENT0000000000000000", "avatar.png")
	if err != nil {
		t.Fatalf("upload avatar: %v", err)
	}

	if result.FileID == "" {
		t.Fatal("expected generated file id")
	}
	if result.AvatarURL != "/api/v1/files/"+result.FileID+"/download" {
		t.Fatalf("expected download url, got %s", result.AvatarURL)
	}
}
