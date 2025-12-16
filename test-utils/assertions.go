package testutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertHelper provides assertion utilities for tests
type AssertHelper struct {
	t *testing.T
}

// NewAssertHelper creates a new assertion helper
func NewAssertHelper(t *testing.T) *AssertHelper {
	return &AssertHelper{t: t}
}

// AssertHTTPStatus checks if the HTTP response has the expected status code
func (ah *AssertHelper) AssertHTTPStatus(resp *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(ah.t, expectedStatus, resp.Code,
		"Expected status code %d, got %d. Response body: %s",
		expectedStatus, resp.Code, resp.Body.String())
}

// AssertJSONResponse checks if the response is valid JSON and matches expected structure
func (ah *AssertHelper) AssertJSONResponse(resp *httptest.ResponseRecorder, expectedData interface{}) {
	assert.Equal(ah.t, "application/json", resp.Header().Get("Content-Type"),
		"Expected Content-Type to be application/json")

	var actual interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &actual)
	require.NoError(ah.t, err, "Response should be valid JSON")

	if expectedData != nil {
		expectedJSON, _ := json.Marshal(expectedData)
		actualJSON, _ := json.Marshal(actual)
		assert.JSONEq(ah.t, string(expectedJSON), string(actualJSON))
	}
}

// AssertJSONFieldEquals checks if a specific JSON field has the expected value
func (ah *AssertHelper) AssertJSONFieldEquals(resp *httptest.ResponseRecorder, fieldPath string, expectedValue interface{}) {
	var data map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &data)
	require.NoError(ah.t, err, "Response should be valid JSON")

	// Navigate nested fields using dot notation (e.g., "data.user.name")
	fields := strings.Split(fieldPath, ".")
	var current interface{} = data

	for i, field := range fields {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[field]
		} else {
			ah.t.Errorf("Field path %s not found (stopped at %s)", fieldPath, fields[:i])
			return
		}
	}

	assert.Equal(ah.t, expectedValue, current,
		"Field %s should equal %v, got %v", fieldPath, expectedValue, current)
}

// AssertJSONFieldExists checks if a specific JSON field exists
func (ah *AssertHelper) AssertJSONFieldExists(resp *httptest.ResponseRecorder, fieldPath string) {
	var data map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &data)
	require.NoError(ah.t, err, "Response should be valid JSON")

	fields := strings.Split(fieldPath, ".")
	var current interface{} = data

	for _, field := range fields {
		if m, ok := current.(map[string]interface{}); ok {
			val, exists := m[field]
			if !exists {
				ah.t.Errorf("Field %s does not exist in response", fieldPath)
				return
			}
			current = val
		} else {
			ah.t.Errorf("Field %s does not exist in response", fieldPath)
			return
		}
	}
}

// AssertJSONArrayLength checks if a JSON array has the expected length
func (ah *AssertHelper) AssertJSONArrayLength(resp *httptest.ResponseRecorder, arrayPath string, expectedLength int) {
	var data map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &data)
	require.NoError(ah.t, err, "Response should be valid JSON")

	// Navigate to the array
	fields := strings.Split(arrayPath, ".")
	var current interface{} = data

	for _, field := range fields {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[field]
		} else {
			ah.t.Errorf("Array path %s not found", arrayPath)
			return
		}
	}

	if arr, ok := current.([]interface{}); ok {
		assert.Equal(ah.t, expectedLength, len(arr),
			"Array %s should have length %d, got %d", arrayPath, expectedLength, len(arr))
	} else {
		ah.t.Errorf("Field %s is not an array", arrayPath)
	}
}

// AssertErrorResponse checks if the response contains an error message
func (ah *AssertHelper) AssertErrorResponse(resp *httptest.ResponseRecorder, expectedStatus int, expectedErrorMsg string) {
	ah.AssertHTTPStatus(resp, expectedStatus)

	var errorResp map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(ah.t, err, "Error response should be valid JSON")

	if expectedErrorMsg != "" {
		errorMsg, ok := errorResp["error"].(string)
		if !ok {
			errorMsg, ok = errorResp["message"].(string)
		}
		require.True(ah.t, ok, "Response should contain error or message field")
		assert.Contains(ah.t, errorMsg, expectedErrorMsg,
			"Error message should contain '%s'", expectedErrorMsg)
	}
}

