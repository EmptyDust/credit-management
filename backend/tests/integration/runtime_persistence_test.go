//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"credit-management/backend/internal/bootstrap"
	"credit-management/backend/internal/platform/database"
	platformredis "credit-management/backend/internal/platform/redis"
)

type runtimeConfig struct {
	databaseDSN      string
	databaseSchema   string
	redisAddr        string
	redisPassword    string
	redisDB          string
	storageLocalRoot string
	storagePublicURL string
}

type harness struct {
	t       *testing.T
	handler http.Handler
}

type apiEnvelope struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type authPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		UserType string `json:"user_type"`
	} `json:"user"`
}

func TestRuntimePersistsRefreshSessionsAcrossAppRecreation(t *testing.T) {
	cfg := newRuntimeConfig(t)
	resetRuntimeState(t, cfg)

	app1 := mustNewRuntimeApp(t, cfg)
	h1 := &harness{t: t, handler: app1.Handler()}

	registerStudent(t, h1, "persist-student-01")
	loginResult := login(t, h1, "persist-student-01")

	app2 := mustNewRuntimeApp(t, cfg)
	h2 := &harness{t: t, handler: app2.Handler()}

	refreshRecorder := h2.requestJSON(http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{
		"refresh_token": loginResult.RefreshToken,
	})
	if refreshRecorder.Code != http.StatusOK {
		t.Fatalf("refresh after app recreation failed: status=%d body=%s", refreshRecorder.Code, refreshRecorder.Body.String())
	}

	refreshed := decodeData[authPayload](t, decodeEnvelope(t, refreshRecorder))
	if refreshed.User.Username != "persist-student-01" {
		t.Fatalf("unexpected refreshed user: %+v", refreshed.User)
	}
}

