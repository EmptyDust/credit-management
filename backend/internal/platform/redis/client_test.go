package redis

import (
	"context"
	"errors"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestOpenRejectsMissingAddress(t *testing.T) {
	client, err := Open(context.Background(), Options{})
	if err == nil {
		t.Fatal("expected missing address to fail")
	}
	if client != nil {
		t.Fatal("expected no client on missing address")
	}
}

func TestOpenClosesClientWhenPingFails(t *testing.T) {
	t.Cleanup(func() {
		pingClient = defaultPingClient
		closeClient = defaultCloseClient
	})

	pingClient = func(context.Context, *goredis.Client) error {
		return errors.New("ping failed")
	}

	closed := false
	closeClient = func(_ *goredis.Client) error {
		closed = true
		return nil
	}

	client, err := Open(context.Background(), Options{Addr: "127.0.0.1:6379"})
	if err == nil {
		t.Fatal("expected ping failure to return error")
	}
	if client != nil {
		t.Fatal("expected no client on ping failure")
	}
	if !closed {
		t.Fatal("expected ping failure path to close the client")
	}
}
