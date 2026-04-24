package e2e

import (
	"net/http"
	"testing"
)

func TestAuthFlow_LoginRefreshLogoutInvalidatesRefreshToken(t *testing.T) {
	h := newHarness(t)

	auth := login(t, h, "student01")

	refreshRecorder := h.requestJSON(http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{
		"refresh_token": auth.RefreshToken,
	})
	if refreshRecorder.Code != http.StatusOK {
		t.Fatalf("refresh failed: status=%d body=%s", refreshRecorder.Code, refreshRecorder.Body.String())
	}

	logoutRecorder := h.requestJSON(http.MethodPost, "/api/v1/auth/logout", auth.AccessToken, map[string]string{})
	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("logout failed: status=%d body=%s", logoutRecorder.Code, logoutRecorder.Body.String())
	}

	refreshAfterLogout := h.requestJSON(http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{
		"refresh_token": auth.RefreshToken,
	})
	if refreshAfterLogout.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked refresh token to fail: status=%d body=%s", refreshAfterLogout.Code, refreshAfterLogout.Body.String())
	}
}
