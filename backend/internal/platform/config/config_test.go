package config

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("BACKEND_HTTP_PORT", "")
	t.Setenv("PORT", "")

	cfg := Load()
	if cfg.HTTPPort != "8082" {
		t.Fatalf("expected default port 8082, got %s", cfg.HTTPPort)
	}
}
