package http

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/http"
)

// RegisterSteps registers all HTTP step definitions
func RegisterSteps(sc *godog.ScenarioContext) {
	registerHTTPSteps(sc)
}

// HTTP Step Definitions
func registerHTTPSteps(sc *godog.ScenarioContext) {
	// Basic HTTP requests
	sc.Step(`^I make a ([A-Z]+) request to "([^"]*)"$`, newHTTPRequestStep)
	sc.Step(`^I make a ([A-Z]+) request to "([^"]*)" with body "([^"]*)"$`, newHTTPRequestWithBodyStep)
	
	// Response status assertions
	sc.Step(`^the HTTP response status should be (\d+)$`, newHTTPResponseStatusStep)
	sc.Step(`^the ([A-Z]+) request to "([^"]*)" should return status (\d+)$`, newHTTPRequestStatusStep)
	
	// Response content assertions
	sc.Step(`^the HTTP response should contain "([^"]*)"$`, newHTTPResponseContainsStep)
	sc.Step(`^the ([A-Z]+) response from "([^"]*)" should contain "([^"]*)"$`, newHTTPRequestContainsStep)
	
	// JSON response assertions
	sc.Step(`^the HTTP response should be valid JSON$`, newHTTPResponseJSONStep)
	sc.Step(`^the ([A-Z]+) response from "([^"]*)" should be valid JSON$`, newHTTPRequestJSONStep)
	
	// Header assertions
	sc.Step(`^the HTTP response header "([^"]*)" should be "([^"]*)"$`, newHTTPResponseHeaderStep)
	sc.Step(`^the ([A-Z]+) response from "([^"]*)" should have header "([^"]*)" with value "([^"]*)"$`, newHTTPRequestHeaderStep)
	
	// File upload
	sc.Step(`^I upload file "([^"]*)" to "([^"]*)" as field "([^"]*)"$`, newHTTPFileUploadStep)
	sc.Step(`^I upload file "([^"]*)" to "([^"]*)" as field "([^"]*)" with form data:$`, newHTTPFileUploadWithDataStep)
	
	// Request with headers
	sc.Step(`^I make a ([A-Z]+) request to "([^"]*)" with headers:$`, newHTTPRequestWithHeadersStep)
	sc.Step(`^I make a ([A-Z]+) request to "([^"]*)" with body "([^"]*)" and headers:$`, newHTTPRequestWithBodyAndHeadersStep)
}

// Store the last request details in context for subsequent assertions
type httpRequestCtxKey struct{}

type httpRequestDetails struct {
	method  string
	url     string
	headers map[string]string
}

// Basic HTTP request step
func newHTTPRequestStep(ctx context.Context, method, url string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads based on feature file location
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	details := &httpRequestDetails{method: method, url: url, headers: make(map[string]string)}
	ctx = context.WithValue(ctx, httpRequestCtxKey{}, details)

	return httpAssert.AssertResponseStatus(method, url, 200, nil)
}

// HTTP request with body
func newHTTPRequestWithBodyStep(ctx context.Context, method, url, body string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	details := &httpRequestDetails{method: method, url: url, headers: make(map[string]string)}
	ctx = context.WithValue(ctx, httpRequestCtxKey{}, details)

	return httpAssert.AssertResponseStatusWithBody(method, url, body, 200, nil)
}

// Response status assertion for the last request
func newHTTPResponseStatusStep(ctx context.Context, statusCode int) error {
	details, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails)
	if !ok {
		return fmt.Errorf("no HTTP request found in context")
	}

	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	return httpAssert.AssertResponseStatus(details.method, details.url, statusCode, details.headers)
}

// Direct request with status assertion
func newHTTPRequestStatusStep(ctx context.Context, method, url string, statusCode int) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	return httpAssert.AssertResponseStatus(method, url, statusCode, nil)
}

// Response contains assertion for the last request
func newHTTPResponseContainsStep(ctx context.Context, expectedContent string) error {
	details, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails)
	if !ok {
		return fmt.Errorf("no HTTP request found in context")
	}

	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	return httpAssert.AssertResponseContains(details.method, details.url, expectedContent, details.headers)
}

// Direct request with content assertion
func newHTTPRequestContainsStep(ctx context.Context, method, url, expectedContent string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	return httpAssert.AssertResponseContains(method, url, expectedContent, nil)
}

// JSON response assertion for the last request
func newHTTPResponseJSONStep(ctx context.Context) error {
	details, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails)
	if !ok {
		return fmt.Errorf("no HTTP request found in context")
	}

	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	return httpAssert.AssertResponseJSON(details.method, details.url, details.headers)
}

// Direct request with JSON assertion
func newHTTPRequestJSONStep(ctx context.Context, method, url string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	return httpAssert.AssertResponseJSON(method, url, nil)
}

// Response header assertion for the last request
func newHTTPResponseHeaderStep(ctx context.Context, headerName, expectedValue string) error {
	details, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails)
	if !ok {
		return fmt.Errorf("no HTTP request found in context")
	}

	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	return httpAssert.AssertResponseHeader(details.method, details.url, headerName, expectedValue, details.headers)
}

// Direct request with header assertion
func newHTTPRequestHeaderStep(ctx context.Context, method, url, headerName, expectedValue string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	return httpAssert.AssertResponseHeader(method, url, headerName, expectedValue, nil)
}

// File upload step
func newHTTPFileUploadStep(ctx context.Context, filePath, url, fieldName string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	return httpAssert.UploadFile(url, fieldName, filePath, nil, nil)
}

// File upload with form data
func newHTTPFileUploadWithDataStep(ctx context.Context, filePath, url, fieldName string, formDataTable *godog.Table) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	// Parse form data from table
	formData := make(map[string]string)
	for _, row := range formDataTable.Rows {
		if len(row.Cells) >= 2 {
			formData[row.Cells[0].Value] = row.Cells[1].Value
		}
	}

	return httpAssert.UploadFile(url, fieldName, filePath, nil, formData)
}

// Request with headers
func newHTTPRequestWithHeadersStep(ctx context.Context, method, url string, headersTable *godog.Table) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	// Parse headers from table
	headers := make(map[string]string)
	for _, row := range headersTable.Rows {
		if len(row.Cells) >= 2 {
			headers[row.Cells[0].Value] = row.Cells[1].Value
		}
	}

	details := &httpRequestDetails{method: method, url: url, headers: headers}
	ctx = context.WithValue(ctx, httpRequestCtxKey{}, details)

	return httpAssert.AssertResponseStatus(method, url, 200, headers)
}

// Request with body and headers
func newHTTPRequestWithBodyAndHeadersStep(ctx context.Context, method, url, body string, headersTable *godog.Table) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	// Parse headers from table
	headers := make(map[string]string)
	for _, row := range headersTable.Rows {
		if len(row.Cells) >= 2 {
			headers[row.Cells[0].Value] = row.Cells[1].Value
		}
	}

	details := &httpRequestDetails{method: method, url: url, headers: headers}
	ctx = context.WithValue(ctx, httpRequestCtxKey{}, details)

	return httpAssert.AssertResponseStatusWithBody(method, url, body, 200, headers)
}

// Helper function to get HTTP asserter from context
func getHTTPAsserter(ctx context.Context) (http.HTTPAssertions, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.HTTP)
	if err != nil {
		return nil, err
	}

	httpAssert, ok := asserter.(http.HTTPAssertions)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement HTTPAssertions")
	}
	return httpAssert, nil
}