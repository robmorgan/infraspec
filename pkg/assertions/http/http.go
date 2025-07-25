package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPAsserter defines HTTP-specific assertions
type HTTPAsserter interface {
	AssertResponseStatus(method, url string, expectedStatus int, headers map[string]string) error
	AssertResponseStatusWithBody(method, url, body string, expectedStatus int, headers map[string]string) error
	AssertResponseContains(method, url, expectedContent string, headers map[string]string) error
	AssertResponseJSON(method, url string, headers map[string]string) error
	AssertResponseHeader(method, url, headerName, expectedValue string, headers map[string]string) error
	AssertResponseContains(expectedContent string) error
	AssertResponseJSON() error
	AssertResponseHeader(headerName, expectedValue string) error
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

// AssertStoredResponseContains checks if the stored HTTP response body contains the expected content
func (h *HTTPAsserter) AssertStoredResponseContains(expectedContent string) error {
	if h.lastBody == nil {
		return fmt.Errorf("no stored response found")
	}

	bodyStr := string(h.lastBody)
	if !strings.Contains(bodyStr, expectedContent) {
		return fmt.Errorf("response body does not contain expected content '%s'. Got: %s", expectedContent, bodyStr)
	}

	return nil
}

// AssertStoredResponseJSON checks if the stored HTTP response is valid JSON
func (h *HTTPAsserter) AssertStoredResponseJSON() error {
	if h.lastBody == nil {
		return fmt.Errorf("no stored response found")
	}

	var jsonData interface{}
	if err := json.Unmarshal(h.lastBody, &jsonData); err != nil {
		return fmt.Errorf("response is not valid JSON: %w. Response body: %s", err, string(h.lastBody))
	}

	return nil
}

// AssertStoredResponseHeader checks if the stored HTTP response has the expected header value
func (h *HTTPAsserter) AssertStoredResponseHeader(headerName, expectedValue string) error {
	if h.lastResp == nil {
		return fmt.Errorf("no stored response found")
	}

	actualValue := h.lastResp.Header.Get(headerName)
	if actualValue != expectedValue {
		return fmt.Errorf("expected header '%s' to be '%s', got '%s'", headerName, expectedValue, actualValue)
	}

	return nil
}
