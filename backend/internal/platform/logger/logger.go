package logger

import (
	"log/slog"
	"os"

	"credit-management/backend/internal/platform/config"
)

func New(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	if cfg.GinMode == "debug" {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
