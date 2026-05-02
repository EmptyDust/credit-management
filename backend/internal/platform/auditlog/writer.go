package auditlog

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"sync"
	"time"

	"credit-management/backend/internal/platform/idgen"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	backend writerBackend
}

func NewWriter() *Writer {
	return &Writer{backend: newMemoryBackend()}
}

func NewGormWriter(db *gorm.DB) *Writer {
	if db == nil {
		return NewWriter()
	}

	return &Writer{backend: newGormBackend(db)}
}

func (w *Writer) Record(_ context.Context, entry Entry) error {
	if w == nil || w.backend == nil {
		return errors.New("audit writer is not initialized")
	}
	return w.backend.Record(entry)
}

func (w *Writer) List(filter Filter) []Entry {
	if w == nil || w.backend == nil {
		return nil
	}
	return w.backend.List(filter)
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

type writerBackend interface {
	Record(entry Entry) error
	List(filter Filter) []Entry
}

type memoryBackend struct {
	mu      sync.RWMutex
	entries []Entry
}

func newMemoryBackend() *memoryBackend {
	return &memoryBackend{entries: make([]Entry, 0)}
}

func (b *memoryBackend) Record(entry Entry) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry = normalizeMemoryEntry(entry, len(b.entries)+1)
	b.entries = append(b.entries, entry)
	return nil
}

func (b *memoryBackend) List(filter Filter) []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	items := make([]Entry, 0, len(b.entries))
	for _, entry := range b.entries {
		if !matchesFilter(entry, filter) {
			continue
		}
		items = append(items, entry)
	}

	return items
}

type gormBackend struct {
	db *gorm.DB
}

type auditLogModel struct {
	ID            string    `gorm:"column:id"`
	ActorID       *string   `gorm:"column:actor_id"`
	ActorRole     string    `gorm:"column:actor_role"`
	TargetType    string    `gorm:"column:target_type"`
	TargetID      *string   `gorm:"column:target_id"`
	Action        string    `gorm:"column:action"`
	Result        string    `gorm:"column:result"`
	SourceIP      *string   `gorm:"column:source_ip"`
	RequestID     *string   `gorm:"column:request_id"`
	SummaryChange string    `gorm:"column:summary_change"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (auditLogModel) TableName() string {
	return "audit_logs"
}

func newGormBackend(db *gorm.DB) *gormBackend {
	return &gormBackend{db: db}
}

func (b *gormBackend) Record(entry Entry) error {
	entry = normalizePersistentEntry(entry)
	model := auditLogModel{
		ID:            entry.ID,
		ActorID:       nullableString(entry.ActorID),
		ActorRole:     entry.ActorRole,
		TargetType:    entry.TargetType,
		TargetID:      nullableString(entry.TargetID),
		Action:        entry.Action,
		Result:        entry.Result,
		RequestID:     nullableString(entry.RequestID),
		SummaryChange: "{}",
		CreatedAt:     entry.CreatedAt,
	}

	return b.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"actor_id", "actor_role", "target_type", "target_id", "action", "result", "request_id", "created_at"}),
	}).Create(&model).Error
}

func (b *gormBackend) List(filter Filter) []Entry {
	query := b.db.Model(&auditLogModel{}).Order("created_at asc").Order("id asc")
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.TargetType != "" {
		query = query.Where("target_type = ?", filter.TargetType)
	}
	if filter.TargetID != "" {
		query = query.Where("target_id = ?", filter.TargetID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.Result != "" {
		query = query.Where("result = ?", filter.Result)
	}

	var rows []auditLogModel
	if err := query.Find(&rows).Error; err != nil {
		return nil
	}

	items := make([]Entry, 0, len(rows))
	for _, row := range rows {
		items = append(items, Entry{
			ID:         row.ID,
			ActorID:    derefString(row.ActorID),
			ActorRole:  row.ActorRole,
			TargetType: row.TargetType,
			TargetID:   derefString(row.TargetID),
			Action:     row.Action,
			Result:     row.Result,
			RequestID:  derefString(row.RequestID),
			CreatedAt:  row.CreatedAt,
		})
	}

	return items
}

func normalizeMemoryEntry(entry Entry, seq int) Entry {
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit-%d", seq)
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	return entry
}

func normalizePersistentEntry(entry Entry) Entry {
	if entry.ID == "" {
		entry.ID = idgen.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	return entry
}

func matchesFilter(entry Entry, filter Filter) bool {
	if filter.ActorID != "" && entry.ActorID != filter.ActorID {
		return false
	}
	if filter.TargetType != "" && entry.TargetType != filter.TargetType {
		return false
	}
	if filter.TargetID != "" && entry.TargetID != filter.TargetID {
		return false
	}
	if filter.Action != "" && entry.Action != filter.Action {
		return false
	}
	if filter.Result != "" && entry.Result != filter.Result {
		return false
	}
	return true
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
