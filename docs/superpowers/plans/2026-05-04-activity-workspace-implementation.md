# Activity Workspace Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the complete activity-domain workspace: category schema APIs, schema-aware draft create/edit, activity stats, import/export, and the frontend `/activities`, `/activities/new`, and `/activities/:id/edit` flows.

**Architecture:** Extend the existing modular monolith without introducing a generic form platform. The backend remains the single validation authority, while the frontend adds an activity-domain schema runtime that mirrors the same closed rule model for usability. The work is decomposed into backend schema/query support, backend activity workflow extensions, frontend runtime/pages, then end-to-end coverage.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, Redis, React, TypeScript, Vite, Vitest

---

## File Map

Backend files likely to change or be created:

- Create: `backend/internal/modules/activities/repo/category_schema_repo.go`
- Create: `backend/internal/modules/activities/repo/category_schema_repo_test.go`
- Modify: `backend/internal/modules/activities/repo/gorm_repo.go`
- Modify: `backend/internal/modules/activities/service/activity_service.go`
- Create: `backend/internal/modules/activities/service/schema_rules.go`
- Create: `backend/internal/modules/activities/service/schema_rules_test.go`
- Modify: `backend/internal/modules/activities/http/handler.go`
- Modify: `backend/internal/httpapi/router/router.go`
- Modify: `backend/internal/bootstrap/app.go`
- Modify: `backend/internal/modules/imports/service/user_import_service.go`
- Create: `backend/internal/modules/imports/service/activity_import_service.go`
- Create: `backend/internal/modules/imports/service/activity_import_service_test.go`
- Modify: `backend/internal/modules/imports/http/user_import_handler.go`
- Create: `backend/internal/modules/imports/http/activity_import_handler.go`
- Create: `backend/internal/modules/imports/http/activity_import_handler_test.go`
- Create: `backend/internal/modules/exports/service/activity_export_service.go`
- Create: `backend/internal/modules/exports/service/activity_export_service_test.go`
- Create: `backend/internal/modules/dashboard/activity_stats_service.go`
- Create: `backend/internal/modules/dashboard/activity_stats_service_test.go`
- Modify: `backend/tests/e2e/activity_flow_test.go`
- Create: `backend/tests/e2e/activity_workspace_flow_test.go`

Frontend files likely to change or be created:

- Create: `frontend/src/modules/activities/api/activity-categories.ts`
- Create: `frontend/src/modules/activities/api/activity-stats.ts`
- Modify: `frontend/src/modules/activities/api/activities.ts`
- Create: `frontend/src/modules/activities/schema/types.ts`
- Create: `frontend/src/modules/activities/schema/rule-evaluator.ts`
- Create: `frontend/src/modules/activities/schema/rule-evaluator.test.ts`
- Create: `frontend/src/modules/activities/components/ActivityDynamicField.tsx`
- Create: `frontend/src/modules/activities/components/ActivitySchemaForm.tsx`
- Create: `frontend/src/modules/activities/components/ActivityImportDialog.tsx`
- Modify: `frontend/src/modules/activities/pages/ActivitiesPage.tsx`
- Create: `frontend/src/modules/activities/pages/ActivityCreatePage.tsx`
- Create: `frontend/src/modules/activities/pages/ActivityEditPage.tsx`
- Create: `frontend/src/modules/activities/pages/ActivityCreatePage.test.tsx`
- Create: `frontend/src/modules/activities/pages/ActivityEditPage.test.tsx`
- Modify: `frontend/src/app/router/index.tsx`

Docs / supporting:

- Modify: `docs/refactor-design.md`
- Modify: `README.md`

---

### Task 1: Add Activity Category Schema Repository and API Contracts

**Files:**
- Create: `backend/internal/modules/activities/repo/category_schema_repo.go`
- Create: `backend/internal/modules/activities/repo/category_schema_repo_test.go`
- Modify: `backend/internal/modules/activities/http/handler.go`
- Modify: `backend/internal/httpapi/router/router.go`
- Modify: `backend/internal/bootstrap/app.go`

