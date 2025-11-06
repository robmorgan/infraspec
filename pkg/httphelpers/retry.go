package httphelpers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RetryConfig holds configuration for the retry mechanism
type RetryConfig struct {
	MaxRetries       int
	InitialDelay     time.Duration
	MaxDelay         time.Duration
	BackoffFactor    float64
	TargetStatusCode int
	TargetString     string
}

// RetryResult holds the result of the retry operation
type RetryResult struct {
	Response   *http.Response
	Body       string
	AttemptNum int
	Success    bool
	Error      error
}

// RetryHTTPRequest retries an HTTP request until conditions are met
func RetryHTTPRequest(ctx context.Context, client *http.Client, req *http.Request, config RetryConfig) *RetryResult {
	var lastResponse *http.Response
	var lastBody string
	var lastError error

	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		// Clone the request for each attempt (important for request body)
		reqClone := req.Clone(ctx)

		// Make the HTTP request
		resp, err := client.Do(reqClone)
		if err != nil {
			lastError = err
			fmt.Printf("Attempt %d failed with error: %v\n", attempt, err)
		} else {
			// Read the response body
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				lastError = fmt.Errorf("failed to read response body: %w", err)
				fmt.Printf("Attempt %d failed to read body: %v\n", attempt, err)
			} else {
				lastResponse = resp
				lastBody = string(body)

				fmt.Printf("Attempt %d: Status %d, Body length: %d\n", attempt, resp.StatusCode, len(body))

				// Check if we've met our success criteria
				statusMatches := config.TargetStatusCode == 0 || resp.StatusCode == config.TargetStatusCode
				stringMatches := config.TargetString == "" || strings.Contains(lastBody, config.TargetString)

				if statusMatches && stringMatches {
					return &RetryResult{
						Response:   lastResponse,
						Body:       lastBody,
						AttemptNum: attempt,
						Success:    true,
						Error:      nil,
					}
				}

				fmt.Printf("Conditions not met - Status: %d (want: %d), String found: %t\n",
					resp.StatusCode, config.TargetStatusCode, stringMatches)
			}
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxRetries {
			select {
			case <-ctx.Done():
				return &RetryResult{
					Response:   lastResponse,
					Body:       lastBody,
					AttemptNum: attempt,
					Success:    false,
					Error:      fmt.Errorf("context cancelled: %w", ctx.Err()),
				}
			case <-time.After(delay):
				// Exponential backoff with jitter
				delay = time.Duration(float64(delay) * config.BackoffFactor)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}
		}
	}

	// All retries exhausted
	return &RetryResult{
		Response:   lastResponse,
		Body:       lastBody,
		AttemptNum: config.MaxRetries,
		Success:    false,
		Error:      fmt.Errorf("max retries (%d) exhausted, last error: %v", config.MaxRetries, lastError),
	}
}
