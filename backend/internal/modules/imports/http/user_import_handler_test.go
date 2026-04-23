package http

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReadRowsFromUploadReturnsFileHeaderAndRows(t *testing.T) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "students.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("username,password,real_name,email,phone,student_id,grade,department_id\nstudent01,Passw0rd!,Student One,student01@example.com,13800138000,20240001,2024,class-01\n")); err != nil {
		t.Fatalf("write csv content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(http.MethodPost, "/imports/students/precheck", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request

	file, rows, err := readRowsFromUpload(ctx, "student")
	if err != nil {
		t.Fatalf("readRowsFromUpload returned error: %v", err)
	}
	if file == nil {
		t.Fatal("expected file header to be returned")
	}
	if file.Filename != "students.csv" {
		t.Fatalf("expected filename students.csv, got %s", file.Filename)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Username != "student01" {
		t.Fatalf("expected username student01, got %s", rows[0].Username)
	}
}
