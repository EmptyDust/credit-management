package health

import (
	"context"
)

type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
	StatusError    Status = "error"
)

type CheckFunc func(context.Context) error

type SummaryResult struct {
	Status  Status `json:"status"`
	Version string `json:"version"`
}

type DependencyResult struct {
	Status Status `json:"status"`
	Error  string `json:"error,omitempty"`
}

type DetailedResult struct {
	Status   Status           `json:"status"`
	Version  string           `json:"version"`
	Database DependencyResult `json:"database"`
	Redis    DependencyResult `json:"redis"`
	Storage  DependencyResult `json:"storage"`
}

type Checker struct {
	version string
	checks  map[string]CheckFunc
}

func NewChecker(version string, checks map[string]CheckFunc) *Checker {
	cloned := make(map[string]CheckFunc, len(checks))
	for name, check := range checks {
		cloned[name] = check
	}

	return &Checker{
		version: version,
		checks:  cloned,
	}
}

func (c *Checker) Summary(ctx context.Context) SummaryResult {
	status, _ := c.runChecks(ctx)
	return SummaryResult{
		Status:  status,
		Version: c.version,
	}
}

func (c *Checker) Detailed(ctx context.Context) DetailedResult {
	status, results := c.runChecks(ctx)
	return DetailedResult{
		Status:   status,
		Version:  c.version,
		Database: results["database"],
		Redis:    results["redis"],
		Storage:  results["storage"],
	}
}

func (c *Checker) runChecks(ctx context.Context) (Status, map[string]DependencyResult) {
	status := StatusOK
	dependencies := make(map[string]DependencyResult, len(c.checks))

	for _, name := range []string{"database", "redis", "storage"} {
		result := DependencyResult{Status: StatusError, Error: "check is not configured"}
		check := c.checks[name]
		if check != nil {
			result = DependencyResult{Status: StatusOK}
			if err := check(ctx); err != nil {
				result = DependencyResult{Status: StatusError, Error: err.Error()}
			}
		}

		dependencies[name] = result
		if result.Status != StatusOK {
			status = StatusDegraded
		}
	}

	return status, dependencies
}
