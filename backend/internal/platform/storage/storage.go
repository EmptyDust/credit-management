package storage

import (
	"context"
	"io"
	"time"
)

type Backend string

const (
	BackendLocal  Backend = "local"
	BackendObject Backend = "object"
)

type PutRequest struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
}

type Object struct {
	Key         string
	Size        int64
	ContentType string
	Backend     Backend
}

type SignedURLOptions struct {
	ExpiresIn time.Duration
}

type Store interface {
	Put(ctx context.Context, req PutRequest) (Object, error)
	Delete(ctx context.Context, key string) error
	SignGetURL(ctx context.Context, key string, opts SignedURLOptions) (string, error)
}
