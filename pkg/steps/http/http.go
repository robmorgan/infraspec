package http

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	httpassert "github.com/robmorgan/infraspec/pkg/assertions/http"
	"github.com/robmorgan/infraspec/pkg/httphelpers"
	"github.com/robmorgan/infraspec/pkg/retry"
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
	sc.Step(`^I set basic auth credentials with username "([^"]*)" and password "([^"]*)"$`, newSetBasicAuthCredentialsStep)
	sc.Step(`^I am authenticated with a valid bearer token$`, newSetBearerTokenFromEnvStep)

	// Basic HTTP requests
	sc.Step(`^I send a ([A-Z]+) request$`, newHTTPRequestStep)
	sc.Step(`^I make a ([A-Z]+) request$`, newHTTPRequestStep)

	// Response status assertions
	sc.Step(`^the HTTP response status should be (\d+)$`, newHTTPResponseStatusStep)

	// Response content assertions
	sc.Step(`^the HTTP response should contain "([^"]*)"$`, newHTTPResponseContainsStep)

	// Retry HTTP requests until response contains content
	sc.Step(`^I retry the HTTP request until the response contains "([^"]*)" with max (\d+) retries and a (\d+) second timeout$`, newRetryHTTPRequestUntilContainsStep)
	sc.Step(`^I retry the HTTP request until the response contains "([^"]*)"$`, newRetryHTTPRequestUntilContainsWithDefaultsStep)

	// JSON response assertions
	sc.Step(`^the HTTP response should be valid JSON$`, newHTTPResponseJSONStep)
	sc.Step(`^the response should be valid JSON$`, newHTTPResponseJSONStep)

	// Header assertions
	sc.Step(`^the HTTP response header "([^"]*)" should be "([^"]*)"$`, newHTTPResponseHeaderStep)
}

// Basic HTTP request step (uses endpoint from scenario state)
func newHTTPRequestStep(ctx context.Context, method string) (context.Context, error) {
	options := contexthelpers.GetHttpRequestOptions(ctx)
	if options == nil || options.Endpoint == "" {
		return ctx, fmt.Errorf("no HTTP endpoint set. Use 'Given I have a HTTP endpoint at' step first")
	}
	options.Method = method

	// If the user is uploading a file, we resolve the filepath relative to the feature file location
	if options.File != nil {
		base := filepath.Dir(contexthelpers.GetUri(ctx))
		absPath, err := filepath.Abs(filepath.Join(base, options.File.FilePath))
		if err != nil {
			return ctx, fmt.Errorf("failed to get absolute path for %s: %w", options.File.FilePath, err)
		}
		options.File.FilePath = absPath
	}

	client := httphelpers.NewHttpClient()
	resp, err := client.Do(ctx, options)
	if err != nil {
		return ctx, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	// Store the response in the context for later assertions
	ctx = context.WithValue(ctx, contexthelpers.HttpResponseCtxKey{}, resp)
	return ctx, nil
}

// Response status assertion for the last request
func newHTTPResponseStatusStep(ctx context.Context, statusCode int) error {
	resp := contexthelpers.GetHttpResponse(ctx)
	if resp == nil {
		return fmt.Errorf("no HTTP response found in context")
	}
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}
	return httpAssert.AssertResponseStatus(resp, statusCode)
}

// Response contains assertion for the last request
func newHTTPResponseContainsStep(ctx context.Context, expectedContent string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}
	resp := contexthelpers.GetHttpResponse(ctx)
	if resp == nil {
		return fmt.Errorf("no HTTP response found in context")
	}
	return httpAssert.AssertResponseContains(resp, expectedContent)
}

// JSON response assertion for the last request
func newHTTPResponseJSONStep(ctx context.Context) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}
	resp := contexthelpers.GetHttpResponse(ctx)
	if resp == nil {
		return fmt.Errorf("no HTTP response found in context")
	}
	return httpAssert.AssertResponseJSON(resp)
}

// Response header assertion for the last request
func newHTTPResponseHeaderStep(ctx context.Context, headerName, expectedValue string) error {
	httpAssert, err := getHTTPAsserter(ctx)
	if err != nil {
		return err
	}
	resp := contexthelpers.GetHttpResponse(ctx)
	if resp == nil {
		return fmt.Errorf("no HTTP response found in context")
	}
	return httpAssert.AssertResponseHeader(resp, headerName, expectedValue)
}

