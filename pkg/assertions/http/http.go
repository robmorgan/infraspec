package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HTTPAsserter implements assertions for HTTP requests
type HTTPAsserter struct {
	client   *http.Client
	baseDir  string // Directory where feature files are located for relative file uploads
	lastResp *http.Response
	lastBody []byte
}

// HTTPAssertions defines HTTP-specific assertions
type HTTPAssertions interface {
	SetBaseDirectory(dir string)
	AssertResponseStatus(method, url string, expectedStatus int, headers map[string]string) error
	AssertResponseStatusWithBody(method, url, body string, expectedStatus int, headers map[string]string) error
	AssertResponseContains(method, url, expectedContent string, headers map[string]string) error
	AssertResponseJSON(method, url string, headers map[string]string) error
	AssertResponseHeader(method, url, headerName, expectedValue string, headers map[string]string) error
	UploadFile(url, fieldName, filePath string, headers map[string]string, formData map[string]string) error
}

// NewHTTPAsserter creates a new HTTPAsserter instance
func NewHTTPAsserter() *HTTPAsserter {
	return &HTTPAsserter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetBaseDirectory sets the base directory for relative file paths
func (h *HTTPAsserter) SetBaseDirectory(dir string) {
	h.baseDir = dir
}

// AssertExists checks if an HTTP endpoint is reachable (returns non-error status)
func (h *HTTPAsserter) AssertExists(resourceType, resourceName string) error {
	if resourceType != "http_endpoint" {
		return fmt.Errorf("unsupported resource type for HTTP asserter: %s", resourceType)
	}

	return h.AssertResponseStatus("GET", resourceName, 200, nil)
}

// AssertTags is not applicable for HTTP resources
func (h *HTTPAsserter) AssertTags(resourceType, resourceName string, tags map[string]string) error {
	return fmt.Errorf("tags are not supported for HTTP resources")
}

// AssertResponseStatus checks if an HTTP request returns the expected status code
func (h *HTTPAsserter) AssertResponseStatus(method, url string, expectedStatus int, headers map[string]string) error {
	return h.AssertResponseStatusWithBody(method, url, "", expectedStatus, headers)
}

// AssertResponseStatusWithBody checks if an HTTP request with body returns the expected status code
func (h *HTTPAsserter) AssertResponseStatusWithBody(method, url, body string, expectedStatus int, headers map[string]string) error {
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Store response for subsequent assertions
	h.lastResp = resp
	h.lastBody, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d for %s %s", expectedStatus, resp.StatusCode, method, url)
	}

	return nil
}

// AssertResponseContains checks if the HTTP response body contains the expected content
func (h *HTTPAsserter) AssertResponseContains(method, url, expectedContent string, headers map[string]string) error {
	err := h.AssertResponseStatus(method, url, 200, headers)
	if err != nil {
		return err
	}

	bodyStr := string(h.lastBody)
	if !strings.Contains(bodyStr, expectedContent) {
		return fmt.Errorf("response body does not contain expected content '%s'. Got: %s", expectedContent, bodyStr)
	}

	return nil
}

// AssertResponseJSON checks if the HTTP response is valid JSON
func (h *HTTPAsserter) AssertResponseJSON(method, url string, headers map[string]string) error {
	err := h.AssertResponseStatus(method, url, 200, headers)
	if err != nil {
		return err
	}

	var jsonData interface{}
	if err := json.Unmarshal(h.lastBody, &jsonData); err != nil {
		return fmt.Errorf("response is not valid JSON: %w. Response body: %s", err, string(h.lastBody))
	}

	return nil
}

// AssertResponseHeader checks if the HTTP response has the expected header value
func (h *HTTPAsserter) AssertResponseHeader(method, url, headerName, expectedValue string, headers map[string]string) error {
	err := h.AssertResponseStatus(method, url, 200, headers)
	if err != nil {
		return err
	}

	actualValue := h.lastResp.Header.Get(headerName)
	if actualValue != expectedValue {
		return fmt.Errorf("expected header '%s' to be '%s', got '%s'", headerName, expectedValue, actualValue)
	}

	return nil
}

// UploadFile uploads a file using multipart/form-data
func (h *HTTPAsserter) UploadFile(url, fieldName, filePath string, headers map[string]string, formData map[string]string) error {
	// Resolve file path relative to base directory if needed
	fullPath := filePath
	if !filepath.IsAbs(filePath) && h.baseDir != "" {
		fullPath = filepath.Join(h.baseDir, filePath)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fullPath, err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add other form fields
	for key, value := range formData {
		err = writer.WriteField(key, value)
		if err != nil {
			return fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set additional headers
	for key, value := range headers {
		if strings.ToLower(key) != "content-type" {
			req.Header.Set(key, value)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP file upload failed: %w", err)
	}
	defer resp.Body.Close()

	// Store response for subsequent assertions
	h.lastResp = resp
	h.lastBody, _ = io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("file upload failed with status %d: %s", resp.StatusCode, string(h.lastBody))
	}

	return nil
}
