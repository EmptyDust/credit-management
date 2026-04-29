# Credit Management Refactor Program Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current gateway plus three Go services plus page-oriented frontend with the confirmed modular-monolith backend and module-oriented frontend, while preserving the approved business rules from `docs/refactor-requirements.md` and `docs/refactor-design.md`.

**Architecture:** Build the new `backend/` modular monolith in parallel with the legacy services, migrate business domains in slices, then switch the frontend to the new `/api/v1` contract and retire the old gateway/service composition. The frontend is refactored in place under `frontend/`, but reorganized into `app/`, `modules/`, and `shared/` boundaries.

**Tech Stack:** Go 1.26, Gin, GORM, PostgreSQL, Redis, React 18, TypeScript, Vite, React Router, React Hook Form, Zod, TanStack Query, Tailwind CSS, Radix UI, pnpm, Docker Compose.

**Status:** Completed on `master` as of `2026-04-28`.

**Completion Notes:**
- The modular-monolith backend, module-oriented frontend, cutover tooling, and OrbStack runtime validation were completed and merged locally into `master`.
- Final merged commit at completion time: `d1a007bada2097d57fce23fffc34bda03a0a7f6f`.
- Regression fixes discovered during browser smoke validation were included before merge: auth-page `401` redirect handling, activity API response field mapping, and approved-application credit sync.

---

## Scope Decomposition

This rewrite is too large to execute as a single uninterrupted task batch. The implementation must be delivered in nine staged tasks, each producing a working and testable slice:

1. Repository bootstrap and new runtime skeleton
2. Database migrations and shared platform services
3. Auth, session, profile, and options backbone
4. User administration and user import/export
5. Activities, participants, files, and review flow
6. Applications, exports, dashboard, audit, and devtools
7. Frontend shell, routing, auth client, and shared UI migration
8. Frontend business pages and import flows
9. Cutover, legacy shutdown, and full verification

## File Structure Map

### New Backend Tree

- `backend/go.mod`: new monolith module definition
- `backend/cmd/server/main.go`: process entrypoint
- `backend/internal/bootstrap/app.go`: dependency wiring and lifecycle
- `backend/internal/platform/config/`: environment and config loading
- `backend/internal/platform/database/`: DB connection, migration runner, transactions
- `backend/internal/platform/logger/`: structured logger setup
- `backend/internal/platform/authn/`: JWT issue/verify/refresh/session logic
- `backend/internal/platform/authz/`: role and activity visibility checks
- `backend/internal/platform/storage/`: local/object storage abstraction
- `backend/internal/platform/auditlog/`: append-only audit writer
- `backend/internal/httpapi/router/`: top-level `/api/v1` router registration
- `backend/internal/httpapi/middleware/`: request ID, logging, auth recovery, error mapping
- `backend/internal/modules/<domain>/`: per-domain service, repo, dto, and handler logic
- `backend/migrations/`: SQL migrations for the new schema

### Frontend Refactor Tree

- `frontend/src/app/providers/`: router, query client, auth bootstrap
- `frontend/src/app/layouts/`: `AuthLayout`, `AppLayout`, `AdminLayout`
- `frontend/src/app/router/`: route definitions and guards
- `frontend/src/modules/auth/`: login, register, session hooks
- `frontend/src/modules/dashboard/`: overview widgets and navigation
- `frontend/src/modules/users/`: students, teachers, profile
- `frontend/src/modules/activities/`: list, detail, editor, review, transfer owner
- `frontend/src/modules/applications/`: application list and detail drawer
- `frontend/src/modules/imports/`: reusable import wizard
- `frontend/src/modules/devtools/`: logs, audit logs, API tester
- `frontend/src/shared/api/`: axios client, query keys, response parsing
- `frontend/src/shared/ui/`: reused tables, dialogs, cards, forms

### Existing Files To Modify

