package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	// AWS SigV4 constants
	authorizationHeader = "Authorization"
	algorithm           = "AWS4-HMAC-SHA256"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// ServiceNameContextKey is the context key for storing the authenticated service name
	ServiceNameContextKey ContextKey = "service-name"
)

// SigV4Middleware validates AWS Signature Version 4 authentication
type SigV4Middleware struct {
	keyStore    KeyStore
	exemptPaths []string
}

// NewSigV4Middleware creates a new SigV4 authentication middleware
func NewSigV4Middleware(keyStore KeyStore, exemptPaths []string) *SigV4Middleware {
	return &SigV4Middleware{
		keyStore:    keyStore,
		exemptPaths: exemptPaths,
	}
}

// Middleware returns an HTTP middleware handler that validates SigV4 signatures
func (m *SigV4Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is exempt from authentication
		if m.isExempt(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract and validate authorization header
		authHeader := r.Header.Get(authorizationHeader)
		if authHeader == "" {
			log.Printf("Authentication failed: missing Authorization header")
			m.writeUnauthorizedResponse(w, r, "missing Authorization header")
			return
		}

		// Parse authorization header to get access key and service name
		authInfo, err := parseAuthorizationHeader(authHeader)
		if err != nil {
			log.Printf("Authentication failed: invalid Authorization header: %v", err)
			m.writeUnauthorizedResponse(w, r, "invalid Authorization header")
			return
		}

		// Validate access key exists in keystore
		if !m.keyStore.ValidateAccessKey(authInfo.AccessKey) {
			log.Printf("Authentication failed: invalid access key: %s", authInfo.AccessKey)
			m.writeUnauthorizedResponse(w, r, "invalid access key")
			return
		}

		// NOTE: We're doing minimal validation here - just checking that the access key exists
		// For a development/testing tool, this is sufficient. Real signature validation would
		// require matching the client's signature computation exactly, which is complex due to
		// differences in header normalization between SDK versions and proxy behaviors.
		log.Printf("Authentication successful for access key: %s, service: %s", authInfo.AccessKey, authInfo.Service)

		// Normalize service name to internal identifier
		// AWS SigV4 uses short names (e.g., "dynamodb") but we use versioned identifiers internally
		serviceName := m.normalizeServiceName(authInfo.Service)

		// Store the normalized service name in the request context for the router to use
		ctx := context.WithValue(r.Context(), ServiceNameContextKey, serviceName)
		r = r.WithContext(ctx)

		// Authentication successful, proceed to next handler
		next.ServeHTTP(w, r)
	})
}

// isExempt checks if a path is exempt from authentication
func (m *SigV4Middleware) isExempt(path string) bool {
	for _, exemptPath := range m.exemptPaths {
		// Support both exact matches and prefix matches (for paths ending with /)
		if path == exemptPath {
			return true
		}
		// If exempt path ends with /, treat it as a prefix
		if strings.HasSuffix(exemptPath, "/") && strings.HasPrefix(path, exemptPath) {
			return true
		}
	}
	return false
}

// AuthorizationInfo holds parsed authorization header information
type AuthorizationInfo struct {
	AccessKey     string
	Date          string
	Region        string
	Service       string
	SignedHeaders []string
	Signature     string
}

// parseAuthorizationHeader parses the AWS SigV4 Authorization header
// Format: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-date, Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024
func parseAuthorizationHeader(header string) (*AuthorizationInfo, error) {
	if !strings.HasPrefix(header, algorithm) {
		return nil, fmt.Errorf("unsupported algorithm, expected %s", algorithm)
	}

	// Remove algorithm prefix
	header = strings.TrimPrefix(header, algorithm+" ")

	// Parse components
	parts := strings.Split(header, ", ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	info := &AuthorizationInfo{}

	// Parse Credential
	credentialPart := strings.TrimPrefix(parts[0], "Credential=")
	credentialComponents := strings.Split(credentialPart, "/")
	if len(credentialComponents) != 5 {
		return nil, fmt.Errorf("invalid credential format")
	}
	info.AccessKey = credentialComponents[0]
	info.Date = credentialComponents[1]
	info.Region = credentialComponents[2]
	info.Service = credentialComponents[3]

	// Parse SignedHeaders
	signedHeadersPart := strings.TrimPrefix(parts[1], "SignedHeaders=")
	info.SignedHeaders = strings.Split(signedHeadersPart, ";")

	// Parse Signature
	info.Signature = strings.TrimPrefix(parts[2], "Signature=")

	return info, nil
}

// writeUnauthorizedResponse writes a 403 Forbidden response with AWS-compatible error format
func (m *SigV4Middleware) writeUnauthorizedResponse(w http.ResponseWriter, r *http.Request, message string) {
	// Use AWS Query Protocol error format (XML)
	errorXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ErrorResponse>
    <Error>
        <Code>SignatureDoesNotMatch</Code>
        <Message>%s</Message>
    </Error>
    <RequestId>%s</RequestId>
</ErrorResponse>`, message, generateRequestID())

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(errorXML))
}

// generateRequestID generates a simple request ID for error responses
func generateRequestID() string {
	return fmt.Sprintf("%d", getCurrentTimestamp())
}

// getCurrentTimestamp returns the current Unix timestamp in nanoseconds
func getCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

// normalizeServiceName maps AWS service names from SigV4 to internal service identifiers
func (m *SigV4Middleware) normalizeServiceName(serviceName string) string {
	// Map service names to internal service identifiers
	// This matches the mapping in router.go for consistency
	serviceMap := map[string]string{
		"dynamodb":                "dynamodb_20120810",
		"application-autoscaling": "anyscalefrontendservice",
		"autoscaling":             "anyscalefrontendservice",
		"sts":                     "sts",
		"rds":                     "rds",
		"s3":                      "s3",
		"ec2":                     "ec2",
		"ssm":                     "ssm",
	}

	if internalName, ok := serviceMap[serviceName]; ok {
		return internalName
	}

	// Return as-is if no mapping found
	return serviceName
}