// AssertSuccessResponse checks if the response indicates success
func (ah *AssertHelper) AssertSuccessResponse(resp *httptest.ResponseRecorder) {
	assert.True(ah.t, resp.Code >= 200 && resp.Code < 300,
		"Expected success status code (2xx), got %d", resp.Code)
}

// AssertValidationError checks for validation errors in response
func (ah *AssertHelper) AssertValidationError(resp *httptest.ResponseRecorder, field string) {
	ah.AssertHTTPStatus(resp, http.StatusBadRequest)

	var errorResp map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(ah.t, err, "Error response should be valid JSON")

	// Check if errors field exists and contains the field
	if errors, ok := errorResp["errors"].(map[string]interface{}); ok {
		_, fieldExists := errors[field]
		assert.True(ah.t, fieldExists,
			"Validation errors should include field '%s'", field)
	} else {
		// Check if error message mentions the field
		errorMsg := fmt.Sprintf("%v", errorResp["error"])
		assert.Contains(ah.t, errorMsg, field,
			"Error message should mention field '%s'", field)
	}
}

// AssertContainsHeader checks if response contains a specific header
func (ah *AssertHelper) AssertContainsHeader(resp *httptest.ResponseRecorder, headerName, expectedValue string) {
	actualValue := resp.Header().Get(headerName)
	assert.Equal(ah.t, expectedValue, actualValue,
		"Header %s should be %s, got %s", headerName, expectedValue, actualValue)
}

// AssertRedirect checks if the response is a redirect to the expected location
func (ah *AssertHelper) AssertRedirect(resp *httptest.ResponseRecorder, expectedLocation string) {
	assert.True(ah.t, resp.Code >= 300 && resp.Code < 400,
		"Expected redirect status code (3xx), got %d", resp.Code)

	location := resp.Header().Get("Location")
	assert.Equal(ah.t, expectedLocation, location,
		"Expected redirect to %s, got %s", expectedLocation, location)
}

// AssertTimeAlmostEqual checks if two times are almost equal (within tolerance)
func (ah *AssertHelper) AssertTimeAlmostEqual(expected, actual time.Time, tolerance time.Duration) {
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	assert.True(ah.t, diff <= tolerance,
		"Time difference %v exceeds tolerance %v", diff, tolerance)
}

// AssertSliceContains checks if a slice contains a specific element
func (ah *AssertHelper) AssertSliceContains(slice interface{}, element interface{}) {
	sliceValue := reflect.ValueOf(slice)
	require.Equal(ah.t, reflect.Slice, sliceValue.Kind(),
		"Expected a slice, got %T", slice)

	found := false
	for i := 0; i < sliceValue.Len(); i++ {
		if reflect.DeepEqual(sliceValue.Index(i).Interface(), element) {
			found = true
			break
		}
	}

	assert.True(ah.t, found,
		"Slice should contain element %v", element)
}

// AssertSliceNotContains checks if a slice does not contain a specific element
func (ah *AssertHelper) AssertSliceNotContains(slice interface{}, element interface{}) {
	sliceValue := reflect.ValueOf(slice)
	require.Equal(ah.t, reflect.Slice, sliceValue.Kind(),
		"Expected a slice, got %T", slice)

	for i := 0; i < sliceValue.Len(); i++ {
		if reflect.DeepEqual(sliceValue.Index(i).Interface(), element) {
			ah.t.Errorf("Slice should not contain element %v", element)
			return
		}
	}
}

// AssertMapContainsKey checks if a map contains a specific key
func (ah *AssertHelper) AssertMapContainsKey(m interface{}, key interface{}) {
	mapValue := reflect.ValueOf(m)
	require.Equal(ah.t, reflect.Map, mapValue.Kind(),
		"Expected a map, got %T", m)

	keyValue := reflect.ValueOf(key)
	found := mapValue.MapIndex(keyValue).IsValid()

	assert.True(ah.t, found,
		"Map should contain key %v", key)
}

// AssertNil is a convenience wrapper for assert.Nil
func (ah *AssertHelper) AssertNil(object interface{}, msgAndArgs ...interface{}) {
	assert.Nil(ah.t, object, msgAndArgs...)
}

