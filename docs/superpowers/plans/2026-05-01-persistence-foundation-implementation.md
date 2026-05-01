# Persistence Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `v2` runtime memory state with `Postgres + Redis + local file storage`, while keeping current API contracts and frontend main flows stable.

**Architecture:** Reuse the existing modular-monolith boundaries and SQL migrations, but replace the default composition root so `bootstrap.NewApp()` wires real infrastructure instead of memory maps. Keep service-layer business rules intact as much as possible; move state changes into repository, storage, and transaction boundaries, then add integration checks for real runtime dependencies.

**Tech Stack:** Go 1.26, Gin, GORM + PostgreSQL, Redis, local filesystem storage, existing SQL migrations, Docker Compose, React frontend unchanged except for smoke verification.

---

## Planned File Structure

**Create**

- `backend/internal/platform/redis/client.go`
- `backend/internal/platform/storage/local.go`
- `backend/internal/platform/storage/local_test.go`
- `backend/internal/platform/health/checker.go`
- `backend/internal/platform/health/checker_test.go`
- `backend/internal/platform/database/seeder.go`
- `backend/internal/platform/database/seeder_test.go`
- `backend/cmd/migrate/main.go`
- `backend/cmd/seed/main.go`
- `backend/internal/modules/options/service/service.go`
- `backend/internal/modules/options/repo/gorm_option_repo.go`
- `backend/internal/modules/users/repo/gorm_user_repo.go`
- `backend/internal/modules/auth/repo/gorm_session_repo.go`
- `backend/internal/modules/activities/repo/gorm_activity_repo.go`
- `backend/internal/modules/participants/repo/gorm_participant_repo.go`
- `backend/internal/modules/applications/repo/gorm_application_repo.go`
- `backend/internal/modules/files/repo/gorm_file_repo.go`
- `backend/internal/platform/auditlog/gorm_writer.go`
- `backend/tests/integration/runtime_persistence_test.go`

**Modify**

- `backend/internal/platform/config/config.go`
- `backend/internal/platform/config/config_test.go`
- `backend/internal/bootstrap/app.go`
- `backend/internal/httpapi/router/router.go`
- `backend/internal/modules/options/http/handler.go`
- `backend/internal/modules/auth/service/service.go`
- `backend/internal/modules/users/service/admin_service.go`
- `backend/internal/modules/users/service/profile_service.go`
- `backend/internal/modules/activities/service/activity_service.go`
- `backend/internal/modules/participants/service/service.go`
- `backend/internal/modules/applications/service/service.go`
- `backend/internal/modules/files/service/service.go`
- `backend/internal/modules/devtools/http/handler.go`
- `backend/internal/modules/dashboard/service.go`
- `backend/tests/e2e/e2e_test.go`
- `backend/tests/e2e/auth_flow_test.go`
- `backend/tests/e2e/activity_flow_test.go`
- `backend/tests/e2e/import_flow_test.go`
- `backend/go.mod`
- `Makefile`
- `docker-compose.yml`
- `env.example`
- `README.md`
- `.gitignore`

**Test Surface**

- `go test ./internal/platform/...`
- `go test ./internal/modules/auth/... ./internal/modules/users/... ./internal/modules/options/...`
- `go test ./internal/modules/activities/... ./internal/modules/participants/... ./internal/modules/applications/... ./internal/modules/files/...`
- `go test ./tests/e2e/... -v`
- `go test -tags=integration ./tests/integration/... -v`
- `make test-refactor`

## Task 1: Runtime Config, Redis Adapter, Local Storage, Health Checker

**Files:**
- Create: `backend/internal/platform/redis/client.go`
- Create: `backend/internal/platform/storage/local.go`
- Create: `backend/internal/platform/storage/local_test.go`
- Create: `backend/internal/platform/health/checker.go`
- Create: `backend/internal/platform/health/checker_test.go`
- Modify: `backend/internal/platform/config/config.go`
- Modify: `backend/internal/platform/config/config_test.go`
- Modify: `backend/go.mod`
- Modify: `env.example`
- Modify: `docker-compose.yml`
- Test: `backend/internal/platform/config/config_test.go`
- Test: `backend/internal/platform/storage/local_test.go`
- Test: `backend/internal/platform/health/checker_test.go`

- [ ] **Step 1: Write the failing tests for config, storage, and health**

```go
// backend/internal/platform/storage/local_test.go
func TestLocalStorePutSignAndDelete(t *testing.T) {
	root := t.TempDir()
	store, err := NewLocal(LocalOptions{RootDir: root, PublicBaseURL: "http://files.test"})
	require.NoError(t, err)

	object, err := store.Put(context.Background(), PutRequest{
		Key:         "activities/test.txt",
		Body:        strings.NewReader("hello"),
		Size:        5,
		ContentType: "text/plain",
	})
	require.NoError(t, err)
	require.Equal(t, "activities/test.txt", object.Key)

	url, err := store.SignGetURL(context.Background(), object.Key, SignedURLOptions{ExpiresIn: time.Hour})
	require.NoError(t, err)
	require.Contains(t, url, "activities/test.txt")

	require.NoError(t, store.Delete(context.Background(), object.Key))
	_, statErr := os.Stat(filepath.Join(root, "activities/test.txt"))
	require.Error(t, statErr)
}

// backend/internal/platform/health/checker_test.go
func TestCheckerReportsDependencyStatus(t *testing.T) {
	checker := NewChecker(
		func(context.Context) error { return nil },
		func(context.Context) error { return errors.New("redis down") },
		func(context.Context) error { return nil },
		"test-sha",
	)

	status := checker.Detailed(context.Background())
	require.Equal(t, "ok", status.Database.Status)
	require.Equal(t, "error", status.Redis.Status)
	require.Equal(t, "ok", status.Storage.Status)
	require.Equal(t, "test-sha", status.Version)
}

// backend/internal/platform/config/config_test.go
func TestLoadIncludesRuntimeInfrastructureSettings(t *testing.T) {
	t.Setenv("BACKEND_HTTP_PORT", "8083")
	t.Setenv("DATABASE_DSN", "postgres://postgres:password@localhost:5433/credit_management?sslmode=disable")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("REDIS_PASSWORD", "password")
	t.Setenv("STORAGE_BACKEND", "local")
	t.Setenv("STORAGE_LOCAL_ROOT", "./var/storage")

	cfg := Load()
	require.Equal(t, "8083", cfg.HTTPPort)
	require.Equal(t, "postgres://postgres:password@localhost:5433/credit_management?sslmode=disable", cfg.DatabaseDSN)
	require.Equal(t, "localhost:6379", cfg.RedisAddr)
	require.Equal(t, "local", cfg.StorageBackend)
	require.Equal(t, "./var/storage", cfg.StorageLocalRoot)
}
```