func TestRuntimePersistsActivityApplicationsAndFilesAcrossAppRecreation(t *testing.T) {
	cfg := newRuntimeConfig(t)
	resetRuntimeState(t, cfg)

	app1 := mustNewRuntimeApp(t, cfg)
	h1 := &harness{t: t, handler: app1.Handler()}

	registerStudent(t, h1, "persist-student-02")
	student := login(t, h1, "persist-student-02")

	createRecorder := h1.requestJSON(http.MethodPost, "/api/v1/activities", student.AccessToken, map[string]any{
		"title":         "Persistent robotics lab",
		"description":   "runtime persistence integration",
		"category":      "general",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details": map[string]any{
			"location": "B-201",
		},
	})
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create activity failed: status=%d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	created := decodeData[struct {
		Activity struct {
			ID string `json:"id"`
		} `json:"activity"`
	}](t, decodeEnvelope(t, createRecorder))

	uploadRecorder := h1.requestMultipart(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/files", student.AccessToken, "file", "proof.txt", []byte("persistent attachment"))
	if uploadRecorder.Code != http.StatusCreated {
		t.Fatalf("upload file failed: status=%d body=%s", uploadRecorder.Code, uploadRecorder.Body.String())
	}

	addParticipantsRecorder := h1.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/participants", student.AccessToken, map[string]any{
		"user_ids": []string{student.User.ID},
		"credits":  2.5,
	})
	if addParticipantsRecorder.Code != http.StatusOK {
		t.Fatalf("add participants failed: status=%d body=%s", addParticipantsRecorder.Code, addParticipantsRecorder.Body.String())
	}

	submitRecorder := h1.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/submit", student.AccessToken, map[string]any{})
	if submitRecorder.Code != http.StatusOK {
		t.Fatalf("submit activity failed: status=%d body=%s", submitRecorder.Code, submitRecorder.Body.String())
	}

	admin := login(t, h1, "admin01")
	reviewRecorder := h1.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/review", admin.AccessToken, map[string]string{
		"target_status":   "approved",
		"review_comments": "persist it",
	})
	if reviewRecorder.Code != http.StatusOK {
		t.Fatalf("review activity failed: status=%d body=%s", reviewRecorder.Code, reviewRecorder.Body.String())
	}

	app2 := mustNewRuntimeApp(t, cfg)
	h2 := &harness{t: t, handler: app2.Handler()}

	detailRecorder := h2.request(http.MethodGet, "/api/v1/activities/"+created.Activity.ID, student.AccessToken, nil, "")
	if detailRecorder.Code != http.StatusOK {
		t.Fatalf("get persisted activity failed: status=%d body=%s", detailRecorder.Code, detailRecorder.Body.String())
	}
	detail := decodeData[struct {
		Activity struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			Status       string `json:"status"`
			AcademicYear string `json:"academic_year"`
			Semester     string `json:"semester"`
		} `json:"activity"`
	}](t, decodeEnvelope(t, detailRecorder))
	if detail.Activity.ID != created.Activity.ID || detail.Activity.Status != "approved" {
		t.Fatalf("unexpected persisted activity: %+v", detail.Activity)
	}

	applicationsRecorder := h2.request(http.MethodGet, "/api/v1/applications", student.AccessToken, nil, "")
	if applicationsRecorder.Code != http.StatusOK {
		t.Fatalf("list applications failed: status=%d body=%s", applicationsRecorder.Code, applicationsRecorder.Body.String())
	}
	applications := decodeData[struct {
		Items []struct {
			Activity struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"activity"`
			AwardedCredits float64 `json:"awarded_credits"`
		} `json:"items"`
	}](t, decodeEnvelope(t, applicationsRecorder))
	if len(applications.Items) != 1 || applications.Items[0].Activity.ID != created.Activity.ID || applications.Items[0].AwardedCredits != 2.5 {
		t.Fatalf("unexpected persisted applications: %+v", applications.Items)
	}

	filesRecorder := h2.request(http.MethodGet, "/api/v1/activities/"+created.Activity.ID+"/files", student.AccessToken, nil, "")
	if filesRecorder.Code != http.StatusOK {
		t.Fatalf("list files failed: status=%d body=%s", filesRecorder.Code, filesRecorder.Body.String())
	}
	files := decodeData[struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}](t, decodeEnvelope(t, filesRecorder))
	if len(files.Items) != 1 || files.Items[0].Name != "proof.txt" {
		t.Fatalf("unexpected persisted files: %+v", files.Items)
	}

	downloadRecorder := h2.request(http.MethodGet, "/api/v1/files/"+files.Items[0].ID+"/download", student.AccessToken, nil, "")
	if downloadRecorder.Code != http.StatusFound {
		t.Fatalf("download redirect failed: status=%d body=%s", downloadRecorder.Code, downloadRecorder.Body.String())
	}
	location := downloadRecorder.Header().Get("Location")
	if !strings.HasPrefix(location, cfg.storagePublicURL+"/") {
		t.Fatalf("unexpected signed url location: %q", location)
	}
}

func newRuntimeConfig(t *testing.T) runtimeConfig {
	t.Helper()

	tempRoot := t.TempDir()
	schema := getenvDefault("TEST_DATABASE_SCHEMA", "integration_runtime")
	return runtimeConfig{
		databaseDSN:      ensureSearchPath(mustEnv(t, "TEST_DATABASE_DSN"), schema),
		databaseSchema:   schema,
		redisAddr:        getenvDefault("TEST_REDIS_ADDR", "127.0.0.1:6379"),
		redisPassword:    os.Getenv("TEST_REDIS_PASSWORD"),
		redisDB:          getenvDefault("TEST_REDIS_DB", "15"),
		storageLocalRoot: filepath.Join(tempRoot, "storage"),
		storagePublicURL: "http://files.local.test/storage",
	}
}

func mustNewRuntimeApp(t *testing.T, cfg runtimeConfig) *bootstrap.App {
	t.Helper()

	t.Setenv("DATABASE_DSN", cfg.databaseDSN)
	t.Setenv("REDIS_ADDR", cfg.redisAddr)
	t.Setenv("REDIS_PASSWORD", cfg.redisPassword)
	t.Setenv("REDIS_DB", cfg.redisDB)
	t.Setenv("STORAGE_BACKEND", "local")
	t.Setenv("STORAGE_LOCAL_ROOT", cfg.storageLocalRoot)
	t.Setenv("STORAGE_PUBLIC_URL", cfg.storagePublicURL)
	t.Setenv("BACKEND_HTTP_PORT", "18082")
	t.Setenv("GIN_MODE", "test")
	t.Setenv("BUILD_VERSION", "integration-test")

	app, err := bootstrap.NewApp()
	if err != nil {
		t.Fatalf("bootstrap runtime app: %v", err)
	}

	return app
}

func resetRuntimeState(t *testing.T, cfg runtimeConfig) {
	t.Helper()

	ctx := context.Background()

	db, err := database.Open(ctx, database.Options{DSN: cfg.databaseDSN})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("database handle: %v", err)
	}
	defer sqlDB.Close()
	if err := db.Exec(`
		DROP SCHEMA IF EXISTS ` + cfg.databaseSchema + ` CASCADE;
		CREATE SCHEMA ` + cfg.databaseSchema + `;
	`).Error; err != nil {
		t.Fatalf("reset integration schema: %v", err)
	}
	if err := database.NewMigrator(resolveMigrationsDir(t)).Apply(ctx, sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	if err := database.NewSeeder().Run(ctx, db); err != nil {
		t.Fatalf("seed runtime tables: %v", err)
	}

	redisDB, err := parseRedisDB(cfg.redisDB)
	if err != nil {
		t.Fatalf("parse redis db: %v", err)
	}
	redisClient, err := platformredis.Open(ctx, platformredis.Options{
		Addr:     cfg.redisAddr,
		Password: cfg.redisPassword,
		DB:       redisDB,
	})
	if err != nil {
		t.Fatalf("open redis: %v", err)
	}
	defer redisClient.Close()
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}

	if err := os.RemoveAll(cfg.storageLocalRoot); err != nil {
		t.Fatalf("remove storage root: %v", err)
	}
	if err := os.MkdirAll(cfg.storageLocalRoot, 0o755); err != nil {
		t.Fatalf("create storage root: %v", err)
	}
}

func resolveMigrationsDir(t *testing.T) string {
	t.Helper()

	for _, candidate := range []string{"../../migrations", "../../../migrations", "./migrations"} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	t.Fatalf("migrations directory not found")
	return ""
}

func parseRedisDB(value string) (int, error) {
	return strconv.Atoi(value)
}

func mustEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("%s is required for integration runtime tests", key)
	}
	return value
}

func getenvDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func ensureSearchPath(dsn string, schema string) string {
	if strings.Contains(dsn, "search_path=") {
		return dsn
	}
	searchPath := schema + ",public"

	parsed, err := url.Parse(dsn)
	if err != nil {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		return dsn + separator + "search_path=" + url.QueryEscape(searchPath)
	}

	query := parsed.Query()
	query.Set("search_path", searchPath)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func (h *harness) request(method string, path string, token string, body []byte, contentType string) *httptest.ResponseRecorder {
	h.t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	h.handler.ServeHTTP(recorder, req)
	return recorder
}

func (h *harness) requestJSON(method string, path string, token string, payload any) *httptest.ResponseRecorder {
	h.t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		h.t.Fatalf("marshal payload: %v", err)
	}

	return h.request(method, path, token, body, "application/json")
}

func (h *harness) requestMultipart(method string, path string, token string, fieldName string, filename string, content []byte) *httptest.ResponseRecorder {
	h.t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		h.t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		h.t.Fatalf("write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		h.t.Fatalf("close multipart writer: %v", err)
	}

	return h.request(method, path, token, body.Bytes(), writer.FormDataContentType())
}

func registerStudent(t *testing.T, h *harness, username string) {
	t.Helper()

	recorder := h.requestJSON(http.MethodPost, "/api/v1/auth/register/student", "", map[string]string{
		"username":      username,
		"password":      "Passw0rd!",
		"real_name":     "Persistence Student",
		"email":         username + "@example.com",
		"phone":         "1380000" + username[len(username)-4:],
		"student_id":    "SID-" + strings.ToUpper(username),
		"grade":         "2024",
		"department_id": "class-01",
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("register student failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func login(t *testing.T, h *harness, loginID string) authPayload {
	t.Helper()

	recorder := h.requestJSON(http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"login_id": loginID,
		"password": "Passw0rd!",
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}

	return decodeData[authPayload](t, decodeEnvelope(t, recorder))
}

func decodeEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) apiEnvelope {
	t.Helper()

	var envelope apiEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}

	return envelope
}

func decodeData[T any](t *testing.T, envelope apiEnvelope) T {
	t.Helper()

	var data T
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v\nraw=%s", err, string(envelope.Data))
	}

	return data
}
