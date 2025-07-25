package http

import (
	"context"
	"fmt"

	// "path/filepath" // Not needed for now

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	httpassert "github.com/robmorgan/infraspec/pkg/assertions/http"
	"github.com/robmorgan/infraspec/pkg/httphelpers"
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

	// Response status assertions
	sc.Step(`^the HTTP response status should be (\d+)$`, newHTTPResponseStatusStep)

	// Response content assertions
	sc.Step(`^the HTTP response should contain "([^"]*)"$`, newHTTPResponseContainsStep)

	// JSON response assertions
	sc.Step(`^the HTTP response should be valid JSON$`, newHTTPResponseJSONStep)
	sc.Step(`^the response should be valid JSON$`, newHTTPResponseJSONStep)

	// Header assertions
	sc.Step(`^the HTTP response header "([^"]*)" should be "([^"]*)"$`, newHTTPResponseHeaderStep)
}

// Basic HTTP request step (uses endpoint from scenario state)
func newHTTPRequestStep(ctx context.Context, method string) error {
	options := contexthelpers.GetHttpRequestOptions(ctx)
	if options == nil || options.Endpoint == "" {
		return fmt.Errorf("no HTTP endpoint set. Use 'Given I have a HTTP endpoint at' step first")
	}
	options.Method = method
	client := httphelpers.NewHttpClient("")
	resp, err := client.Do(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to make HTTP request: %w", err)
	}
	// Store the response in the context for later assertions
	ctx = context.WithValue(ctx, contexthelpers.HttpResponseCtxKey{}, resp)
	return nil
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
	opts.Endpoint = url
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetHeadersStep(ctx context.Context, headersTable *godog.Table) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	opts.Headers = make(map[string]string)
	for _, row := range headersTable.Rows {
		opts.Headers[row.Cells[0].Value] = row.Cells[1].Value
	}
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetFileStep(ctx context.Context, filePath, fieldName string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	opts.File = &httphelpers.File{
		FieldName: fieldName,
		FilePath:  filePath,
	}
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetContentTypeStep(ctx context.Context, contentType string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	opts.ContentType = contentType
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetFormDataStep(ctx context.Context, formDataTable *godog.Table) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	opts.FormData = make(map[string]string)
	for _, row := range formDataTable.Rows {
		opts.FormData[row.Cells[0].Value] = row.Cells[1].Value
	}
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
}

func newSetRequestBodyStep(ctx context.Context, body string) (context.Context, error) {
	opts := contexthelpers.GetHttpRequestOptions(ctx)
	opts.RequestBody = []byte(body)
	return context.WithValue(ctx, contexthelpers.HttpRequestOptionsCtxKey{}, opts), nil
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