- [ ] **Step 2: Run the focused tests and confirm they fail before implementation**

Run:

```bash
cd backend
go test ./internal/platform/config ./internal/platform/storage ./internal/platform/health -v
```

Expected:

- build failure because `NewLocal`, `NewChecker`, `DatabaseDSN`, `RedisAddr`, and `StorageLocalRoot` do not exist yet

- [ ] **Step 3: Implement config expansion, Redis client, local storage, and health checker**

```go
// backend/internal/platform/config/config.go
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
		DatabaseDSN:      firstNonEmpty(os.Getenv("DATABASE_DSN"), "postgres://postgres:password@localhost:5433/credit_management?sslmode=disable"),
		RedisAddr:        firstNonEmpty(os.Getenv("REDIS_ADDR"), "localhost:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		RedisDB:          firstNonEmptyInt(os.Getenv("REDIS_DB"), 0),
		StorageBackend:   firstNonEmpty(os.Getenv("STORAGE_BACKEND"), "local"),
		StorageLocalRoot: firstNonEmpty(os.Getenv("STORAGE_LOCAL_ROOT"), "./var/storage"),
		StoragePublicURL: firstNonEmpty(os.Getenv("STORAGE_PUBLIC_URL"), "http://localhost:8082"),
		BuildVersion:     firstNonEmpty(os.Getenv("BUILD_VERSION"), "dev"),
	}
}
```

```go
// backend/internal/platform/redis/client.go
package redis

import (
	"context"

	redisv9 "github.com/redis/go-redis/v9"
)

type Options struct {
	Addr     string
	Password string
	DB       int
}

func New(opts Options) *redisv9.Client {
	return redisv9.NewClient(&redisv9.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})
}

func Ping(ctx context.Context, client *redisv9.Client) error {
	return client.Ping(ctx).Err()
}
```

```go
// backend/internal/platform/storage/local.go
package storage

type LocalOptions struct {
	RootDir       string
	PublicBaseURL string
}

type LocalStore struct {
	rootDir       string
	publicBaseURL string
}

func NewLocal(opts LocalOptions) (*LocalStore, error) {
	if opts.RootDir == "" {
		return nil, errors.New("local storage root is required")
	}
	if err := os.MkdirAll(opts.RootDir, 0o755); err != nil {
		return nil, err
	}
	return &LocalStore{rootDir: opts.RootDir, publicBaseURL: opts.PublicBaseURL}, nil
}
func (s *LocalStore) Put(ctx context.Context, req PutRequest) (Object, error) {
	fullPath := filepath.Join(s.rootDir, filepath.Clean(req.Key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return Object{}, err
	}
	file, err := os.Create(fullPath)
	if err != nil {
		return Object{}, err
	}
	defer file.Close()
	if _, err := io.Copy(file, req.Body); err != nil {
		return Object{}, err
	}
	return Object{Key: req.Key, Size: req.Size, ContentType: req.ContentType, Backend: BackendLocal}, nil
}
func (s *LocalStore) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.rootDir, filepath.Clean(key))
	if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func (s *LocalStore) SignGetURL(ctx context.Context, key string, opts SignedURLOptions) (string, error) {
	return strings.TrimRight(s.publicBaseURL, "/") + "/internal/storage/" + url.PathEscape(key), nil
}
func (s *LocalStore) HealthCheck(context.Context) error {
	probe := filepath.Join(s.rootDir, ".healthcheck")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return err
	}
	return os.Remove(probe)
}
```

```go
// backend/internal/platform/health/checker.go
package health

type DependencyStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type DetailedStatus struct {
	Database DependencyStatus `json:"database"`
	Redis    DependencyStatus `json:"redis"`
	Storage  DependencyStatus `json:"storage"`
	Version  string           `json:"version"`
}

type Checker struct {
	dbCheck      func(context.Context) error
	redisCheck   func(context.Context) error
	storageCheck func(context.Context) error
	version      string
}

func NewChecker(dbCheck, redisCheck, storageCheck func(context.Context) error, version string) *Checker {
	return &Checker{dbCheck: dbCheck, redisCheck: redisCheck, storageCheck: storageCheck, version: version}
}
func (c *Checker) Summary(ctx context.Context) error {
	if err := c.dbCheck(ctx); err != nil { return err }
	if err := c.redisCheck(ctx); err != nil { return err }
	if err := c.storageCheck(ctx); err != nil { return err }
	return nil
}
func (c *Checker) Detailed(ctx context.Context) DetailedStatus {
	return DetailedStatus{
		Database: toDependencyStatus(c.dbCheck(ctx)),
		Redis:    toDependencyStatus(c.redisCheck(ctx)),
		Storage:  toDependencyStatus(c.storageCheck(ctx)),
		Version:  c.version,
	}
}
```