- [ ] **Step 1: Write the failing repository test**

```go
func TestCategorySchemaRepoGetsEnabledCategorySchema(t *testing.T) {
	t.Skip("replace with TEST_DATABASE_DSN-backed repository test")
}
```

Add assertions for:
- enabled categories only
- schema fields sorted by `display_order`
- rules JSON decoded into strongly-typed structs

- [ ] **Step 2: Run the repository test to verify it fails**

Run: `cd backend && go test ./internal/modules/activities/repo -run TestCategorySchemaRepoGetsEnabledCategorySchema -v`

Expected:
- FAIL because the repository file or implementation does not exist yet

- [ ] **Step 3: Implement the category schema repository**

Create a focused repository with these shapes:

```go
type CategorySummary struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

type SchemaField struct {
	Key          string          `json:"key"`
	Label        string          `json:"label"`
	Type         string          `json:"type"`
	Required     bool            `json:"required"`
	DefaultValue json.RawMessage `json:"default_value"`
	Placeholder  string          `json:"placeholder"`
	Options      []SchemaOption  `json:"options"`
	Rules        FieldRules      `json:"rules"`
	DisplayOrder int             `json:"display_order"`
}
```

Use `activity_category_templates` and `activity_field_schemas`. Decode `validation_rules` JSON into `FieldRules`.

- [ ] **Step 4: Add API handlers and routes**

Add routes:

```go
api.GET("/activity-categories", deps.ActivityHandler.ListCategories)
api.GET("/activity-categories/:category/schema", deps.ActivityHandler.GetCategorySchema)
```

Handler response shape must match the approved spec.

- [ ] **Step 5: Run focused backend tests**

Run:

```bash
cd backend
go test ./internal/modules/activities/repo ./internal/modules/activities/http -v
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/modules/activities/repo/category_schema_repo.go \
        backend/internal/modules/activities/repo/category_schema_repo_test.go \
        backend/internal/modules/activities/http/handler.go \
        backend/internal/httpapi/router/router.go \
        backend/internal/bootstrap/app.go
git commit -m "feat: add activity category schema api"
```

---

### Task 2: Add Schema Rule Evaluation to Activity Create and Edit

**Files:**
- Create: `backend/internal/modules/activities/service/schema_rules.go`
- Create: `backend/internal/modules/activities/service/schema_rules_test.go`
- Modify: `backend/internal/modules/activities/service/activity_service.go`
- Modify: `backend/internal/modules/activities/repo/gorm_repo.go`
- Modify: `backend/internal/modules/activities/http/handler.go`

- [ ] **Step 1: Write failing service tests for schema validation**

Add tests for:
- hidden field values are ignored or stripped
- enabled/disabled rule is respected
- cross-field `end_date >= start_date` validation fails correctly
- unsupported field type returns `ACTIVITY_SCHEMA_INVALID`

Example test skeleton:

```go
func TestValidateDetailsRejectsInvalidCrossFieldRange(t *testing.T) {
	schema := CategorySchema{Fields: []SchemaField{
		{Key: "start_date", Type: "date"},
		{Key: "end_date", Type: "date", Rules: FieldRules{
			Validate: []ValidateRule{{Type: "compare", Left: "end_date", Op: "gte", Right: "start_date", Message: "end date cannot be earlier than start date"}},
		}},
	}}

	err := ValidateActivityDetails(schema, map[string]any{
		"start_date": "2026-06-01",
		"end_date":   "2026-05-01",
	})

	if err == nil {
		t.Fatal("expected validation error")
	}
}
```

- [ ] **Step 2: Run the focused service tests**

Run: `cd backend && go test ./internal/modules/activities/service -run 'TestValidateDetails|TestCreate|TestUpdate' -v`

Expected:
- FAIL before implementation

