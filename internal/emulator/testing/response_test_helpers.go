package testing

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// AssertResponseStatus validates the HTTP status code
func AssertResponseStatus(t *testing.T, resp *emulator.AWSResponse, expectedStatus int) {
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Errorf("expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}
}

// AssertHeader validates that a header exists and has the expected value
func AssertHeader(t *testing.T, resp *emulator.AWSResponse, headerName, expectedValue string) {
	t.Helper()
	actualValue := resp.Headers[headerName]
	if actualValue != expectedValue {
		t.Errorf("expected header %s=%s, got %s", headerName, expectedValue, actualValue)
	}
}

// AssertHeaderContains validates that a header exists and contains the expected substring
func AssertHeaderContains(t *testing.T, resp *emulator.AWSResponse, headerName, expectedSubstring string) {
	t.Helper()
	actualValue := resp.Headers[headerName]
	if !strings.Contains(actualValue, expectedSubstring) {
		t.Errorf("expected header %s to contain %s, got %s", headerName, expectedSubstring, actualValue)
	}
}

// AssertContentType validates the Content-Type header
func AssertContentType(t *testing.T, resp *emulator.AWSResponse, expectedContentType string) {
	t.Helper()
	AssertHeader(t, resp, "Content-Type", expectedContentType)
}

// AssertRequestID validates that a RequestId is present (in header or body)
func AssertRequestID(t *testing.T, resp *emulator.AWSResponse) {
	t.Helper()
	// Check headers first
	if resp.Headers["x-amzn-RequestId"] != "" || resp.Headers["x-amz-request-id"] != "" {
		return
	}

	// Check XML body for RequestId (PascalCase for Query protocol, lowercase for EC2)
	body := string(resp.Body)
	if strings.Contains(body, "<RequestId>") || strings.Contains(body, "<requestId>") {
		return
	}

	t.Error("RequestId not found in headers or body")
}

// ValidateResponse validates a response using the response validator
func ValidateResponse(t *testing.T, validator *emulator.ResponseValidator, serviceName, action string, resp *emulator.AWSResponse) {
	t.Helper()
	if err := validator.ValidateResponse(serviceName, action, resp); err != nil {
		t.Errorf("response validation failed: %v", err)
	}

	if err := validator.ValidateResponseHeaders(resp, serviceName); err != nil {
		t.Errorf("response headers validation failed: %v", err)
	}

	if err := validator.ValidateResponseStatusCode(resp.StatusCode); err != nil {
		t.Errorf("response status code validation failed: %v", err)
	}
}

// CompareWithGoldenFile compares a response body with a golden file
func CompareWithGoldenFile(t *testing.T, resp *emulator.AWSResponse, goldenFilePath string) {
	t.Helper()
	goldenData, err := os.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenFilePath, err)
	}

	// Normalize both responses before comparison
	normalizedResp := normalizeResponseBody(resp.Body, resp.Headers["Content-Type"])
	normalizedGolden := normalizeResponseBody(goldenData, resp.Headers["Content-Type"])

	if string(normalizedResp) != string(normalizedGolden) {
		t.Errorf("response body does not match golden file\nExpected:\n%s\nGot:\n%s", string(normalizedGolden), string(normalizedResp))
	}
}

// normalizeResponseBody normalizes a response body by removing dynamic fields
// like RequestId, timestamps, etc. for comparison
func normalizeResponseBody(body []byte, contentType string) []byte {
	bodyStr := string(body)

	// Remove RequestId from XML
	bodyStr = strings.ReplaceAll(bodyStr, `<RequestId>.*?</RequestId>`, `<RequestId>NORMALIZED</RequestId>`)
	bodyStr = strings.ReplaceAll(bodyStr, `<RequestId>.*?</RequestId>`, `<RequestId>NORMALIZED</RequestId>`)

	// Remove RequestId from JSON
	if strings.Contains(contentType, "json") {
		// Try to parse and normalize JSON
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(bodyStr), &jsonData); err == nil {
			// Remove dynamic fields
			delete(jsonData, "x-amzn-RequestId")
			delete(jsonData, "RequestId")
			normalized, _ := json.Marshal(jsonData)
			return normalized
		}
	}

	return []byte(bodyStr)
}

// AssertXMLStructure validates that XML response has expected structure
func AssertXMLStructure(t *testing.T, resp *emulator.AWSResponse, expectedRootElement string) {
	t.Helper()
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, fmt.Sprintf("<%s", expectedRootElement)) {
		t.Errorf("expected XML root element <%s>, not found in response", expectedRootElement)
	}
}

