package testutils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateID generates a random unique ID
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateShortID generates a shorter random ID
func GenerateShortID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateJSONRequest creates a test HTTP request with JSON body
func CreateJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// CreateFormRequest creates a test HTTP request with form data
func CreateFormRequest(method, url string, formData map[string]string) (*http.Request, error) {
	form := strings.NewReader("")
	if len(formData) > 0 {
		values := make([]string, 0, len(formData))
		for key, value := range formData {
			values = append(values, fmt.Sprintf("%s=%s", key, value))
		}
		form = strings.NewReader(strings.Join(values, "&"))
	}

	req, err := http.NewRequest(method, url, form)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// CreateMultipartRequest creates a test HTTP request with multipart form data
func CreateMultipartRequest(method, url string, fields map[string]string, files map[string]string) (*http.Request, *bytes.Buffer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add fields
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, nil, err
		}
	}

	// Add files
	for fieldName, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			return nil, nil, err
		}

		if _, err := io.Copy(part, file); err != nil {
			return nil, nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, body, nil
}

// CreateMultipartRequestWithBytes creates a multipart request with file bytes
func CreateMultipartRequestWithBytes(method, url string, fields map[string]string, files map[string][]byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add fields
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}

	// Add files
	for fieldName, fileBytes := range files {
		part, err := writer.CreateFormFile(fieldName, fieldName)
		if err != nil {
			return nil, err
		}

		if _, err := part.Write(fileBytes); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

// AddAuthHeader adds an authorization header to a request
func AddAuthHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

// AddCustomHeader adds a custom header to a request
func AddCustomHeader(req *http.Request, key, value string) {
	req.Header.Set(key, value)
}

// PerformRequest performs a request on a Gin engine and returns the response
func PerformRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ParseJSONResponse parses the JSON response body into a target interface
func ParseJSONResponse(resp *httptest.ResponseRecorder, target interface{}) error {
	return json.Unmarshal(resp.Body.Bytes(), target)
}

// CreateTempFile creates a temporary file with given content
func CreateTempFile(content []byte, filename string) (string, error) {
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

// CleanupTempFile removes a temporary file
func CleanupTempFile(filePath string) error {
	return os.Remove(filePath)
}

// MockTime mocks the current time for testing
type MockTime struct {
	currentTime time.Time
}

// NewMockTime creates a new mock time instance
func NewMockTime(t time.Time) *MockTime {
	return &MockTime{currentTime: t}
}

// Now returns the mocked current time
func (m *MockTime) Now() time.Time {
	return m.currentTime
}

// Advance advances the mock time by a duration
func (m *MockTime) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

// Set sets the mock time to a specific time
func (m *MockTime) Set(t time.Time) {
	m.currentTime = t
}

// WaitForCondition waits for a condition to be true or times out
func WaitForCondition(condition func() bool, timeout time.Duration, checkInterval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(checkInterval)
	}
	return false
}

// Retry retries a function until it succeeds or max retries reached
func Retry(fn func() error, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// RandomEmail generates a random email address
func RandomEmail() string {
	return fmt.Sprintf("%s@example.com", RandomString(10))
}

// RandomUsername generates a random username
func RandomUsername() string {
	return "user_" + RandomString(8)
}

// ToPtr converts a value to a pointer
func ToPtr[T any](v T) *T {
	return &v
}

// FromPtr returns the value from a pointer or a default value if nil
func FromPtr[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// ContainsString checks if a string slice contains a specific string
func ContainsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// StringSliceEqual checks if two string slices are equal
func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// MergeMaps merges multiple maps into one
func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// GetEnvOrDefault gets an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestEnv sets up test environment variables
func SetupTestEnv(env map[string]string) func() {
	original := make(map[string]string)
	for key, value := range env {
		original[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key, value := range original {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}
}

// CompareFloats compares two floats with a tolerance
func CompareFloats(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Min returns the minimum of two comparable values
func Min[T int | int64 | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two comparable values
func Max[T int | int64 | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}
