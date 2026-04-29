package auditlog

import (
	"context"
	"encoding/csv"
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actor_id,omitempty"`
	ActorRole  string    `json:"actor_role,omitempty"`
	TargetType string    `json:"target_type,omitempty"`
	TargetID   string    `json:"target_id,omitempty"`
	Action     string    `json:"action"`
	Result     string    `json:"result,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Filter struct {
	ActorID    string
	TargetType string
	TargetID   string
	Action     string
	Result     string
}

type Writer struct {
	mu      sync.RWMutex
	entries []Entry
}

func NewWriter() *Writer {
	return &Writer{entries: make([]Entry, 0)}
}

func (w *Writer) Record(_ context.Context, entry Entry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit-%d", len(w.entries)+1)
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	w.entries = append(w.entries, entry)
	return nil
}

func (w *Writer) List(filter Filter) []Entry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	items := make([]Entry, 0, len(w.entries))
	for _, entry := range w.entries {
		if filter.ActorID != "" && entry.ActorID != filter.ActorID {
			continue
		}
		if filter.TargetType != "" && entry.TargetType != filter.TargetType {
			continue
		}
		if filter.TargetID != "" && entry.TargetID != filter.TargetID {
			continue
		}
		if filter.Action != "" && entry.Action != filter.Action {
			continue
		}
		if filter.Result != "" && entry.Result != filter.Result {
			continue
		}

		items = append(items, entry)
	}

	return items
}

func (w *Writer) ExportCSV(filter Filter) ([]byte, error) {
	entries := w.List(filter)
	buf := &bytesBuffer{}
	writer := csv.NewWriter(buf)
	if err := writer.Write([]string{"id", "actor_id", "actor_role", "target_type", "target_id", "action", "result", "request_id", "created_at"}); err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if err := writer.Write([]string{
			entry.ID,
			entry.ActorID,
			entry.ActorRole,
			entry.TargetType,
			entry.TargetID,
			entry.Action,
			entry.Result,
			entry.RequestID,
			entry.CreatedAt.Format(time.RFC3339),
		}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuffer) Bytes() []byte {
	return append([]byte(nil), b.data...)
}
