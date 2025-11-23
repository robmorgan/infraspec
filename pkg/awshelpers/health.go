package awshelpers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/robmorgan/infraspec/internal/config"
)

const (
	// HealthCheckTimeout is the maximum time to wait for the health check
	HealthCheckTimeout = 5 * time.Second
	// HealthCheckPath is the path to the health check endpoint
	HealthCheckPath = "/_health"
)

// CheckVirtualCloudHealth verifies that the InfraSpec Virtual Cloud API is accessible.
// It skips the check if endpoints are overridden for localhost testing.
// Returns nil if the health check passes or if it should be skipped, otherwise returns an error.
func CheckVirtualCloudHealth() error {
	if !config.UseInfraspecVirtualCloud() {
		return nil
	}

	baseEndpoint := getBaseEndpoint()

	// Skip health check for localhost endpoints (used for local testing)
	if isLocalhostEndpoint(baseEndpoint) {
		config.Logging.Logger.Debug("Skipping health check for localhost endpoint")
		return nil
	}

	apiEndpoint := buildAPIEndpoint(baseEndpoint)
	healthURL := apiEndpoint + HealthCheckPath

	config.Logging.Logger.Debug("Checking InfraSpec Virtual Cloud API health", "url", healthURL)

	ctx, cancel := context.WithTimeout(context.Background(), HealthCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	client := &http.Client{
		Timeout: HealthCheckTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("InfraSpec Virtual Cloud API is not accessible at %s: %w\n\nPlease verify:\n  - The API is running and accessible\n  - Your network connection is working\n  - No firewall is blocking the connection", healthURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("InfraSpec Virtual Cloud API health check failed with status %d at %s\n\nPlease verify the API is running correctly", resp.StatusCode, healthURL)
	}

	config.Logging.Logger.Info("InfraSpec Virtual Cloud API is healthy")
	return nil
}

// getBaseEndpoint returns the base endpoint URL for InfraSpec Virtual Cloud
func getBaseEndpoint() string {
	// Check for general AWS_ENDPOINT_URL first
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		return endpoint
	}

	// Default to InfraSpec Cloud endpoint
	return InfraspecCloudDefaultEndpointURL
}

// isLocalhostEndpoint checks if the endpoint is a localhost URL
func isLocalhostEndpoint(endpoint string) bool {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return false
	}

	hostname := parsedURL.Hostname()
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || strings.HasSuffix(hostname, ".localhost")
}

// buildAPIEndpoint constructs the API endpoint from the base endpoint
// For example: "https://infraspec.sh" -> "https://api.infraspec.sh"
//
//	"http://localhost:8000" -> "http://localhost:8000"
func buildAPIEndpoint(baseEndpoint string) string {
	parsedURL, err := url.Parse(baseEndpoint)
	if err != nil {
		// If parsing fails, return the base endpoint as-is
		return baseEndpoint
	}

	// For localhost, return as-is
	if isLocalhostEndpoint(baseEndpoint) {
		return baseEndpoint
	}

	// For production endpoints (infraspec.sh), add "api." subdomain
	host := parsedURL.Hostname()
	port := parsedURL.Port()

	// Add "api." prefix to the hostname
	newHost := "api." + host
	if port != "" {
		newHost = newHost + ":" + port
	}

	// Reconstruct the URL with the new host
	parsedURL.Host = newHost

	return parsedURL.String()
}
