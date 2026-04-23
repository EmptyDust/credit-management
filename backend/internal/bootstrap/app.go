package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"time"

	"credit-management/backend/internal/httpapi/router"
	activityhttp "credit-management/backend/internal/modules/activities/http"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	activityservice "credit-management/backend/internal/modules/activities/service"
	apphttp "credit-management/backend/internal/modules/applications/http"
	appservice "credit-management/backend/internal/modules/applications/service"
	auditservice "credit-management/backend/internal/modules/audit/service"
	authhttp "credit-management/backend/internal/modules/auth/http"
	authservice "credit-management/backend/internal/modules/auth/service"
	dashboard "credit-management/backend/internal/modules/dashboard"
	dashboardhttp "credit-management/backend/internal/modules/dashboard/http"
	devtoolshttp "credit-management/backend/internal/modules/devtools/http"
	exportservice "credit-management/backend/internal/modules/exports/service"
	filehttp "credit-management/backend/internal/modules/files/http"
	fileservice "credit-management/backend/internal/modules/files/service"
	importhttp "credit-management/backend/internal/modules/imports/http"
	importservice "credit-management/backend/internal/modules/imports/service"
	optionshttp "credit-management/backend/internal/modules/options/http"
	participanthttp "credit-management/backend/internal/modules/participants/http"
	participantservice "credit-management/backend/internal/modules/participants/service"
	usershttp "credit-management/backend/internal/modules/users/http"
	usersrepo "credit-management/backend/internal/modules/users/repo"
	userservice "credit-management/backend/internal/modules/users/service"
	"credit-management/backend/internal/platform/auditlog"
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
	participantSvc := participantservice.New(activityRepo, store)
	fileSvc := fileservice.New(activityRepo)
	appSvc := appservice.New()
	appSvc.SetActivityRepository(activityRepo)
	appSvc.SetFileService(fileSvc)
	appSvc.SetParticipantService(participantSvc)
	activitySvc := activityservice.New(activityRepo, appservice.NewHook(appSvc))
	auditWriter := auditlog.NewWriter()
	auditSvc := auditservice.New(auditWriter)
	dashboardSvc := dashboard.NewService(
		func() int {
			users, _ := store.ListUsers(context.Background(), usersrepo.ListFilter{})
			return len(users)
		},
		func() int {
			activities, _ := activityRepo.List(context.Background(), activityrepo.ListFilter{})
			return len(activities)
		},
		appSvc.Count,
	)
	authHandler := authhttp.NewHandler(authSvc)
	activityHandler := activityhttp.NewHandler(activitySvc)
	appHandler := apphttp.NewHandler(appSvc)
	meHandler := usershttp.NewMeHandler(profileSvc)
	adminHandler := usershttp.NewAdminHandler(adminSvc, exportSvc)
	dashboardHandler := dashboardhttp.NewHandler(dashboardSvc)
	devtoolsHandler := devtoolshttp.NewHandler(auditSvc)
	fileHandler := filehttp.NewHandler(fileSvc)
	importHandler := importhttp.NewUserImportHandler(importSvc)
	optionsHandler := optionshttp.NewHandler()
	participantHandler := participanthttp.NewHandler(participantSvc)

	gin.SetMode(cfg.GinMode)
	engine := router.New(router.Dependencies{
		Logger:             logger,
		AuthHandler:        authHandler,
		ActivityHandler:    activityHandler,
		ApplicationHandler: appHandler,
		AdminHandler:       adminHandler,
		DashboardHandler:   dashboardHandler,
		DevtoolsHandler:    devtoolsHandler,
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
