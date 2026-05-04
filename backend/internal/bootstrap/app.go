package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"credit-management/backend/internal/httpapi/router"
	activityhttp "credit-management/backend/internal/modules/activities/http"
	activityrepo "credit-management/backend/internal/modules/activities/repo"
	activityservice "credit-management/backend/internal/modules/activities/service"
	apphttp "credit-management/backend/internal/modules/applications/http"
	appservice "credit-management/backend/internal/modules/applications/service"
	auditservice "credit-management/backend/internal/modules/audit/service"
	authhttp "credit-management/backend/internal/modules/auth/http"
	authrepo "credit-management/backend/internal/modules/auth/repo"
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
	optionsrepo "credit-management/backend/internal/modules/options/repo"
	optionsservice "credit-management/backend/internal/modules/options/service"
	participanthttp "credit-management/backend/internal/modules/participants/http"
	participantservice "credit-management/backend/internal/modules/participants/service"
	usershttp "credit-management/backend/internal/modules/users/http"
	usersrepo "credit-management/backend/internal/modules/users/repo"
	userservice "credit-management/backend/internal/modules/users/service"
	"credit-management/backend/internal/platform/auditlog"
	"credit-management/backend/internal/platform/config"
	"credit-management/backend/internal/platform/database"
	"credit-management/backend/internal/platform/health"
	platformlogger "credit-management/backend/internal/platform/logger"
	platformredis "credit-management/backend/internal/platform/redis"
	"credit-management/backend/internal/platform/storage"

	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	server *http.Server
}

func NewApp() (*App, error) {
	cfg := config.Load()
	logger := platformlogger.New(cfg)

	ctx := context.Background()
	db, err := database.Open(ctx, database.Options{DSN: cfg.DatabaseDSN})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := database.NewMigrator(resolveMigrationsDir()).Apply(ctx, sqlDB); err != nil {
		return nil, err
	}
	if err := database.NewSeeder().Run(ctx, db); err != nil {
		return nil, err
	}

	redisClient, err := platformredis.Open(ctx, platformredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		return nil, err
	}

	localStore := storage.NewLocalStore(cfg.StorageLocalRoot, cfg.StoragePublicURL)
	if err := localStore.HealthCheck(ctx); err != nil {
		return nil, err
	}

	optionHandler := optionshttp.NewHandler(
		optionsservice.NewService(optionsrepo.NewGorm(db)),
	)

	userRepo := usersrepo.NewGorm(db)
	sessionRepo := authrepo.NewGorm(db, redisClient)

	authSvc := authservice.New(userRepo, sessionRepo)
	profileSvc := userservice.NewProfileService(userRepo, sessionRepo)
	adminSvc := userservice.NewAdminService(userRepo, sessionRepo)
	exportSvc := exportservice.NewUserExportService(userRepo)
	importSvc := importservice.NewUserImportService(userRepo)

	return newConfiguredApp(cfg, logger, configuredAppDeps{
		authSvc:     authSvc,
		profileSvc:  profileSvc,
		adminSvc:    adminSvc,
		exportSvc:   exportSvc,
		importSvc:   importSvc,
		db:          db,
		auditWriter: auditlog.NewGormWriter(db),
		healthChecker: health.NewChecker("dev", map[string]health.CheckFunc{
			"database": func(context.Context) error {
				return sqlDB.PingContext(ctx)
			},
			"redis": func(context.Context) error {
				return redisClient.Ping(ctx).Err()
			},
			"storage": localStore.HealthCheck,
		}),
		storageStore:         localStore,
		optionsHandler:       optionHandler,
		activityRepo:         activityrepo.NewGorm(db),
		activityCategoryRepo: activityrepo.NewCategorySchemaGorm(db),
		participantRepo:      participantservice.NewGormRepository(db),
		applicationRepo:      appservice.NewGormRepository(db),
		userRepo:             userRepo,
		userCount: func() int {
			users, _ := userRepo.ListUsers(context.Background(), usersrepo.ListFilter{})
			return len(users)
		},
	})
}

func NewTestApp() (*App, error) {
	cfg := config.Load()
	cfg.GinMode = gin.TestMode
	logger := platformlogger.New(cfg)

	optionHandler := optionshttp.NewHandler(
		optionsservice.NewService(memoryOptionRepo{}),
	)

	return newMemoryBackedApp(cfg, logger, optionHandler)
}