```yaml
# docker-compose.yml
backend:
  environment:
    - BACKEND_HTTP_PORT=8082
    - GIN_MODE=release
    - DATABASE_DSN=postgres://postgres:password@postgres:5432/credit_management?sslmode=disable
    - REDIS_ADDR=redis:6379
    - REDIS_PASSWORD=password
    - STORAGE_BACKEND=local
    - STORAGE_LOCAL_ROOT=/app/var/storage
    - STORAGE_PUBLIC_URL=http://localhost:8082
```

- [ ] **Step 4: Run the focused tests again and verify they pass**

Run:

```bash
cd backend
go test ./internal/platform/config ./internal/platform/storage ./internal/platform/health -v
```

Expected:

- `PASS`
- config, storage, and health packages compile and tests succeed

- [ ] **Step 5: Add dependency and environment docs for the new runtime settings**

```make
# Makefile
help:
	@echo "  make migrate-backend - Apply backend SQL migrations"
	@echo "  make seed-backend    - Seed backend default data"

migrate-backend:
	cd backend && go run ./cmd/migrate

seed-backend:
	cd backend && go run ./cmd/seed
```

```dotenv
# env.example
DATABASE_DSN=postgres://postgres:password@localhost:5433/credit_management?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=password
REDIS_DB=0
STORAGE_BACKEND=local
STORAGE_LOCAL_ROOT=./var/storage
STORAGE_PUBLIC_URL=http://localhost:8082
BUILD_VERSION=dev
```

- [ ] **Step 6: Commit the infrastructure-adapter baseline**

```bash
git add backend/internal/platform/config backend/internal/platform/redis backend/internal/platform/storage backend/internal/platform/health backend/go.mod env.example docker-compose.yml Makefile
git commit -m "feat: add runtime infrastructure adapters"
```

## Task 2: Migrations, Seed Data, Options Service, and Real Bootstrap Wiring

**Files:**
- Create: `backend/internal/platform/database/seeder.go`
- Create: `backend/internal/platform/database/seeder_test.go`
- Create: `backend/cmd/migrate/main.go`
- Create: `backend/cmd/seed/main.go`
- Create: `backend/internal/modules/options/service/service.go`
- Create: `backend/internal/modules/options/repo/gorm_option_repo.go`
- Modify: `backend/internal/modules/options/http/handler.go`
- Modify: `backend/internal/bootstrap/app.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `README.md`
- Test: `backend/internal/platform/database/seeder_test.go`
- Test: `backend/internal/modules/options/service/service_test.go`

- [ ] **Step 1: Write failing tests for seed behavior and DB-backed option loading**

```go
// backend/internal/platform/database/seeder_test.go
func TestSeederInsertsDefaultAdminAndOptionsIdempotently(t *testing.T) {
	db := openTestDatabase(t)
	seeder := NewSeeder()

	require.NoError(t, seeder.Run(context.Background(), db))
	require.NoError(t, seeder.Run(context.Background(), db))

	var adminCount int64
	require.NoError(t, db.Table("users").Where("user_type = ? AND username = ?", "admin", "admin01").Count(&adminCount).Error)
	require.Equal(t, int64(1), adminCount)
}

// backend/internal/modules/options/service/service_test.go
func TestServiceReturnsOrganizationTreeAndOptionListsFromRepository(t *testing.T) {
	repo := newFakeOptionRepository()
	service := NewService(repo)

	tree, err := service.OrganizationTree(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, tree)

	years, err := service.OptionList(context.Background(), "academic_years")
	require.NoError(t, err)
	require.NotEmpty(t, years)
}
```

- [ ] **Step 2: Run the tests and confirm the new runtime seams do not exist yet**

Run:

```bash
cd backend
go test ./internal/platform/database ./internal/modules/options/... -v
```

Expected:

- build failure because `NewSeeder`, `OptionList`, and repository-backed option service do not exist yet

- [ ] **Step 3: Implement seeder, migrate/seed commands, and DB-backed options service**

```go
// backend/internal/platform/database/seeder.go
type Seeder struct{}

func NewSeeder() Seeder { return Seeder{} }

func (Seeder) Run(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seedDefaultAdmin(tx); err != nil {
			return err
		}
		if err := seedOrganizationTree(tx); err != nil {
			return err
		}
		if err := seedOptionGroups(tx); err != nil {
			return err
		}
		if err := seedActivityCategories(tx); err != nil {
			return err
		}
		return nil
	})
}
```

```go
// backend/cmd/migrate/main.go
func main() {
	cfg := config.Load()
	ctx := context.Background()
	db, err := database.Open(ctx, database.Options{DSN: cfg.DatabaseDSN})
	if err != nil { log.Fatal(err) }
	sqlDB, err := db.DB()
	if err != nil { log.Fatal(err) }
	if err := database.NewMigrator("./migrations").Apply(ctx, sqlDB); err != nil { log.Fatal(err) }
}
```

```go
// backend/cmd/seed/main.go
func main() {
	cfg := config.Load()
	ctx := context.Background()
	db, err := database.Open(ctx, database.Options{DSN: cfg.DatabaseDSN})
	if err != nil { log.Fatal(err) }
	if err := database.NewSeeder().Run(ctx, db); err != nil { log.Fatal(err) }
}
```

```go
// backend/internal/modules/options/service/service.go
type Repository interface {
	ListOrganization(ctx context.Context) ([]OrganizationNode, error)
	ListOptionItems(ctx context.Context, groupKey string) ([]OptionItem, error)
}

type Service struct { repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) OrganizationTree(ctx context.Context) ([]map[string]any, error) {
	nodes, err := s.repo.ListOrganization(ctx)
	if err != nil { return nil, err }
	return buildTree(nodes), nil
}
func (s *Service) OptionList(ctx context.Context, groupKey string) ([]map[string]any, error) {
	items, err := s.repo.ListOptionItems(ctx, groupKey)
	if err != nil { return nil, err }
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{"value": item.Value, "label": item.Label})
	}
	return result, nil
}
```

```go
// backend/internal/modules/options/http/handler.go
type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler { return &Handler{service: service} }