// AssertNotNil is a convenience wrapper for assert.NotNil
func (ah *AssertHelper) AssertNotNil(object interface{}, msgAndArgs ...interface{}) {
	assert.NotNil(ah.t, object, msgAndArgs...)
}

// AssertEqual is a convenience wrapper for assert.Equal
func (ah *AssertHelper) AssertEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.Equal(ah.t, expected, actual, msgAndArgs...)
}

// AssertNotEqual is a convenience wrapper for assert.NotEqual
func (ah *AssertHelper) AssertNotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.NotEqual(ah.t, expected, actual, msgAndArgs...)
}

// AssertTrue is a convenience wrapper for assert.True
func (ah *AssertHelper) AssertTrue(value bool, msgAndArgs ...interface{}) {
	assert.True(ah.t, value, msgAndArgs...)
}

// AssertFalse is a convenience wrapper for assert.False
func (ah *AssertHelper) AssertFalse(value bool, msgAndArgs ...interface{}) {
	assert.False(ah.t, value, msgAndArgs...)
}

// AssertNoError is a convenience wrapper for assert.NoError
func (ah *AssertHelper) AssertNoError(err error, msgAndArgs ...interface{}) {
	assert.NoError(ah.t, err, msgAndArgs...)
}

// AssertError is a convenience wrapper for assert.Error
func (ah *AssertHelper) AssertError(err error, msgAndArgs ...interface{}) {
	assert.Error(ah.t, err, msgAndArgs...)
}

// AssertContains is a convenience wrapper for assert.Contains
func (ah *AssertHelper) AssertContains(container, element interface{}, msgAndArgs ...interface{}) {
	assert.Contains(ah.t, container, element, msgAndArgs...)
}

// AssertNotContains is a convenience wrapper for assert.NotContains
func (ah *AssertHelper) AssertNotContains(container, element interface{}, msgAndArgs ...interface{}) {
	assert.NotContains(ah.t, container, element, msgAndArgs...)
}

// AssertEmpty is a convenience wrapper for assert.Empty
func (ah *AssertHelper) AssertEmpty(object interface{}, msgAndArgs ...interface{}) {
	assert.Empty(ah.t, object, msgAndArgs...)
}

// AssertNotEmpty is a convenience wrapper for assert.NotEmpty
func (ah *AssertHelper) AssertNotEmpty(object interface{}, msgAndArgs ...interface{}) {
	assert.NotEmpty(ah.t, object, msgAndArgs...)
}

// AssertLen is a convenience wrapper for assert.Len
func (ah *AssertHelper) AssertLen(object interface{}, length int, msgAndArgs ...interface{}) {
	assert.Len(ah.t, object, length, msgAndArgs...)
}

// RequireNoError is a convenience wrapper for require.NoError (stops test on failure)
func (ah *AssertHelper) RequireNoError(err error, msgAndArgs ...interface{}) {
	require.NoError(ah.t, err, msgAndArgs...)
}

// RequireNotNil is a convenience wrapper for require.NotNil (stops test on failure)
func (ah *AssertHelper) RequireNotNil(object interface{}, msgAndArgs ...interface{}) {
	require.NotNil(ah.t, object, msgAndArgs...)
}

// AssertPaginationResponse checks if the response contains valid pagination metadata
func (ah *AssertHelper) AssertPaginationResponse(resp *httptest.ResponseRecorder, expectedTotal int) {
	var data map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &data)
	require.NoError(ah.t, err, "Response should be valid JSON")

	// Check for pagination fields
	if pagination, ok := data["pagination"].(map[string]interface{}); ok {
		if expectedTotal >= 0 {
			total := int(pagination["total"].(float64))
			assert.Equal(ah.t, expectedTotal, total,
				"Total items should be %d, got %d", expectedTotal, total)
		}

		// Check for common pagination fields
		assert.Contains(ah.t, pagination, "page", "Pagination should include page")
		assert.Contains(ah.t, pagination, "page_size", "Pagination should include page_size")
		assert.Contains(ah.t, pagination, "total", "Pagination should include total")
	} else {
		ah.t.Error("Response should contain pagination metadata")
	}
}
