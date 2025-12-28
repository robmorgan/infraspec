package generator

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic cases
		{"DeleteScheduledAction", "delete_scheduled_action"},
		{"GetCallerIdentity", "get_caller_identity"},
		{"PutObject", "put_object"},
		{"ListBuckets", "list_buckets"},

		// Acronym handling - keeps acronyms together
		{"CreateDBInstance", "create_db_instance"},
		{"DescribeDBInstances", "describe_db_instances"},
		{"GetAPIKey", "get_api_key"},
		{"CreateVPCEndpoint", "create_vpc_endpoint"},

		// All uppercase acronyms
		{"S3", "s3"},
		{"EC2", "ec2"},
		{"IAM", "iam"},
		{"SQS", "sqs"},
		{"RDS", "rds"},

		// Edge cases
		{"", ""},
		{"lowercase", "lowercase"},
		{"A", "a"},
		{"AB", "ab"},
		{"ABC", "abc"},
		{"AbCd", "ab_cd"},
		{"HTTPServer", "http_server"},
		{"XMLParser", "xml_parser"},
		{"AWSService", "aws_service"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
