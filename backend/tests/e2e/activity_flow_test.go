package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreateActivityResponseUsesAPIFieldNames(t *testing.T) {
	h := newHarness(t)

	student := login(t, h, "student01")

	createRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities", student.AccessToken, map[string]any{
		"title":         "API contract smoke",
		"description":   "assert raw field names",
		"category":      "innovation",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details":       map[string]any{},
	})
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create activity failed: status=%d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	envelope := decodeEnvelope(t, createRecorder)

	var payload map[string]map[string]any
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatalf("decode raw payload: %v\nraw=%s", err, string(envelope.Data))
	}

	activity := payload["activity"]
	if activity["id"] == nil {
		t.Fatalf("expected activity.id in response, got %v", activity)
	}
	if activity["ID"] != nil {
		t.Fatalf("unexpected PascalCase field in response: %v", activity)
	}
}

func TestDraftToApprovedGeneratesApplications(t *testing.T) {
	h := newHarness(t)

	student := login(t, h, "student01")

	createRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities", student.AccessToken, map[string]any{
		"title":         "Robotics Expo",
		"description":   "Annual showcase",
		"category":      "innovation",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details": map[string]any{
			"location": "Hall A",
		},
	})
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create activity failed: status=%d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	var created struct {
		Activity struct {
			ID string `json:"id"`
		} `json:"activity"`
	}
	created = decodeData[struct {
		Activity struct {
			ID string `json:"id"`
		} `json:"activity"`
	}](t, decodeEnvelope(t, createRecorder))

	addParticipants := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/participants", student.AccessToken, map[string]any{
		"user_ids": []string{student.User.ID},
		"credits":  2.5,
	})
	if addParticipants.Code != http.StatusOK {
		t.Fatalf("add participants failed: status=%d body=%s", addParticipants.Code, addParticipants.Body.String())
	}

	submitRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/submit", student.AccessToken, map[string]any{})
	if submitRecorder.Code != http.StatusOK {
		t.Fatalf("submit activity failed: status=%d body=%s", submitRecorder.Code, submitRecorder.Body.String())
	}

	teacher := login(t, h, "teacher01")
	reviewRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/review", teacher.AccessToken, map[string]string{
		"target_status":   "approved",
		"review_comments": "approved for credit sync",
	})
	if reviewRecorder.Code != http.StatusOK {
		t.Fatalf("review activity failed: status=%d body=%s", reviewRecorder.Code, reviewRecorder.Body.String())
	}

	applicationsRecorder := h.request(http.MethodGet, "/api/v1/applications", student.AccessToken, nil, "")
	if applicationsRecorder.Code != http.StatusOK {
		t.Fatalf("list applications failed: status=%d body=%s", applicationsRecorder.Code, applicationsRecorder.Body.String())
	}

	var applications struct {
		Items []struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Activity struct {
				Title string `json:"title"`
			} `json:"activity"`
			AcademicYear string `json:"academic_year"`
			Semester     string `json:"semester"`
		} `json:"items"`
	}
	applications = decodeData[struct {
		Items []struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Activity struct {
				Title string `json:"title"`
			} `json:"activity"`
			AcademicYear string `json:"academic_year"`
			Semester     string `json:"semester"`
		} `json:"items"`
	}](t, decodeEnvelope(t, applicationsRecorder))

	if len(applications.Items) != 1 {
		t.Fatalf("expected 1 application, got %d", len(applications.Items))
	}
	if applications.Items[0].Activity.Title != "Robotics Expo" {
		t.Fatalf("unexpected application title: %+v", applications.Items[0])
	}
	if applications.Items[0].AcademicYear != "2025-2026" || applications.Items[0].Semester != "autumn" {
		t.Fatalf("unexpected academic dimensions: %+v", applications.Items[0])
	}

	filteredRecorder := h.request(http.MethodGet, "/api/v1/applications?academic_year=2025-2026&semester=autumn&keyword=Robotics", student.AccessToken, nil, "")
	if filteredRecorder.Code != http.StatusOK {
		t.Fatalf("filtered applications failed: status=%d body=%s", filteredRecorder.Code, filteredRecorder.Body.String())
	}
	filtered := decodeData[struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}](t, decodeEnvelope(t, filteredRecorder))
	if len(filtered.Items) != 1 {
		t.Fatalf("expected 1 filtered application, got %+v", filtered.Items)
	}

	emptyFilteredRecorder := h.request(http.MethodGet, "/api/v1/applications?semester=spring", student.AccessToken, nil, "")
	if emptyFilteredRecorder.Code != http.StatusOK {
		t.Fatalf("empty filtered applications failed: status=%d body=%s", emptyFilteredRecorder.Code, emptyFilteredRecorder.Body.String())
	}
	emptyFiltered := decodeData[struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}](t, decodeEnvelope(t, emptyFilteredRecorder))
	if len(emptyFiltered.Items) != 0 {
		t.Fatalf("expected no spring applications, got %+v", emptyFiltered.Items)
	}

	filteredStatsRecorder := h.request(http.MethodGet, "/api/v1/applications/stats?semester=autumn", student.AccessToken, nil, "")
	if filteredStatsRecorder.Code != http.StatusOK {
		t.Fatalf("filtered application stats failed: status=%d body=%s", filteredStatsRecorder.Code, filteredStatsRecorder.Body.String())
	}
	filteredStats := decodeData[struct {
		Total         int     `json:"total"`
		ApprovedCount int     `json:"approved_count"`
		TotalCredits  float64 `json:"total_credits"`
	}](t, decodeEnvelope(t, filteredStatsRecorder))
	if filteredStats.Total != 1 || filteredStats.ApprovedCount != 1 || filteredStats.TotalCredits != 2.5 {
		t.Fatalf("unexpected filtered stats: %+v", filteredStats)
	}

	exportRecorder := h.request(http.MethodGet, "/api/v1/applications/export?semester=autumn", student.AccessToken, nil, "")
	if exportRecorder.Code != http.StatusOK {
		t.Fatalf("filtered application export failed: status=%d body=%s", exportRecorder.Code, exportRecorder.Body.String())
	}
	if body := exportRecorder.Body.String(); !strings.Contains(body, "academic_year,semester") || !strings.Contains(body, "2025-2026,autumn") {
		t.Fatalf("unexpected filtered application export: %s", body)
	}
}

