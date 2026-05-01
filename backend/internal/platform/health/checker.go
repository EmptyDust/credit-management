package health

import (
	"context"
	"sort"
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
	Status       Status                      `json:"status"`
	Version      string                      `json:"version"`
	Dependencies map[string]DependencyResult `json:"dependencies"`
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
	status, dependencies := c.runChecks(ctx)
	return DetailedResult{
		Status:       status,
		Version:      c.version,
		Dependencies: dependencies,
	}
}

func (c *Checker) runChecks(ctx context.Context) (Status, map[string]DependencyResult) {
	status := StatusOK
	dependencies := make(map[string]DependencyResult, len(c.checks))

	names := make([]string, 0, len(c.checks))
	for name := range c.checks {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		check := c.checks[name]
		if check == nil {
			dependencies[name] = DependencyResult{Status: StatusError, Error: "check is not configured"}
			status = StatusDegraded
			continue
		}

		if err := check(ctx); err != nil {
			dependencies[name] = DependencyResult{Status: StatusError, Error: err.Error()}
			status = StatusDegraded
			continue
		}

		dependencies[name] = DependencyResult{Status: StatusOK}
	}

	return status, dependencies
}