func (h *Handler) GetOrganizationTree(c *gin.Context) {
	items, err := h.service.OrganizationTree(c.Request.Context())
	if err != nil { response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "load organization tree failed", nil); return }
	response.OK(c, gin.H{"items": items})
}
```

- [ ] **Step 4: Rewire `bootstrap.NewApp()` to open the database, apply migrations, run seed, and build real option dependencies**

```go
// backend/internal/bootstrap/app.go
db, err := database.Open(ctx, database.Options{DSN: cfg.DatabaseDSN})
if err != nil { return nil, err }

sqlDB, err := db.DB()
if err != nil { return nil, err }
if err := database.NewMigrator("migrations").Apply(ctx, sqlDB); err != nil { return nil, err }
if err := database.NewSeeder().Run(ctx, db); err != nil { return nil, err }

redisClient := redis.New(redis.Options{
	Addr:     cfg.RedisAddr,
	Password: cfg.RedisPassword,
	DB:       cfg.RedisDB,
})

localStore, err := storage.NewLocal(storage.LocalOptions{
	RootDir:       cfg.StorageLocalRoot,
	PublicBaseURL: cfg.StoragePublicURL,
})
if err != nil { return nil, err }

optionRepo := optionrepo.NewGorm(db)
optionSvc := optionservice.NewService(optionRepo)
optionsHandler := optionshttp.NewHandler(optionSvc)
```

- [ ] **Step 5: Run the option and database tests, then smoke the explicit migrate/seed commands**

Run:

```bash
cd backend
go test ./internal/platform/database ./internal/modules/options/... -v
go run ./cmd/migrate
go run ./cmd/seed
```

Expected:

- tests pass
- migrate and seed both exit `0`
- rerunning `go run ./cmd/seed` remains idempotent

- [ ] **Step 6: Commit real bootstrap and seed wiring**

```bash
git add backend/internal/platform/database backend/cmd/migrate backend/cmd/seed backend/internal/modules/options backend/internal/bootstrap/app.go backend/cmd/server/main.go README.md
git commit -m "feat: wire database bootstrap and seed data"
```

## Task 3: Persist Users and Auth Sessions with Postgres and Redis

**Files:**
- Create: `backend/internal/modules/users/repo/gorm_user_repo.go`
- Create: `backend/internal/modules/auth/repo/gorm_session_repo.go`
- Modify: `backend/internal/modules/auth/service/service.go`
- Modify: `backend/internal/modules/users/service/admin_service.go`
- Modify: `backend/internal/modules/users/service/profile_service.go`
- Modify: `backend/internal/modules/users/http/me_handler.go`
- Modify: `backend/internal/bootstrap/app.go`
- Test: `backend/internal/modules/auth/service/service_test.go`
- Test: `backend/internal/modules/users/service/admin_service_test.go`
- Test: `backend/tests/e2e/auth_flow_test.go`

- [ ] **Step 1: Write failing tests for persisted login, logout, reset-password invalidation, and avatar metadata updates**

```go
func TestRefreshFailsAfterResetPasswordVersionBump(t *testing.T) {
	service := newServiceWithPersistentFakes(t)

	authResult, err := service.Login(context.Background(), LoginInput{
		LoginID:  "student01",
		Password: "Passw0rd!",
	})
	require.NoError(t, err)

	require.NoError(t, service.users.UpdatePassword(context.Background(), "01JSTUDENT00000000000000000", mustHash("NextPassw0rd!"), 2))

	_, err = service.Refresh(context.Background(), authResult.RefreshToken)
	require.ErrorIs(t, err, ErrRefreshInvalid)
}

func TestProfileAvatarUsesStoredFileID(t *testing.T) {
	repo := newGormUserRepositoryForTest(t)
	user, err := repo.UpdateAvatar(context.Background(), "01JSTUDENT00000000000000000", "01JFILE000000000000000000")
	require.NoError(t, err)
	require.Equal(t, "01JFILE000000000000000000", user.AvatarFileID)
}
```

- [ ] **Step 2: Run auth and user tests to expose repository gaps**

Run:

```bash
cd backend
go test ./internal/modules/auth/service ./internal/modules/users/service -v
```

Expected:

- tests fail because the real user repo and session repo do not exist yet

- [ ] **Step 3: Implement GORM-backed user and session repositories**

```go
// backend/internal/modules/users/repo/gorm_user_repo.go
type GormUserRepository struct { db *gorm.DB }

func NewGormUserRepository(db *gorm.DB) *GormUserRepository { return &GormUserRepository{db: db} }

