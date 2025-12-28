package metadata

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

const (
	// HeaderIMDSv2Token is the header for IMDSv2 session token
	HeaderIMDSv2Token = "X-aws-ec2-metadata-token"
	// HeaderIMDSv2TokenTTL is the header for requesting a token with specific TTL
	HeaderIMDSv2TokenTTL = "X-aws-ec2-metadata-token-ttl-seconds"
)

// Handler handles EC2 metadata service requests
type Handler struct {
	state    emulator.StateManager
	endpoint *MetadataEndpoint
}

// NewHandler creates a new metadata service handler
func NewHandler(state emulator.StateManager) *Handler {
	return &Handler{
		state:    state,
		endpoint: NewMetadataEndpoint(state),
	}
}

// ServeHTTP handles HTTP requests for the metadata service
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Metadata] %s %s", r.Method, r.URL.Path)

	// Handle IMDSv2 token generation
	if r.URL.Path == "/latest/api/token" {
		h.handleTokenRequest(w, r)
		return
	}

	// For all other requests, validate IMDSv2 token if present
	token := r.Header.Get(HeaderIMDSv2Token)
	if token != "" {
		valid, err := ValidateToken(h.state, token)
		if err != nil {
			log.Printf("[Metadata] Error validating token: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !valid {
			log.Printf("[Metadata] Invalid or expired token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	// If no token is present, allow request (IMDSv1 compatibility)

	// Handle metadata requests
	h.handleMetadataRequest(w, r)
}

// handleTokenRequest handles IMDSv2 token generation (PUT /latest/api/token)
func (h *Handler) handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse TTL from header
	ttlHeader := r.Header.Get(HeaderIMDSv2TokenTTL)
	ttl, err := ParseTTL(ttlHeader)
	if err != nil {
		log.Printf("[Metadata] Invalid TTL: %v", err)
		http.Error(w, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
		return
	}

	// Generate token
	token, err := GenerateToken(h.state, ttl)
	if err != nil {
		log.Printf("[Metadata] Failed to generate token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[Metadata] Generated IMDSv2 token with TTL %d seconds", ttl)

	// Return token as plain text
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(token.Token))
}

// handleMetadataRequest handles metadata retrieval requests
func (h *Handler) handleMetadataRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get metadata for the requested path
	path := r.URL.Path
	content, err := h.endpoint.GetMetadata(path)
	if err != nil {
		log.Printf("[Metadata] Not found: %s - %v", path, err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Determine content type based on path
	contentType := "text/plain"
	if strings.Contains(path, "iam/security-credentials/") && !strings.HasSuffix(path, "/") {
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Server", "EC2ws")
	w.WriteHeader(http.StatusOK)

	// For HEAD requests, don't write body
	if r.Method == http.MethodHead {
		return
	}

	_, _ = w.Write([]byte(content))
}
