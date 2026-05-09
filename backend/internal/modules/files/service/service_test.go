package service

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
	activityservice "credit-management/backend/internal/modules/activities/service"
	"credit-management/backend/internal/platform/storage"
)

func TestUploadAndDownloadWithMemoryRepository(t *testing.T) {
	activityRepo := activityservice.NewMemoryRepository()
	activity := &activitydomain.Activity{
		ID:      "act-1",
		Title:   "draft activity",
		Status:  activitydomain.StatusDraft,
		OwnerID: "owner-1",
	}
	if err := activityRepo.Save(context.Background(), activity); err != nil {
		t.Fatalf("seed activity: %v", err)
	}

	svc := New(activityRepo)
	fileHeader := newMultipartHeader(t, "proof.txt", "proof")

	file, err := svc.Upload(context.Background(), "owner-1", "student", "act-1", fileHeader)
	if err != nil {
		t.Fatalf("upload file: %v", err)
	}
	if file.ID == "" {
		t.Fatal("expected file id")
	}

	items, err := svc.List(context.Background(), "owner-1", "student", "act-1")
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one file, got %d", len(items))
	}

	download, err := svc.Download(context.Background(), "owner-1", "student", file.ID)
	if err != nil {
		t.Fatalf("download file: %v", err)
	}
	if string(download.Content) != "proof" {
		t.Fatalf("expected proof content, got %q", string(download.Content))
	}
}

func TestDownloadReturnsSignedURLWhenStoreIsConfigured(t *testing.T) {
	activityRepo := activityservice.NewMemoryRepository()
	activity := &activitydomain.Activity{
		ID:      "act-2",
		Title:   "draft activity",
		Status:  activitydomain.StatusDraft,
		OwnerID: "owner-2",
	}
	if err := activityRepo.Save(context.Background(), activity); err != nil {
		t.Fatalf("seed activity: %v", err)
	}

	store := &fakeStore{signedURL: "https://example.test/files/object"}
	repo := newMemoryMetadataRepository(make(map[string]*storedFile), make(map[string][]string))
	svc := NewWithStore(activityRepo, repo, store)

	fileHeader := newMultipartHeader(t, "proof.txt", "proof")
	file, err := svc.Upload(context.Background(), "owner-2", "student", "act-2", fileHeader)
	if err != nil {
		t.Fatalf("upload file: %v", err)
	}

	download, err := svc.Download(context.Background(), "owner-2", "student", file.ID)
	if err != nil {
		t.Fatalf("download file: %v", err)
	}
	if download.SignedURL != "https://example.test/files/object" {
		t.Fatalf("expected signed url, got %q", download.SignedURL)
	}
	if len(download.Content) != 0 {
		t.Fatalf("expected no inline content when signed url is returned")
	}
}

func TestDeleteRemovesStoredObject(t *testing.T) {
	activityRepo := activityservice.NewMemoryRepository()
	activity := &activitydomain.Activity{
		ID:      "act-3",
		Title:   "draft activity",
		Status:  activitydomain.StatusDraft,
		OwnerID: "owner-3",
	}
	if err := activityRepo.Save(context.Background(), activity); err != nil {
		t.Fatalf("seed activity: %v", err)
	}

	store := &fakeStore{signedURL: "https://example.test/files/object"}
	repo := newMemoryMetadataRepository(make(map[string]*storedFile), make(map[string][]string))
	svc := NewWithStore(activityRepo, repo, store)

	fileHeader := newMultipartHeader(t, "proof.txt", "proof")
	file, err := svc.Upload(context.Background(), "owner-3", "student", "act-3", fileHeader)
	if err != nil {
		t.Fatalf("upload file: %v", err)
	}

	if err := svc.Delete(context.Background(), "owner-3", "student", "act-3", file.ID); err != nil {
		t.Fatalf("delete file: %v", err)
	}
	if len(store.deletedKeys) != 1 {
		t.Fatalf("expected one deleted key, got %d", len(store.deletedKeys))
	}
}

func newMultipartHeader(t *testing.T, filename string, content string) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", "text/plain")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(int64(body.Len()))
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one file header, got %d", len(files))
	}
	return files[0]
}

type fakeStore struct {
	putKeys     []string
	deletedKeys []string
	signedURL   string
}

func (s *fakeStore) Put(_ context.Context, req storage.PutRequest) (storage.Object, error) {
	s.putKeys = append(s.putKeys, req.Key)
	return storage.Object{
		Key:         req.Key,
		Size:        req.Size,
		ContentType: req.ContentType,
		Backend:     storage.BackendLocal,
	}, nil
}

func (s *fakeStore) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte("stored object"))), nil
}

func (s *fakeStore) Delete(_ context.Context, key string) error {
	s.deletedKeys = append(s.deletedKeys, key)
	return nil
}

func (s *fakeStore) SignGetURL(_ context.Context, _ string, _ storage.SignedURLOptions) (string, error) {
	return s.signedURL, nil
}