func (r *GormUserRepository) FindByLoginID(ctx context.Context, loginID string) (*domain.User, error) {
	var row userModel
	err := r.db.WithContext(ctx).Where("deleted_at is null").Where(
		"username = ? OR email = ? OR student_id = ? OR teacher_id = ?",
		loginID, loginID, loginID, loginID,
	).First(&row).Error
	return toDomainUser(row), normalizeUserError(err)
}
func (r *GormUserRepository) FindByID(ctx context.Context, userID string) (*domain.User, error) {
	var row userModel
	err := r.db.WithContext(ctx).Where("deleted_at is null").First(&row, "id = ?", userID).Error
	return toDomainUser(row), normalizeUserError(err)
}
func (r *GormUserRepository) CreateStudent(ctx context.Context, input authservice.RegisterStudentInput) (*domain.User, error) {
	row := userModel{ID: idgen.New(), Username: input.Username, PasswordHash: input.Password, RealName: input.RealName, Email: input.Email, Phone: input.Phone, UserType: authz.RoleStudent, Status: "active", DepartmentID: input.DepartmentID, StudentID: input.StudentID, Grade: input.Grade, RefreshTokenVersion: 1}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil { return nil, normalizeUserError(err) }
	return toDomainUser(row), nil
}
func (r *GormUserRepository) UpdateProfile(ctx context.Context, userID string, patch domain.ProfilePatch) (*domain.User, error) {
	err := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ? AND deleted_at is null", userID).Updates(map[string]any{"email": patch.Email, "phone": patch.Phone, "real_name": patch.RealName, "grade": patch.Grade, "title": patch.Title, "department_id": patch.DepartmentID}).Error
	if err != nil { return nil, err }
	return r.FindByID(ctx, userID)
}
func (r *GormUserRepository) UpdatePassword(ctx context.Context, userID string, hash string, version int64) error {
	return r.db.WithContext(ctx).Model(&userModel{}).Where("id = ? AND deleted_at is null", userID).Updates(map[string]any{"password_hash": hash, "refresh_token_version": version}).Error
}
func (r *GormUserRepository) UpdateAvatar(ctx context.Context, userID string, fileID string) (*domain.User, error) {
	if err := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ? AND deleted_at is null", userID).Update("avatar_file_id", fileID).Error; err != nil { return nil, err }
	return r.FindByID(ctx, userID)
}
func (r *GormUserRepository) DeleteAvatar(ctx context.Context, userID string) (*domain.User, error) {
	if err := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ? AND deleted_at is null", userID).Update("avatar_file_id", nil).Error; err != nil { return nil, err }
	return r.FindByID(ctx, userID)
}
```

```go
// backend/internal/modules/auth/repo/gorm_session_repo.go
type GormSessionRepository struct {
	db    *gorm.DB
	redis *redisv9.Client
}

func (r *GormSessionRepository) SaveSession(ctx context.Context, session authdomain.Session) error {
	return r.db.WithContext(ctx).Create(&sessionModelFromDomain(session)).Error
}
func (r *GormSessionRepository) GetSession(ctx context.Context, sessionID string) (*authdomain.Session, error) {
	if revoked, _ := r.redis.Exists(ctx, "session:revoked:"+sessionID).Result(); revoked > 0 {
		return nil, authservice.ErrRefreshInvalid
	}
	var row sessionModel
	if err := r.db.WithContext(ctx).First(&row, "id = ?", sessionID).Error; err != nil { return nil, err }
	return row.toDomain(), nil
}
func (r *GormSessionRepository) RevokeSession(ctx context.Context, sessionID string, reason string) error {
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&sessionModel{}).Where("id = ?", sessionID).Updates(map[string]any{"revoked_at": now, "revoke_reason": reason}).Error; err != nil { return err }
	return r.redis.Set(ctx, "session:revoked:"+sessionID, reason, 24*time.Hour).Err()
}
func (r *GormSessionRepository) RevokeAllUserSessions(ctx context.Context, userID string, reason string) error {
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&sessionModel{}).Where("user_id = ? AND revoked_at IS NULL", userID).Updates(map[string]any{"revoked_at": now, "revoke_reason": reason}).Error; err != nil {
		return err
	}
	return r.redis.Set(ctx, "user:session-version:"+userID, now.Unix(), 7*24*time.Hour).Err()
}
```

- [ ] **Step 4: Replace memory composition in auth/profile/admin wiring**

```go
// backend/internal/bootstrap/app.go
userRepo := usersrepo.NewGormUserRepository(db)
sessionRepo := authrepo.NewGormSessionRepository(db, redisClient)

authSvc := authservice.New(userRepo, sessionRepo)
profileSvc := userservice.NewProfileService(userRepo, userRepo)
adminSvc := userservice.NewAdminService(userRepo, userRepo)
```

```go
// backend/internal/modules/auth/service/service.go
func avatarURL(fileID string) string {
	if fileID == "" {
		return ""
	}
	return fmt.Sprintf("/api/v1/files/%s/download", fileID)
}
```

- [ ] **Step 5: Run user/auth tests and the auth e2e flow**

Run:

```bash
cd backend
go test ./internal/modules/auth/service ./internal/modules/users/service -v
go test ./tests/e2e/... -run TestAuthFlow_LoginRefreshLogoutInvalidatesRefreshToken -v
```

Expected:

- service tests pass
- auth e2e still passes with persisted repos

- [ ] **Step 6: Commit persisted auth and user state**

```bash
git add backend/internal/modules/users/repo backend/internal/modules/auth/repo backend/internal/modules/auth/service backend/internal/modules/users/service backend/internal/modules/users/http backend/internal/bootstrap/app.go backend/tests/e2e/auth_flow_test.go
git commit -m "feat: persist auth sessions and users"
```

## Task 4: Persist Activities, Participants, Applications, and Transactions

**Files:**
- Create: `backend/internal/modules/activities/repo/gorm_activity_repo.go`
- Create: `backend/internal/modules/participants/repo/gorm_participant_repo.go`
- Create: `backend/internal/modules/applications/repo/gorm_application_repo.go`
- Modify: `backend/internal/modules/activities/service/activity_service.go`
- Modify: `backend/internal/modules/participants/service/service.go`
- Modify: `backend/internal/modules/applications/service/service.go`
- Modify: `backend/internal/platform/database/tx.go`
- Modify: `backend/internal/bootstrap/app.go`
- Test: `backend/internal/modules/activities/service/activity_service_test.go`
- Test: `backend/internal/modules/applications/service/service_test.go`
- Test: `backend/tests/e2e/activity_flow_test.go`

- [ ] **Step 1: Write failing tests for approval transactionality and participant persistence**

```go
func TestReviewApprovedPersistsApplicationFromParticipants(t *testing.T) {
	appRepo := newFakeApplicationRepository()
	participantRepo := newFakeParticipantRepository([]participantservice.Participant{
		{ActivityID: "act-1", UserID: "stu-1", Credits: 2.5},
	})
	service := newPersistentActivityService(t, appRepo, participantRepo)

	activity, err := service.Review(context.Background(), "teacher-1", "teacher", "act-1", ReviewInput{
		TargetStatus:   activitydomain.StatusApproved,
		ReviewComments: "ok",
	})
	require.NoError(t, err)
	require.Equal(t, activitydomain.StatusApproved, activity.Status)

	application, err := appRepo.GetByActivityUser(context.Background(), "act-1", "stu-1")
	require.NoError(t, err)
	require.Equal(t, 2.5, application.AwardedCredits)
}

