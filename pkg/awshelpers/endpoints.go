package awshelpers

import (
	"net/url"
	"os"
	"strings"

	"github.com/robmorgan/infraspec/internal/config"
)

// GetVirtualCloudEndpoint returns the endpoint URL to use for the given AWS service when
// InfraSpec Virtual Cloud mode is enabled. The function looks for a service-specific
// environment variable (e.g. AWS_ENDPOINT_URL_RDS) and falls back to AWS_ENDPOINT_URL,
// finally defaulting to the InfraSpec Cloud endpoint with service-specific subdomain.
//
// If service is empty, returns the base endpoint URL without subdomain construction.
// Otherwise, constructs a service-specific subdomain endpoint (e.g. https://dynamodb.infraspec.sh).
func GetVirtualCloudEndpoint(service string) (string, bool) {
	if !config.UseInfraspecVirtualCloud() {
		return "", false
	}

	// Check for service-specific environment variable first
	if service != "" {
		if endpoint := os.Getenv("AWS_ENDPOINT_URL_" + strings.ToUpper(service)); endpoint != "" {
			return endpoint, true
		}
	}

	// Check for general AWS_ENDPOINT_URL
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		// If a service is specified, build service-specific subdomain endpoint
		if service != "" {
			return BuildServiceEndpoint(endpoint, service), true
		}
		return endpoint, true
	}

	// Default to InfraSpec Cloud endpoint
	baseEndpoint := InfraspecCloudDefaultEndpointURL
	if service != "" {
		return BuildServiceEndpoint(baseEndpoint, service), true
	}
	return baseEndpoint, true
}

// BuildServiceEndpoint constructs a service-specific endpoint URL by adding a subdomain
// to the base endpoint. For example:
//   - Base: "https://infraspec.sh" + Subdomain: "s3" = "https://s3.infraspec.sh"
//   - Base: "https://infraspec.sh" + Subdomain: "dynamodb" = "https://dynamodb.infraspec.sh"
//   - Base: "http://localhost:8000" + Subdomain: "s3" = "http://localhost:8000" (no subdomain for localhost)
//   - Base: "http://127.0.0.1:8000" + Subdomain: "sts" = "http://127.0.0.1:8000" (no subdomain for 127.0.0.1)
func BuildServiceEndpoint(baseEndpoint, subdomain string) string {
	parsedURL, err := url.Parse(baseEndpoint)
	if err != nil {
		// If parsing fails, return the base endpoint as-is
		return baseEndpoint
	}

	// Extract host and port
	host := parsedURL.Hostname()

	// For localhost or 127.0.0.1, don't use subdomains as they don't resolve properly
	// The emulator handles all services on the same endpoint
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return baseEndpoint
	}

	port := parsedURL.Port()

	// Build new host with subdomain prefix
	newHost := subdomain + "." + host

	if port != "" {
		newHost = newHost + ":" + port
	}

	// Reconstruct the URL with the new host
	parsedURL.Host = newHost

	return parsedURL.String()
}
