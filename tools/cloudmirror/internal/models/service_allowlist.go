// Package models re-exports the service allowlist from the shared package.
package models

import (
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/allowlist"
)

// ServiceConfig is an alias to the shared service config type.
type ServiceConfig = allowlist.ServiceConfig

// GetAllowedServices returns all enabled services from the allowlist.
func GetAllowedServices() []ServiceConfig {
	return allowlist.GetAllowedServices()
}

// GetAllowedServiceNames returns the names of all enabled services.
func GetAllowedServiceNames() []string {
	return allowlist.GetAllowedServiceNames()
}

// GetAllowedAWSNames returns the AWS SDK model names of all enabled services.
func GetAllowedAWSNames() []string {
	return allowlist.GetAllowedAWSNames()
}

// IsServiceAllowed checks if a service is in the allowlist and enabled.
func IsServiceAllowed(name string) bool {
	return allowlist.IsServiceAllowed(name)
}

// GetServiceConfig returns the configuration for a service by name.
func GetServiceConfig(name string) *ServiceConfig {
	return allowlist.GetServiceConfig(name)
}

// GetServicesByPriority returns enabled services sorted by priority (highest first).
func GetServicesByPriority() []ServiceConfig {
	return allowlist.GetServicesByPriority()
}

// ServiceNameToAWS converts an InfraSpec service name to AWS SDK model name.
func ServiceNameToAWS(name string) string {
	return allowlist.ServiceNameToAWS(name)
}

// AWSNameToService converts an AWS SDK model name to InfraSpec service name.
func AWSNameToService(awsName string) string {
	return allowlist.AWSNameToService(awsName)
}
