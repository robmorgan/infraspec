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
// finally defaulting to the InfraSpec Cloud endpoint if none are provided.
func GetVirtualCloudEndpoint(service string) (string, bool) {
	if !config.UseInfraspecVirtualCloud() {
		return "", false
	}

	if service != "" {
		if endpoint := os.Getenv("AWS_ENDPOINT_URL_" + strings.ToUpper(service)); endpoint != "" {
			return endpoint, true
		}
	}

	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		return endpoint, true
	}

	return InfraspecCloudDefaultEndpointURL, true
}

// BuildServiceEndpoint constructs a service-specific endpoint URL by adding a subdomain
// to the base endpoint. For example:
//   - Base: "https://api.infraspec.sh" + Subdomain: "dynamodb" = "https://dynamodb.api.infraspec.sh"
//   - Base: "http://localhost:8000" + Subdomain: "sts" = "http://sts.localhost:8000"
func BuildServiceEndpoint(baseEndpoint, subdomain string) string {
	parsedURL, err := url.Parse(baseEndpoint)
	if err != nil {
		// If parsing fails, return the base endpoint as-is
		return baseEndpoint
	}

	// Extract host and port
	host := parsedURL.Hostname()
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
