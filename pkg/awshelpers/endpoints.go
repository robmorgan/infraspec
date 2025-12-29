package awshelpers

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// GetVirtualCloudEndpoint returns the endpoint URL to use for the given AWS service when
// embedded emulator mode is enabled (detected via AWS_ENDPOINT_URL environment variable).
// The function looks for a service-specific environment variable (e.g. AWS_ENDPOINT_URL_RDS)
// and falls back to AWS_ENDPOINT_URL.
//
// If service is empty, returns the base endpoint URL without subdomain construction.
// Otherwise, constructs a service-specific subdomain endpoint for non-localhost URLs.
func GetVirtualCloudEndpoint(service string) (string, bool) {
	// Check for service-specific environment variable first
	if service != "" {
		if endpoint := os.Getenv("AWS_ENDPOINT_URL_" + strings.ToUpper(service)); endpoint != "" {
			return endpoint, true
		}
	}

	// Check for general AWS_ENDPOINT_URL (set by embedded emulator)
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		// If a service is specified, build service-specific subdomain endpoint
		if service != "" {
			return BuildServiceEndpoint(endpoint, service), true
		}
		return endpoint, true
	}

	// No endpoint configured - use real AWS
	return "", false
}

// BuildServiceEndpoint constructs a service-specific endpoint URL by adding a subdomain
// to the base endpoint. For example:
//   - Base: "https://infraspec.sh" + Subdomain: "s3" = "https://s3.infraspec.sh"
//   - Base: "https://infraspec.sh" + Subdomain: "dynamodb" = "https://dynamodb.infraspec.sh"
//   - Base: "http://localhost:3687" + Subdomain: "s3" = "http://s3.127.0.0.1.nip.io:3687"
//   - Base: "http://127.0.0.1:3687" + Subdomain: "s3" = "http://s3.127.0.0.1.nip.io:3687"
//
// For localhost/127.0.0.1, nip.io is used to enable wildcard DNS resolution for virtual-hosted
// style S3 addressing (e.g., bucket.s3.127.0.0.1.nip.io resolves to 127.0.0.1).
func BuildServiceEndpoint(baseEndpoint, subdomain string) string {
	parsedURL, err := url.Parse(baseEndpoint)
	if err != nil {
		// If parsing fails, return the base endpoint as-is
		return baseEndpoint
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()

	// For localhost or 127.0.0.1, use nip.io for wildcard DNS support
	// This enables virtual-hosted style S3 addressing (bucket.s3.127.0.0.1.nip.io)
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		// Convert to IP format that nip.io understands
		if host == "localhost" || host == "::1" {
			host = "127.0.0.1"
		}
		// Build: subdomain.IP.nip.io (e.g., s3.127.0.0.1.nip.io)
		newHost := fmt.Sprintf("%s.%s.nip.io", subdomain, host)
		if port != "" {
			newHost = newHost + ":" + port
		}
		parsedURL.Host = newHost
		return parsedURL.String()
	}

	// For remote endpoints, add subdomain as prefix
	newHost := subdomain + "." + host

	if port != "" {
		newHost = newHost + ":" + port
	}

	// Reconstruct the URL with the new host
	parsedURL.Host = newHost

	return parsedURL.String()
}
