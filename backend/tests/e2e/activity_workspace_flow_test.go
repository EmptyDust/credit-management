package e2e

import (
	"net/http"
	"strings"
	"testing"
)

func TestActivityWorkspaceSchemaStatsExportAndImport(t *testing.T) {
	h := newHarness(t)
	student := login(t, h, "student01")

	categoriesRecorder := h.request(http.MethodGet, "/api/v1/activity-categories", student.AccessToken, nil, "")
	if categoriesRecorder.Code != http.StatusOK {
		t.Fatalf("list categories failed: status=%d body=%s", categoriesRecorder.Code, categoriesRecorder.Body.String())
	}

	categories := decodeData[struct {
		Items []struct {
			Key   string `json:"key"`
			Label string `json:"label"`
		} `json:"items"`
	}](t, decodeEnvelope(t, categoriesRecorder))
	if !containsE2ECategory(categories.Items, "innovation") {
		t.Fatalf("expected innovation category in %+v", categories.Items)
	}

	schemaRecorder := h.request(http.MethodGet, "/api/v1/activity-categories/innovation/schema", student.AccessToken, nil, "")
	if schemaRecorder.Code != http.StatusOK {
		t.Fatalf("get category schema failed: status=%d body=%s", schemaRecorder.Code, schemaRecorder.Body.String())
	}

	schema := decodeData[struct {
		Schema struct {
			Fields []struct {
				Key  string `json:"key"`
				Type string `json:"type"`
			} `json:"fields"`
		} `json:"schema"`
	}](t, decodeEnvelope(t, schemaRecorder))
	if !containsE2ESchemaField(schema.Schema.Fields, "location") {
		t.Fatalf("expected location field in %+v", schema.Schema.Fields)
	}

	createRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities", student.AccessToken, map[string]any{
		"title":         "Workspace Contract Activity",
		"description":   "created from schema",
		"category":      "innovation",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details": map[string]any{
			"location": "Room 101",
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

	updateRecorder := h.requestJSON(http.MethodPatch, "/api/v1/activities/"+created.Activity.ID, student.AccessToken, map[string]any{
		"title":         "Workspace Contract Activity",
		"description":   "updated from schema",
		"category":      "innovation",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details": map[string]any{
			"location": "Room 202",
		},
	})
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("update activity failed: status=%d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}

	updated := decodeData[struct {
		Activity struct {
			Details map[string]any `json:"details"`
		} `json:"activity"`
	}](t, decodeEnvelope(t, updateRecorder))
	if updated.Activity.Details["location"] != "Room 202" {
		t.Fatalf("expected updated schema detail, got %+v", updated.Activity.Details)
	}

	statsRecorder := h.request(http.MethodGet, "/api/v1/activities/stats?category=innovation", student.AccessToken, nil, "")
	if statsRecorder.Code != http.StatusOK {
		t.Fatalf("activity stats failed: status=%d body=%s", statsRecorder.Code, statsRecorder.Body.String())
	}

	stats := decodeData[struct {
		StatusCounts map[string]int `json:"status_counts"`
		MineCount    int            `json:"mine_count"`
	}](t, decodeEnvelope(t, statsRecorder))
	if stats.StatusCounts["draft"] == 0 || stats.MineCount == 0 {
		t.Fatalf("unexpected stats payload: %+v", stats)
	}

	exportRecorder := h.request(http.MethodGet, "/api/v1/activities/export?category=innovation", student.AccessToken, nil, "")
	if exportRecorder.Code != http.StatusOK {
		t.Fatalf("activity export failed: status=%d body=%s", exportRecorder.Code, exportRecorder.Body.String())
	}
	if got := exportRecorder.Header().Get("Content-Disposition"); got != `attachment; filename="activities.csv"` {
		t.Fatalf("unexpected export disposition: %q", got)
	}
	exportCSV := exportRecorder.Body.String()
	if !strings.Contains(exportCSV, "details.location") || !strings.Contains(exportCSV, "Room 202") {
		t.Fatalf("expected flattened detail in export, got %q", exportCSV)
	}

	submitRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/submit", student.AccessToken, map[string]any{})
	if submitRecorder.Code != http.StatusOK {
		t.Fatalf("submit activity failed: status=%d body=%s", submitRecorder.Code, submitRecorder.Body.String())
	}

	importCSV := []byte(strings.Join([]string{
		"title,description,category,academic_year,semester,start_date,end_date,details.location",
		"Imported Workspace Activity,Imported,innovation,2025-2026,spring,2026-05-01,2026-05-03,Room 303",
		"",
	}, "\n"))

	precheckRecorder := h.requestMultipart(
		http.MethodPost,
		"/api/v1/imports/activities/precheck",
		student.AccessToken,
		"file",
		"activities.csv",
		importCSV,
	)
	if precheckRecorder.Code != http.StatusOK {
		t.Fatalf("activity import precheck failed: status=%d body=%s", precheckRecorder.Code, precheckRecorder.Body.String())
	}

	precheck := decodeData[struct {
		JobID    string `json:"job_id"`
		FileHash string `json:"file_hash"`
		Summary  struct {
			ValidRows   int `json:"valid_rows"`
			InvalidRows int `json:"invalid_rows"`
		} `json:"summary"`
	}](t, decodeEnvelope(t, precheckRecorder))
	if precheck.Summary.ValidRows != 1 || precheck.Summary.InvalidRows != 0 {
		t.Fatalf("unexpected activity import precheck summary: %+v", precheck.Summary)
	}

	confirmRecorder := h.requestJSON(http.MethodPost, "/api/v1/imports/activities/confirm", student.AccessToken, map[string]string{
		"job_id":    precheck.JobID,
		"file_hash": precheck.FileHash,
	})
	if confirmRecorder.Code != http.StatusOK {
		t.Fatalf("activity import confirm failed: status=%d body=%s", confirmRecorder.Code, confirmRecorder.Body.String())
	}

	confirm := decodeData[struct {
		Stage              string   `json:"stage"`
		CreatedActivityIDs []string `json:"created_activity_ids"`
	}](t, decodeEnvelope(t, confirmRecorder))
	if confirm.Stage != "completed" || len(confirm.CreatedActivityIDs) != 1 {
		t.Fatalf("unexpected activity import confirm result: %+v", confirm)
	}

	secondConfirmRecorder := h.requestJSON(http.MethodPost, "/api/v1/imports/activities/confirm", student.AccessToken, map[string]string{
		"job_id":    precheck.JobID,
		"file_hash": precheck.FileHash,
	})
	if secondConfirmRecorder.Code != http.StatusOK {
		t.Fatalf("second activity import confirm failed: status=%d body=%s", secondConfirmRecorder.Code, secondConfirmRecorder.Body.String())
	}
}

func containsE2ECategory(items []struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}, key string) bool {
	for _, item := range items {
		if item.Key == key {
			return true
		}
	}

	return false
}

func containsE2ESchemaField(fields []struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}, key string) bool {
	for _, field := range fields {
		if field.Key == key {
			return true
		}
	}

	return false
}