- [ ] **Step 3: Implement the backend schema validator**

Implement a closed rule evaluator:

```go
type FieldCondition struct {
	Field string          `json:"field"`
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value"`
}

type ValidateRule struct {
	Type    string `json:"type"`
	Left    string `json:"left"`
	Op      string `json:"op"`
	Right   string `json:"right"`
	Message string `json:"message"`
}
```

Validation contract:
- evaluate `visible_when`
- evaluate `enabled_when`
- strip or ignore values for fields not visible
- enforce required only when field is visible
- enforce cross-field validation using the approved operator set

- [ ] **Step 4: Wire schema validation into create and update**

Before persisting `details_json`, the service must:
- load schema by category key or category id
- validate base fields
- normalize `details`
- persist only normalized `details`

Return stable errors mapped to:
- `ACTIVITY_CATEGORY_NOT_FOUND`
- `ACTIVITY_SCHEMA_INVALID`
- `ACTIVITY_DETAILS_INVALID`

- [ ] **Step 5: Run focused tests and then all activity service tests**

Run:

```bash
cd backend
go test ./internal/modules/activities/service -v
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/modules/activities/service/schema_rules.go \
        backend/internal/modules/activities/service/schema_rules_test.go \
        backend/internal/modules/activities/service/activity_service.go \
        backend/internal/modules/activities/repo/gorm_repo.go \
        backend/internal/modules/activities/http/handler.go
git commit -m "feat: validate activity details against category schema"
```

---

### Task 3: Add Activity Statistics and Export

**Files:**
- Create: `backend/internal/modules/dashboard/activity_stats_service.go`
- Create: `backend/internal/modules/dashboard/activity_stats_service_test.go`
- Create: `backend/internal/modules/exports/service/activity_export_service.go`
- Create: `backend/internal/modules/exports/service/activity_export_service_test.go`
- Modify: `backend/internal/modules/activities/http/handler.go`
- Modify: `backend/internal/bootstrap/app.go`

- [ ] **Step 1: Write failing tests for activity stats aggregation**

Add tests covering:
- status counts
- category counts
- academic year counts
- semester counts
- `mine_count`
- `participated_count`

- [ ] **Step 2: Write failing tests for CSV export flattening**

Add tests covering:
- base columns
- dynamic `details.<field_key>` columns
- filtered export result set

- [ ] **Step 3: Run focused tests to verify failure**

Run:

```bash
cd backend
go test ./internal/modules/dashboard ./internal/modules/exports/service -v
```

Expected:
- FAIL before implementation

- [ ] **Step 4: Implement stats service and export service**

Stats response shape must match the approved spec. Export must flatten dynamic fields using:

```text
details.location
details.activity_mode
details.max_participants
```

Do not implement a report engine. Keep it to list-page statistics and CSV export only.

- [ ] **Step 5: Expose `/activities/stats` and `/activities/export`**

Update the activity HTTP handler with:
- `Stats`
- `Export`

Ensure filter parsing matches the current list-page query model.

- [ ] **Step 6: Run focused tests**

Run:

```bash
cd backend
go test ./internal/modules/dashboard ./internal/modules/exports/service ./internal/modules/activities/http -v
```

Expected:
- PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/modules/dashboard/activity_stats_service.go \
        backend/internal/modules/dashboard/activity_stats_service_test.go \
        backend/internal/modules/exports/service/activity_export_service.go \
        backend/internal/modules/exports/service/activity_export_service_test.go \
        backend/internal/modules/activities/http/handler.go \
        backend/internal/bootstrap/app.go
git commit -m "feat: add activity stats and export"
```

---

### Task 4: Add Activity Import Precheck and Confirm

**Files:**
- Create: `backend/internal/modules/imports/service/activity_import_service.go`
- Create: `backend/internal/modules/imports/service/activity_import_service_test.go`
- Create: `backend/internal/modules/imports/http/activity_import_handler.go`
- Create: `backend/internal/modules/imports/http/activity_import_handler_test.go`
- Modify: `backend/internal/httpapi/router/router.go`
- Modify: `backend/internal/bootstrap/app.go`