func TestCopyActivityIncludesParticipantsAndFiles(t *testing.T) {
	h := newHarness(t)

	student := login(t, h, "student01")

	createRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities", student.AccessToken, map[string]any{
		"title":         "Copy source activity",
		"description":   "source for copy contract",
		"category":      "innovation",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details": map[string]any{
			"location": "Lab 1",
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

	addParticipants := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/participants", student.AccessToken, map[string]any{
		"user_ids": []string{student.User.ID},
		"credits":  3,
	})
	if addParticipants.Code != http.StatusOK {
		t.Fatalf("add participants failed: status=%d body=%s", addParticipants.Code, addParticipants.Body.String())
	}

	uploadRecorder := h.requestMultipart(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/files", student.AccessToken, "file", "copy-proof.txt", []byte("copy proof"))
	if uploadRecorder.Code != http.StatusCreated {
		t.Fatalf("upload file failed: status=%d body=%s", uploadRecorder.Code, uploadRecorder.Body.String())
	}

	copyRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/copy", student.AccessToken, map[string]any{})
	if copyRecorder.Code != http.StatusCreated {
		t.Fatalf("copy activity failed: status=%d body=%s", copyRecorder.Code, copyRecorder.Body.String())
	}
	copied := decodeData[struct {
		Activity struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"activity"`
	}](t, decodeEnvelope(t, copyRecorder))
	if copied.Activity.Status != "draft" {
		t.Fatalf("expected copied activity to be draft, got %+v", copied.Activity)
	}

	participantsRecorder := h.request(http.MethodGet, "/api/v1/activities/"+copied.Activity.ID+"/participants", student.AccessToken, nil, "")
	if participantsRecorder.Code != http.StatusOK {
		t.Fatalf("list copied participants failed: status=%d body=%s", participantsRecorder.Code, participantsRecorder.Body.String())
	}
	participants := decodeData[struct {
		Items []struct {
			UserID  string  `json:"user_id"`
			Credits float64 `json:"credits"`
		} `json:"items"`
	}](t, decodeEnvelope(t, participantsRecorder))
	if len(participants.Items) != 1 || participants.Items[0].UserID != student.User.ID || participants.Items[0].Credits != 3 {
		t.Fatalf("unexpected copied participants: %+v", participants.Items)
	}

	filesRecorder := h.request(http.MethodGet, "/api/v1/activities/"+copied.Activity.ID+"/files", student.AccessToken, nil, "")
	if filesRecorder.Code != http.StatusOK {
		t.Fatalf("list copied files failed: status=%d body=%s", filesRecorder.Code, filesRecorder.Body.String())
	}
	files := decodeData[struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}](t, decodeEnvelope(t, filesRecorder))
	if len(files.Items) != 1 || files.Items[0].Name != "copy-proof.txt" {
		t.Fatalf("unexpected copied files: %+v", files.Items)
	}
}

func TestActivityReviewAndTransferWriteAuditLogs(t *testing.T) {
	h := newHarness(t)

	student := login(t, h, "student01")
	teacher := login(t, h, "teacher01")
	admin := login(t, h, "admin01")

	createRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities", student.AccessToken, map[string]any{
		"title":         "Audited activity",
		"description":   "audit contract",
		"category":      "innovation",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details": map[string]any{
			"location": "Audit Lab",
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

	submitRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/submit", student.AccessToken, map[string]any{})
	if submitRecorder.Code != http.StatusOK {
		t.Fatalf("submit activity failed: status=%d body=%s", submitRecorder.Code, submitRecorder.Body.String())
	}

	reviewRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/review", teacher.AccessToken, map[string]string{
		"target_status":   "approved",
		"review_comments": "audited approval",
	})
	if reviewRecorder.Code != http.StatusOK {
		t.Fatalf("review activity failed: status=%d body=%s", reviewRecorder.Code, reviewRecorder.Body.String())
	}

	transferRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/transfer-owner", admin.AccessToken, map[string]string{
		"new_owner_id": teacher.User.ID,
	})
	if transferRecorder.Code != http.StatusOK {
		t.Fatalf("transfer activity failed: status=%d body=%s", transferRecorder.Code, transferRecorder.Body.String())
	}

	auditRecorder := h.request(http.MethodGet, "/api/v1/devtools/audit-logs?target_type=activity&target_id="+created.Activity.ID, admin.AccessToken, nil, "")
	if auditRecorder.Code != http.StatusOK {
		t.Fatalf("list audit logs failed: status=%d body=%s", auditRecorder.Code, auditRecorder.Body.String())
	}
	auditLogs := decodeData[struct {
		Items []struct {
			ActorID    string `json:"actor_id"`
			ActorRole  string `json:"actor_role"`
			TargetType string `json:"target_type"`
			TargetID   string `json:"target_id"`
			Action     string `json:"action"`
			Result     string `json:"result"`
		} `json:"items"`
	}](t, decodeEnvelope(t, auditRecorder))

	if !containsAuditAction(auditLogs.Items, teacher.User.ID, "teacher", "activity.review", "success") {
		t.Fatalf("expected activity.review audit log, got %+v", auditLogs.Items)
	}
	if !containsAuditAction(auditLogs.Items, admin.User.ID, "admin", "activity.transfer_owner", "success") {
		t.Fatalf("expected activity.transfer_owner audit log, got %+v", auditLogs.Items)
	}
}

func containsAuditAction(items []struct {
	ActorID    string `json:"actor_id"`
	ActorRole  string `json:"actor_role"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Action     string `json:"action"`
	Result     string `json:"result"`
}, actorID string, actorRole string, action string, result string) bool {
	for _, item := range items {
		if item.ActorID == actorID && item.ActorRole == actorRole && item.TargetType == "activity" && item.Action == action && item.Result == result {
			return true
		}
	}
	return false
}

func TestParticipantManagementStatsAndExport(t *testing.T) {
	h := newHarness(t)

	student := login(t, h, "student01")

	createRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities", student.AccessToken, map[string]any{
		"title":         "Participant management activity",
		"description":   "participant contract",
		"category":      "innovation",
		"academic_year": "2025-2026",
		"semester":      "autumn",
		"details": map[string]any{
			"location": "Participant Lab",
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

	addRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/participants", student.AccessToken, map[string]any{
		"user_ids": []string{student.User.ID},
		"credits":  1.5,
	})
	if addRecorder.Code != http.StatusOK {
		t.Fatalf("add participants failed: status=%d body=%s", addRecorder.Code, addRecorder.Body.String())
	}

	listRecorder := h.request(http.MethodGet, "/api/v1/activities/"+created.Activity.ID+"/participants", student.AccessToken, nil, "")
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list participants failed: status=%d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	participants := decodeData[struct {
		Items []struct {
			ID      string  `json:"id"`
			UserID  string  `json:"user_id"`
			Credits float64 `json:"credits"`
		} `json:"items"`
	}](t, decodeEnvelope(t, listRecorder))
	if len(participants.Items) != 1 {
		t.Fatalf("expected 1 participant, got %+v", participants.Items)
	}
	participantID := participants.Items[0].ID

	updateRecorder := h.requestJSON(http.MethodPatch, "/api/v1/activities/"+created.Activity.ID+"/participants/"+participantID, student.AccessToken, map[string]any{
		"credits": 4.5,
	})
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("update participant failed: status=%d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}
	updated := decodeData[struct {
		Participant struct {
			ID      string  `json:"id"`
			Credits float64 `json:"credits"`
		} `json:"participant"`
	}](t, decodeEnvelope(t, updateRecorder))
	if updated.Participant.ID != participantID || updated.Participant.Credits != 4.5 {
		t.Fatalf("unexpected updated participant: %+v", updated.Participant)
	}

	statsRecorder := h.request(http.MethodGet, "/api/v1/activities/"+created.Activity.ID+"/participants/stats", student.AccessToken, nil, "")
	if statsRecorder.Code != http.StatusOK {
		t.Fatalf("participant stats failed: status=%d body=%s", statsRecorder.Code, statsRecorder.Body.String())
	}
	stats := decodeData[struct {
		TotalCount     int     `json:"total_count"`
		TotalCredits   float64 `json:"total_credits"`
		AverageCredits float64 `json:"average_credits"`
		RecentNewCount int     `json:"recent_new_count"`
	}](t, decodeEnvelope(t, statsRecorder))
	if stats.TotalCount != 1 || stats.TotalCredits != 4.5 || stats.AverageCredits != 4.5 || stats.RecentNewCount != 1 {
		t.Fatalf("unexpected participant stats: %+v", stats)
	}

	exportRecorder := h.request(http.MethodGet, "/api/v1/activities/"+created.Activity.ID+"/participants/export", student.AccessToken, nil, "")
	if exportRecorder.Code != http.StatusOK {
		t.Fatalf("participant export failed: status=%d body=%s", exportRecorder.Code, exportRecorder.Body.String())
	}
	if body := exportRecorder.Body.String(); !strings.Contains(body, "participant_id,user_id,credits,joined_at") || !strings.Contains(body, participantID) {
		t.Fatalf("unexpected participant export: %s", body)
	}

	batchSetRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/participants/batch-set-credits", student.AccessToken, map[string]any{
		"participant_ids": []string{participantID},
		"credits":         2,
	})
	if batchSetRecorder.Code != http.StatusOK {
		t.Fatalf("batch set credits failed: status=%d body=%s", batchSetRecorder.Code, batchSetRecorder.Body.String())
	}
	batchSet := decodeData[struct {
		AffectedCount int `json:"affected_count"`
	}](t, decodeEnvelope(t, batchSetRecorder))
	if batchSet.AffectedCount != 1 {
		t.Fatalf("unexpected batch set result: %+v", batchSet)
	}

	batchDeleteRecorder := h.requestJSON(http.MethodPost, "/api/v1/activities/"+created.Activity.ID+"/participants/batch-delete", student.AccessToken, map[string]any{
		"participant_ids": []string{participantID},
	})
	if batchDeleteRecorder.Code != http.StatusOK {
		t.Fatalf("batch delete failed: status=%d body=%s", batchDeleteRecorder.Code, batchDeleteRecorder.Body.String())
	}
	batchDelete := decodeData[struct {
		AffectedCount int `json:"affected_count"`
	}](t, decodeEnvelope(t, batchDeleteRecorder))
	if batchDelete.AffectedCount != 1 {
		t.Fatalf("unexpected batch delete result: %+v", batchDelete)
	}

	emptyRecorder := h.request(http.MethodGet, "/api/v1/activities/"+created.Activity.ID+"/participants", student.AccessToken, nil, "")
	if emptyRecorder.Code != http.StatusOK {
		t.Fatalf("list participants after delete failed: status=%d body=%s", emptyRecorder.Code, emptyRecorder.Body.String())
	}
	empty := decodeData[struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}](t, decodeEnvelope(t, emptyRecorder))
	if len(empty.Items) != 0 {
		t.Fatalf("expected no participants after delete, got %+v", empty.Items)
	}
}
