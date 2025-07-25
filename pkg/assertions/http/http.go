package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/robmorgan/infraspec/pkg/httphelpers"
)

// HTTPAsserter defines HTTP-specific assertions
type HTTPAsserter interface {
	AssertResponseStatus(resp *httphelpers.HttpResponse, expectedStatus int) error
	AssertResponseHeader(resp *httphelpers.HttpResponse, headerName, expectedValue string) error
	AssertResponseContains(resp *httphelpers.HttpResponse, expectedContent string) error
	AssertResponseJSON(resp *httphelpers.HttpResponse) error
}

// HTTPAsserter implements HTTP-specific assertions
type httpAsserter struct{}

// NewHTTPAsserter creates a new AWSAsserter instance
func NewHTTPAsserter() *httpAsserter {
	return &httpAsserter{}
}

// AssertResponseStatus checks if an HTTP request returns the expected status code
func (h *httpAsserter) AssertResponseStatus(resp *httphelpers.HttpResponse, expectedStatus int) error {
	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d for %s %s", expectedStatus, resp.StatusCode, method, url)
	}
	return nil
}

// AssertResponseContains checks if the HTTP response body contains the expected content
func (h *httpAsserter) AssertResponseContains(resp *httphelpers.HttpResponse, expectedContent string) error {
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, expectedContent) {
		return fmt.Errorf("response body does not contain expected content '%s'. Got: %s", expectedContent, bodyStr)
	}

	return nil
}

// AssertResponseJSON checks if the HTTP response is valid JSON
func (h *httpAsserter) AssertResponseJSON(resp *httphelpers.HttpResponse) error {
	var jsonData interface{}
	if err := json.Unmarshal(resp.Body, &jsonData); err != nil {
		return fmt.Errorf("response is not valid JSON: %w. Response body: %s", err, string(resp.Body))
	}

	return nil
}

// AssertResponseHeader checks if the HTTP response has the expected header value
func (h *httpAsserter) AssertResponseHeader(resp *httphelpers.HttpResponse, headerName, expectedValue string) error {
	actualValue := resp.Headers.Get(headerName)
	if actualValue != expectedValue {
		return fmt.Errorf("expected header '%s' to be '%s', got '%s'", headerName, expectedValue, actualValue)
	}

	return nil
}
