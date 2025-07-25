package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/http"
	"github.com/robmorgan/infraspec/test/httpserver"
)

func TestHTTPAssertions(t *testing.T) {
	// Create mock server
	mockServer := httpserver.NewMockHTTPServer()
	defer mockServer.Close()

	// Create HTTP asserter
	httpAsserter := http.NewHTTPAsserter()

	t.Run("AssertResponseStatus", func(t *testing.T) {
		// Test successful request
		err := httpAsserter.AssertResponseStatus("GET", mockServer.URL()+"/json", 200, nil)
		assert.NoError(t, err)

		// Test 404
		err = httpAsserter.AssertResponseStatus("GET", mockServer.URL()+"/status/404", 404, nil)
		assert.NoError(t, err)

		// Test wrong status expectation
		err = httpAsserter.AssertResponseStatus("GET", mockServer.URL()+"/json", 404, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected status 404, got 200")
	})

	t.Run("AssertResponseStatusWithBody", func(t *testing.T) {
		err := httpAsserter.AssertResponseStatusWithBody("POST", mockServer.URL()+"/echo", "test body", 200, nil)
		assert.NoError(t, err)
	})

	t.Run("AssertResponseContains", func(t *testing.T) {
		err := httpAsserter.AssertResponseContains("GET", mockServer.URL()+"/text", "Hello, World!", nil)
		assert.NoError(t, err)

		// Test content not found
		err = httpAsserter.AssertResponseContains("GET", mockServer.URL()+"/text", "Not Found", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not contain expected content")
	})

	t.Run("AssertResponseJSON", func(t *testing.T) {
		// Valid JSON
		err := httpAsserter.AssertResponseJSON("GET", mockServer.URL()+"/json", nil)
		assert.NoError(t, err)

		// Invalid JSON
		err = httpAsserter.AssertResponseJSON("GET", mockServer.URL()+"/text", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response is not valid JSON")
	})

	t.Run("AssertResponseHeader", func(t *testing.T) {
		err := httpAsserter.AssertResponseHeader("GET", mockServer.URL()+"/json", "Content-Type", "application/json", nil)
		assert.NoError(t, err)

		// Wrong header value
		err = httpAsserter.AssertResponseHeader("GET", mockServer.URL()+"/json", "Content-Type", "text/plain", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected header 'Content-Type' to be 'text/plain'")
	})

	t.Run("RequestWithHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer test-token",
			"X-Custom":      "test-value",
		}

		err := httpAsserter.AssertResponseStatus("GET", mockServer.URL()+"/headers", 200, headers)
		assert.NoError(t, err)
	})

	t.Run("UploadFile", func(t *testing.T) {
		// Create a temporary file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		content := "Hello, World!"
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(t, err)

		// Set base directory
		httpAsserter.SetBaseDirectory(tempDir)

		// Test file upload
		formData := map[string]string{
			"uuid": "test-uuid",
			"type": "document",
		}

		err = httpAsserter.UploadFile(mockServer.URL()+"/upload", "file", "test.txt", nil, formData)
		assert.NoError(t, err)
	})

	t.Run("AssertExists", func(t *testing.T) {
		// Valid endpoint
		err := httpAsserter.AssertExists("http_endpoint", mockServer.URL()+"/json")
		assert.NoError(t, err)

		// Invalid resource type
		err = httpAsserter.AssertExists("invalid_type", mockServer.URL()+"/json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported resource type")
	})

	t.Run("AssertTags", func(t *testing.T) {
		// Tags are not supported for HTTP resources
		err := httpAsserter.AssertTags("http_endpoint", "test", map[string]string{"key": "value"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tags are not supported for HTTP resources")
	})
}

func TestHTTPAsserterImplementsInterface(t *testing.T) {
	// Ensure HTTPAsserter implements the base Asserter interface
	var _ assertions.Asserter = (*http.HTTPAsserter)(nil)

	// Ensure HTTPAsserter implements the HTTPAssertions interface
	var _ http.HTTPAssertions = (*http.HTTPAsserter)(nil)
}

func TestHTTPAsserterFactory(t *testing.T) {
	asserter, err := assertions.New("http")
	require.NoError(t, err)
	require.NotNil(t, asserter)

	// Test that it can be cast to HTTPAssertions
	httpAsserter, ok := asserter.(http.HTTPAssertions)
	assert.True(t, ok)
	assert.NotNil(t, httpAsserter)
}

func TestHTTPAsserterWithTimeout(t *testing.T) {
	httpAsserter := http.NewHTTPAsserter()

	// Test with a non-existent endpoint (should timeout/fail)
	err := httpAsserter.AssertResponseStatus("GET", "http://localhost:99999/nonexistent", 200, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP request failed")
}
