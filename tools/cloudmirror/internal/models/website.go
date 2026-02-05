package models

import "time"

// WebsiteReport represents the combined coverage report for the InfraSpec website
type WebsiteReport struct {
	GeneratedAt time.Time        `json:"generatedAt"`
	Services    []ServiceSummary `json:"services"`
}

// ServiceSummary represents a service's compatibility status for the website
type ServiceSummary struct {
	Name         string              `json:"name"`
	FullName     string              `json:"fullName"`
	Status       ServiceStatus       `json:"status"`
	InfraSpec    *InfraSpecCoverage  `json:"infraspec,omitempty"`
	VirtualCloud *VirtualCloudStatus `json:"virtualCloud,omitempty"`
}

// ServiceStatus represents the implementation status of a service
type ServiceStatus string

const (
	StatusImplemented ServiceStatus = "implemented"
	StatusPlanned     ServiceStatus = "planned"
	StatusUnsupported ServiceStatus = "unsupported"
)

// InfraSpecCoverage represents InfraSpec assertion coverage
type InfraSpecCoverage struct {
	Status     ServiceStatus        `json:"status"`
	Operations []InfraSpecOperation `json:"operations"`
}

// InfraSpecOperation represents an InfraSpec assertion operation
type InfraSpecOperation struct {
	Name        string `json:"name"`
	Implemented bool   `json:"implemented"`
	Description string `json:"description,omitempty"`
}

// VirtualCloudStatus represents Virtual Cloud (InfraSpec API) coverage
type VirtualCloudStatus struct {
	Status          ServiceStatus           `json:"status"`
	CoveragePercent float64                 `json:"coveragePercent"`
	TotalOperations int                     `json:"totalOperations"`
	Implemented     int                     `json:"implemented"`
	Operations      []VirtualCloudOperation `json:"operations"`
}

// VirtualCloudOperation represents a Virtual Cloud operation
type VirtualCloudOperation struct {
	Name        string   `json:"name"`
	Implemented bool     `json:"implemented"`
	Priority    Priority `json:"priority"`
}

// PlannedService represents a service that is planned but not yet implemented
type PlannedService struct {
	Name     string `json:"name"`
	FullName string `json:"fullName"`
}

// PlannedServices lists AWS services planned for implementation
var PlannedServices = []PlannedService{
	{Name: "ec2", FullName: "Amazon Elastic Compute Cloud"},
	{Name: "lambda", FullName: "AWS Lambda"},
	{Name: "sqs", FullName: "Amazon Simple Queue Service"},
	{Name: "sns", FullName: "Amazon Simple Notification Service"},
	{Name: "iam", FullName: "AWS Identity and Access Management"},
}

// ServiceFullNames maps service names to their full AWS names
var ServiceFullNames = map[string]string{
	"s3":                     "Amazon Simple Storage Service",
	"rds":                    "Amazon Relational Database Service",
	"dynamodb":               "Amazon DynamoDB",
	"sts":                    "AWS Security Token Service",
	"applicationautoscaling": "AWS Application Auto Scaling",
	"ec2":                    "Amazon Elastic Compute Cloud",
	"lambda":                 "AWS Lambda",
	"sqs":                    "Amazon Simple Queue Service",
	"sns":                    "Amazon Simple Notification Service",
	"iam":                    "AWS Identity and Access Management",
}
