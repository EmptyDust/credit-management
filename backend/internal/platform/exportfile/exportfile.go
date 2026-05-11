package exportfile

import (
	"bytes"
	"encoding/csv"

	"github.com/xuri/excelize/v2"
)

const (
	FormatCSV  = "csv"
	FormatXLSX = "xlsx"

	ContentTypeCSV  = "text/csv; charset=utf-8"
	ContentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

type File struct {
	Filename    string
	ContentType string
	Content     []byte
}

func NormalizeFormat(format string) string {
	switch format {
	case FormatXLSX:
		return FormatXLSX
	default:
		return FormatCSV
	}
}

func WriteCSV(rows [][]string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func WriteXLSX(sheetName string, rows [][]string) ([]byte, error) {
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	workbook := excelize.NewFile()
	defer func() {
		_ = workbook.Close()
	}()

	defaultSheet := workbook.GetSheetName(0)
	if err := workbook.SetSheetName(defaultSheet, sheetName); err != nil {
		return nil, err
	}

	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				return nil, err
			}
			if err := workbook.SetCellValue(sheetName, cell, value); err != nil {
				return nil, err
			}
		}
	}

	buffer := &bytes.Buffer{}
	if err := workbook.Write(buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
