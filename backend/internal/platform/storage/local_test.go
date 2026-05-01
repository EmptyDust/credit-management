package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalStorePutAndDelete(t *testing.T) {
	root := t.TempDir()
	store := NewLocalStore(root, "http://localhost:8082/uploads")

	obj, err := store.Put(context.Background(), PutRequest{
		Key:         "avatars/alice.txt",
		Body:        strings.NewReader("hello"),
		Size:        int64(len("hello")),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("expected put to succeed, got %v", err)
	}
	if obj.Key != "avatars/alice.txt" {
		t.Fatalf("expected object key avatars/alice.txt, got %s", obj.Key)
	}
	if obj.Size != int64(len("hello")) {
		t.Fatalf("expected object size %d, got %d", len("hello"), obj.Size)
	}
	if obj.ContentType != "text/plain" {
		t.Fatalf("expected content type text/plain, got %s", obj.ContentType)
	}
	if obj.Backend != BackendLocal {
		t.Fatalf("expected backend %s, got %s", BackendLocal, obj.Backend)
	}

	content, err := os.ReadFile(filepath.Join(root, "avatars", "alice.txt"))
	if err != nil {
		t.Fatalf("expected file to be written, got %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("expected stored file contents hello, got %s", string(content))
	}

	if err := store.Delete(context.Background(), "avatars/alice.txt"); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "avatars", "alice.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, got err=%v", err)
	}
}

func TestLocalStoreSignGetURL(t *testing.T) {
	store := NewLocalStore("/tmp/storage", "https://cdn.example.com/public")

	url, err := store.SignGetURL(context.Background(), "reports/credit summary.pdf", SignedURLOptions{
		ExpiresIn: 15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("expected sign url to succeed, got %v", err)
	}
	if url != "https://cdn.example.com/public/reports/credit%20summary.pdf" {
		t.Fatalf("expected encoded public url, got %s", url)
	}
}

func TestLocalStoreHealthCheck(t *testing.T) {
	store := NewLocalStore(t.TempDir(), "http://localhost:8082/uploads")

	if err := store.HealthCheck(context.Background()); err != nil {
		t.Fatalf("expected health check to succeed, got %v", err)
	}
}

func TestLocalStorePutDoesNotLeaveObjectOnCopyFailure(t *testing.T) {
	root := t.TempDir()
	store := NewLocalStore(root, "")

	_, err := store.Put(context.Background(), PutRequest{
		Key:  "avatars/broken.txt",
		Body: errReader{err: errors.New("copy failed")},
	})
	if err == nil {
		t.Fatal("expected put to fail on copy error")
	}

	finalPath := filepath.Join(root, "avatars", "broken.txt")
	if _, statErr := os.Stat(finalPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no final object after copy failure, got err=%v", statErr)
	}

	entries, readErr := os.ReadDir(filepath.Join(root, "avatars"))
	if readErr != nil {
		t.Fatalf("expected parent directory to remain readable, got %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no temp files to remain, got %d entries", len(entries))
	}
}

func TestLocalStoreRejectsBackslashTraversalKey(t *testing.T) {
	store := NewLocalStore(t.TempDir(), "")

	_, err := store.Put(context.Background(), PutRequest{
		Key:  "..\\escape.txt",
		Body: strings.NewReader("escape"),
	})
	if err == nil {
		t.Fatal("expected backslash traversal key to be rejected")
	}
}

type errReader struct {
	err error
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
