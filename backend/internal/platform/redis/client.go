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

var (
	pingClient  = defaultPingClient
	closeClient = defaultCloseClient
)

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
		if err := pingClient(ctx, client); err != nil {
			_ = closeClient(client)
			return nil, fmt.Errorf("ping redis: %w", err)
		}
	}

	return client, nil
}

func defaultPingClient(ctx context.Context, client *goredis.Client) error {
	return client.Ping(ctx).Err()
}

func defaultCloseClient(client *goredis.Client) error {
	return client.Close()
}
