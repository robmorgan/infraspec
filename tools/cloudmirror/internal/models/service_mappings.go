package models

// ServiceNameMappings maps InfraSpec service names to AWS SDK model file names.
// This is the canonical source of truth for service name translations.
var ServiceNameMappings = map[string]string{
	"rds":                      "rds",
	"s3":                       "s3",
	"dynamodb":                 "dynamodb",
	"sts":                      "sts",
	"applicationautoscaling":   "application-auto-scaling",
	"application-auto-scaling": "application-auto-scaling",
	"ec2":                      "ec2",
	"lambda":                   "lambda",
	"sqs":                      "sqs",
	"sns":                      "sns",
	"iam":                      "iam",
}

// ReverseServiceNameMappings maps AWS SDK model file names to InfraSpec service names.
// Use this when converting from AWS model names to InfraSpec directory names.
var ReverseServiceNameMappings = map[string]string{
	"application-auto-scaling": "applicationautoscaling",
}

// GetAWSModelName returns the AWS SDK model file name for a given InfraSpec service name.
func GetAWSModelName(serviceName string) string {
	if mapped, ok := ServiceNameMappings[serviceName]; ok {
		return mapped
	}
	return serviceName
}

// GetInfraSpecServiceName returns the InfraSpec service directory name for a given AWS model name.
func GetInfraSpecServiceName(awsModelName string) string {
	if mapped, ok := ReverseServiceNameMappings[awsModelName]; ok {
		return mapped
	}
	return awsModelName
}
