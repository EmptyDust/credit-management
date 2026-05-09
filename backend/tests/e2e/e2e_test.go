package e2e

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"credit-management/backend/internal/bootstrap"
)

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

type multipartUploadFile struct {
	FieldName string
	Filename  string
	Content   []byte
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	app, err := bootstrap.NewTestApp()
	if err != nil {
		t.Fatalf("bootstrap app: %v", err)
	}

	return &harness{
		t:       t,
		handler: app.Handler(),
	}
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

func (h *harness) requestMultipartFiles(method string, path string, token string, files []multipartUploadFile, fields map[string]string) *httptest.ResponseRecorder {
	h.t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			h.t.Fatalf("write field %s: %v", key, err)
		}
	}
	for _, file := range files {
		part, err := writer.CreateFormFile(file.FieldName, file.Filename)
		if err != nil {
			h.t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(file.Content); err != nil {
			h.t.Fatalf("write multipart content: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		h.t.Fatalf("close multipart writer: %v", err)
	}

	return h.request(method, path, token, body.Bytes(), writer.FormDataContentType())
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
