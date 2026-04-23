package bootstrap

import (
	"errors"
	"net/http"
	"time"

	"credit-management/backend/internal/httpapi/router"
	activityhttp "credit-management/backend/internal/modules/activities/http"
	activityservice "credit-management/backend/internal/modules/activities/service"
	authhttp "credit-management/backend/internal/modules/auth/http"
	authservice "credit-management/backend/internal/modules/auth/service"
	exportservice "credit-management/backend/internal/modules/exports/service"
	filehttp "credit-management/backend/internal/modules/files/http"
	fileservice "credit-management/backend/internal/modules/files/service"
	importhttp "credit-management/backend/internal/modules/imports/http"
	importservice "credit-management/backend/internal/modules/imports/service"
	optionshttp "credit-management/backend/internal/modules/options/http"
	participanthttp "credit-management/backend/internal/modules/participants/http"
	participantservice "credit-management/backend/internal/modules/participants/service"
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
	adminSvc := userservice.NewAdminService(store, store)
	exportSvc := exportservice.NewUserExportService(store)
	importSvc := importservice.NewUserImportService(store)
	activityRepo := activityservice.NewMemoryRepository()
	activitySvc := activityservice.New(activityRepo, nil)
	participantSvc := participantservice.New(activityRepo, store)
	fileSvc := fileservice.New(activityRepo)
	authHandler := authhttp.NewHandler(authSvc)
	activityHandler := activityhttp.NewHandler(activitySvc)
	meHandler := usershttp.NewMeHandler(profileSvc)
	adminHandler := usershttp.NewAdminHandler(adminSvc, exportSvc)
	fileHandler := filehttp.NewHandler(fileSvc)
	importHandler := importhttp.NewUserImportHandler(importSvc)
	optionsHandler := optionshttp.NewHandler()
	participantHandler := participanthttp.NewHandler(participantSvc)

	gin.SetMode(cfg.GinMode)
	engine := router.New(router.Dependencies{
		Logger:             logger,
		AuthHandler:        authHandler,
		ActivityHandler:    activityHandler,
		AdminHandler:       adminHandler,
		FileHandler:        fileHandler,
		ImportHandler:      importHandler,
		MeHandler:          meHandler,
		OptionsHandler:     optionsHandler,
		ParticipantHandler: participantHandler,
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
