package health

import (
	"context"
	"errors"
	"testing"
)

func TestCheckerSummaryReportsHealthyWhenDependenciesPass(t *testing.T) {
	checker := NewChecker("1.2.3", map[string]CheckFunc{
		"database": func(context.Context) error { return nil },
		"redis":    func(context.Context) error { return nil },
		"storage":  func(context.Context) error { return nil },
	})

	summary := checker.Summary(context.Background())
	if summary.Status != StatusOK {
		t.Fatalf("expected summary status %s, got %s", StatusOK, summary.Status)
	}
	if summary.Version != "1.2.3" {
		t.Fatalf("expected summary version 1.2.3, got %s", summary.Version)
	}
}

func TestCheckerDetailedReportsDependencyFailures(t *testing.T) {
	checker := NewChecker("dev", map[string]CheckFunc{
		"database": func(context.Context) error { return nil },
		"redis":    func(context.Context) error { return errors.New("dial tcp timeout") },
		"storage":  func(context.Context) error { return nil },
	})

	detailed := checker.Detailed(context.Background())
	if detailed.Status != StatusDegraded {
		t.Fatalf("expected detailed status %s, got %s", StatusDegraded, detailed.Status)
	}
	if detailed.Version != "dev" {
		t.Fatalf("expected detailed version dev, got %s", detailed.Version)
	}

	if detailed.Redis.Status != StatusError {
		t.Fatalf("expected redis status %s, got %s", StatusError, detailed.Redis.Status)
	}
	if detailed.Redis.Error != "dial tcp timeout" {
		t.Fatalf("expected redis error dial tcp timeout, got %s", detailed.Redis.Error)
	}

	if detailed.Database.Status != StatusOK {
		t.Fatalf("expected database status %s, got %s", StatusOK, detailed.Database.Status)
	}
	if detailed.Database.Error != "" {
		t.Fatalf("expected empty database error, got %s", detailed.Database.Error)
	}

	if detailed.Storage.Status != StatusOK {
		t.Fatalf("expected storage status %s, got %s", StatusOK, detailed.Storage.Status)
	}
}