func newMemoryBackedApp(cfg config.Config, logger *slog.Logger, optionHandler *optionshttp.Handler) (*App, error) {
	authStore, err := authservice.NewMemoryStore()
	if err != nil {
		return nil, err
	}
	authSvc := authservice.New(authStore, authStore)
	profileSvc := userservice.NewProfileService(authStore, authStore)
	adminSvc := userservice.NewAdminService(authStore, authStore)
	exportSvc := exportservice.NewUserExportService(authStore)
	importSvc := importservice.NewUserImportService(authStore)

	return newConfiguredApp(cfg, logger, configuredAppDeps{
		authSvc:        authSvc,
		profileSvc:     profileSvc,
		adminSvc:       adminSvc,
		exportSvc:      exportSvc,
		importSvc:      importSvc,
		optionsHandler: optionHandler,
		userRepo:       authStore,
		userCount: func() int {
			users, _ := authStore.ListUsers(context.Background(), usersrepo.ListFilter{})
			return len(users)
		},
	})
}

type configuredAppDeps struct {
	authSvc              *authservice.Service
	profileSvc           *userservice.Service
	adminSvc             *userservice.AdminService
	exportSvc            *exportservice.UserExportService
	importSvc            *importservice.UserImportService
	optionsHandler       *optionshttp.Handler
	db                   *gorm.DB
	auditWriter          *auditlog.Writer
	healthChecker        *health.Checker
	storageStore         storage.Store
	activityRepo         activityrepo.Repository
	activityCategoryRepo activityrepo.CategorySchemaReader
	participantRepo      participantservice.Repository
	applicationRepo      appservice.Repository
	userRepo             usersrepo.UserRepository
	userCount            func() int
}

