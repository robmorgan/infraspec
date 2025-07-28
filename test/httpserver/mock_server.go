package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

// MockHTTPServer provides a configurable mock HTTP server for testing
type MockHTTPServer struct {
	server    *httptest.Server
	routes    map[string]http.HandlerFunc
	responses map[string]MockResponse
}

// MockResponse defines a mock HTTP response
type MockResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer() *MockHTTPServer {
	mockServer := &MockHTTPServer{
		routes:    make(map[string]http.HandlerFunc),
		responses: make(map[string]MockResponse),
	}

	// Set up default routes
	mockServer.setupDefaultRoutes()

	// Create the HTTP test server
	mux := http.NewServeMux()
	mux.HandleFunc("/", mockServer.handleRequest)
	mockServer.server = httptest.NewServer(mux)

	return mockServer
}

// URL returns the base URL of the mock server
func (m *MockHTTPServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockHTTPServer) Close() {
	m.server.Close()
}

// AddResponse adds a mock response for a specific path and method
func (m *MockHTTPServer) AddResponse(method, path string, response MockResponse) {
	key := fmt.Sprintf("%s %s", method, path)
	m.responses[key] = response
}

// setupDefaultRoutes sets up commonly used test routes
func (m *MockHTTPServer) setupDefaultRoutes() {
	// JSON endpoint
	m.AddResponse("GET", "/json", MockResponse{
		StatusCode: 200,
		Body:       `{"message": "Hello, World!", "status": "ok"}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
	})

	// Plain text endpoint
	m.AddResponse("GET", "/text", MockResponse{
		StatusCode: 200,
		Body:       "Hello, World!",
		Headers:    map[string]string{"Content-Type": "text/plain"},
	})

	// Status code test endpoints
	for _, code := range []int{200, 201, 400, 404, 500} {
		path := fmt.Sprintf("/status/%d", code)
		m.AddResponse("GET", path, MockResponse{
			StatusCode: code,
			Body:       fmt.Sprintf("Status: %d", code),
			Headers:    map[string]string{"Content-Type": "text/plain"},
		})
	}

	// Echo endpoint that returns request details
	m.AddResponse("POST", "/echo", MockResponse{
		StatusCode: 200,
		Body:       "", // Will be filled dynamically
		Headers:    map[string]string{"Content-Type": "application/json"},
	})

	// File upload endpoint
	m.AddResponse("POST", "/upload", MockResponse{
		StatusCode: 200,
		Body:       "", // Will be filled dynamically
		Headers:    map[string]string{"Content-Type": "application/json"},
	})

	// Headers test endpoint
	m.AddResponse("GET", "/headers", MockResponse{
		StatusCode: 200,
		Body:       "", // Will be filled dynamically
		Headers:    map[string]string{"Content-Type": "application/json"},
	})

	// Bearer token authentication endpoint
	m.AddResponse("GET", "/bearer", MockResponse{
		StatusCode: 200,
		Body:       "", // Will be filled dynamically
		Headers:    map[string]string{"Content-Type": "application/json"},
	})
}

// handleRequest handles all incoming requests to the mock server
func (m *MockHTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	// Handle dynamic endpoints
	switch {
	case path == "/echo" && method == "POST":
		m.handleEcho(w, r)
		return
	case path == "/upload" && method == "POST":
		m.handleUpload(w, r)
		return
	case path == "/headers" && method == "GET":
		m.handleHeaders(w, r)
		return
	case path == "/bearer" && method == "GET":
		m.handleBearer(w, r)
		return
	case strings.HasPrefix(path, "/status/"):
		m.handleStatus(w, r)
		return
	}

	// Look for configured response
	key := fmt.Sprintf("%s %s", method, path)
	response, exists := m.responses[key]
	if !exists {
		// Default 404 response
		http.NotFound(w, r)
		return
	}

	// Set headers
	for name, value := range response.Headers {
		w.Header().Set(name, value)
	}

	// Set status code
	w.WriteHeader(response.StatusCode)

	// Write body
	w.Write([]byte(response.Body))
}

// handleEcho returns request details as JSON
func (m *MockHTTPServer) handleEcho(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	response := map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"headers": r.Header,
		"body":    string(body),
	}

	jsonResponse, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResponse)
}

// handleUpload handles file upload requests
func (m *MockHTTPServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	files := make(map[string]string)
	formData := make(map[string]string)

	// Get form fields
	for key, values := range r.MultipartForm.Value {
		if len(values) > 0 {
			formData[key] = values[0]
		}
	}

	// Get uploaded files
	for fieldName, fileHeaders := range r.MultipartForm.File {
		if len(fileHeaders) > 0 {
			file, err := fileHeaders[0].Open()
			if err == nil {
				content, _ := io.ReadAll(file)
				files[fieldName] = string(content)
				file.Close()
			}
		}
	}

	response := map[string]interface{}{
		"files":     files,
		"form_data": formData,
		"status":    "uploaded",
	}

	jsonResponse, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResponse)
}

// handleHeaders returns request headers as JSON
func (m *MockHTTPServer) handleHeaders(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"headers": r.Header,
	}

	jsonResponse, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResponse)
}

// handleBearer handles Bearer token authentication
func (m *MockHTTPServer) handleBearer(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// For testing purposes, accept any non-empty token
	if token == "" {
		http.Error(w, "Empty Bearer token", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"authenticated": true,
		"token":         token,
		"message":       "Bearer token authentication successful",
	}

	jsonResponse, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResponse)
}

// handleStatus returns responses based on the status code in the URL
func (m *MockHTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Extract status code from path like /status/404
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid status path", http.StatusBadRequest)
		return
	}

	statusCode, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, "Invalid status code", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf("Status: %d", statusCode)))
}
