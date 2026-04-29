package auditlog

import (
	"context"
	"testing"
)

func TestWriterRecordAppendsEntries(t *testing.T) {
	writer := NewWriter()

	if err := writer.Record(context.Background(), Entry{Action: "create"}); err != nil {
		t.Fatalf("expected first record to succeed, got %v", err)
	}
	if err := writer.Record(context.Background(), Entry{Action: "delete"}); err != nil {
		t.Fatalf("expected second record to succeed, got %v", err)
	}

	entries := writer.List(Filter{})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Action != "create" || entries[1].Action != "delete" {
		t.Fatalf("expected append-only order, got %+v", entries)
	}
}
