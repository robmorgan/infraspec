package integration

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/http"
	"github.com/robmorgan/infraspec/pkg/httphelpers"
	"github.com/robmorgan/infraspec/test/httpserver"
)

func TestHTTPAssertions(t *testing.T) {
	// Create mock server
	mockServer := httpserver.NewMockHTTPServer()
	defer mockServer.Close()

	// Create HTTP asserter
	httpAsserter := http.NewHTTPAsserter()
	client := httphelpers.NewHttpClient("")
	ctx := context.Background()

	t.Run("AssertResponseStatus", func(t *testing.T) {
		// Test successful request
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/json",
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseStatus(resp, 200)
		assert.NoError(t, err)

		// Test 404
		resp, err = client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/status/404",
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseStatus(resp, 404)
		assert.NoError(t, err)

		// Test wrong status expectation
		resp, err = client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/json",
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseStatus(resp, 404)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected status 404, got 200")
	})

	t.Run("AssertResponseContains", func(t *testing.T) {
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/text",
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseContains(resp, "Hello, World!")
		assert.NoError(t, err)

		// Test content not found
		err = httpAsserter.AssertResponseContains(resp, "Not Found")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not contain expected content")
	})

	t.Run("AssertResponseJSON", func(t *testing.T) {
		// Valid JSON
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/json",
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseJSON(resp)
		assert.NoError(t, err)

		// Invalid JSON
		resp, err = client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/text",
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseJSON(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response is not valid JSON")
	})

	t.Run("AssertResponseHeader", func(t *testing.T) {
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/json",
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseHeader(resp, "Content-Type", "application/json")
		assert.NoError(t, err)

		// Wrong header value
		err = httpAsserter.AssertResponseHeader(resp, "Content-Type", "text/plain")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected header 'Content-Type' to be 'text/plain'")
	})

	t.Run("RequestWithHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer test-token",
			"X-Custom":      "test-value",
		}
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/headers",
			Headers:  headers,
		})
		require.NoError(t, err)
		err = httpAsserter.AssertResponseStatus(resp, 200)
		assert.NoError(t, err)
	})
}

func TestHTTPAsserterImplementsInterface(t *testing.T) {
	// Use reflect to check that the concrete type returned by NewHTTPAsserter implements the interfaces
	var _ assertions.Asserter = http.NewHTTPAsserter()
	var _ http.HTTPAsserter = http.NewHTTPAsserter()
	// Additionally, check at runtime
	if !reflect.TypeOf(http.NewHTTPAsserter()).Implements(reflect.TypeOf((*assertions.Asserter)(nil)).Elem()) {
		t.Errorf("httpAsserter does not implement assertions.Asserter")
	}
	if !reflect.TypeOf(http.NewHTTPAsserter()).Implements(reflect.TypeOf((*http.HTTPAsserter)(nil)).Elem()) {
		t.Errorf("httpAsserter does not implement http.HTTPAsserter")
	}
}

func TestHTTPAsserterFactory(t *testing.T) {
	asserter, err := assertions.New("http")
	require.NoError(t, err)
	require.NotNil(t, asserter)
	// Test that it can be asserted as HTTPAsserter
	_, ok := asserter.(http.HTTPAsserter)
	assert.True(t, ok)
}

func TestHTTPAsserterWithTimeout(t *testing.T) {
	httpAsserter := http.NewHTTPAsserter()
	client := httphelpers.NewHttpClient("")
	ctx := context.Background()
	// Test with a non-existent endpoint (should timeout/fail)
	resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
		Method:   "GET",
		Endpoint: "http://localhost:99999/nonexistent",
	})
	// The request should fail, so resp may be nil
	assert.Error(t, err)
	if resp != nil {
		err2 := httpAsserter.AssertResponseStatus(resp, 200)
		assert.Error(t, err2)
	}
}
