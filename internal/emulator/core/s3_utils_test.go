package emulator

import (
	"testing"
)

func TestParseS3Host(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		wantVirtual    bool
		wantBucketName string
	}{
		// Virtual-hosted style patterns
		{
			name:           "virtual-hosted with infraspec.sh domain",
			host:           "my-bucket.s3.infraspec.sh",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "virtual-hosted with localhost",
			host:           "my-bucket.s3.localhost",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "virtual-hosted with port",
			host:           "my-bucket.s3.localhost:3687",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "legacy virtual-hosted with localhost",
			host:           "my-bucket.localhost",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "legacy virtual-hosted with port",
			host:           "my-bucket.localhost:3687",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "bucket with hyphens",
			host:           "my-test-bucket-123.s3.infraspec.sh",
			wantVirtual:    true,
			wantBucketName: "my-test-bucket-123",
		},
		// nip.io DNS service patterns
		{
			name:           "virtual-hosted with nip.io",
			host:           "my-bucket.s3.127.0.0.1.nip.io",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "virtual-hosted with nip.io and port",
			host:           "my-bucket.s3.127.0.0.1.nip.io:3687",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "virtual-hosted with sslip.io",
			host:           "my-bucket.s3.127.0.0.1.sslip.io",
			wantVirtual:    true,
			wantBucketName: "my-bucket",
		},
		{
			name:           "path-style with nip.io",
			host:           "s3.127.0.0.1.nip.io",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "path-style with nip.io and port",
			host:           "s3.127.0.0.1.nip.io:3687",
			wantVirtual:    false,
			wantBucketName: "",
		},
		// Path-style patterns (not virtual-hosted)
		{
			name:           "path-style s3.infraspec.sh",
			host:           "s3.infraspec.sh",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "path-style s3.localhost",
			host:           "s3.localhost",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "path-style s3.localhost with port",
			host:           "s3.localhost:3687",
			wantVirtual:    false,
			wantBucketName: "",
		},
		// Non-S3 patterns
		{
			name:           "empty host",
			host:           "",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "plain localhost",
			host:           "localhost",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "localhost with port",
			host:           "localhost:3687",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "other service subdomain",
			host:           "dynamodb.infraspec.sh",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "localhost.localhost should not be virtual",
			host:           "localhost.localhost",
			wantVirtual:    false,
			wantBucketName: "",
		},
		{
			name:           "s3.localhost is path-style, not virtual",
			host:           "s3.localhost",
			wantVirtual:    false,
			wantBucketName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ParseS3Host(tt.host)
			if info.IsVirtualHosted != tt.wantVirtual {
				t.Errorf("ParseS3Host(%q).IsVirtualHosted = %v, want %v",
					tt.host, info.IsVirtualHosted, tt.wantVirtual)
			}
			if info.BucketName != tt.wantBucketName {
				t.Errorf("ParseS3Host(%q).BucketName = %q, want %q",
					tt.host, info.BucketName, tt.wantBucketName)
			}
		})
	}
}

func TestIsS3VirtualHostedRequest(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		{"virtual-hosted bucket.s3.domain", "my-bucket.s3.infraspec.sh", true},
		{"virtual-hosted bucket.s3.localhost", "my-bucket.s3.localhost", true},
		{"legacy bucket.localhost", "my-bucket.localhost", true},
		{"virtual-hosted nip.io", "my-bucket.s3.127.0.0.1.nip.io", true},
		{"virtual-hosted sslip.io", "my-bucket.s3.127.0.0.1.sslip.io", true},
		{"path-style s3.domain", "s3.infraspec.sh", false},
		{"path-style s3.localhost", "s3.localhost", false},
		{"path-style nip.io", "s3.127.0.0.1.nip.io", false},
		{"empty", "", false},
		{"localhost", "localhost", false},
		{"other service", "dynamodb.infraspec.sh", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsS3VirtualHostedRequest(tt.host)
			if got != tt.want {
				t.Errorf("IsS3VirtualHostedRequest(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestIsS3Request(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		// S3 requests (both virtual-hosted and path-style)
		{"virtual-hosted bucket.s3.domain", "my-bucket.s3.infraspec.sh", true},
		{"virtual-hosted bucket.s3.localhost", "my-bucket.s3.localhost", true},
		{"legacy bucket.localhost", "my-bucket.localhost", true},
		{"path-style s3.domain", "s3.infraspec.sh", true},
		{"path-style s3.localhost", "s3.localhost", true},
		{"path-style with port", "s3.localhost:3687", true},
		// nip.io DNS service patterns
		{"virtual-hosted nip.io", "my-bucket.s3.127.0.0.1.nip.io", true},
		{"virtual-hosted nip.io with port", "my-bucket.s3.127.0.0.1.nip.io:3687", true},
		{"virtual-hosted sslip.io", "my-bucket.s3.127.0.0.1.sslip.io", true},
		{"path-style nip.io", "s3.127.0.0.1.nip.io", true},
		{"path-style nip.io with port", "s3.127.0.0.1.nip.io:3687", true},
		// Non-S3 requests
		{"empty", "", false},
		{"plain localhost", "localhost", false},
		{"localhost with port", "localhost:3687", false},
		{"other service dynamodb", "dynamodb.infraspec.sh", false},
		{"other service rds", "rds.infraspec.sh", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsS3Request(tt.host)
			if got != tt.want {
				t.Errorf("IsS3Request(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestExtractBucketNameFromHost(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"virtual-hosted bucket.s3.domain", "my-bucket.s3.infraspec.sh", "my-bucket"},
		{"virtual-hosted bucket.s3.localhost", "my-bucket.s3.localhost", "my-bucket"},
		{"legacy bucket.localhost", "my-bucket.localhost", "my-bucket"},
		{"with port", "my-bucket.s3.localhost:3687", "my-bucket"},
		{"virtual-hosted nip.io", "my-bucket.s3.127.0.0.1.nip.io", "my-bucket"},
		{"virtual-hosted nip.io with port", "my-bucket.s3.127.0.0.1.nip.io:3687", "my-bucket"},
		{"virtual-hosted sslip.io", "my-bucket.s3.127.0.0.1.sslip.io", "my-bucket"},
		{"path-style s3.domain", "s3.infraspec.sh", ""},
		{"path-style s3.localhost", "s3.localhost", ""},
		{"path-style nip.io", "s3.127.0.0.1.nip.io", ""},
		{"empty", "", ""},
		{"localhost", "localhost", ""},
		{"other service", "dynamodb.infraspec.sh", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractBucketNameFromHost(tt.host)
			if got != tt.want {
				t.Errorf("ExtractBucketNameFromHost(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}
