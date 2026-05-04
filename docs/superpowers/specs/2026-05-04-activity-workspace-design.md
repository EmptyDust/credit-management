# Activity Workspace Design

## 1. Context

`v2` has completed the persistence foundation:

- `Postgres + Redis + local storage` are the default runtime dependencies.
- users, sessions, activities, participants, applications, files, audit logs, dashboard, and devtools health are persisted.
- runtime integration verification already covers restart-stable state.

The largest remaining gap between the current implementation and the product/design baseline is the activity domain. The current system still lacks:

- `/activities/new`
- `/activities/:id/edit`
- category schema-driven details editing
- activity statistics
- activity import/export
- complete frontend wiring for the full activity workspace

This spec defines the next project slice: a complete activity workspace for creation, editing, statistics, import, and export.

## 2. Goal

Deliver a complete activity-domain workspace across backend and frontend so that authorized users can:

- create draft activities from category-specific schemas
- edit draft activities in a dedicated workspace
- use conditional field visibility, field dependencies, and cross-field validation
- view activity statistics on the list page
- import and export activities through the formal import flow

This slice must stay inside the activity domain. It must not expand into a general low-code platform, broader devtools work, or unrelated application-domain changes.

## 3. Scope

### 3.1 In Scope

Backend:

- `GET /api/v1/activity-categories`
- `GET /api/v1/activity-categories/:category/schema`
- `GET /api/v1/activities/stats`
- `GET /api/v1/activities/export`
- `POST /api/v1/imports/activities/precheck`
- `POST /api/v1/imports/activities/confirm`
- complete schema-aware `POST /api/v1/activities`
- complete schema-aware `PATCH /api/v1/activities/:id`
- `GET /api/v1/activities/:id` aligned for the edit workspace

Frontend:

- `/activities`
- `/activities/new`
- `/activities/:id/edit`
- activity import dialog
- activity export action
- statistics cards and filters on the activity list page
- category-schema-driven activity editing runtime

Rules:

- only `draft` activities allow structural editing
- owner can edit own draft activities
- teacher/admin can edit any activity according to the already confirmed permission baseline
- schema rules support:
  - conditional visibility
  - enabled/disabled state derived from other fields
  - cross-field validation

### 3.2 Out of Scope

- generic form-engine or low-code platform extraction
- activity template system
- object storage upgrade
- audit/devtools feature expansion
- application-domain new features
- activity workflow changes outside the already confirmed rules

## 4. Approaches Considered

### 4.1 Full Generic Dynamic Form Engine

Build a reusable schema platform that could later power multiple domains.

Trade-offs:

- strongest long-term abstraction
- highest scope and risk
- would turn this slice into a platform project instead of an activity-domain delivery

Rejected because it is too broad for the immediate product gap.

### 4.2 Activity-Domain Dedicated Schema Runtime

Build a schema runtime specifically for activity categories, with a constrained rule model.

Trade-offs:

- satisfies the activity workspace goal now
- keeps rule syntax explicit and testable
- avoids over-engineering
- may need future extraction if other domains later need the same runtime

Recommended because it delivers the required behavior without creating a larger platform prematurely.

### 4.3 Fixed Forms First, Schema Later

Ship `/activities/new` and `/activities/:id/edit` as hand-built forms first, then abstract schema support later.

Trade-offs:

- fastest short-term UI delivery
- creates immediate duplication and likely rework
- conflicts with the confirmed requirement to support schema-driven conditional behavior now

Rejected because it pushes the real requirement into a second rewrite.

## 5. Architecture Decision

This slice will implement an **activity-domain dedicated schema runtime**.

Key implications:

- schema interpretation is limited to the activity domain
- rule evaluation is declarative, not script-based
- backend remains the final validation source
- frontend performs matching pre-validation for usability but does not become the source of truth

## 6. Backend Design

### 6.1 New / Expanded Endpoints

#### `GET /api/v1/activity-categories`

Purpose:

- return enabled activity categories for list filters and new-activity entry

Response shape:

```json
{
  "items": [
    {
      "key": "innovation",
      "label": "Innovation",
      "sort_order": 10
    }
  ]
}
```

#### `GET /api/v1/activity-categories/:category/schema`

Purpose:

- return the schema for one activity category

Response shape:

```json
{
  "category": {
    "key": "innovation",
    "label": "Innovation"
  },
  "schema": {
    "version": 1,
    "fields": [
      {
        "key": "location",
        "label": "Location",
        "type": "text",
        "required": true,
        "default_value": "",
        "placeholder": "Enter location",
        "options": [],
        "rules": {
          "visible_when": [],
          "enabled_when": [],
          "validate": []
        },
        "display_order": 10
      }
    ]
  }
}
```

