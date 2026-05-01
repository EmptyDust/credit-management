package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort         string
	GinMode          string
	DatabaseDSN      string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	StorageBackend   string
	StorageLocalRoot string
	StoragePublicURL string
	BuildVersion     string
}

func Load() Config {
	return Config{
		HTTPPort:         firstNonEmpty(os.Getenv("BACKEND_HTTP_PORT"), os.Getenv("PORT"), "8082"),
		GinMode:          firstNonEmpty(os.Getenv("GIN_MODE"), "release"),
		DatabaseDSN:      os.Getenv("DATABASE_DSN"),
		RedisAddr:        firstNonEmpty(os.Getenv("REDIS_ADDR"), "127.0.0.1:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		RedisDB:          envInt("REDIS_DB", 0),
		StorageBackend:   firstNonEmpty(os.Getenv("STORAGE_BACKEND"), "local"),
		StorageLocalRoot: firstNonEmpty(os.Getenv("STORAGE_LOCAL_ROOT"), "./tmp/storage"),
		StoragePublicURL: os.Getenv("STORAGE_PUBLIC_URL"),
		BuildVersion:     firstNonEmpty(os.Getenv("BUILD_VERSION"), "dev"),
	}
}

func (c Config) Address() string {
	return ":" + c.HTTPPort
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
