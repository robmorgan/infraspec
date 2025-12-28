package metadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func setupTest(t *testing.T) (*Handler, emulator.StateManager) {
	state := emulator.NewMemoryStateManager()
	if err := InitializeDefaults(state); err != nil {
		t.Fatalf("Failed to initialize defaults: %v", err)
	}
	handler := NewHandler(state)
	return handler, state
}

// TestIMDSv1_InstanceID tests IMDSv1 (no token) access to instance-id
func TestIMDSv1_InstanceID(t *testing.T) {
	handler, _ := setupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/instance-id", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if !strings.HasPrefix(body, "i-") {
		t.Errorf("Expected instance ID to start with 'i-', got %s", body)
	}
}

// TestIMDSv2_TokenGeneration tests IMDSv2 token generation
func TestIMDSv2_TokenGeneration(t *testing.T) {
	handler, _ := setupTest(t)

	req := httptest.NewRequest(http.MethodPut, "/latest/api/token", nil)
	req.Header.Set(HeaderIMDSv2TokenTTL, "300")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	token := strings.TrimSpace(w.Body.String())
	if token == "" {
		t.Error("Expected non-empty token")
	}

	if len(token) < 10 {
		t.Errorf("Token too short: %s", token)
	}
}

// TestIMDSv2_TokenValidation tests using a token to access metadata
func TestIMDSv2_TokenValidation(t *testing.T) {
	handler, _ := setupTest(t)

	// Generate token
	tokenReq := httptest.NewRequest(http.MethodPut, "/latest/api/token", nil)
	tokenReq.Header.Set(HeaderIMDSv2TokenTTL, "60")
	tokenW := httptest.NewRecorder()
	handler.ServeHTTP(tokenW, tokenReq)

	token := strings.TrimSpace(tokenW.Body.String())

	// Use token to access metadata
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/instance-id", nil)
	req.Header.Set(HeaderIMDSv2Token, token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestIMDSv2_InvalidToken tests that invalid tokens are rejected
func TestIMDSv2_InvalidToken(t *testing.T) {
	handler, _ := setupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/instance-id", nil)
	req.Header.Set(HeaderIMDSv2Token, "invalid-token-12345")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestIMDSv2_ExpiredToken tests that expired tokens are rejected
func TestIMDSv2_ExpiredToken(t *testing.T) {
	handler, state := setupTest(t)

	// Create an expired token directly in state
	expiredToken := &IMDSv2Token{
		Token:     "expired-token-test",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		TTL:       300,
	}
	state.Set("metadata:tokens:expired-token-test", expiredToken)

	// Try to use expired token
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/instance-id", nil)
	req.Header.Set(HeaderIMDSv2Token, "expired-token-test")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestMetadataEndpoints tests various metadata endpoints
func TestMetadataEndpoints(t *testing.T) {
	handler, _ := setupTest(t)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedPrefix string
	}{
		{"Instance ID", "/latest/meta-data/instance-id", 200, "i-"},
		{"Instance Type", "/latest/meta-data/instance-type", 200, "t3."},
		{"AMI ID", "/latest/meta-data/ami-id", 200, "ami-"},
		{"Local IPv4", "/latest/meta-data/local-ipv4", 200, "172."},
		{"Public IPv4", "/latest/meta-data/public-ipv4", 200, "54."},
		{"MAC Address", "/latest/meta-data/mac", 200, "0e:"},
		{"Availability Zone", "/latest/meta-data/placement/availability-zone", 200, "us-east-"},
		{"Region", "/latest/meta-data/placement/region", 200, "us-east-"},
		{"Not Found", "/latest/meta-data/nonexistent", 404, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == 200 && tt.expectedPrefix != "" {
				body := strings.TrimSpace(w.Body.String())
				if !strings.HasPrefix(body, tt.expectedPrefix) {
					t.Errorf("Expected body to start with '%s', got '%s'", tt.expectedPrefix, body)
				}
			}
		})
	}
}

// TestDirectoryListing tests metadata directory listings
func TestDirectoryListing(t *testing.T) {
	handler, _ := setupTest(t)

	tests := []struct {
		name             string
		path             string
		expectedContains []string
	}{
		{
			"Meta-data root",
			"/latest/meta-data",
			[]string{"instance-id", "instance-type", "ami-id", "placement/", "network/", "iam/"},
		},
		{
			"Placement directory",
			"/latest/meta-data/placement/",
			[]string{"availability-zone", "region"},
		},
		{
			"IAM directory",
			"/latest/meta-data/iam/",
			[]string{"security-credentials/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			body := w.Body.String()
			for _, expected := range tt.expectedContains {
				if !strings.Contains(body, expected) {
					t.Errorf("Expected response to contain '%s', got: %s", expected, body)
				}
			}
		})
	}
}

// TestIAMCredentials tests IAM role credentials endpoint
func TestIAMCredentials(t *testing.T) {
	handler, state := setupTest(t)

	// Get role name
	var roleName string
	if err := state.Get("metadata:instance:iam-role", &roleName); err != nil {
		t.Fatalf("Failed to get role name: %v", err)
	}

	// Test listing roles
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/iam/security-credentials/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if !strings.Contains(body, roleName) {
		t.Errorf("Expected role listing to contain '%s', got: %s", roleName, body)
	}

	// Test getting credentials
	req = httptest.NewRequest(http.MethodGet, "/latest/meta-data/iam/security-credentials/"+roleName, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify it's valid JSON
	var creds IAMCredentials
	if err := json.Unmarshal(w.Body.Bytes(), &creds); err != nil {
		t.Errorf("Failed to parse credentials JSON: %v", err)
	}

	// Verify credential fields
	if creds.AccessKeyID == "" {
		t.Error("Expected AccessKeyID to be non-empty")
	}
	if creds.SecretAccessKey == "" {
		t.Error("Expected SecretAccessKey to be non-empty")
	}
	if creds.Token == "" {
		t.Error("Expected Token to be non-empty")
	}
	if creds.Code != "Success" {
		t.Errorf("Expected Code to be 'Success', got '%s'", creds.Code)
	}
}

// TestUserData tests user data endpoint
func TestUserData(t *testing.T) {
	handler, state := setupTest(t)

	// Initially user data is empty, should return 404
	req := httptest.NewRequest(http.MethodGet, "/latest/user-data", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for empty user data, got %d", w.Code)
	}

	// Set user data
	userData := "#!/bin/bash\necho 'Hello World'"
	if err := state.Set("metadata:instance:user-data", userData); err != nil {
		t.Fatalf("Failed to set user data: %v", err)
	}

	// Now should return user data
	req = httptest.NewRequest(http.MethodGet, "/latest/user-data", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != userData {
		t.Errorf("Expected user data '%s', got '%s'", userData, body)
	}
}

// TestNetworkInterfaces tests network interface metadata
func TestNetworkInterfaces(t *testing.T) {
	handler, state := setupTest(t)

	// Get MAC address
	var mac string
	if err := state.Get("metadata:instance:mac", &mac); err != nil {
		t.Fatalf("Failed to get MAC: %v", err)
	}

	// Test network interfaces directory
	req := httptest.NewRequest(http.MethodGet, "/latest/meta-data/network/interfaces/macs/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if !strings.Contains(body, mac) {
		t.Errorf("Expected MAC listing to contain '%s', got: %s", mac, body)
	}

	// Test specific interface metadata
	tests := []struct {
		name string
		path string
	}{
		{"Subnet ID", "/latest/meta-data/network/interfaces/macs/" + mac + "/subnet-id"},
		{"VPC ID", "/latest/meta-data/network/interfaces/macs/" + mac + "/vpc-id"},
		{"Local IPv4s", "/latest/meta-data/network/interfaces/macs/" + mac + "/local-ipv4s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			if w.Body.Len() == 0 {
				t.Error("Expected non-empty response")
			}
		})
	}
}

// TestMethodNotAllowed tests that only GET/HEAD/PUT are allowed
func TestMethodNotAllowed(t *testing.T) {
	handler, _ := setupTest(t)

	req := httptest.NewRequest(http.MethodPost, "/latest/meta-data/instance-id", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHEADRequest tests HEAD requests
func TestHEADRequest(t *testing.T) {
	handler, _ := setupTest(t)

	req := httptest.NewRequest(http.MethodHead, "/latest/meta-data/instance-id", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// HEAD should have no body
	if w.Body.Len() > 0 {
		t.Error("Expected empty body for HEAD request")
	}
}

// TestTokenTTLValidation tests token TTL validation
func TestTokenTTLValidation(t *testing.T) {
	handler, _ := setupTest(t)

	tests := []struct {
		name           string
		ttl            string
		expectedStatus int
	}{
		{"Valid TTL", "300", 200},
		{"Minimum TTL", "1", 200},
		{"Maximum TTL", "21600", 200},
		{"Too Low", "0", 400},
		{"Too High", "21601", 400},
		{"Invalid", "abc", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/latest/api/token", nil)
			req.Header.Set(HeaderIMDSv2TokenTTL, tt.ttl)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
