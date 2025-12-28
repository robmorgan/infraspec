package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

type EmulatorHandler struct {
	router emulator.RequestRouter
}

func NewEmulatorHandler(router emulator.RequestRouter) *EmulatorHandler {
	return &EmulatorHandler{
		router: router,
	}
}

func (h *EmulatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	service, err := h.router.Route(r)
	if err != nil {
		log.Printf("Failed to route request: %v", err)
		h.writeErrorResponseForRequest(w, r, 400, "InvalidService", err.Error())
		return
	}

	awsReq, err := h.convertHTTPRequest(r)
	if err != nil {
		log.Printf("Failed to convert HTTP request: %v", err)
		h.writeErrorResponseForService(w, r, service, 400, "InvalidRequest", err.Error())
		return
	}

	// If the service implements ActionExtractor, let it extract the action
	// before we log. This is needed for REST-based services like S3.
	if actionExtractor, ok := service.(emulator.ActionExtractor); ok {
		awsReq.Action = actionExtractor.ExtractAction(awsReq)
	}

	// Log the service and action for each request
	log.Printf("Service: %s, Action: %s", service.ServiceName(), awsReq.Action)

	awsResp, err := service.HandleRequest(ctx, awsReq)
	if err != nil {
		log.Printf("Service error: %v", err)
		h.writeErrorResponseForService(w, r, service, 500, "InternalFailure", err.Error())
		return
	}

	h.writeAWSResponse(w, awsResp)
}

func (h *EmulatorHandler) convertHTTPRequest(r *http.Request) (*emulator.AWSRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Add Host header explicitly as it's not in r.Header
	headers["Host"] = r.Host

	action := h.extractAction(r, headers, body)

	// Include query string in path for S3 operations like ?publicAccessBlock, ?versioning, etc.
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path = path + "?" + r.URL.RawQuery
	}

	return &emulator.AWSRequest{
		Method:  r.Method,
		Path:    path,
		Headers: headers,
		Body:    body,
		Action:  action,
	}, nil
}

func (h *EmulatorHandler) extractAction(r *http.Request, headers map[string]string, body []byte) string {
	target := headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	action := r.URL.Query().Get("Action")
	if action != "" {
		return action
	}

	// Check form data for Action parameter (used by Terraform and AWS CLI)
	contentType := r.Header.Get("Content-Type")
	if r.Method == "POST" && strings.Contains(contentType, "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(string(body))
		if err == nil {
			if action := values.Get("Action"); action != "" {
				return action
			}
		}
	}

	return ""
}

func (h *EmulatorHandler) writeAWSResponse(w http.ResponseWriter, resp *emulator.AWSResponse) {
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(resp.Body)
}

func (h *EmulatorHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	errorXML := `<?xml version="1.0" encoding="UTF-8"?>
<ErrorResponse>
    <Error>
        <Code>` + code + `</Code>
        <Message>` + message + `</Message>
    </Error>
    <RequestId>00000000-0000-0000-0000-000000000000</RequestId>
</ErrorResponse>`

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(statusCode)
	w.Write([]byte(errorXML))
}

// writeErrorResponseForRequest writes an error response when we haven't determined the service yet
func (h *EmulatorHandler) writeErrorResponseForRequest(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	// Detect service from request headers to determine error format
	if h.isJSONProtocolService(r) {
		h.writeJSONErrorResponse(w, statusCode, code, message)
	} else {
		h.writeErrorResponse(w, statusCode, code, message)
	}
}

// writeErrorResponseForService writes an error response based on the service protocol
func (h *EmulatorHandler) writeErrorResponseForService(w http.ResponseWriter, r *http.Request, service emulator.Service, statusCode int, code, message string) {
	serviceName := service.ServiceName()

	// JSON protocol services
	if serviceName == "dynamodb_20120810" {
		h.writeJSONErrorResponse(w, statusCode, code, message)
		return
	}

	// Default to XML for Query protocol services (RDS, STS, etc.)
	h.writeErrorResponse(w, statusCode, code, message)
}

// isJSONProtocolService checks if the request is for a JSON protocol service
func (h *EmulatorHandler) isJSONProtocolService(r *http.Request) bool {
	// Check for X-Amz-Target header (used by DynamoDB and other JSON protocol services)
	if target := r.Header.Get("X-Amz-Target"); target != "" {
		return strings.HasPrefix(target, "DynamoDB_")
	}
	return false
}

// writeJSONErrorResponse writes a JSON error response for JSON protocol services
func (h *EmulatorHandler) writeJSONErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	errorData := map[string]interface{}{
		"__type": code,
		"message": message,
	}

	body, _ := json.Marshal(errorData)

	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.Header().Set("x-amzn-RequestId", "00000000-0000-0000-0000-000000000000")
	w.Header().Set("x-amzn-ErrorType", code)
	w.WriteHeader(statusCode)
	w.Write(body)
}

func (h *EmulatorHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "aws-emulator",
	}

	// Include Railway git commit SHA if present
	if gitSHA := os.Getenv("RAILWAY_GIT_COMMIT_SHA"); gitSHA != "" {
		response["git_commit_sha"] = gitSHA
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

func (h *EmulatorHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	services := h.router.GetServices()

	serviceNames := make([]string, 0, len(services))
	for _, service := range services {
		serviceNames = append(serviceNames, service.ServiceName())
	}

	// Create JSON response
	response := map[string]interface{}{
		"services": serviceNames,
		"count":    len(serviceNames),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

func (h *EmulatorHandler) RootStatus(w http.ResponseWriter, r *http.Request) {
	// Check if this is an S3 virtual-hosted style request
	// S3 virtual-hosted requests have patterns like: bucket-name.s3.infraspec.sh or bucket-name.s3.localhost
	// These should be forwarded to the S3 handler, not return the root status
	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	if emulator.IsS3VirtualHostedRequest(host) {
		// This is an S3 virtual-hosted request, forward to the emulator handler
		h.ServeHTTP(w, r)
		return
	}

	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
