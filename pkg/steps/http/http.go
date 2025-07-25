package http

import (
	"context"
	"fmt"
	"path/filepath"

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
	// Setup steps
	sc.Step(`^I have a HTTP endpoint at "([^"]*)"$`, newHTTPEndpointStep)
	sc.Step(`^I set the headers to$`, newSetHeadersStep)
	sc.Step(`^I have a file "([^"]*)" as field "([^"]*)"$`, newSetFileStep)
	sc.Step(`^I set content type to "([^"]*)"$`, newSetContentTypeStep)
	sc.Step(`^I set the form data to:$`, newSetFormDataStep)
	sc.Step(`^I set the request body to "([^"]*)"$`, newSetRequestBodyStep)

	// Basic HTTP requests
	sc.Step(`^I make a ([A-Z]+) request$`, newHTTPRequestStep)
	sc.Step(`^I make a ([A-Z]+) request to "([^"]*)"$`, newHTTPRequestWithURLStep)

	// Response status assertions
	sc.Step(`^the HTTP response status should be (\d+)$`, newHTTPResponseStatusStep)
	sc.Step(`^the ([A-Z]+) request to "([^"]*)" should return status (\d+)$`, newHTTPRequestStatusStep)

	// Response content assertions
	sc.Step(`^the HTTP response should contain "([^"]*)"$`, newHTTPResponseContainsStep)

	// JSON response assertions
	sc.Step(`^the HTTP response should be valid JSON$`, newHTTPResponseJSONStep)
	sc.Step(`^the response should be valid JSON$`, newHTTPResponseJSONStep)

	// Header assertions
	sc.Step(`^the HTTP response header "([^"]*)" should be "([^"]*)"$`, newHTTPResponseHeaderStep)

	// File upload
	sc.Step(`^I upload file "([^"]*)" to "([^"]*)" as field "([^"]*)"$`, newHTTPFileUploadStep)

	// Request with headers
	sc.Step(`^I make a ([A-Z]+) request to "([^"]*)" with headers:$`, newHTTPRequestWithHeadersStep)
}

// Basic HTTP request step (uses endpoint from scenario state)
func newHTTPRequestStep(ctx context.Context, method string) error {
	options := contexthelpers.GetHttpRequestOptions(ctx)
	if options.Url == "" {
		return fmt.Errorf("no HTTP endpoint set. Use 'Given I have a HTTP endpoint at' step first")
	}

	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// Set base directory for file uploads based on feature file location
	uri := contexthelpers.GetUri(ctx)
	if uri != "" {
		httpAssert.SetBaseDirectory(filepath.Dir(uri))
	}

	// Use headers from scenario state if available
	headers := scenarioState.headers
	if headers == nil {
		headers = make(map[string]string)
	}

	details := &httpRequestDetails{method: method, url: scenarioState.endpoint, headers: headers}

	// Store details in global state since context doesn't persist between steps
	globalRequestDetails = details

	// Handle file upload if file is set in scenario state
	if scenarioState.file != nil {
		if scenarioState.contentType != "" && scenarioState.formData != nil {
			return httpAssert.UploadFile(scenarioState.endpoint, scenarioState.file.fieldName, scenarioState.file.path, headers, scenarioState.formData)
		} else {
			return httpAssert.UploadFile(scenarioState.endpoint, scenarioState.file.fieldName, scenarioState.file.path, headers, nil)
		}
	}

	// For requests without file upload, just store the details for later assertions
	// The actual HTTP request will be made by the response assertion steps
	return nil
}

// HTTP request step with URL provided directly
func newHTTPRequestWithURLStep(ctx context.Context, method, url string) error {
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

	// Store details in global state since context doesn't persist between steps
	globalRequestDetails = details

	return httpAssert.AssertResponseStatus(method, url, 200, nil)
}

// Response status assertion for the last request
func newHTTPResponseStatusStep(ctx context.Context, statusCode int) error {
	details := globalRequestDetails
	if details == nil {
		// Try to get from context as fallback
		if ctxDetails, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails); ok {
			details = ctxDetails
		} else {
			return fmt.Errorf("no HTTP request found in context")
		}
	}

	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// If we have a request body, use the body-aware method
	if scenarioState.requestBody != "" {
		hasStoredResponse = true
		return httpAssert.AssertResponseStatusWithBody(details.method, details.url, scenarioState.requestBody, statusCode, details.headers)
	}

	hasStoredResponse = true
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
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// If we have a stored response, use it
	if hasStoredResponse {
		return httpAssert.AssertStoredResponseContains(expectedContent)
	}

	// Otherwise fall back to making a new request
	details := globalRequestDetails
	if details == nil {
		// Try to get from context as fallback
		if ctxDetails, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails); ok {
			details = ctxDetails
		} else {
			return fmt.Errorf("no HTTP request found in context")
		}
	}

	return httpAssert.AssertResponseContains(details.method, details.url, expectedContent, details.headers)
}

// JSON response assertion for the last request
func newHTTPResponseJSONStep(ctx context.Context) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// If we have a stored response, use it
	if hasStoredResponse {
		return httpAssert.AssertStoredResponseJSON()
	}

	// Otherwise fall back to making a new request
	details := globalRequestDetails
	if details == nil {
		// Try to get from context as fallback
		if ctxDetails, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails); ok {
			details = ctxDetails
		} else {
			return fmt.Errorf("no HTTP request found in context")
		}
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
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}

	// If we have a stored response, use it
	if hasStoredResponse {
		return httpAssert.AssertStoredResponseHeader(headerName, expectedValue)
	}

	// Otherwise fall back to making a new request
	details := globalRequestDetails
	if details == nil {
		// Try to get from context as fallback
		if ctxDetails, ok := ctx.Value(httpRequestCtxKey{}).(*httpRequestDetails); ok {
			details = ctxDetails
		} else {
			return fmt.Errorf("no HTTP request found in context")
		}
	}

	return httpAssert.AssertResponseHeader(details.method, details.url, headerName, expectedValue, details.headers)
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

// Setup step functions
func newHTTPEndpointStep(ctx context.Context, url string) error {
	scenarioState.endpoint = url
	return nil
}

func newSetHeadersStep(ctx context.Context, headersTable *godog.Table) error {
	headers := make(map[string]string)
	for _, row := range headersTable.Rows {
		if len(row.Cells) >= 2 {
			headers[row.Cells[0].Value] = row.Cells[1].Value
		}
	}
	scenarioState.headers = headers
	return nil
}

func newSetFileStep(ctx context.Context, filePath, fieldName string) error {
	fileDetails := &httpFileDetails{path: filePath, fieldName: fieldName}
	scenarioState.file = fileDetails
	return nil
}

func newSetContentTypeStep(ctx context.Context, contentType string) error {
	scenarioState.contentType = contentType
	return nil
}

func newSetFormDataStep(ctx context.Context, formDataTable *godog.Table) error {
	formData := make(map[string]string)
	for _, row := range formDataTable.Rows {
		if len(row.Cells) >= 2 {
			formData[row.Cells[0].Value] = row.Cells[1].Value
		}
	}
	scenarioState.formData = formData
	return nil
}

func newSetRequestBodyStep(ctx context.Context, body string) error {
	scenarioState.requestBody = body
	return nil
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