// Setup step functions
func newHTTPEndpointStep(ctx context.Context, url string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.Endpoint = url
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetHeadersStep(ctx context.Context, headersTable *godog.Table) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.Headers = make(map[string]string)
	for _, row := range headersTable.Rows {
		opts.Headers[row.Cells[0].Value] = row.Cells[1].Value
	}
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetFileStep(ctx context.Context, filePath, fieldName string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.File = &httphelpers.File{
		FieldName: fieldName,
		FilePath:  filePath,
	}
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetContentTypeStep(ctx context.Context, contentType string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.ContentType = contentType
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetFormDataStep(ctx context.Context, formDataTable *godog.Table) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.FormData = make(map[string]string)
	for _, row := range formDataTable.Rows {
		opts.FormData[row.Cells[0].Value] = row.Cells[1].Value
	}
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetRequestBodyStep(ctx context.Context, body string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.RequestBody = []byte(body)
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetBasicAuthCredentialsStep(ctx context.Context, username, password string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.BasicAuth = &httphelpers.BasicAuth{Username: username, Password: password}
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetBearerTokenFromEnvStep(ctx context.Context) (context.Context, error) {
	return NewSetBearerTokenFromEnvStep(ctx)
}

// NewSetBearerTokenFromEnvStep sets the Bearer token from the BEARER_TOKEN environment variable
func NewSetBearerTokenFromEnvStep(ctx context.Context) (context.Context, error) {
	// Check for environment variable containing the Bearer token
	token := os.Getenv("INFRASPEC_BEARER_TOKEN")
	if token == "" {
		return ctx, fmt.Errorf("INFRASPEC_BEARER_TOKEN environment variable is not set. Please set it before running this scenario")
	}

	opts := contexthelpers.GetHttpRequestOptions(ctx)
	if opts == nil {
		opts = &httphelpers.HttpRequestOptions{}
	}
	opts.BearerToken = token
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

// Retry HTTP request until response contains specified content
func newRetryHTTPRequestUntilContainsStep(ctx context.Context, expectedContent string, maxRetries, timeoutSeconds int) (context.Context, error) {
	options := contexthelpers.GetHttpRequestOptions(ctx)
	if options == nil || options.Endpoint == "" {
		return ctx, fmt.Errorf("no HTTP endpoint set. Use 'Given I have a HTTP endpoint at' step first")
	}

	// If the user is uploading a file, we resolve the filepath relative to the feature file location
	if options.File != nil {
		base := filepath.Dir(contexthelpers.GetUri(ctx))
		absPath, err := filepath.Abs(filepath.Join(base, options.File.FilePath))
		if err != nil {
			return ctx, fmt.Errorf("failed to get absolute path for %s: %w", options.File.FilePath, err)
		}
		options.File.FilePath = absPath
	}

	client := httphelpers.NewHttpClient()
	timeout := time.Duration(timeoutSeconds) * time.Second
	sleepBetweenRetries := time.Second // 1 second between retries

	actionDescription := fmt.Sprintf("HTTP request to %s until response contains '%s'", options.Endpoint, expectedContent)

	// Use the retry package to handle the retry logic
	result, err := retry.DoWithRetry(actionDescription, maxRetries, sleepBetweenRetries, func() (string, error) {
		// Create a context with timeout for this individual request
		requestCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := client.Do(requestCtx, options)
		if err != nil {
			return "", fmt.Errorf("HTTP request failed: %w", err)
		}

		// Check if response contains the expected content
		bodyStr := string(resp.Body)
		if !strings.Contains(bodyStr, expectedContent) {
			return "", fmt.Errorf("response does not contain expected content '%s'. Got: %s", expectedContent, bodyStr)
		}

		// Success - return the response body as the result
		return bodyStr, nil
	})

	if err != nil {
		return ctx, fmt.Errorf("failed to get response containing '%s' after %d retries: %w", expectedContent, maxRetries, err)
	}

	// Create a mock response to store in context for subsequent assertions
	// We'll create a simple response with the successful result
	mockResp := &httphelpers.HttpResponse{
		Status:     "200 OK",
		StatusCode: 200,
		Headers:    make(map[string][]string),
		Body:       []byte(result),
	}

	// Store the response in the context for later assertions
	ctx = context.WithValue(ctx, contexthelpers.HttpResponseCtxKey{}, mockResp)
	return ctx, nil
}

// Retry HTTP request until response contains specified content with default values
func newRetryHTTPRequestUntilContainsWithDefaultsStep(ctx context.Context, expectedContent string) (context.Context, error) {
	// Default values: 5 retries, 30 second timeout
	return newRetryHTTPRequestUntilContainsStep(ctx, expectedContent, 5, 30)
}

// Helper function to get HTTP asserter from context
func getHTTPAsserter(ctx context.Context) (httpassert.HTTPAsserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.HTTP)
	if err != nil {
		return nil, err
	}
	httpAssert, ok := asserter.(httpassert.HTTPAsserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement HTTPAsserter")
	}
	return httpAssert, nil
}
