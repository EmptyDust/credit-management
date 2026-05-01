package config

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("BACKEND_HTTP_PORT", "")
	t.Setenv("PORT", "")
	t.Setenv("GIN_MODE", "")
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_DB", "")
	t.Setenv("STORAGE_BACKEND", "")
	t.Setenv("STORAGE_LOCAL_ROOT", "")
	t.Setenv("STORAGE_PUBLIC_URL", "")
	t.Setenv("BUILD_VERSION", "")

	cfg := Load()
	if cfg.HTTPPort != "8082" {
		t.Fatalf("expected default port 8082, got %s", cfg.HTTPPort)
	}
	if cfg.GinMode != "release" {
		t.Fatalf("expected default gin mode release, got %s", cfg.GinMode)
	}
	if cfg.RedisDB != 0 {
		t.Fatalf("expected default redis db 0, got %d", cfg.RedisDB)
	}
	if cfg.StorageBackend != "local" {
		t.Fatalf("expected default storage backend local, got %s", cfg.StorageBackend)
	}
	if cfg.StorageLocalRoot != "./tmp/storage" {
		t.Fatalf("expected default storage root ./tmp/storage, got %s", cfg.StorageLocalRoot)
	}
	if cfg.BuildVersion != "dev" {
		t.Fatalf("expected default build version dev, got %s", cfg.BuildVersion)
	}
}

func TestLoadConfigReadsRuntimeSettings(t *testing.T) {
	t.Setenv("BACKEND_HTTP_PORT", "9090")
	t.Setenv("PORT", "3000")
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("DATABASE_DSN", "postgres://postgres:password@localhost:5433/credit_management?sslmode=disable")
	t.Setenv("REDIS_ADDR", "127.0.0.1:6380")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "4")
	t.Setenv("STORAGE_BACKEND", "object")
	t.Setenv("STORAGE_LOCAL_ROOT", "/var/lib/credit-management")
	t.Setenv("STORAGE_PUBLIC_URL", "https://cdn.example.com/files")
	t.Setenv("BUILD_VERSION", "1.2.3")

	cfg := Load()
	if cfg.HTTPPort != "9090" {
		t.Fatalf("expected explicit backend port 9090, got %s", cfg.HTTPPort)
	}
	if cfg.GinMode != "debug" {
		t.Fatalf("expected gin mode debug, got %s", cfg.GinMode)
	}
	if cfg.DatabaseDSN != "postgres://postgres:password@localhost:5433/credit_management?sslmode=disable" {
		t.Fatalf("expected database dsn to be loaded, got %s", cfg.DatabaseDSN)
	}
	if cfg.RedisAddr != "127.0.0.1:6380" {
		t.Fatalf("expected redis addr 127.0.0.1:6380, got %s", cfg.RedisAddr)
	}
	if cfg.RedisPassword != "secret" {
		t.Fatalf("expected redis password secret, got %s", cfg.RedisPassword)
	}
	if cfg.RedisDB != 4 {
		t.Fatalf("expected redis db 4, got %d", cfg.RedisDB)
	}
	if cfg.StorageBackend != "object" {
		t.Fatalf("expected storage backend object, got %s", cfg.StorageBackend)
	}
	if cfg.StorageLocalRoot != "/var/lib/credit-management" {
		t.Fatalf("expected storage root /var/lib/credit-management, got %s", cfg.StorageLocalRoot)
	}
	if cfg.StoragePublicURL != "https://cdn.example.com/files" {
		t.Fatalf("expected storage public url https://cdn.example.com/files, got %s", cfg.StoragePublicURL)
	}
	if cfg.BuildVersion != "1.2.3" {
		t.Fatalf("expected build version 1.2.3, got %s", cfg.BuildVersion)
	}
}

func TestLoadConfigFallsBackForInvalidRedisDB(t *testing.T) {
	t.Setenv("REDIS_DB", "not-a-number")

	cfg := Load()
	if cfg.RedisDB != 0 {
		t.Fatalf("expected invalid redis db to fall back to 0, got %d", cfg.RedisDB)
	}
}
