package bootstrap

import (
	"errors"
	"net/http"
	"time"

	"credit-management/backend/internal/httpapi/router"
	authhttp "credit-management/backend/internal/modules/auth/http"
	authservice "credit-management/backend/internal/modules/auth/service"
	optionshttp "credit-management/backend/internal/modules/options/http"
	usershttp "credit-management/backend/internal/modules/users/http"
	userservice "credit-management/backend/internal/modules/users/service"
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
	store, err := authservice.NewMemoryStore()
	if err != nil {
		return nil, err
	}
	authSvc := authservice.New(store, store)
	profileSvc := userservice.NewProfileService(store, store)
	authHandler := authhttp.NewHandler(authSvc)
	meHandler := usershttp.NewMeHandler(profileSvc)
	optionsHandler := optionshttp.NewHandler()

	gin.SetMode(cfg.GinMode)
	engine := router.New(router.Dependencies{
		Logger:         logger,
		AuthHandler:    authHandler,
		MeHandler:      meHandler,
		OptionsHandler: optionsHandler,
	})

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