#### `POST /api/v1/activities`

Purpose:

- first draft creation

Behavior:

- validates base fields and schema-driven `details`
- creates the draft activity
- returns the created activity payload with stable API field naming

#### `PATCH /api/v1/activities/:id`

Purpose:

- update an existing draft activity

Behavior:

- allowed only for structurally editable activities
- re-validates `details` against the current category schema
- persists only valid schema fields

#### `GET /api/v1/activities/stats`

Purpose:

- power statistics cards and filter summary on the list page

Response shape:

```json
{
  "status_counts": {
    "draft": 3,
    "pending_review": 5,
    "approved": 12,
    "rejected": 1
  },
  "category_counts": [
    { "key": "general", "label": "General", "count": 8 },
    { "key": "innovation", "label": "Innovation", "count": 13 }
  ],
  "academic_year_counts": [
    { "value": "2025-2026", "count": 15 }
  ],
  "semester_counts": [
    { "value": "autumn", "count": 11 },
    { "value": "spring", "count": 4 }
  ],
  "mine_count": 6,
  "participated_count": 9
}
```

#### `GET /api/v1/activities/export`

Purpose:

- export filtered activities as CSV

Export includes:

- fixed activity fields
- flattened dynamic fields using `details.<field_key>` column naming

#### `POST /api/v1/imports/activities/precheck`

Purpose:

- validate activity import file and produce a formal precheck job

Behavior:

- uses the existing import-job pattern
- validates both fixed fields and flattened dynamic fields

#### `POST /api/v1/imports/activities/confirm`

Purpose:

- formally persist prechecked activity imports

Behavior:

- keeps the existing job-id + file-hash binding rule
- remains idempotent for repeated confirm requests

### 6.2 Schema Model

This slice does not introduce a scriptable DSL. Rules stay declarative and closed.

#### Field Types

Supported field types:

- `text`
- `textarea`
- `number`
- `date`
- `select`
- `radio`
- `checkbox`
- `boolean`

#### Rule Model

Schema rules use an activity-specific structure:

```json
{
  "visible_when": [
    { "field": "activity_mode", "op": "eq", "value": "offline" }
  ],
  "enabled_when": [
    { "field": "approval_needed", "op": "eq", "value": true }
  ],
  "validate": [
    {
      "type": "compare",
      "left": "end_date",
      "op": "gte",
      "right": "start_date",
      "message": "End date cannot be earlier than start date"
    }
  ]
}
```

Supported condition operators:

- `eq`
- `neq`
- `in`
- `not_in`
- `gt`
- `gte`
- `lt`
- `lte`
- `contains`
- `not_empty`

Supported validation rule categories:

- required-field validation
- numeric/date comparison
- allowed-option validation
- dependency-based validation

Hard limits:

- no arbitrary script execution
- no cyclical field dependency evaluation
- no nested object editor
- no multi-level dependency graph beyond current activity-form needs

### 6.3 Data Persistence

Persistent storage model remains:

- fixed fields in `activities`
- category binding via `category_id`
- schema-driven values in `details_json`

This slice adds one persistence refinement:

- store `schema_version` with the activity record

Reason:

- preserve the version of the schema used when the activity was last saved
- create room for future schema evolution without needing another storage rewrite

If the column does not exist yet, it should be introduced by migration in the implementation phase.

### 6.4 Validation Model

Validation is split into two layers:

1. fixed-field validation
2. schema-field validation

Fixed-field validation covers:

- title
- category
- academic year
- semester
- start date
- end date

Schema-field validation covers:

- field existence
- type coercion
- required checks
- conditional visibility semantics
- cross-field rules
- allowed-option checks

Backend is always authoritative. Frontend validation must mirror it closely, but backend remains the final decision point.

### 6.5 Import / Export Shape

Import/export will stay flat, not nested.

Column strategy:

- fixed columns: `title`, `description`, `category`, `academic_year`, `semester`, `start_date`, `end_date`
- dynamic columns: `details.<field_key>`

This keeps templates readable and avoids introducing nested spreadsheet parsing complexity.

## 7. Frontend Design

### 7.1 Activity List `/activities`

First-screen queries:

- `GET /activities`
- `GET /activities/stats`
- `GET /activity-categories`
- `GET /options/academic-years`
- `GET /options/semesters`

Page areas:

- statistics cards
- filter bar
- activity table
- actions: `New Activity`, `Export`, `Import`

Interactions:

- `New Activity` routes to `/activities/new`
- row click routes to `/activities/:id`
- `Edit` routes to `/activities/:id/edit` when allowed