- [ ] **Step 1: Write failing tests for activity import precheck**

Cover:
- required fixed columns
- unknown category handling
- schema validation on dynamic columns
- per-row error output

- [ ] **Step 2: Write failing tests for confirm idempotency**

Cover:
- same `job_id + file_hash` confirm returns stable result
- no duplicate activities created

- [ ] **Step 3: Run focused tests**

Run:

```bash
cd backend
go test ./internal/modules/imports/service ./internal/modules/imports/http -run 'Activity' -v
```

Expected:
- FAIL before implementation

- [ ] **Step 4: Implement activity import service**

Implementation rules:
- follow the existing import job pattern
- use flat spreadsheet columns
- map dynamic columns by `details.<field_key>`
- validate against category schema before confirm

Suggested row shape:

```go
type ActivityImportRow struct {
	Title        string
	Description  string
	Category     string
	AcademicYear string
	Semester     string
	StartDate    string
	EndDate      string
	Details      map[string]any
}
```

- [ ] **Step 5: Implement activity import handler and routes**

Add routes:

```go
protected.POST("/imports/activities/precheck", deps.ImportHandler.ActivityPrecheck)
protected.POST("/imports/activities/confirm", deps.ImportHandler.ActivityConfirm)
```

- [ ] **Step 6: Run focused import tests**

Run:

```bash
cd backend
go test ./internal/modules/imports/service ./internal/modules/imports/http -v
```

Expected:
- PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/modules/imports/service/activity_import_service.go \
        backend/internal/modules/imports/service/activity_import_service_test.go \
        backend/internal/modules/imports/http/activity_import_handler.go \
        backend/internal/modules/imports/http/activity_import_handler_test.go \
        backend/internal/httpapi/router/router.go \
        backend/internal/bootstrap/app.go
git commit -m "feat: add activity import workflow"
```

---

### Task 5: Build the Frontend Activity Schema Runtime

**Files:**
- Create: `frontend/src/modules/activities/schema/types.ts`
- Create: `frontend/src/modules/activities/schema/rule-evaluator.ts`
- Create: `frontend/src/modules/activities/schema/rule-evaluator.test.ts`
- Create: `frontend/src/modules/activities/components/ActivityDynamicField.tsx`
- Create: `frontend/src/modules/activities/components/ActivitySchemaForm.tsx`

- [ ] **Step 1: Write failing rule-evaluator tests**

Cover:
- `visible_when`
- `enabled_when`
- `compare` validation
- `in` / `not_empty` handling

Example test shape:

```ts
it("hides offline-only fields when mode is online", () => {
  const visible = isFieldVisible(
    [{ field: "activity_mode", op: "eq", value: "offline" }],
    { activity_mode: "online" },
  )
  expect(visible).toBe(false)
})
```

- [ ] **Step 2: Run the failing tests**

Run:

```bash
cd frontend
npm exec --yes pnpm -- exec vitest run src/modules/activities/schema/rule-evaluator.test.ts
```

Expected:
- FAIL before implementation

- [ ] **Step 3: Implement the rule evaluator**

Expose focused helpers:

```ts
export function isFieldVisible(...)
export function isFieldEnabled(...)
export function validateSchemaValues(...)
```

Do not build a general-purpose rules engine. Keep the API activity-domain-specific.

- [ ] **Step 4: Implement the dynamic field renderer**

Map schema field types to concrete UI components:
- text input
- textarea
- number input
- date input
- select
- radio group
- checkbox group
- boolean switch

- [ ] **Step 5: Implement the activity schema form wrapper**

Responsibilities:
- render fields by schema order
- recalculate visibility/enabled state on change
- normalize values before submit
- surface field-level validation messages

- [ ] **Step 6: Run focused frontend tests**

Run:

```bash
cd frontend
npm exec --yes pnpm -- exec vitest run src/modules/activities/schema/rule-evaluator.test.ts
```

Expected:
- PASS

- [ ] **Step 7: Commit**

```bash
git add frontend/src/modules/activities/schema/types.ts \
        frontend/src/modules/activities/schema/rule-evaluator.ts \
        frontend/src/modules/activities/schema/rule-evaluator.test.ts \
        frontend/src/modules/activities/components/ActivityDynamicField.tsx \
        frontend/src/modules/activities/components/ActivitySchemaForm.tsx
