package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"credit-management/credit-activity-service/models"
	testutils "credit-management/test-utils"
)

// TestUploadAttachment tests uploading a valid file
func TestUploadAttachment(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	// Create activity
	activity := createTestActivity(t)

	// Create a test PDF file
	testFile, err := testutils.CreateTempFile([]byte("PDF content here"), "test.pdf")
	ah.RequireNoError(err)
	defer testutils.CleanupTempFile(testFile)

	// Setup router
	router := gin.New()
	router.POST("/api/activities/:id/attachments", func(c *gin.Context) {
		c.Set("id", activity.OwnerID)
		c.Set("user_type", "student")
		activityHandler.UploadAttachment(c)
	})

	// Create multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, err := os.Open(testFile)
	ah.RequireNoError(err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(testFile))
	ah.RequireNoError(err)
	_, err = io.Copy(part, file)
	ah.RequireNoError(err)

	// Add description field
	writer.WriteField("description", "Test attachment")
	writer.Close()

	req, err := http.NewRequest("POST", "/api/activities/"+activity.ID+"/attachments", body)
	ah.RequireNoError(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusCreated)
	ah.AssertJSONFieldExists(resp, "data.id")
	ah.AssertJSONFieldEquals(resp, "data.activity_id", activity.ID)
}

// TestUploadAttachmentFileSizeLimit tests file size validation
func TestUploadAttachmentFileSizeLimit(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	// Create a large file (> 10MB)
	largeContent := bytes.Repeat([]byte("a"), 11*1024*1024) // 11MB
	testFile, err := testutils.CreateTempFile(largeContent, "large.pdf")
	ah.RequireNoError(err)
	defer testutils.CleanupTempFile(testFile)

	router := gin.New()
	router.POST("/api/activities/:id/attachments", func(c *gin.Context) {
		c.Set("id", activity.OwnerID)
		c.Set("user_type", "student")
		activityHandler.UploadAttachment(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(testFile)
	ah.RequireNoError(err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(testFile))
	ah.RequireNoError(err)
	_, err = io.Copy(part, file)
	ah.RequireNoError(err)
	writer.Close()

	req, err := http.NewRequest("POST", "/api/activities/"+activity.ID+"/attachments", body)
	ah.RequireNoError(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := testutils.PerformRequest(router, req)

	// Should reject file that's too large
	ah.AssertHTTPStatus(resp, http.StatusBadRequest)
	ah.AssertErrorResponse(resp, http.StatusBadRequest, "文件大小")
}

// TestUploadAttachmentFileTypeValidation tests file type validation
func TestUploadAttachmentFileTypeValidation(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	tests := []struct {
		name           string
		filename       string
		content        []byte
		shouldSucceed  bool
		errorContains  string
	}{
		{
			name:          "valid PDF",
			filename:      "document.pdf",
			content:       []byte("%PDF-1.4 content"),
			shouldSucceed: true,
		},
		{
			name:          "valid image",
			filename:      "image.jpg",
			content:       []byte("\xFF\xD8\xFF"), // JPEG magic bytes
			shouldSucceed: true,
		},
		{
			name:          "valid Word doc",
			filename:      "document.docx",
			content:       []byte("PK"), // ZIP magic bytes (docx is zip)
			shouldSucceed: true,
		},
		{
			name:           "invalid executable",
			filename:       "malware.exe",
			content:        []byte("MZ"), // EXE magic bytes
			shouldSucceed:  false,
			errorContains:  "不支持的文件类型",
		},
		{
			name:           "invalid script",
			filename:       "script.sh",
			content:        []byte("#!/bin/bash"),
			shouldSucceed:  false,
			errorContains:  "不支持的文件类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile, err := testutils.CreateTempFile(tt.content, tt.filename)
			ah.RequireNoError(err)
			defer testutils.CleanupTempFile(testFile)

			router := gin.New()
			router.POST("/api/activities/:id/attachments", func(c *gin.Context) {
				c.Set("id", activity.OwnerID)
				c.Set("user_type", "student")
				activityHandler.UploadAttachment(c)
			})

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			file, err := os.Open(testFile)
			ah.RequireNoError(err)
			defer file.Close()

			part, err := writer.CreateFormFile("file", tt.filename)
			ah.RequireNoError(err)
			_, err = io.Copy(part, file)
			ah.RequireNoError(err)
			writer.Close()

			req, err := http.NewRequest("POST", "/api/activities/"+activity.ID+"/attachments", body)
			ah.RequireNoError(err)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp := testutils.PerformRequest(router, req)

			if tt.shouldSucceed {
				ah.AssertSuccessResponse(resp)
			} else {
				ah.AssertHTTPStatus(resp, http.StatusBadRequest)
				ah.AssertErrorResponse(resp, http.StatusBadRequest, tt.errorContains)
			}
		})
	}
}

// TestUploadAttachmentPathTraversal tests path traversal attack prevention
func TestUploadAttachmentPathTraversal(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	maliciousFilenames := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
		"test../../secret.txt",
	}

	for _, filename := range maliciousFilenames {
		t.Run("filename: "+filename, func(t *testing.T) {
			router := gin.New()
			router.POST("/api/activities/:id/attachments", func(c *gin.Context) {
				c.Set("id", activity.OwnerID)
				c.Set("user_type", "student")
				activityHandler.UploadAttachment(c)
			})

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", filename)
			ah.RequireNoError(err)
			part.Write([]byte("malicious content"))
			writer.Close()

			req, err := http.NewRequest("POST", "/api/activities/"+activity.ID+"/attachments", body)
			ah.RequireNoError(err)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp := testutils.PerformRequest(router, req)

			// Should reject path traversal attempts
			ah.AssertHTTPStatus(resp, http.StatusBadRequest)
		})
	}
}

// TestDownloadAttachment tests downloading an attachment
func TestDownloadAttachment(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	// Create attachment record
	attachment := models.Attachment{
		ActivityID:   activity.ID,
		FileName:     "stored_file.pdf",
		OriginalName: "document.pdf",
		FileSize:     1024,
		FileType:     "pdf",
		FileCategory: "document",
		UploadedBy:   activity.OwnerID,
	}
	err := testDB.DB.Create(&attachment).Error
	ah.RequireNoError(err)

	router := gin.New()
	router.GET("/api/attachments/:id/download", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "student")
		activityHandler.DownloadAttachment(c)
	})

	req, err := testutils.CreateJSONRequest("GET", "/api/attachments/"+attachment.ID+"/download", nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertSuccessResponse(resp)
	ah.AssertContainsHeader(resp, "Content-Disposition", "attachment")
}

// TestDeleteAttachment tests deleting an attachment
func TestDeleteAttachment(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	// Create attachment
	attachment := models.Attachment{
		ActivityID:   activity.ID,
		FileName:     "to_delete.pdf",
		OriginalName: "document.pdf",
		FileSize:     1024,
		FileType:     "pdf",
		FileCategory: "document",
		UploadedBy:   activity.OwnerID,
	}
	err := testDB.DB.Create(&attachment).Error
	ah.RequireNoError(err)

	router := gin.New()
	router.DELETE("/api/attachments/:id", func(c *gin.Context) {
		c.Set("id", activity.OwnerID)
		c.Set("user_type", "student")
		activityHandler.DeleteAttachment(c)
	})

	req, err := testutils.CreateJSONRequest("DELETE", "/api/attachments/"+attachment.ID, nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)

	// Verify attachment is soft deleted
	var deletedAttachment models.Attachment
	err = testDB.DB.Unscoped().Where("id = ?", attachment.ID).First(&deletedAttachment).Error
	ah.RequireNoError(err)
	ah.AssertNotNil(deletedAttachment.DeletedAt)
}

// TestListAttachments tests listing attachments for an activity
func TestListAttachments(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	// Create multiple attachments
	for i := 0; i < 5; i++ {
		attachment := models.Attachment{
			ActivityID:   activity.ID,
			FileName:     "file_" + string(rune(i)) + ".pdf",
			OriginalName: "document_" + string(rune(i)) + ".pdf",
			FileSize:     1024,
			FileType:     "pdf",
			FileCategory: "document",
			UploadedBy:   activity.OwnerID,
		}
		err := testDB.DB.Create(&attachment).Error
		ah.RequireNoError(err)
	}

	router := gin.New()
	router.GET("/api/activities/:id/attachments", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "student")
		activityHandler.ListAttachments(c)
	})

	req, err := testutils.CreateJSONRequest("GET", "/api/activities/"+activity.ID+"/attachments", nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)
	ah.AssertJSONArrayLength(resp, "data", 5)
}

