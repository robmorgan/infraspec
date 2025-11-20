package awshelpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLocalhostEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{
			name:     "localhost",
			endpoint: "http://localhost:8000",
			want:     true,
		},
		{
			name:     "127.0.0.1",
			endpoint: "http://127.0.0.1:8000",
			want:     true,
		},
		{
			name:     "::1 (IPv6 localhost)",
			endpoint: "http://[::1]:8000",
			want:     true,
		},
		{
			name:     "subdomain.localhost",
			endpoint: "http://api.localhost:8000",
			want:     true,
		},
		{
			name:     "production endpoint",
			endpoint: "https://infraspec.sh",
			want:     false,
		},
		{
			name:     "production API endpoint",
			endpoint: "https://api.infraspec.sh",
			want:     false,
		},
		{
			name:     "custom domain",
			endpoint: "https://example.com",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLocalhostEndpoint(tt.endpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildAPIEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		baseEndpoint string
		want         string
	}{
		{
			name:         "infraspec.sh to api.infraspec.sh",
			baseEndpoint: "https://infraspec.sh",
			want:         "https://api.infraspec.sh",
		},
		{
			name:         "localhost stays localhost",
			baseEndpoint: "http://localhost:8000",
			want:         "http://localhost:8000",
		},
		{
			name:         "127.0.0.1 stays the same",
			baseEndpoint: "http://127.0.0.1:8000",
			want:         "http://127.0.0.1:8000",
		},
		{
			name:         "custom domain gets api prefix",
			baseEndpoint: "https://example.com",
			want:         "https://api.example.com",
		},
		{
			name:         "custom domain with port gets api prefix",
			baseEndpoint: "https://example.com:8443",
			want:         "https://api.example.com:8443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAPIEndpoint(tt.baseEndpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBaseEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		want      string
		setupFunc func()
		cleanFunc func()
	}{
		{
			name:     "default endpoint when no env var",
			envValue: "",
			want:     InfraspecCloudDefaultEndpointURL,
			setupFunc: func() {
				os.Unsetenv("AWS_ENDPOINT_URL")
			},
			cleanFunc: func() {},
		},
		{
			name:     "custom endpoint from env var",
			envValue: "http://localhost:8000",
			want:     "http://localhost:8000",
			setupFunc: func() {
				os.Setenv("AWS_ENDPOINT_URL", "http://localhost:8000")
			},
			cleanFunc: func() {
				os.Unsetenv("AWS_ENDPOINT_URL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			defer tt.cleanFunc()

			got := getBaseEndpoint()
			assert.Equal(t, tt.want, got)
		})
	}
}