git commit -m "feat: add activity schema runtime"
```

---

### Task 6: Build Activity Create and Edit Pages

**Files:**
- Create: `frontend/src/modules/activities/api/activity-categories.ts`
- Modify: `frontend/src/modules/activities/api/activities.ts`
- Create: `frontend/src/modules/activities/pages/ActivityCreatePage.tsx`
- Create: `frontend/src/modules/activities/pages/ActivityEditPage.tsx`
- Create: `frontend/src/modules/activities/pages/ActivityCreatePage.test.tsx`
- Create: `frontend/src/modules/activities/pages/ActivityEditPage.test.tsx`
- Modify: `frontend/src/app/router/index.tsx`

- [ ] **Step 1: Write failing page tests**

Create tests for:
- category selection loads schema on `/activities/new`
- save draft redirects to `/activities/:id/edit`
- edit page hydrates fixed fields and schema-driven fields
- submit for review is available only in draft flow

- [ ] **Step 2: Run the failing page tests**

Run:

```bash
cd frontend
npm exec --yes pnpm -- exec vitest run \
  src/modules/activities/pages/ActivityCreatePage.test.tsx \
  src/modules/activities/pages/ActivityEditPage.test.tsx
```

Expected:
- FAIL before implementation

- [ ] **Step 3: Implement activity categories and schema API client**

Add focused API helpers:

```ts
export async function listActivityCategories()
export async function getActivityCategorySchema(categoryKey: string)
```

- [ ] **Step 4: Implement `/activities/new`**

The page must:
- load categories
- fetch schema after category selection
- render fixed form section + dynamic schema section
- call `POST /activities`
- redirect to `/activities/:id/edit`

- [ ] **Step 5: Implement `/activities/:id/edit`**

The page must:
- load activity detail
- load category schema
- populate fixed and dynamic values
- save with `PATCH /activities/:id`
- allow submit for review

- [ ] **Step 6: Wire routes**

Add:

```tsx
<Route path="/activities/new" element={<ActivityCreatePage />} />
<Route path="/activities/:id/edit" element={<ActivityEditPage />} />
```

- [ ] **Step 7: Run focused tests**

Run:

```bash
cd frontend
npm exec --yes pnpm -- exec vitest run \
  src/modules/activities/pages/ActivityCreatePage.test.tsx \
  src/modules/activities/pages/ActivityEditPage.test.tsx \
  src/app/router/index.test.tsx
```

Expected:
- PASS

- [ ] **Step 8: Commit**

```bash
git add frontend/src/modules/activities/api/activity-categories.ts \
        frontend/src/modules/activities/api/activities.ts \
        frontend/src/modules/activities/pages/ActivityCreatePage.tsx \
        frontend/src/modules/activities/pages/ActivityEditPage.tsx \
        frontend/src/modules/activities/pages/ActivityCreatePage.test.tsx \
        frontend/src/modules/activities/pages/ActivityEditPage.test.tsx \
        frontend/src/app/router/index.tsx