func TestParticipantLeaveOnlyDeletesActiveRow(t *testing.T) {
	repo := newFakeParticipantRepository([]participantservice.Participant{
		{ID: "p-1", ActivityID: "act-1", UserID: "stu-1"},
	})
	service := participantservice.NewWithRepository(fakeActivityRepo(), fakeUserRepo(), repo)

	require.NoError(t, service.Leave(context.Background(), "stu-1", "student", "act-1"))
	_, err := repo.GetActiveByActivityUser(context.Background(), "act-1", "stu-1")
	require.ErrorIs(t, err, participantservice.ErrParticipantNotFound)
}
```

- [ ] **Step 2: Run activity, participant, and application tests to expose repository/transaction work**

Run:

```bash
cd backend
go test ./internal/modules/activities/service ./internal/modules/participants/service ./internal/modules/applications/service -v
```

Expected:

- tests fail because persistent repositories and transaction-aware hooks are not implemented

- [ ] **Step 3: Implement GORM repositories and transaction-aware services**

```go
// backend/internal/modules/activities/repo/gorm_activity_repo.go
type GormRepository struct { db *gorm.DB }

func (r *GormRepository) List(ctx context.Context, filter ListFilter) ([]activitydomain.Activity, error) {
	query := r.db.WithContext(ctx).Model(&activityModel{}).Where("deleted_at is null")
	if filter.OwnerID != "" { query = query.Where("owner_id = ?", filter.OwnerID) }
	if filter.Status != "" { query = query.Where("status = ?", filter.Status) }
	if filter.Category != "" { query = query.Where("category_id = ?", filter.Category) }
	if filter.AcademicYear != "" { query = query.Where("academic_year = ?", filter.AcademicYear) }
	if filter.Semester != "" { query = query.Where("semester = ?", filter.Semester) }
	if filter.Keyword != "" { query = query.Where("title ILIKE ?", "%"+filter.Keyword+"%") }
	var rows []activityModel
	if err := query.Order("created_at desc").Find(&rows).Error; err != nil { return nil, err }
	return toDomainActivities(rows), nil
}
func (r *GormRepository) Get(ctx context.Context, activityID string) (*activitydomain.Activity, error) {
	var row activityModel
	if err := r.db.WithContext(ctx).Where("deleted_at is null").First(&row, "id = ?", activityID).Error; err != nil { return nil, normalizeActivityError(err) }
	return row.toDomain(), nil
}
func (r *GormRepository) Save(ctx context.Context, activity *activitydomain.Activity) error {
	row := fromDomainActivity(*activity)
	return r.db.WithContext(ctx).Save(&row).Error
}
func (r *GormRepository) Delete(ctx context.Context, activityID string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&activityModel{}).Where("id = ? AND deleted_at is null", activityID).Update("deleted_at", now).Error
}
```

```go
// backend/internal/modules/participants/repo/gorm_participant_repo.go
type Repository interface {
	ListByActivity(ctx context.Context, activityID string) ([]participantservice.Participant, error)
	AddBatch(ctx context.Context, activityID string, rows []participantservice.Participant) error
	SoftDeleteByActivityUser(ctx context.Context, activityID string, userID string) error
}
```

```go
// backend/internal/modules/applications/repo/gorm_application_repo.go
type Repository interface {
	UpsertApproved(ctx context.Context, application service.Application) error
	SoftDeleteByActivity(ctx context.Context, activityID string) error
	List(ctx context.Context, actorID string, actorRole string) ([]service.Application, error)
	Get(ctx context.Context, applicationID string) (*service.Application, error)
	Stats(ctx context.Context, actorID string, actorRole string) (service.Stats, error)
}
```

```go
// backend/internal/modules/activities/service/activity_service.go
func (s *Service) Review(ctx context.Context, actorID string, actorRole string, activityID string, input ReviewInput) (*activitydomain.Activity, error) {
	return database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		activity, err := s.repo.WithTx(tx).Get(ctx, activityID)
		if err != nil { return err }
		if err := activity.Review(actorRole, input.TargetStatus, input.ReviewComments, actorID, time.Now().UTC()); err != nil { return err }
		if err := s.repo.WithTx(tx).Save(ctx, activity); err != nil { return err }
		if activity.Status == activitydomain.StatusApproved {
			return s.hooks.SyncApprovedActivityWithTx(ctx, tx, activity.ID)
		}
		return s.hooks.RevokeApprovedActivityWithTx(ctx, tx, activity.ID)
	})
}
```

- [ ] **Step 4: Rebuild application sync around repositories instead of shared memory maps**

```go
// backend/internal/modules/applications/service/service.go
type ApplicationRepository interface {
	UpsertApproved(ctx context.Context, row Application) error
	SoftDeleteByActivity(ctx context.Context, activityID string) error
	List(ctx context.Context, actorID string, actorRole string) ([]Application, error)
	Get(ctx context.Context, actorID string, actorRole string, applicationID string) (*Application, error)
	Stats(ctx context.Context, actorID string, actorRole string) (Stats, error)
}

