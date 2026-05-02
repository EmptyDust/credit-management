package service

import (
	"context"
	"strings"
	"testing"

	"credit-management/backend/internal/platform/auditlog"
)

func TestServiceRecordListAndExport(t *testing.T) {
	writer := auditlog.NewWriter()
	service := New(writer)

	if err := service.Record(context.Background(), auditlog.Entry{
		ActorID:    "01JADMIN000000000000000000",
		ActorRole:  "admin",
		TargetType: "activity",
		TargetID:   "01JACTIVITY000000000000000",
		Action:     "approve",
		Result:     "success",
		RequestID:  "req-1",
	}); err != nil {
		t.Fatalf("record audit entry: %v", err)
	}

	items := service.List(auditlog.Filter{Action: "approve"})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Action != "approve" {
		t.Fatalf("expected action approve, got %s", items[0].Action)
	}

	content, err := service.Export(auditlog.Filter{Action: "approve"})
	if err != nil {
		t.Fatalf("export audit log: %v", err)
	}
	if !strings.Contains(string(content), "approve") {
		t.Fatalf("expected export to contain action, got %q", string(content))
	}
}
