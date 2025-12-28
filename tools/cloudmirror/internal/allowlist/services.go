// Package allowlist provides the canonical list of AWS services that InfraSpec tracks.
// This package is shared between CloudMirror and autotrack tools.
package allowlist

// ServiceConfig defines configuration for a tracked AWS service.
type ServiceConfig struct {
	// Name is the canonical service name used in InfraSpec (e.g., "rds", "s3")
	Name string `json:"name"`

	// AWSName is the AWS SDK model name (e.g., "rds", "s3", "application-auto-scaling")
	AWSName string `json:"aws_name"`

	// FullName is the human-readable service name
	FullName string `json:"full_name"`

	// Enabled controls whether this service is actively tracked
	Enabled bool `json:"enabled"`

	// Priority determines the order of processing (higher = more important)
	Priority int `json:"priority"`

	// Notes provides additional context about the service
	Notes string `json:"notes,omitempty"`
}

// ServiceAllowlist is the canonical list of AWS services that InfraSpec tracks.
// Only services in this list will be:
// - Analyzed for SDK changes
// - Processed for AI-assisted implementation generation
// - Included in coverage reports
//
// To add a new service:
// 1. Add an entry to this list with Enabled: true
// 2. Ensure the AWSName matches the AWS SDK model directory name
// 3. Set an appropriate Priority (100 = core, 80 = important, 60 = standard)
var ServiceAllowlist = []ServiceConfig{
	// Core services (Priority 100) - Most commonly used with Terraform
	{
		Name:     "iam",
		AWSName:  "iam",
		FullName: "AWS Identity and Access Management",
		Enabled:  true,
		Priority: 100,
		Notes:    "Core service for all AWS operations",
	},
	{
		Name:     "sts",
		AWSName:  "sts",
		FullName: "AWS Security Token Service",
		Enabled:  true,
		Priority: 100,
		Notes:    "Required for authentication and assume role",
	},
	{
		Name:     "s3",
		AWSName:  "s3",
		FullName: "Amazon Simple Storage Service",
		Enabled:  true,
		Priority: 100,
		Notes:    "Object storage, Terraform state backend",
	},
	{
		Name:     "ec2",
		AWSName:  "ec2",
		FullName: "Amazon Elastic Compute Cloud",
		Enabled:  true,
		Priority: 100,
		Notes:    "Compute instances, VPCs, networking",
	},

	// Important services (Priority 80) - Frequently used
	{
		Name:     "rds",
		AWSName:  "rds",
		FullName: "Amazon Relational Database Service",
		Enabled:  true,
		Priority: 80,
		Notes:    "Managed relational databases",
	},
	{
		Name:     "dynamodb",
		AWSName:  "dynamodb",
		FullName: "Amazon DynamoDB",
		Enabled:  true,
		Priority: 80,
		Notes:    "NoSQL database service",
	},
	{
		Name:     "sqs",
		AWSName:  "sqs",
		FullName: "Amazon Simple Queue Service",
		Enabled:  true,
		Priority: 80,
		Notes:    "Message queuing service",
	},
	{
		Name:     "lambda",
		AWSName:  "lambda",
		FullName: "AWS Lambda",
		Enabled:  true,
		Priority: 80,
		Notes:    "Serverless compute",
	},

	// Standard services (Priority 60)
	{
		Name:     "applicationautoscaling",
		AWSName:  "application-auto-scaling",
		FullName: "AWS Application Auto Scaling",
		Enabled:  true,
		Priority: 60,
		Notes:    "Auto scaling for various AWS services",
	},

	// Planned services (Priority 40) - Not yet implemented but tracked
	// Uncomment and set Enabled: true when ready to implement
	/*
		{
			Name:     "sns",
			AWSName:  "sns",
			FullName: "Amazon Simple Notification Service",
			Enabled:  false,
			Priority: 60,
			Notes:    "Pub/sub messaging - planned",
		},
		{
			Name:     "cloudwatch",
			AWSName:  "cloudwatch",
			FullName: "Amazon CloudWatch",
			Enabled:  false,
			Priority: 60,
			Notes:    "Monitoring and observability - planned",
		},
		{
			Name:     "secretsmanager",
			AWSName:  "secretsmanager",
			FullName: "AWS Secrets Manager",
			Enabled:  false,
			Priority: 60,
			Notes:    "Secrets management - planned",
		},
		{
			Name:     "kms",
			AWSName:  "kms",
			FullName: "AWS Key Management Service",
			Enabled:  false,
			Priority: 60,
			Notes:    "Key management - planned",
		},
	*/
}

// GetAllowedServices returns all enabled services from the allowlist.
func GetAllowedServices() []ServiceConfig {
	var allowed []ServiceConfig
	for _, svc := range ServiceAllowlist {
		if svc.Enabled {
			allowed = append(allowed, svc)
		}
	}
	return allowed
}

// GetAllowedServiceNames returns the names of all enabled services.
func GetAllowedServiceNames() []string {
	var names []string
	for _, svc := range ServiceAllowlist {
		if svc.Enabled {
			names = append(names, svc.Name)
		}
	}
	return names
}

// GetAllowedAWSNames returns the AWS SDK model names of all enabled services.
func GetAllowedAWSNames() []string {
	var names []string
	for _, svc := range ServiceAllowlist {
		if svc.Enabled {
			names = append(names, svc.AWSName)
		}
	}
	return names
}

// IsServiceAllowed checks if a service is in the allowlist and enabled.
// Accepts either the InfraSpec name or AWS SDK name.
func IsServiceAllowed(name string) bool {
	for _, svc := range ServiceAllowlist {
		if svc.Enabled && (svc.Name == name || svc.AWSName == name) {
			return true
		}
	}
	return false
}

// GetServiceConfig returns the configuration for a service by name.
// Returns nil if the service is not in the allowlist.
func GetServiceConfig(name string) *ServiceConfig {
	for i := range ServiceAllowlist {
		if ServiceAllowlist[i].Name == name || ServiceAllowlist[i].AWSName == name {
			return &ServiceAllowlist[i]
		}
	}
	return nil
}

// GetServicesByPriority returns enabled services sorted by priority (highest first).
func GetServicesByPriority() []ServiceConfig {
	allowed := GetAllowedServices()

	// Simple bubble sort since list is small
	for i := 0; i < len(allowed)-1; i++ {
		for j := 0; j < len(allowed)-i-1; j++ {
			if allowed[j].Priority < allowed[j+1].Priority {
				allowed[j], allowed[j+1] = allowed[j+1], allowed[j]
			}
		}
	}

	return allowed
}

// ServiceNameToAWS converts an InfraSpec service name to AWS SDK model name.
func ServiceNameToAWS(name string) string {
	for _, svc := range ServiceAllowlist {
		if svc.Name == name {
			return svc.AWSName
		}
	}
	return name
}

// AWSNameToService converts an AWS SDK model name to InfraSpec service name.
func AWSNameToService(awsName string) string {
	for _, svc := range ServiceAllowlist {
		if svc.AWSName == awsName {
			return svc.Name
		}
	}
	return awsName
}
