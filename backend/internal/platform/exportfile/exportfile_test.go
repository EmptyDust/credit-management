package exportfile

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestWriteXLSXCreatesWorkbookWithRows(t *testing.T) {
	content, err := WriteXLSX("Activities", [][]string{
		{"title", "category"},
		{"Innovation Lab", "innovation"},
	})
	if err != nil {
		t.Fatalf("WriteXLSX returned error: %v", err)
	}

	workbook, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("open xlsx content: %v", err)
	}
	defer func() {
		if closeErr := workbook.Close(); closeErr != nil {
			t.Fatalf("close workbook: %v", closeErr)
		}
	}()

	rows, err := workbook.GetRows("Activities")
	if err != nil {
		t.Fatalf("read sheet rows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "title" || rows[1][0] != "Innovation Lab" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}