- `.gitignore`
- `Makefile`
- `docker-compose.yml`
- `frontend/package.json`
- `frontend/vite.config.ts`
- `frontend/src/App.tsx`
- `frontend/src/lib/api.ts`
- `frontend/src/contexts/AuthContext.tsx`
- `docs/refactor-design.md`

### Legacy Files To Keep Read-Only During Migration

- `auth-service/**`
- `credit-activity-service/**`
- `user-service/**`
- `api-gateway/**`

These old services are migration references only until Task 9.

### Task 1: Repository Bootstrap And New Runtime Skeleton

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/bootstrap/app.go`
- Create: `backend/internal/platform/config/config.go`
- Create: `backend/internal/platform/logger/logger.go`
- Create: `backend/internal/httpapi/router/router.go`
- Create: `backend/internal/httpapi/middleware/request_id.go`
- Create: `backend/internal/httpapi/middleware/recovery.go`
- Modify: `.gitignore`
- Modify: `docker-compose.yml`
- Modify: `Makefile`
- Test: `backend/internal/platform/config/config_test.go`

- [x] **Step 1: Scaffold the new backend Go module**

```go
module credit-management/backend

go 1.26.0

require (
    github.com/gin-gonic/gin v1.11.0
    github.com/joho/godotenv v1.5.1
    gorm.io/driver/postgres v1.6.0
    gorm.io/gorm v1.31.1
)
```

- [x] **Step 2: Add the minimal process bootstrap**

```go
package main

import (
    "log"

    "credit-management/backend/internal/bootstrap"
)