func newConfiguredApp(cfg config.Config, logger *slog.Logger, deps configuredAppDeps) (*App, error) {
	activityRepo := deps.activityRepo
	if activityRepo == nil {
		activityRepo = activityservice.NewMemoryRepository()
	}
	activityCategoryRepo := deps.activityCategoryRepo
	if activityCategoryRepo == nil {
		activityCategoryRepo = memoryCategorySchemaRepo{}
	}
	participantSvc := participantservice.New(activityRepo, deps.userRepo)
	if deps.participantRepo != nil {
		participantSvc = participantservice.NewWithRepository(activityRepo, deps.userRepo, deps.participantRepo)
	}
	fileSvc := fileservice.New(activityRepo)
	if deps.storageStore != nil && deps.db != nil {
		fileSvc = fileservice.NewWithStore(activityRepo, fileservice.NewGormMetadataRepository(deps.db), deps.storageStore)
	}
	appSvc := appservice.New()
	if deps.applicationRepo != nil {
		appSvc = appservice.NewWithRepository(deps.applicationRepo)
	}
	appSvc.SetActivityRepository(activityRepo)
	appSvc.SetFileService(fileSvc)
	appSvc.SetParticipantService(participantSvc)
	activityStatsSvc := dashboard.NewActivityStatsService(activityRepo, activityCategoryRepo, participantSvc)
	activityExportSvc := exportservice.NewActivityExportService(activityRepo)
	activitySvc := activityservice.New(activityRepo, appservice.NewHook(appSvc))
	auditWriter := deps.auditWriter
	if auditWriter == nil {
		auditWriter = auditlog.NewWriter()
	}
	auditSvc := auditservice.New(auditWriter)
	dashboardSvc := dashboard.NewService(
		deps.userCount,
		func() int {
			activities, _ := activityRepo.List(context.Background(), activityrepo.ListFilter{})
			return len(activities)
		},
		appSvc.Count,
	)
	authHandler := authhttp.NewHandler(deps.authSvc)
	activityHandler := activityhttp.NewHandler(activitySvc, activityCategoryRepo).
		UseStatsService(activityStatsSvc).
		UseExportService(activityExportSvc)
	appHandler := apphttp.NewHandler(appSvc)
	meHandler := usershttp.NewMeHandler(deps.profileSvc)
	adminHandler := usershttp.NewAdminHandler(deps.adminSvc, deps.exportSvc)
	dashboardHandler := dashboardhttp.NewHandler(dashboardSvc)
	devtoolsHandler := devtoolshttp.NewHandlerWithHealth(auditSvc, logger, deps.healthChecker)
	fileHandler := filehttp.NewHandler(fileSvc)
	importHandler := importhttp.NewUserImportHandler(deps.importSvc)
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
		OptionsHandler:     deps.optionsHandler,
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

func resolveMigrationsDir() string {
	for _, candidate := range []string{"./migrations", "../migrations", "../../migrations"} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return "./migrations"
}

type memoryOptionRepo struct{}

type memoryCategorySchemaRepo struct{}

func (memoryOptionRepo) ListOrganization(context.Context) ([]optionsservice.OrganizationNode, error) {
	return []optionsservice.OrganizationNode{
		{ID: "school-01", Name: "Demo University", NodeType: "school", Path: "school-01", Level: 1, SortOrder: 1},
		{ID: "college-01", ParentID: "school-01", Name: "Engineering College", NodeType: "college", Path: "school-01/college-01", Level: 2, SortOrder: 1},
		{ID: "major-01", ParentID: "college-01", Name: "Software Engineering", NodeType: "major", Path: "school-01/college-01/major-01", Level: 3, SortOrder: 1},
		{ID: "class-01", ParentID: "major-01", Name: "2024 Class 1", NodeType: "class", Path: "school-01/college-01/major-01/class-01", Level: 4, SortOrder: 1},
	}, nil
}

func (memoryOptionRepo) ListOptionItems(_ context.Context, groupKey string) ([]optionsservice.OptionItem, error) {
	switch groupKey {
	case "grades":
		return []optionsservice.OptionItem{
			{Value: "2024", Label: "2024"},
			{Value: "2025", Label: "2025"},
			{Value: "2026", Label: "2026"},
		}, nil
	case "academic_years":
		return []optionsservice.OptionItem{
			{Value: "2024-2025", Label: "2024-2025"},
			{Value: "2025-2026", Label: "2025-2026"},
			{Value: "2026-2027", Label: "2026-2027"},
		}, nil
	case "semesters":
		return []optionsservice.OptionItem{
			{Value: "spring", Label: "Spring"},
			{Value: "autumn", Label: "Autumn"},
		}, nil
	case "teacher_titles":
		return []optionsservice.OptionItem{
			{Value: "assistant", Label: "Assistant"},
			{Value: "lecturer", Label: "Lecturer"},
			{Value: "associate_professor", Label: "Associate Professor"},
			{Value: "professor", Label: "Professor"},
		}, nil
	default:
		return []optionsservice.OptionItem{}, nil
	}
}

func (memoryCategorySchemaRepo) ListCategories(context.Context) ([]activityrepo.CategorySummary, error) {
	return []activityrepo.CategorySummary{
		{Key: "general", Label: "General", SortOrder: 1},
		{Key: "innovation", Label: "Innovation", SortOrder: 2},
	}, nil
}

func (memoryCategorySchemaRepo) GetCategorySchema(_ context.Context, key string) (*activityrepo.CategorySchemaEnvelope, error) {
	switch key {
	case "general":
		return &activityrepo.CategorySchemaEnvelope{
			Category: activityrepo.CategoryDescriptor{
				Key:   "general",
				Label: "General",
			},
			Schema: activityrepo.CategorySchema{
				Version: 1,
				Fields:  []activityrepo.SchemaField{},
			},
		}, nil
	case "innovation":
		return &activityrepo.CategorySchemaEnvelope{
			Category: activityrepo.CategoryDescriptor{
				Key:   "innovation",
				Label: "Innovation",
			},
			Schema: activityrepo.CategorySchema{
				Version: 1,
				Fields: []activityrepo.SchemaField{
					{
						Key:   "location",
						Label: "Location",
						Type:  "text",
					},
				},
			},
		}, nil
	default:
		return nil, activityrepo.ErrCategorySchemaNotFound
	}
}

func (a *App) Run() error {
	a.logger.Info("starting backend server", "addr", a.server.Addr)

	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (a *App) Handler() http.Handler {
	return a.server.Handler
}