// AssertJSONField validates that a JSON response contains a field with expected value
func AssertJSONField(t *testing.T, resp *emulator.AWSResponse, fieldPath string, expectedValue interface{}) {
	t.Helper()
	var jsonData map[string]interface{}
	if err := json.Unmarshal(resp.Body, &jsonData); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	// Navigate nested fields using dot notation
	value := getNestedField(jsonData, fieldPath)
	if !reflect.DeepEqual(value, expectedValue) {
		t.Errorf("expected field %s=%v, got %v", fieldPath, expectedValue, value)
	}
}

// getNestedField gets a nested field from a map using dot notation
func getNestedField(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// AssertErrorResponse validates an error response structure
func AssertErrorResponse(t *testing.T, resp *emulator.AWSResponse, expectedCode string, protocol emulator.ProtocolType) {
	t.Helper()
	if resp.StatusCode < 400 {
		t.Errorf("expected error status code (>=400), got %d", resp.StatusCode)
	}

	switch protocol {
	case emulator.ProtocolQuery, emulator.ProtocolRESTXML:
		// XML error format
		bodyStr := string(resp.Body)
		if !strings.Contains(bodyStr, fmt.Sprintf("<Code>%s</Code>", expectedCode)) {
			t.Errorf("expected error code %s in XML response", expectedCode)
		}

	case emulator.ProtocolJSON:
		// JSON error format with __type
		var jsonData map[string]interface{}
		if err := json.Unmarshal(resp.Body, &jsonData); err != nil {
			t.Fatalf("failed to parse JSON error response: %v", err)
		}
		if jsonData["__type"] != expectedCode {
			t.Errorf("expected error code %s, got %v", expectedCode, jsonData["__type"])
		}

	case emulator.ProtocolRESTJSON:
		// REST-JSON error format
		var jsonData map[string]interface{}
		if err := json.Unmarshal(resp.Body, &jsonData); err != nil {
			t.Fatalf("failed to parse JSON error response: %v", err)
		}
		// REST-JSON errors may have Type field
		if jsonData["Type"] == nil && jsonData["__type"] == nil {
			t.Error("expected error response to have Type or __type field")
		}
	}
}

// LoadGoldenFile loads a golden file from the testdata directory
func LoadGoldenFile(t *testing.T, relativePath string) []byte {
	t.Helper()
	// Try multiple possible locations
	possiblePaths := []string{
		filepath.Join("testdata", relativePath),
		filepath.Join("..", "testdata", relativePath),
		filepath.Join("..", "..", "testdata", relativePath),
	}

	for _, path := range possiblePaths {
		if data, err := os.ReadFile(path); err == nil {
			return data
		}
	}

	t.Fatalf("golden file not found: %s (tried: %v)", relativePath, possiblePaths)
	return nil
}

// WriteGoldenFile writes a response to a golden file (for updating golden files)
func WriteGoldenFile(t *testing.T, relativePath string, data []byte) {
	t.Helper()
	fullPath := filepath.Join("testdata", relativePath)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		t.Fatalf("failed to write golden file: %v", err)
	}
}

// AssertResponseMatchesType validates that response body matches expected Go type structure
func AssertResponseMatchesType(t *testing.T, resp *emulator.AWSResponse, expectedType reflect.Type) {
	t.Helper()
	contentType := resp.Headers["Content-Type"]

	if strings.Contains(contentType, "json") {
		// Try to unmarshal into the expected type
		value := reflect.New(expectedType).Interface()
		if err := json.Unmarshal(resp.Body, value); err != nil {
			t.Errorf("failed to unmarshal JSON into expected type %s: %v", expectedType.Name(), err)
		}
	} else if strings.Contains(contentType, "xml") {
		// Try to unmarshal XML into the expected type
		value := reflect.New(expectedType).Interface()
		if err := xml.Unmarshal(resp.Body, value); err != nil {
			t.Errorf("failed to unmarshal XML into expected type %s: %v", expectedType.Name(), err)
		}
	}
}

// ExtractRequestID extracts RequestId from response (header or body)
func ExtractRequestID(resp *emulator.AWSResponse) string {
	// Check headers first
	if id := resp.Headers["x-amzn-RequestId"]; id != "" {
		return id
	}
	if id := resp.Headers["x-amz-request-id"]; id != "" {
		return id
	}

	body := string(resp.Body)

	// Try to extract from XML body (PascalCase for Query protocol)
	if strings.Contains(body, "<RequestId>") {
		start := strings.Index(body, "<RequestId>")
		if start != -1 {
			start += len("<RequestId>")
			end := strings.Index(body[start:], "</RequestId>")
			if end != -1 {
				return body[start : start+end]
			}
		}
	}

	// Try to extract from XML body (lowercase for EC2 protocol)
	if strings.Contains(body, "<requestId>") {
		start := strings.Index(body, "<requestId>")
		if start != -1 {
			start += len("<requestId>")
			end := strings.Index(body[start:], "</requestId>")
			if end != -1 {
				return body[start : start+end]
			}
		}
	}

	return ""
}

