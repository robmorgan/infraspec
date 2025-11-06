package httphelpers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robmorgan/infraspec/pkg/retry"
)

// HttpRequestOptions holds the options for an HTTP request
type HttpRequestOptions struct {
	Endpoint    string
	Method      string
	Headers     map[string]string
	ContentType string
	FormData    map[string]string
	File        *File
	RequestBody []byte
	BasicAuth   *BasicAuth
	BearerToken string
	RetryConfig *RetryConfig
}

// RetryConfig holds configuration for the retry mechanism
type RetryConfig struct {
	MaxRetries       int
	InitialDelay     time.Duration
	MaxDelay         time.Duration
	BackoffFactor    float64
	TargetStatusCode int
	TargetBody       string
}

type BasicAuth struct {
	Username string
	Password string
}

type File struct {
	FieldName string
	FilePath  string
}

type HttpResponse struct {
	Status     string
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// HttpClient handles HTTP requests and file uploads
type HttpClient struct {
	client *http.Client
}

// NewHttpClient creates a new HttpClient instance
func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{},
	}
}

// DoWithRetry performs an HTTP request with retry logic.
//
// It uses the same logic as Do but wraps it with retry functionality.
// The retry behavior is controlled by maxRetries and sleepBetweenRetries parameters.
func (h *HttpClient) DoWithRetry(ctx context.Context, opts *HttpRequestOptions, expectedStatusCode int, expectedBody string, maxRetries int, sleepBetweenRetries time.Duration) (*HttpResponse, error) {
	description := fmt.Sprintf("HTTP %s request to %s", opts.Method, opts.Endpoint)

	resp, err := retry.DoWithRetryInterface(description, maxRetries, sleepBetweenRetries, func() (interface{}, error) {
		resp, err := h.Do(ctx, opts)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != expectedStatusCode {
			return nil, fmt.Errorf("expected status code %d but got %d", expectedStatusCode, resp.StatusCode)
		}
		if expectedBody != "" && !strings.Contains(string(resp.Body), expectedBody) {
			return nil, fmt.Errorf("expected body %s but got %s", expectedBody, string(resp.Body))
		}
		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Do performs an HTTP request.
//
// If the request has a body, it will be sent as a multipart/form-data request.
// If the request has a file, it will be uploaded as a file.
// If the request has a form data, it will be added as form fields.
// If the request has headers, they will be added to the request.
// If the request has a content type, it will be set as the content type of the request.
// If the request has a method, it will be set as the method of the request.
func (h *HttpClient) Do(ctx context.Context, opts *HttpRequestOptions) (*HttpResponse, error) {
	req, err := h.createRequest(ctx, opts)
	if err != nil {
		return nil, err
	}

	return h.executeRequest(req)
}

// createRequest creates an HTTP request based on the provided options
func (h *HttpClient) createRequest(ctx context.Context, opts *HttpRequestOptions) (*http.Request, error) {
	var buf bytes.Buffer
	var req *http.Request
	var err error

	// Determine if we need multipart form data
	needsMultipart := opts.File != nil || opts.FormData != nil

	if needsMultipart {
		req, err = h.createMultipartRequest(ctx, opts, &buf)
	} else {
		req, err = h.createRegularRequest(ctx, opts)
	}

	if err != nil {
		return nil, err
	}

	// Set additional headers
	h.setHeaders(req, opts)

	// Set authentication
	h.setAuthentication(req, opts)

	return req, nil
}

// createMultipartRequest creates a multipart form data request
func (h *HttpClient) createMultipartRequest(ctx context.Context, opts *HttpRequestOptions, buf *bytes.Buffer) (*http.Request, error) {
	writer := multipart.NewWriter(buf)

	// Handle file upload if specified
	if opts.File != nil {
		if err := h.addFileToRequest(writer, opts.File); err != nil {
			return nil, err
		}
	}

	// Add form data fields
	if opts.FormData != nil {
		if err := h.addFormDataToRequest(writer, opts.FormData); err != nil {
			return nil, err
		}
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request with multipart body
	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.Endpoint, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set multipart content type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

// createRegularRequest creates a regular HTTP request
func (h *HttpClient) createRegularRequest(ctx context.Context, opts *HttpRequestOptions) (*http.Request, error) {
	var body io.Reader
	if opts.RequestBody != nil {
		body = bytes.NewReader(opts.RequestBody)
	}

	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.Endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set content type if specified
	if opts.ContentType != "" {
		req.Header.Set("Content-Type", opts.ContentType)
	}

	return req, nil
}

// addFileToRequest adds a file to the multipart request
func (h *HttpClient) addFileToRequest(writer *multipart.Writer, file *File) error {
	fileHandle, err := os.Open(file.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", file.FilePath, err)
	}
	defer fileHandle.Close()

	// Add file field
	fieldName := file.FieldName
	if fieldName == "" {
		fieldName = "file"
	}
	fileWriter, err := writer.CreateFormFile(fieldName, filepath.Base(file.FilePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, fileHandle)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// addFormDataToRequest adds form data fields to the multipart request
func (h *HttpClient) addFormDataToRequest(writer *multipart.Writer, formData map[string]string) error {
	for key, value := range formData {
		err := writer.WriteField(key, value)
		if err != nil {
			return fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}
	return nil
}

// setHeaders sets additional headers on the request
func (h *HttpClient) setHeaders(req *http.Request, opts *HttpRequestOptions) {
	if opts.Headers != nil {
		for key, value := range opts.Headers {
			if strings.ToLower(key) != "content-type" {
				req.Header.Set(key, value)
			}
		}
	}
}

// setAuthentication sets authentication on the request
func (h *HttpClient) setAuthentication(req *http.Request, opts *HttpRequestOptions) {
	// Set basic auth credentials if specified
	if opts.BasicAuth != nil {
		req.SetBasicAuth(opts.BasicAuth.Username, opts.BasicAuth.Password)
	}

	// Set Bearer token if specified
	if opts.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+opts.BearerToken)
	}
}

// executeRequest executes the HTTP request and returns the response
func (h *HttpClient) executeRequest(req *http.Request) (*HttpResponse, error) {
	// Send request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HttpResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}, nil
}
