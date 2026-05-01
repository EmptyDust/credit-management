package redis

import (
	"context"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Options struct {
	Addr     string
	Password string
	DB       int
}

func Open(ctx context.Context, opts Options) (*goredis.Client, error) {
	if opts.Addr == "" {
		return nil, errors.New("redis addr is required")
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	if ctx != nil {
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("ping redis: %w", err)
		}
	}

	return client, nil
}
