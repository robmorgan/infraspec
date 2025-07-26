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
)

type HttpRequestOptions struct {
	Endpoint    string
	Method      string
	Headers     map[string]string
	ContentType string
	FormData    map[string]string
	BaseDir     string // BaseDir for file uploads based on feature file location
	File        *File
	RequestBody []byte
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
	client  *http.Client
	baseDir string
}

// NewHttpClient creates a new HttpClient instance
func NewHttpClient(baseDir string) *HttpClient {
	return &HttpClient{
		client:  &http.Client{},
		baseDir: baseDir,
	}
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
	var buf bytes.Buffer
	var writer *multipart.Writer
	var req *http.Request
	var err error

	// Determine if we need multipart form data
	needsMultipart := opts.File != nil || opts.FormData != nil

	if needsMultipart {
		// Create multipart writer for form data and/or file uploads
		writer = multipart.NewWriter(&buf)

		// Handle file upload if specified
		if opts.File != nil {
			// Resolve file path relative to base directory if needed
			fullPath := opts.File.FilePath
			if !filepath.IsAbs(opts.File.FilePath) && h.baseDir != "" {
				fullPath = filepath.Join(h.baseDir, opts.File.FilePath)
			}

			file, err := os.Open(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %w", fullPath, err)
			}
			defer file.Close()

			// Add file field
			fieldName := opts.File.FieldName
			if fieldName == "" {
				fieldName = "file"
			}
			fileWriter, err := writer.CreateFormFile(fieldName, filepath.Base(opts.File.FilePath))
			if err != nil {
				return nil, fmt.Errorf("failed to create form file: %w", err)
			}

			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return nil, fmt.Errorf("failed to copy file content: %w", err)
			}
		}

		// Add form data fields
		if opts.FormData != nil {
			for key, value := range opts.FormData {
				err = writer.WriteField(key, value)
				if err != nil {
					return nil, fmt.Errorf("failed to write form field %s: %w", key, err)
				}
			}
		}

		// Close the multipart writer
		err = writer.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close multipart writer: %w", err)
		}

		// Create request with multipart body
		req, err = http.NewRequestWithContext(ctx, opts.Method, opts.Endpoint, &buf)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set multipart content type
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		// Create request with regular body or no body
		var body io.Reader
		if opts.RequestBody != nil {
			body = bytes.NewReader(opts.RequestBody)
		}
		req, err = http.NewRequestWithContext(ctx, opts.Method, opts.Endpoint, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set content type if specified
		if opts.ContentType != "" {
			req.Header.Set("Content-Type", opts.ContentType)
		}
	}

	// Set additional headers
	if opts.Headers != nil {
		for key, value := range opts.Headers {
			if strings.ToLower(key) != "content-type" {
				req.Header.Set(key, value)
			}
		}
	}

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
