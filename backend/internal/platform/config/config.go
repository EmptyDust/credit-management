package config

import "os"

type Config struct {
	HTTPPort string
	GinMode  string
}

func Load() Config {
	return Config{
		HTTPPort: firstNonEmpty(os.Getenv("BACKEND_HTTP_PORT"), os.Getenv("PORT"), "8082"),
		GinMode:  firstNonEmpty(os.Getenv("GIN_MODE"), "release"),
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
