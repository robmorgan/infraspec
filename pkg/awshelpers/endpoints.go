package awshelpers

import (
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