git commit -m "feat: add activity create and edit pages"
```

---

### Task 7: Upgrade Activity List Page with Stats, Import, and Export

**Files:**
- Create: `frontend/src/modules/activities/api/activity-stats.ts`
- Create: `frontend/src/modules/activities/components/ActivityImportDialog.tsx`
- Modify: `frontend/src/modules/activities/pages/ActivitiesPage.tsx`
- Modify: `frontend/src/modules/imports/api/user-imports.ts` (only if shared upload utility reuse makes sense; otherwise keep activity import separate)

- [ ] **Step 1: Write failing list-page tests**

Cover:
- stats cards render from API
- new activity button routes to `/activities/new`
- export action calls the export endpoint
- import dialog performs precheck then confirm

- [ ] **Step 2: Run the failing tests**

Run:

```bash
cd frontend
npm exec --yes pnpm -- exec vitest run src/modules/activities/pages/ActivitiesPage.test.tsx
```

Expected:
- FAIL before implementation

- [ ] **Step 3: Implement stats API and import dialog**

The import dialog must follow the existing formal import pattern:
- upload file
- show precheck summary/errors
- explicit confirm

- [ ] **Step 4: Upgrade `ActivitiesPage`**

Add:
- first-screen stats fetch
- stats cards
- category/year/semester filters
- export button
- import button
- `New Activity` button

- [ ] **Step 5: Run focused frontend tests**

Run:

```bash
cd frontend
npm exec --yes pnpm -- exec vitest run src/modules/activities/pages/ActivitiesPage.test.tsx
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/modules/activities/api/activity-stats.ts \
        frontend/src/modules/activities/components/ActivityImportDialog.tsx \
        frontend/src/modules/activities/pages/ActivitiesPage.tsx
git commit -m "feat: complete activity list workspace"
```

---

### Task 8: End-to-End Coverage and Docs Alignment

**Files:**
- Modify: `backend/tests/e2e/activity_flow_test.go`
- Create: `backend/tests/e2e/activity_workspace_flow_test.go`
- Modify: `docs/refactor-design.md`
- Modify: `README.md`

- [ ] **Step 1: Write or extend e2e coverage**

Add an e2e flow covering:
- create draft from category schema
- edit draft values
- submit after schema validation passes
- export activities
- import activities through precheck + confirm

- [ ] **Step 2: Run the new e2e tests to confirm initial failure**

Run:

```bash
cd backend
go test ./tests/e2e/... -run 'ActivityWorkspace|Activity' -v
```

Expected:
- FAIL until backend/frontend/API assumptions are aligned

- [ ] **Step 3: Fix any last-mile contract mismatches**

Typical final-mile items:
- payload shape mismatches
- export header names
- import error-code shape
- schema response naming inconsistencies

Keep fixes minimal and local.

- [ ] **Step 4: Update docs**

Update:
- `docs/refactor-design.md` to reflect the delivered activity workspace endpoints/pages
- `README.md` to mention the activity workspace flows if needed

- [ ] **Step 5: Run full verification**

Run:

```bash
make test-backend
make test-e2e
make test-frontend
git diff --check
```

Expected:
- all commands exit `0`

- [ ] **Step 6: Commit**

```bash
git add backend/tests/e2e/activity_flow_test.go \
        backend/tests/e2e/activity_workspace_flow_test.go \
        docs/refactor-design.md \
        README.md
git commit -m "test: cover activity workspace flow"
```

---

## Self-Review

Spec coverage check:

- category/schema APIs: Task 1
- schema-aware create/edit validation: Task 2
- activity stats and export: Task 3
- activity import workflow: Task 4
- frontend activity schema runtime: Task 5
- `/activities/new` and `/activities/:id/edit`: Task 6
- upgraded activity list with stats/import/export: Task 7
- e2e and doc alignment: Task 8

Placeholder scan:

- no `TBD` / `TODO` placeholders are left in task steps
- every task includes files, commands, and expected outcomes

Type consistency check:

- backend rule model uses `FieldRules`, `FieldCondition`, `ValidateRule`
- frontend runtime mirrors the same concepts in `types.ts` and `rule-evaluator.ts`
- import/export column strategy is consistently `details.<field_key>`