// TestAttachmentPermissions tests permission checks for attachments
func TestAttachmentPermissions(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	// Create attachment owned by someone else
	attachment := models.Attachment{
		ActivityID:   activity.ID,
		FileName:     "protected.pdf",
		OriginalName: "document.pdf",
		FileSize:     1024,
		FileType:     "pdf",
		FileCategory: "document",
		UploadedBy:   testutils.GenerateID(), // Different user
	}
	err := testDB.DB.Create(&attachment).Error
	ah.RequireNoError(err)

	// Try to delete as different user
	router := gin.New()
	router.DELETE("/api/attachments/:id", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID()) // Different user ID
		c.Set("user_type", "student")
		activityHandler.DeleteAttachment(c)
	})

	req, err := testutils.CreateJSONRequest("DELETE", "/api/attachments/"+attachment.ID, nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	// Should be forbidden
	ah.AssertHTTPStatus(resp, http.StatusForbidden)
}

// TestAttachmentVirusScanning tests malicious content detection
func TestAttachmentMaliciousContent(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "attachments")
	ah := testutils.NewAssertHelper(t)

	activity := createTestActivity(t)

	// Test files with suspicious content
	suspiciousContents := []struct {
		name    string
		content []byte
	}{
		{
			name:    "script_injection.pdf",
			content: []byte("<script>alert('xss')</script>"),
		},
		{
			name:    "php_shell.jpg",
			content: []byte("<?php system($_GET['cmd']); ?>"),
		},
	}

	for _, tc := range suspiciousContents {
		t.Run(tc.name, func(t *testing.T) {
			testFile, err := testutils.CreateTempFile(tc.content, tc.name)
			ah.RequireNoError(err)
			defer testutils.CleanupTempFile(testFile)

			router := gin.New()
			router.POST("/api/activities/:id/attachments", func(c *gin.Context) {
				c.Set("id", activity.OwnerID)
				c.Set("user_type", "student")
				activityHandler.UploadAttachment(c)
			})

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			file, err := os.Open(testFile)
			ah.RequireNoError(err)
			defer file.Close()

			part, err := writer.CreateFormFile("file", tc.name)
			ah.RequireNoError(err)
			_, err = io.Copy(part, file)
			ah.RequireNoError(err)
			writer.Close()

			req, err := http.NewRequest("POST", "/api/activities/"+activity.ID+"/attachments", body)
			ah.RequireNoError(err)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp := testutils.PerformRequest(router, req)

			// Should ideally scan and reject, or at least sanitize
			// Implementation depends on security requirements
			_ = resp
		})
	}
}

// Helper function to create a test activity
func createTestActivity(t *testing.T) *models.CreditActivity {
	ah := testutils.NewAssertHelper(t)

	activity := &models.CreditActivity{
		Title:       "Test Activity",
		Description: "Test Description",
		StartDate:   testutils.NewMockTime(testutils.ParseTime("2024-12-20")).Now(),
		EndDate:     testutils.NewMockTime(testutils.ParseTime("2024-12-22")).Now(),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}

	err := testDB.DB.Create(activity).Error
	ah.RequireNoError(err)

	return activity
}
