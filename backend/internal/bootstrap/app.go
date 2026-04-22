package bootstrap

import (
	"errors"
	"net/http"
	"time"

	"credit-management/backend/internal/httpapi/router"
	"credit-management/backend/internal/platform/config"
	platformlogger "credit-management/backend/internal/platform/logger"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	server *http.Server
}

func NewApp() (*App, error) {
	cfg := config.Load()
	logger := platformlogger.New(cfg)

	gin.SetMode(cfg.GinMode)
	engine := router.New(logger)

	return &App{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:              cfg.Address(),
			Handler:           engine,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("starting backend server", "addr", a.server.Addr)

	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