### 7.2 New Activity `/activities/new`

Flow:

1. load activity categories
2. user selects category
3. fetch category schema
4. render fixed fields + dynamic fields
5. save draft with `POST /activities`
6. redirect to `/activities/:id/edit`

The page is a two-layer form:

- fixed section:
  - title
  - description
  - category
  - academic year
  - semester
  - start date
  - end date
- dynamic section:
  - schema-driven `details`

### 7.3 Edit Activity `/activities/:id/edit`

Flow:

1. load activity detail
2. load its category schema
3. populate the fixed and dynamic form state
4. allow:
  - save draft
  - submit for review
  - participant management
  - file upload

Edit page remains structurally writable only for `draft`.

### 7.4 Frontend Activity Schema Runtime

This is an activity-only runtime, not a generic platform.

Responsibilities:

- render field components by schema type
- compute `visible_when`
- compute `enabled_when`
- perform pre-submit validation
- re-evaluate affected fields when dependencies change

Implementation shape:

- a dedicated activity form-state layer
- dedicated field renderer registry for activity field types
- dedicated rule evaluator limited to the supported operator set

### 7.5 Activity Import Dialog

Flow:

1. download template
2. upload file
3. call `POST /imports/activities/precheck`
4. display errors and summary
5. explicit confirm
6. call `POST /imports/activities/confirm`
7. refresh list + stats

### 7.6 Export

The list page export action uses current filters and downloads CSV from `GET /activities/export`.

## 8. Permissions and Workflow Boundaries

This slice does not redefine the activity workflow. It reuses the already confirmed rules:

- only `draft` allows structural edit
- owner can submit and withdraw per existing requirement
- teacher/admin have their already confirmed elevated permissions
- category schema changes do not widen edit permissions

For frontend UX:

- users without edit permission must not see edit actions
- if stale state still allows a click, backend remains the enforcement point

## 9. Error Model

New or unified error codes:

- `ACTIVITY_CATEGORY_NOT_FOUND`
- `ACTIVITY_SCHEMA_INVALID`
- `ACTIVITY_DETAILS_INVALID`
- `ACTIVITY_IMPORT_PRECHECK_FAILED`
- `ACTIVITY_IMPORT_CONFIRM_FAILED`

Behavior:

- frontend uses these codes for field-level or summary-level feedback
- backend returns stable structured error payloads

## 10. Testing Strategy

### 10.1 Backend

- repository tests for schema queries
- service tests for:
  - conditional rule validation
  - cross-field validation
  - import precheck validation
  - export flattening
- HTTP tests for:
  - category/schema endpoints
  - activity create/edit error behavior

### 10.2 Frontend

- page tests for:
  - `/activities/new`
  - `/activities/:id/edit`
  - activity list statistics and actions
- schema runtime tests for:
  - conditional visibility
  - dependency recalculation
  - cross-field validation messaging

### 10.3 End-to-End

Add or extend end-to-end coverage for:

- create draft activity from category schema
- edit draft activity
- submit draft after passing schema validation
- export activities
- import activities through precheck + confirm

## 11. Risks and Controls

### 11.1 Rule Complexity

Risk:

- schema rules become too broad and unstable

Control:

- keep the operator set closed
- avoid script execution
- reject cyclical dependencies in implementation

### 11.2 Import Complexity

Risk:

- spreadsheets become difficult to parse or explain

Control:

- keep import/export flat
- use `details.<field_key>` naming

### 11.3 Frontend / Backend Validation Drift

Risk:

- frontend permits values that backend rejects

Control:

- backend remains authoritative
- frontend mirrors only the same closed rule set
- test rule parity at service/runtime level

### 11.4 Oversized Edit Page

Risk:

- edit page becomes a monolith

Control:

- isolate:
  - form runtime
  - participant section
  - file section
- do not merge all concerns into one large page-level component

## 12. Recommended Implementation Order

1. backend category schema repository and endpoints
2. backend schema-aware activity create/update validation
3. backend activity stats and export
4. backend activity import precheck/confirm
5. frontend `/activities/new`
6. frontend `/activities/:id/edit`
7. frontend list-page stats/import/export wiring
8. end-to-end integration coverage

## 13. Success Criteria

This slice is complete when:

- category-driven draft creation works through `/activities/new`
- category-driven draft editing works through `/activities/:id/edit`
- conditional visibility, field dependencies, and cross-field validation are enforced
- activity list page shows usable statistics
- activity import and export complete through the formal workflow
- backend and frontend remain aligned with the confirmed permission and draft-edit rules
