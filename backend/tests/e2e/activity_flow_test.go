package e2e

import (
	"encoding/json"
	"net/http"
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
}