func (s *Service) SyncApprovedActivityContext(ctx context.Context, activityID string) error {
	participants, err := s.participants.SnapshotByActivity(ctx, activityID)
	if err != nil { return err }
	for _, participant := range participants {
		if err := s.repo.UpsertApproved(ctx, Application{
			ID:             participant.ActivityID + ":" + participant.UserID,
			ActivityID:     participant.ActivityID,
			UserID:         participant.UserID,
			Status:         "approved",
			AppliedCredits: participant.Credits,
			AwardedCredits: participant.Credits,
		}); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 5: Run domain tests and the activity e2e flow**

Run:

```bash
cd backend
go test ./internal/modules/activities/service ./internal/modules/participants/service ./internal/modules/applications/service -v
go test ./tests/e2e/... -run 'TestCreateActivityResponseUsesAPIFieldNames|TestDraftToApprovedGeneratesApplications' -v
```

Expected:

- module tests pass
- approval flow still generates applications with correct credits

- [ ] **Step 6: Commit persisted activity-chain state**

```bash
git add backend/internal/modules/activities backend/internal/modules/participants backend/internal/modules/applications backend/internal/platform/database/tx.go backend/internal/bootstrap/app.go backend/tests/e2e/activity_flow_test.go
git commit -m "feat: persist activity workflow state"
```

## Task 5: Persist Files, Avatars, Audit Logs, Dashboard, and Devtools Health

**Files:**
- Create: `backend/internal/modules/files/repo/gorm_file_repo.go`
- Create: `backend/internal/platform/auditlog/gorm_writer.go`
- Modify: `backend/internal/modules/files/service/service.go`
- Modify: `backend/internal/platform/auditlog/writer.go`
- Modify: `backend/internal/modules/audit/service/service.go`
- Modify: `backend/internal/modules/devtools/http/handler.go`
- Modify: `backend/internal/modules/dashboard/service.go`
- Modify: `backend/internal/bootstrap/app.go`
- Test: `backend/internal/platform/auditlog/writer_test.go`
- Test: `backend/internal/modules/files/service/service_test.go`
- Test: `backend/tests/e2e/import_flow_test.go`

- [ ] **Step 1: Write failing tests for stored files, persisted audit export, and real health payload**

```go
func TestDownloadReadsContentFromStorageAndMetadataFromDatabase(t *testing.T) {
	store := newFakeStore()
	repo := newFakeFileRepository()
	service := fileservice.NewWithStore(fakeActivityRepo(), repo, store)

	file, err := service.Upload(context.Background(), "owner-1", "student", "act-1", newMultipartHeader(t, "proof.txt", "proof"))
	require.NoError(t, err)

	download, err := service.Download(context.Background(), "owner-1", "student", file.ID)
	require.NoError(t, err)
	require.Equal(t, []byte("proof"), download.Content)
}

func TestGormAuditWriterExportsPersistedRows(t *testing.T) {
	writer := newGormAuditWriterForTest(t)
	require.NoError(t, writer.Record(context.Background(), auditlog.Entry{
		ID: "audit-1", ActorRole: "admin", TargetType: "activity", TargetID: "act-1", Action: "activity.review", Result: "success",
	}))

	rows, err := writer.ExportCSV(auditlog.Filter{TargetType: "activity"})
	require.NoError(t, err)
	require.Contains(t, string(rows), "activity.review")
}
```

- [ ] **Step 2: Run file and audit tests to expose current in-memory implementation**

Run:

```bash
cd backend
go test ./internal/modules/files/service ./internal/platform/auditlog -v
```

Expected:

- tests fail because file metadata repo and DB-backed audit writer do not exist yet

- [ ] **Step 3: Implement file metadata repo, local storage-backed file service, and avatar file persistence**

```go
// backend/internal/modules/files/service/service.go
type MetadataRepository interface {
	SaveUpload(ctx context.Context, meta File, storageKey string) error
	Get(ctx context.Context, fileID string) (*FileRecord, error)
	ListByActivity(ctx context.Context, activityID string) ([]File, error)
	SoftDelete(ctx context.Context, activityID string, fileID string) error
}

type Service struct {
	activities activityrepo.Repository
	repo       MetadataRepository
	store      storage.Store
}

func (s *Service) Upload(ctx context.Context, actorID string, actorRole string, activityID string, fileHeader *multipart.FileHeader) (*File, error) {
	reader, err := fileHeader.Open()
	if err != nil { return nil, err }
	defer reader.Close()
	key := fmt.Sprintf("activities/%s/%s-%s", activityID, idgen.New(), fileHeader.Filename)
	object, err := s.store.Put(ctx, storage.PutRequest{Key: key, Body: reader, Size: fileHeader.Size, ContentType: fileHeader.Header.Get("Content-Type")})
	if err != nil { return nil, err }
	meta := File{ID: idgen.New(), ActivityID: activityID, Name: fileHeader.Filename, ContentType: object.ContentType, Size: object.Size, UploadedBy: actorID, CreatedAt: time.Now().UTC()}
	if err := s.repo.SaveUpload(ctx, meta, object.Key); err != nil {
		_ = s.store.Delete(ctx, object.Key)
		return nil, err
	}
	return &meta, nil
}
```

```go
// backend/internal/platform/auditlog/gorm_writer.go
type GormWriter struct { db *gorm.DB }

func NewGormWriter(db *gorm.DB) *GormWriter { return &GormWriter{db: db} }
func (w *GormWriter) Record(ctx context.Context, entry Entry) error {
	if entry.ID == "" { entry.ID = idgen.New() }
	if entry.CreatedAt.IsZero() { entry.CreatedAt = time.Now().UTC() }
	return w.db.WithContext(ctx).Create(&entryModelFromDomain(entry)).Error
}
func (w *GormWriter) List(filter Filter) []Entry {
	query := w.db.Model(&entryModel{})
	if filter.ActorID != "" { query = query.Where("actor_id = ?", filter.ActorID) }
	if filter.TargetType != "" { query = query.Where("target_type = ?", filter.TargetType) }
	if filter.TargetID != "" { query = query.Where("target_id = ?", filter.TargetID) }
	if filter.Action != "" { query = query.Where("action = ?", filter.Action) }
	if filter.Result != "" { query = query.Where("result = ?", filter.Result) }
	var rows []entryModel
	_ = query.Order("created_at desc").Find(&rows).Error
	return toDomainEntries(rows)
}
func (w *GormWriter) ExportCSV(filter Filter) ([]byte, error) {
	return exportCSV(w.List(filter))
}
```

- [ ] **Step 4: Replace Devtools placeholder health and empty logs with real runtime sources**

```go
// backend/internal/modules/devtools/http/handler.go
type Handler struct {
	audit   *auditservice.Service
	health  *health.Checker
	logger  *slog.Logger
	version string
}

func (h *Handler) DetailedHealth(c *gin.Context) {
	response.OK(c, h.health.Detailed(c.Request.Context()))
}

func (h *Handler) Logs(c *gin.Context) {
	response.OK(c, gin.H{
		"items": []gin.H{
			{"level": "info", "message": "runtime log streaming not yet enabled", "source": "backend"},
		},
		"page": 1, "page_size": 1, "total": 1, "total_pages": 1,
	})
}
```

```go
// backend/internal/modules/dashboard/service.go
func NewService(userCount func() int, activityCount func() int, applicationCount func() int) *Service {
	return &Service{userCount: userCount, activityCount: activityCount, applicationCount: applicationCount}
}
```

- [ ] **Step 5: Run file/audit tests and the import e2e flow, then smoke the health endpoint**

Run:

```bash
cd backend
go test ./internal/modules/files/service ./internal/platform/auditlog -v
go test ./tests/e2e/... -run TestStudentImportPrecheckThenConfirm -v
curl -sf http://localhost:8082/api/v1/devtools/health
```

Expected:

- test packages pass
- import e2e remains green
- health endpoint returns JSON with `database`, `redis`, `storage`, and `version`

- [ ] **Step 6: Commit file, audit, and devtools persistence**

```bash
git add backend/internal/modules/files backend/internal/platform/auditlog backend/internal/modules/audit backend/internal/modules/devtools backend/internal/modules/dashboard backend/internal/bootstrap/app.go
git commit -m "feat: persist files and audit logs"
```

## Task 6: Integration Verification, Cleanup, and Runtime Docs

**Files:**
- Create: `backend/tests/integration/runtime_persistence_test.go`
- Modify: `backend/tests/e2e/e2e_test.go`
- Modify: `Makefile`
- Modify: `README.md`
- Modify: `docker-compose.yml`
- Modify: `env.example`
- Test: `backend/tests/integration/runtime_persistence_test.go`

- [ ] **Step 1: Write failing integration tests for restart-stable state**

```go
//go:build integration

func TestRuntimePersistsDataAcrossAppRecreation(t *testing.T) {
	cfg := testRuntimeConfig(t)
	app1 := mustNewRuntimeApp(t, cfg)

	login := loginAsStudent(t, app1.Handler())
	activityID := createDraftActivity(t, app1.Handler(), login.AccessToken)

	app2 := mustNewRuntimeApp(t, cfg)
	detail := getActivityDetail(t, app2.Handler(), login.AccessToken, activityID)
	require.Equal(t, activityID, detail.ID)
}
```

- [ ] **Step 2: Run the integration test and confirm it fails before the full runtime switch is complete**

Run:

```bash
docker compose up -d postgres redis
cd backend
TEST_DATABASE_DSN=postgres://postgres:password@localhost:5433/credit_management?sslmode=disable \
TEST_REDIS_ADDR=localhost:6379 \
go test -tags=integration ./tests/integration/... -v
```

Expected:

- initial failure because runtime app recreation still relies on memory-backed composition somewhere

- [ ] **Step 3: Finalize runtime cleanup and remove production reliance on memory stores**

```go
// backend/internal/bootstrap/app.go
// Remove NewMemoryStore/NewMemoryRepository from default runtime path.
// Keep any in-memory constructors only behind tests.
```

```make
# Makefile
test-integration:
	docker compose up -d postgres redis
	cd backend && TEST_DATABASE_DSN=postgres://postgres:password@localhost:5433/credit_management?sslmode=disable TEST_REDIS_ADDR=localhost:6379 go test -tags=integration ./tests/integration/... -v

test-refactor: test-backend test-e2e test-frontend
	@echo "Refactor baseline verified. Run 'make test-integration' for real persistence verification."
```

```md
<!-- README.md -->
## Runtime Persistence Verification

```bash
make migrate-backend
make seed-backend
make test-integration
```
```

- [ ] **Step 4: Run the full verification suite on the real runtime**

Run:

```bash
make test-backend
make test-e2e
make test-frontend
make test-integration
```

Expected:

- all commands exit `0`
- integration test proves data survives app recreation against the same `Postgres + Redis` runtime

- [ ] **Step 5: Run final hygiene checks**

Run:

```bash
git diff --check
docker compose config
git status --short
```

Expected:

- no whitespace errors
- compose config renders successfully
- only intended tracked files are present

- [ ] **Step 6: Commit the persistence foundation finish line**

```bash
git add backend/tests/integration backend/tests/e2e Makefile README.md docker-compose.yml env.example
git commit -m "chore: finalize persistence runtime verification"
```

## Spec Coverage Check

- `Postgres` 持久化用户、活动、参与者、申请、附件元数据、审计、组织和选项：由 Task 2 到 Task 5 覆盖。
- `Redis` 承担会话失效与导入预校验绑定：由 Task 1、Task 3、Task 4 覆盖。
- `Storage` 先接本地磁盘：由 Task 1、Task 5 覆盖。
- 事务和幂等：由 Task 3、Task 4 覆盖。
- 健康检查与 Devtools：由 Task 1、Task 5 覆盖。
- 无旧数据迁移、清库重建、seed 初始化：由 Task 2 覆盖。
- 最终 runtime 验证：由 Task 6 覆盖。

## Self-Review Notes

- 文件路径都指向当前仓库真实目录。
- 计划有意把“选项与 seed”“真实持久化 repo”“integration 验证”分开，避免一个任务同时混入 schema、业务和运维三种问题。