func main() {
    app, err := bootstrap.NewApp()
    if err != nil {
        log.Fatal(err)
    }
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

- [x] **Step 3: Add the first failing config test**

```go
func TestLoadConfigDefaults(t *testing.T) {
    cfg := config.Load()
    if cfg.HTTPPort != "8082" {
        t.Fatalf("expected default port 8082, got %s", cfg.HTTPPort)
    }
}
```

- [x] **Step 4: Run the new backend test to verify the skeleton compiles**

Run: `cd backend && go test ./internal/platform/config -v`
Expected: `PASS`

- [x] **Step 5: Wire the new backend service into local tooling**

```yaml
backend:
  build: ./backend
  container_name: credit_management_backend
  ports:
    - "8082:8082"
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
```

- [x] **Step 6: Update automation targets**

```make
test-backend:
	cd backend && go test ./...

dev-backend:
	cd backend && go run ./cmd/server
```

- [x] **Step 7: Commit**

```bash
git add .gitignore Makefile docker-compose.yml backend/go.mod backend/cmd/server/main.go backend/internal/bootstrap/app.go backend/internal/platform/config/config.go backend/internal/platform/logger/logger.go backend/internal/httpapi/router/router.go backend/internal/httpapi/middleware/request_id.go backend/internal/httpapi/middleware/recovery.go backend/internal/platform/config/config_test.go
git commit -m "chore: scaffold modular monolith backend"
```

### Task 2: Database Migrations And Shared Platform Services

**Files:**
- Create: `backend/migrations/0001_init_extensions.sql`
- Create: `backend/migrations/0002_create_users_and_sessions.sql`
- Create: `backend/migrations/0003_create_org_options_activities.sql`
- Create: `backend/migrations/0004_create_participants_applications_files.sql`
- Create: `backend/migrations/0005_create_imports_audit.sql`
- Create: `backend/internal/platform/database/database.go`
- Create: `backend/internal/platform/database/migrator.go`
- Create: `backend/internal/platform/database/tx.go`
- Create: `backend/internal/platform/idgen/idgen.go`
- Create: `backend/internal/platform/timeutil/timeutil.go`
- Create: `backend/internal/platform/storage/storage.go`
- Test: `backend/internal/platform/database/migrator_test.go`

- [x] **Step 1: Write a failing migration runner test**

```go
func TestMigratorDiscoversSQLFiles(t *testing.T) {
    files, err := database.DiscoverMigrations("../../migrations")
    require.NoError(t, err)
    require.NotEmpty(t, files)
}
```

- [x] **Step 2: Add the migration runner**

```go
func DiscoverMigrations(dir string) ([]string, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    var files []string
    for _, entry := range entries {
        if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
            files = append(files, filepath.Join(dir, entry.Name()))
        }
    }
    sort.Strings(files)
    return files, nil
}
```

- [x] **Step 3: Define the first schema migration slices**

```sql
create table users (
  id varchar(32) primary key,
  username varchar(64) not null unique,
  password_hash text not null,
  user_type varchar(16) not null,
  status varchar(16) not null,
  deleted_at timestamptz null
);
```

- [x] **Step 4: Add imports, audit, and activity-related tables using the confirmed schema from the design doc**

```sql
create table activities (
  id varchar(32) primary key,
  owner_id varchar(32) not null references users(id),
  title varchar(255) not null,
  category varchar(64) not null,
  academic_year varchar(32) not null,
  semester varchar(32) not null,
  status varchar(32) not null,
  details_json jsonb not null default '{}'::jsonb
);
```

- [x] **Step 5: Verify migration discovery and SQL ordering**

Run: `cd backend && go test ./internal/platform/database -v`
Expected: `PASS`

- [x] **Step 6: Commit**

```bash
git add backend/migrations backend/internal/platform/database backend/internal/platform/idgen/idgen.go backend/internal/platform/timeutil/timeutil.go backend/internal/platform/storage/storage.go
git commit -m "feat: add migration runner and core schema"
```

### Task 3: Auth, Session, Profile, And Options Backbone

**Files:**
- Create: `backend/internal/modules/auth/domain/session.go`
- Create: `backend/internal/modules/auth/service/service.go`
- Create: `backend/internal/modules/auth/http/handler.go`
- Create: `backend/internal/modules/users/service/profile_service.go`
- Create: `backend/internal/modules/options/http/handler.go`
- Create: `backend/internal/platform/authn/jwt.go`
- Create: `backend/internal/platform/authz/roles.go`
- Modify: `backend/internal/httpapi/router/router.go`
- Test: `backend/internal/modules/auth/service/service_test.go`
- Test: `backend/internal/platform/authn/jwt_test.go`

- [x] **Step 1: Write the failing JWT claims test**

```go
func TestIssueTokenContainsUserIDClaim(t *testing.T) {
    token, err := authn.IssueAccessToken(authn.ClaimsInput{
        UserID:   "01JUSER",
        Username: "alice",
        UserType: "student",
    })
    require.NoError(t, err)
    claims, err := authn.ParseAccessToken(token)
    require.NoError(t, err)
    require.Equal(t, "01JUSER", claims.UserID)
}
```

- [x] **Step 2: Implement `/api/v1/auth/login`, `/refresh`, and `/logout`**

```go
auth := v1.Group("/auth")
auth.POST("/login", authHandler.Login)
auth.POST("/refresh", authHandler.Refresh)
auth.POST("/logout", authMiddleware.Required(), authHandler.Logout)
auth.POST("/register/student", authHandler.RegisterStudent)
```

- [x] **Step 3: Implement `/api/v1/me/profile`, `/change-password`, and avatar endpoints**

```go
me := v1.Group("/me", authMiddleware.Required())
me.GET("/profile", meHandler.GetProfile)
me.PATCH("/profile", meHandler.UpdateProfile)
me.POST("/change-password", meHandler.ChangePassword)
me.POST("/avatar", meHandler.UploadAvatar)
me.DELETE("/avatar", meHandler.DeleteAvatar)
```

- [x] **Step 4: Implement organization and option endpoints**

```go
v1.GET("/options/organization/tree", optionsHandler.GetOrganizationTree)
v1.GET("/options/grades", optionsHandler.GetGrades)
v1.GET("/options/academic-years", authMiddleware.Required(), optionsHandler.GetAcademicYears)
v1.GET("/options/semesters", authMiddleware.Required(), optionsHandler.GetSemesters)
```

- [x] **Step 5: Run auth and platform tests**

Run: `cd backend && go test ./internal/modules/auth/... ./internal/platform/authn/... -v`
Expected: `PASS`

- [x] **Step 6: Commit**

```bash
git add backend/internal/modules/auth backend/internal/modules/users/service/profile_service.go backend/internal/modules/options backend/internal/platform/authn backend/internal/platform/authz backend/internal/httpapi/router/router.go
git commit -m "feat: implement auth profile and option backbone"
```

### Task 4: User Administration And User Import/Export

**Files:**
- Create: `backend/internal/modules/users/http/admin_handler.go`
- Create: `backend/internal/modules/users/service/admin_service.go`
- Create: `backend/internal/modules/users/repo/user_repo.go`
- Create: `backend/internal/modules/imports/service/user_import_service.go`
- Create: `backend/internal/modules/exports/service/user_export_service.go`
- Create: `backend/internal/modules/imports/http/user_import_handler.go`
- Test: `backend/internal/modules/users/service/admin_service_test.go`
- Test: `backend/internal/modules/imports/service/user_import_service_test.go`

- [x] **Step 1: Write the failing user import precheck test**

```go
func TestStudentImportPrecheckRejectsDuplicateStudentID(t *testing.T) {
    rows := []imports.UserRow{
        {Username: "a", StudentID: "20240001"},
        {Username: "b", StudentID: "20240001"},
    }
    result := service.PrecheckStudents(rows)
    require.Equal(t, 1, result.InvalidRows)
}
```

- [x] **Step 2: Add admin CRUD routes**

```go
v1.GET("/students", teacherOrAdmin, usersHandler.ListStudents)
v1.POST("/students", adminOnly, usersHandler.CreateStudent)
v1.PATCH("/students/:id", adminOnly, usersHandler.UpdateStudent)
v1.DELETE("/students/:id", adminOnly, usersHandler.DeleteStudent)
v1.GET("/teachers", adminOnly, usersHandler.ListTeachers)
v1.POST("/teachers", adminOnly, usersHandler.CreateTeacher)
```

- [x] **Step 3: Add status change, reset password, export, and template download routes**

```go
v1.PATCH("/users/:id/status", adminOnly, usersHandler.UpdateStatus)
v1.POST("/users/:id/reset-password", adminOnly, usersHandler.ResetPassword)
v1.GET("/students/export", teacherOrAdmin, usersHandler.ExportStudents)
v1.GET("/teachers/export", adminOnly, usersHandler.ExportTeachers)
v1.GET("/imports/templates/students", adminOnly, importHandler.StudentTemplate)
```

- [x] **Step 4: Implement two-step import endpoints with `job_id + file_hash`**

```go
v1.POST("/imports/students/precheck", adminOnly, importHandler.PrecheckStudents)
v1.POST("/imports/students/confirm", adminOnly, importHandler.ConfirmStudents)
v1.POST("/imports/teachers/precheck", adminOnly, importHandler.PrecheckTeachers)
v1.POST("/imports/teachers/confirm", adminOnly, importHandler.ConfirmTeachers)
```

- [x] **Step 5: Run user-domain tests**

Run: `cd backend && go test ./internal/modules/users/... ./internal/modules/imports/... ./internal/modules/exports/... -v`
Expected: `PASS`

- [x] **Step 6: Commit**

```bash
git add backend/internal/modules/users backend/internal/modules/imports backend/internal/modules/exports
git commit -m "feat: implement user administration and import export"
```

### Task 5: Activities, Participants, Files, And Review Flow

**Files:**
- Create: `backend/internal/modules/activities/service/activity_service.go`
- Create: `backend/internal/modules/activities/http/handler.go`
- Create: `backend/internal/modules/participants/service/service.go`
- Create: `backend/internal/modules/files/service/service.go`
- Create: `backend/internal/modules/files/http/handler.go`
- Create: `backend/internal/modules/activities/repo/activity_repo.go`
- Test: `backend/internal/modules/activities/service/activity_service_test.go`
- Test: `backend/internal/modules/participants/service/service_test.go`
- Test: `backend/internal/modules/files/service/service_test.go`

- [x] **Step 1: Write the failing activity status rule test**

```go
func TestUpdateDraftOnly(t *testing.T) {
    activity := domain.Activity{Status: domain.StatusApproved}
    err := activity.AllowStructuralEdit()
    require.ErrorIs(t, err, domain.ErrActivityStatusInvalid)
}
```

- [x] **Step 2: Implement activity CRUD plus submit, withdraw, copy, and transfer owner**

```go
v1.GET("/activities", authRequired, activityHandler.List)
v1.POST("/activities", authRequired, activityHandler.Create)
v1.GET("/activities/:id", authRequired, activityHandler.Get)
v1.PATCH("/activities/:id", authRequired, activityHandler.Update)
v1.DELETE("/activities/:id", authRequired, activityHandler.Delete)
v1.POST("/activities/:id/submit", authRequired, activityHandler.Submit)
v1.POST("/activities/:id/withdraw", authRequired, activityHandler.Withdraw)
v1.POST("/activities/:id/copy", authRequired, activityHandler.Copy)
v1.POST("/activities/:id/transfer-owner", teacherOrAdmin, activityHandler.TransferOwner)
```

- [x] **Step 3: Implement participants and file endpoints with draft-only structural edits**

```go
v1.GET("/activities/:id/participants", authRequired, participantHandler.List)
v1.POST("/activities/:id/participants", ownerTeacherAdmin, participantHandler.Add)
v1.POST("/activities/:id/participants/leave", authRequired, participantHandler.Leave)
v1.GET("/activities/:id/files", authRequired, fileHandler.List)
v1.POST("/activities/:id/files", ownerTeacherAdmin, fileHandler.Upload)
v1.GET("/files/:file_id/download", authRequired, fileHandler.Download)
```

- [x] **Step 4: Implement review flow with same-transaction application side effects stubbed behind an interface**

```go
type ApplicationHooks interface {
    SyncApprovedActivity(ctx context.Context, tx *gorm.DB, activityID string) error
}
```

- [x] **Step 5: Run activity-domain tests**

Run: `cd backend && go test ./internal/modules/activities/... ./internal/modules/participants/... ./internal/modules/files/... -v`
Expected: `PASS`

- [x] **Step 6: Commit**

```bash
git add backend/internal/modules/activities backend/internal/modules/participants backend/internal/modules/files
git commit -m "feat: implement activities participants and files"
```

### Task 6: Applications, Exports, Dashboard, Audit, And Devtools

**Files:**
- Create: `backend/internal/modules/applications/service/service.go`
- Create: `backend/internal/modules/applications/http/handler.go`
- Create: `backend/internal/modules/dashboard/http/handler.go`
- Create: `backend/internal/modules/devtools/http/handler.go`
- Create: `backend/internal/modules/audit/service/service.go`
- Create: `backend/internal/platform/auditlog/writer.go`
- Test: `backend/internal/modules/applications/service/service_test.go`
- Test: `backend/internal/modules/dashboard/service_test.go`
- Test: `backend/internal/platform/auditlog/writer_test.go`

- [x] **Step 1: Write the failing application sync test**

```go
func TestApprovedActivityCreatesOrRestoresApplications(t *testing.T) {
    effect := service.SyncApprovedActivity(testActivityID, []string{"u1", "u2"})
    require.Equal(t, 2, effect.AffectedCount)
}
```

- [x] **Step 2: Implement applications list, detail, stats, and export**

```go
v1.GET("/applications", authRequired, applicationHandler.List)
v1.GET("/applications/stats", authRequired, applicationHandler.Stats)
v1.GET("/applications/:id", authRequired, applicationHandler.Get)
v1.GET("/applications/export", authRequired, applicationHandler.Export)
```

- [x] **Step 3: Implement dashboard and devtools**

```go
v1.GET("/dashboard/overview", authRequired, dashboardHandler.Overview)
v1.GET("/healthz", devtoolsHandler.Healthz)
v1.GET("/devtools/health", adminOnly, devtoolsHandler.DetailedHealth)
v1.GET("/devtools/logs", adminOnly, devtoolsHandler.Logs)
v1.GET("/devtools/audit-logs", adminOnly, devtoolsHandler.AuditLogs)
```

- [x] **Step 4: Enforce append-only audit writes and filtered export**

```go
func (w *Writer) Record(ctx context.Context, entry Entry) error {
    return w.db.WithContext(ctx).Create(&entry).Error
}
```

- [x] **Step 5: Run service-level tests**

Run: `cd backend && go test ./internal/modules/applications/... ./internal/modules/dashboard/... ./internal/modules/devtools/... ./internal/platform/auditlog/... -v`
Expected: `PASS`

- [x] **Step 6: Commit**

```bash
git add backend/internal/modules/applications backend/internal/modules/dashboard backend/internal/modules/devtools backend/internal/modules/audit backend/internal/platform/auditlog
git commit -m "feat: implement applications dashboard audit and devtools"
```

### Task 7: Frontend Shell, Routing, Auth Client, And Shared UI Migration

**Files:**
- Modify: `frontend/package.json`
- Create: `frontend/src/app/providers/AppProviders.tsx`
- Create: `frontend/src/app/providers/QueryProvider.tsx`
- Create: `frontend/src/app/router/index.tsx`
- Create: `frontend/src/app/layouts/AuthLayout.tsx`
- Create: `frontend/src/app/layouts/AppLayout.tsx`
- Create: `frontend/src/app/layouts/AdminLayout.tsx`
- Create: `frontend/src/shared/api/http.ts`
- Create: `frontend/src/shared/api/query-client.ts`
- Create: `frontend/src/shared/types/api.ts`
- Modify: `frontend/src/main.tsx`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/app/router/index.test.tsx`

- [x] **Step 1: Add the missing frontend state and test dependencies**

```json
{
  "@tanstack/react-query": "^5.0.0",
  "zustand": "^5.0.0",
  "vitest": "^3.0.0",
  "@testing-library/react": "^16.0.0",
  "@testing-library/user-event": "^14.0.0"
}
```

- [x] **Step 2: Replace the legacy API client with the new `/api/v1` contract**

```ts
export const http = axios.create({
  baseURL: "/api/v1",
  timeout: 15000,
});
```

- [x] **Step 3: Add router and provider shell**

```tsx
export function AppProviders() {
  return (
    <QueryProvider>
      <RouterProvider router={router} />
    </QueryProvider>
  );
}
```

- [x] **Step 4: Add a failing router access test**

```tsx
it("redirects anonymous users to /login", async () => {
  render(<AppProviders />);
  expect(await screen.findByText(/login/i)).toBeInTheDocument();
});
```

- [x] **Step 5: Run frontend typecheck and router test**

Run: `cd frontend && pnpm exec vitest run src/app/router/index.test.tsx && pnpm build`
Expected: `PASS` and a successful Vite build

- [x] **Step 6: Commit**

```bash
git add frontend/package.json frontend/src/app frontend/src/shared/api frontend/src/shared/types frontend/src/App.tsx frontend/src/main.tsx
git commit -m "refactor: add frontend app shell and shared api client"
```

### Task 8: Frontend Business Pages And Import Flows

**Files:**
- Create: `frontend/src/modules/auth/**`
- Create: `frontend/src/modules/dashboard/**`
- Create: `frontend/src/modules/users/**`
- Create: `frontend/src/modules/activities/**`
- Create: `frontend/src/modules/applications/**`
- Create: `frontend/src/modules/imports/**`
- Create: `frontend/src/modules/devtools/**`
- Modify: `frontend/src/components/ui/**` as shared primitives only
- Test: `frontend/src/modules/activities/pages/ActivitiesPage.test.tsx`
- Test: `frontend/src/modules/applications/pages/ApplicationsPage.test.tsx`

- [x] **Step 1: Migrate login and register pages to the new module layout**

```tsx
export function LoginPage() {
  const mutation = useLoginMutation();
  return <LoginForm onSubmit={mutation.mutate} />;
}
```

- [x] **Step 2: Implement dashboard, students, teachers, and profile pages against the new API layer**

```tsx
const { data } = useQuery({
  queryKey: queryKeys.dashboard.overview(search),
  queryFn: api.dashboard.getOverview,
});
```

- [x] **Step 3: Implement activities list, detail, editor, review dialog, and owner transfer dialog**

```tsx
const queries = useSuspenseQueries({
  queries: [
    activitiesQueries.detail(id),
    activitiesQueries.participants(id),
    activitiesQueries.files(id),
  ],
});
```

- [x] **Step 4: Implement application list/detail and unified import wizard**

```tsx
<ImportWizardDialog
  onPrecheck={importsApi.precheckActivities}
  onConfirm={importsApi.confirmActivities}
/>
```

- [x] **Step 5: Run focused frontend tests plus build**

Run: `cd frontend && pnpm exec vitest run src/modules/activities/pages/ActivitiesPage.test.tsx src/modules/applications/pages/ApplicationsPage.test.tsx && pnpm build`
Expected: `PASS`

- [x] **Step 6: Commit**

```bash
git add frontend/src/modules frontend/src/components/ui
git commit -m "refactor: migrate business pages to modular frontend"
```

### Task 9: Cutover, Legacy Shutdown, And Full Verification

**Files:**
- Modify: `docker-compose.yml`
- Modify: `Makefile`
- Modify: `README.md`
- Modify: `docs/refactor-design.md`
- Create: `docs/cutover-checklist.md`
- Create: `backend/tests/e2e/auth_flow_test.go`
- Create: `backend/tests/e2e/activity_flow_test.go`
- Create: `backend/tests/e2e/import_flow_test.go`

- [x] **Step 1: Remove legacy gateway/service runtime from default compose flow**

```yaml
frontend:
  depends_on:
    backend:
      condition: service_started
```

- [x] **Step 2: Add end-to-end tests for the acceptance baseline**

```go
func TestDraftToApprovedGeneratesApplications(t *testing.T) {
    // create draft -> submit -> approve -> assert applications exist
}
```

- [x] **Step 3: Update project automation**

```make
test-e2e:
	cd backend && go test ./tests/e2e/... -v

up-refactor:
	docker compose up -d postgres redis backend frontend
```

- [x] **Step 4: Run the full verification suite**

Run: `make test-backend && cd frontend && pnpm build && cd .. && docker compose up -d postgres redis backend && cd backend && go test ./tests/e2e/... -v`
Expected: backend unit tests pass, frontend build succeeds, e2e tests pass against the new stack

- [x] **Step 5: Commit**

```bash
git add docker-compose.yml Makefile README.md docs/cutover-checklist.md docs/refactor-design.md backend/tests/e2e
git commit -m "chore: cut over to modular monolith stack"
```

## Self-Review Checklist

- Confirm every confirmed rule from `docs/refactor-requirements.md` sections `1.1`, `17`, `19`, `20`, and `21` maps to at least one task above.
- Do not implement any new compatibility layer for `TEST_DATA_MODE`, `X-Internal-Service`, old `/permissions`, or old avatar public URLs.
- Keep every structural edit in activities, participants, and files restricted to `draft`.
- Verify that imports are always `precheck -> confirm -> same file hash -> idempotent`.
- Verify that approved activity review and application sync stay in the same DB transaction.
- Verify that frontend pages consume `/api/v1` only and stop relying on `/api`, `/refresh-token`, `/users/profile`, and other legacy route shapes.
