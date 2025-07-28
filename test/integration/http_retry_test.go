package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robmorgan/infraspec/pkg/httphelpers"
	"github.com/robmorgan/infraspec/pkg/retry"
	"github.com/robmorgan/infraspec/test/httpserver"
)

// TestHTTPRetryFunctionality tests the retry functionality with various scenarios
func TestHTTPRetryFunctionality(t *testing.T) {
	client := httphelpers.NewHttpClient()
	ctx := context.Background()

	t.Run("BasicHTTPRequest", func(t *testing.T) {
		// Test basic HTTP request functionality that the retry will use
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		// Test basic GET request
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/json",
		})

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "Hello, World!")
		assert.Contains(t, string(resp.Body), "status")
	})

	t.Run("HTTPRequestWithHeaders", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		// Test request with headers
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/headers",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
				"X-Custom":      "test-value",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("HTTPRequestWithBasicAuth", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		// Test request with basic auth
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: mockServer.URL() + "/json",
			BasicAuth: &httphelpers.BasicAuth{
				Username: "testuser",
				Password: "testpass",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("HTTPRequestWithBearerToken", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		// Test request with bearer token
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:      "GET",
			Endpoint:    mockServer.URL() + "/bearer",
			BearerToken: "test-token",
		})

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "authenticated")
	})

	t.Run("HTTPRequestWithPOSTAndBody", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		// Test POST request with body
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:      "POST",
			Endpoint:    mockServer.URL() + "/echo",
			RequestBody: []byte(`{"test": "data"}`),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("HTTPRequestTimeout", func(t *testing.T) {
		// Test request to non-existent endpoint (should timeout/fail)
		startTime := time.Now()
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "GET",
			Endpoint: "http://localhost:99999/nonexistent",
		})
		elapsed := time.Since(startTime)

		// Should fail quickly
		assert.Error(t, err)
		assert.Less(t, elapsed, 10*time.Second)
		if resp != nil {
			assert.NotEqual(t, 200, resp.StatusCode)
		}
	})

	t.Run("HTTPRequestWithFormData", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		// Test request with form data
		resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
			Method:   "POST",
			Endpoint: mockServer.URL() + "/upload",
			FormData: map[string]string{
				"field1": "value1",
				"field2": "value2",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

// TestRetryWithHTTPRequests tests the retry package with HTTP requests
func TestRetryWithHTTPRequests(t *testing.T) {
	client := httphelpers.NewHttpClient()
	ctx := context.Background()

	t.Run("RetryUntilSuccess", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		// Create a simple retry scenario
		attemptCount := 0
		actionDescription := "HTTP request until success"

		// Use retry package to retry until we get a successful response
		result, err := retry.DoWithRetry(actionDescription, 3, time.Second, func() (string, error) {
			attemptCount++

			// Simulate different responses based on attempt count
			var endpoint string
			if attemptCount < 3 {
				endpoint = "/status/500" // Will return error status
			} else {
				endpoint = "/json" // Will return success
			}

			resp, err := client.Do(ctx, &httphelpers.HttpRequestOptions{
				Method:   "GET",
				Endpoint: mockServer.URL() + endpoint,
			})

			if err != nil {
				return "", err
			}

			// Consider it successful only if status is 200 and contains "Hello"
			if resp.StatusCode == 200 && string(resp.Body) != "" {
				return string(resp.Body), nil
			}

			return "", fmt.Errorf("not ready yet, attempt %d", attemptCount)
		})

		require.NoError(t, err)
		assert.Contains(t, result, "Hello, World!")
		assert.GreaterOrEqual(t, attemptCount, 3)
	})

	t.Run("RetryWithMaxRetriesExceeded", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		actionDescription := "HTTP request that should fail"

		// Use retry package with a scenario that will always fail
		_, err := retry.DoWithRetry(actionDescription, 2, time.Millisecond*100, func() (string, error) {
			// Always return an error to simulate failure
			return "", fmt.Errorf("always failing")
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsuccessful after 2 retries")
	})

	t.Run("RetryWithTimeout", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		actionDescription := "HTTP request with timeout"

		// Use retry package with timeout
		_, err := retry.DoWithTimeout(actionDescription, time.Millisecond*500, func() (string, error) {
			// Simulate a long-running operation
			time.Sleep(time.Second)
			return "success", nil
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "did not complete before timeout")
	})

	t.Run("RetryWithFatalError", func(t *testing.T) {
		mockServer := httpserver.NewMockHTTPServer()
		defer mockServer.Close()

		actionDescription := "HTTP request with fatal error"

		// Use retry package with a fatal error
		_, err := retry.DoWithRetry(actionDescription, 5, time.Millisecond*100, func() (string, error) {
			// Return a fatal error that should not be retried
			return "", retry.FatalError{Underlying: fmt.Errorf("permanent failure")}
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "FatalError")
	})
}
