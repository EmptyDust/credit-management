package service

import (
	"context"
	"testing"
	"time"

	"credit-management/backend/internal/platform/authn"
)

func TestLoginReturnsTokensForActiveUser(t *testing.T) {
	authn.Configure(authn.Config{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     24 * time.Hour,
		RefreshTTL:    7 * 24 * time.Hour,
	})

	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("expected in-memory store setup to succeed, got %v", err)
	}

	svc := New(store, store)

	result, err := svc.Login(context.Background(), LoginInput{
		LoginID:  "student01",
		Password: "Passw0rd!",
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("expected access token to be populated")
	}

	if result.RefreshToken == "" {
		t.Fatal("expected refresh token to be populated")
	}

	claims, err := authn.ParseAccessToken(result.AccessToken)
	if err != nil {
		t.Fatalf("expected access token to parse, got %v", err)
	}

	if claims.UserID != result.User.ID {
		t.Fatalf("expected access token user_id %s, got %s", result.User.ID, claims.UserID)
	}
}
