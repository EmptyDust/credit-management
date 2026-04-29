package authn

import (
	"testing"
	"time"
)

func TestIssueTokenContainsUserIDClaim(t *testing.T) {
	Configure(Config{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     24 * time.Hour,
		RefreshTTL:    7 * 24 * time.Hour,
	})

	token, err := IssueAccessToken(ClaimsInput{
		UserID:    "01JUSER",
		Username:  "alice",
		UserType:  "student",
		SessionID: "01JSESSION",
	})
	if err != nil {
		t.Fatalf("expected no error issuing token, got %v", err)
	}

	claims, err := ParseAccessToken(token)
	if err != nil {
		t.Fatalf("expected no error parsing token, got %v", err)
	}

	if claims.UserID != "01JUSER" {
		t.Fatalf("expected user_id claim 01JUSER, got %s", claims.UserID)
	}
}
