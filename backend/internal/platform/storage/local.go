package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type LocalStore struct {
	root      string
	publicURL string
}

func NewLocalStore(root string, publicURL string) *LocalStore {
	return &LocalStore{
		root:      root,
		publicURL: strings.TrimRight(publicURL, "/"),
	}
}

func (s *LocalStore) Put(ctx context.Context, req PutRequest) (Object, error) {
	if err := ctx.Err(); err != nil {
		return Object{}, err
	}
	if err := validateKey(req.Key); err != nil {
		return Object{}, err
	}
	if req.Body == nil {
		return Object{}, errors.New("put body is required")
	}
	if err := os.MkdirAll(filepath.Dir(s.pathForKey(req.Key)), 0o755); err != nil {
		return Object{}, fmt.Errorf("create local storage directory: %w", err)
	}

	file, err := os.Create(s.pathForKey(req.Key))
	if err != nil {
		return Object{}, fmt.Errorf("create local storage object: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, req.Body)
	if err != nil {
		return Object{}, fmt.Errorf("write local storage object: %w", err)
	}

	size := req.Size
	if size <= 0 {
		size = written
	}

	return Object{
		Key:         req.Key,
		Size:        size,
		ContentType: req.ContentType,
		Backend:     BackendLocal,
	}, nil
}

func (s *LocalStore) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}

	err := os.Remove(s.pathForKey(key))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete local storage object: %w", err)
	}

	return nil
}

func (s *LocalStore) SignGetURL(ctx context.Context, key string, _ SignedURLOptions) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := validateKey(key); err != nil {
		return "", err
	}
	if s.publicURL == "" {
		return "", errors.New("storage public url is required")
	}

	segments := strings.Split(key, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}

	return s.publicURL + "/" + strings.Join(segments, "/"), nil
}

func (s *LocalStore) HealthCheck(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return fmt.Errorf("ensure local storage root: %w", err)
	}

	probePath := filepath.Join(s.root, fmt.Sprintf(".healthcheck-%d", time.Now().UnixNano()))
	if err := os.WriteFile(probePath, []byte("ok"), 0o600); err != nil {
		return fmt.Errorf("write local storage probe: %w", err)
	}
	if err := os.Remove(probePath); err != nil {
		return fmt.Errorf("remove local storage probe: %w", err)
	}

	return nil
}

func (s *LocalStore) pathForKey(key string) string {
	return filepath.Join(s.root, filepath.FromSlash(key))
}

func validateKey(key string) error {
	if key == "" {
		return errors.New("storage key is required")
	}

	for _, segment := range strings.Split(key, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("invalid storage key %q", key)
		}
	}

	if strings.HasPrefix(path.Clean("/"+key), "/../") {
		return fmt.Errorf("invalid storage key %q", key)
	}

	return nil
}
