package service

import (
	"context"

	"credit-management/backend/internal/platform/auditlog"
)

type Service struct {
	writer *auditlog.Writer
}

func New(writer *auditlog.Writer) *Service {
	return &Service{writer: writer}
}

func (s *Service) Record(ctx context.Context, entry auditlog.Entry) error {
	return s.writer.Record(ctx, entry)
}

func (s *Service) List(filter auditlog.Filter) []auditlog.Entry {
	return s.writer.List(filter)
}

func (s *Service) Export(filter auditlog.Filter) ([]byte, error) {
	return s.writer.ExportCSV(filter)
}
