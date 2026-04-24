package e2e

import (
	"net/http"
	"testing"
)

func TestStudentImportPrecheckThenConfirm(t *testing.T) {
	h := newHarness(t)
	admin := login(t, h, "admin01")

	csv := []byte("username,password,real_name,email,phone,student_id,grade,department_id\nimportstudent,Passw0rd!,Import Student,importstudent@example.com,13800138123,20245555,2024,class-02\n")

	precheckRecorder := h.requestMultipart(
		http.MethodPost,
		"/api/v1/imports/students/precheck",
		admin.AccessToken,
		"file",
		"students.csv",
		csv,
	)
	if precheckRecorder.Code != http.StatusOK {
		t.Fatalf("precheck failed: status=%d body=%s", precheckRecorder.Code, precheckRecorder.Body.String())
	}

	var precheck struct {
		JobID    string `json:"job_id"`
		FileHash string `json:"file_hash"`
		Summary  struct {
			TotalRows   int `json:"total_rows"`
			ValidRows   int `json:"valid_rows"`
			InvalidRows int `json:"invalid_rows"`
		} `json:"summary"`
	}
	precheck = decodeData[struct {
		JobID    string `json:"job_id"`
		FileHash string `json:"file_hash"`
		Summary  struct {
			TotalRows   int `json:"total_rows"`
			ValidRows   int `json:"valid_rows"`
			InvalidRows int `json:"invalid_rows"`
		} `json:"summary"`
	}](t, decodeEnvelope(t, precheckRecorder))

	if precheck.Summary.ValidRows != 1 || precheck.Summary.InvalidRows != 0 {
		t.Fatalf("unexpected precheck summary: %+v", precheck.Summary)
	}

	confirmRecorder := h.requestJSON(http.MethodPost, "/api/v1/imports/students/confirm", admin.AccessToken, map[string]string{
		"job_id":    precheck.JobID,
		"file_hash": precheck.FileHash,
	})
	if confirmRecorder.Code != http.StatusOK {
		t.Fatalf("confirm import failed: status=%d body=%s", confirmRecorder.Code, confirmRecorder.Body.String())
	}

	var confirm struct {
		Stage          string   `json:"stage"`
		CreatedUserIDs []string `json:"created_user_ids"`
	}
	confirm = decodeData[struct {
		Stage          string   `json:"stage"`
		CreatedUserIDs []string `json:"created_user_ids"`
	}](t, decodeEnvelope(t, confirmRecorder))

	if confirm.Stage != "completed" || len(confirm.CreatedUserIDs) != 1 {
		t.Fatalf("unexpected confirm result: %+v", confirm)
	}

	listRecorder := h.request(http.MethodGet, "/api/v1/students", admin.AccessToken, nil, "")
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list students failed: status=%d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
}
