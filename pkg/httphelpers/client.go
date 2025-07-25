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

// Do performs a HTTP request.
//
// If the request has a body, it will be sent as a multipart/form-data request.
// If the request has a file, it will be uploaded as a file.
// If the request has a form data, it will be added as form fields.
// If the request has headers, they will be added to the request.
// If the request has a content type, it will be set as the content type of the request.
// If the request has a method, it will be set as the method of the request.
func (h *HttpClient) Do(ctx context.Context, opts *HttpRequestOptions) (*HttpResponse, error) {
	var buf bytes.Buffer

	// Create request
	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.Url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", opts.ContentType)

	// Set additional headers
	if opts.Headers != nil {
		for key, value := range opts.Headers {
			if strings.ToLower(key) != "content-type" {
				req.Header.Set(key, value)
			}
		}
	}

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

// UploadFile uploads a file using multipart/form-data
func (h *HttpClient) UploadFile(ctx context.Context, opts *HttpRequestOptions) (*HttpResponse, error) {
	// Resolve file path relative to base directory if needed
	var fullPath string
	if req.File != nil {
		fullPath = req.File.FilePath
		if !filepath.IsAbs(req.File.FilePath) && h.baseDir != "" {
			fullPath = filepath.Join(h.baseDir, req.File.FilePath)
		}
	} else {
		return nil, fmt.Errorf("no file specified for upload")
	}

	// TODO - only upload a file if one was specified
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fullPath, err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fieldName := req.File.FieldName
	if fieldName == "" {
		fieldName = "file"
	}
	fileWriter, err := writer.CreateFormFile(fieldName, filepath.Base(req.File.FilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add other form fields
	if req.FormData != nil {
		for key, value := range req.FormData {
			err = writer.WriteField(key, value)
			if err != nil {
				return nil, fmt.Errorf("failed to write form field %s: %w", key, err)
			}
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", opts.Url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set additional headers
	if opts.Headers != nil {
		for key, value := range opts.Headers {
			if strings.ToLower(key) != "content-type" {
				req.Header.Set(key, value)
			}
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP file upload failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("file upload failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return &HttpResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       responseBody,
	}, nil
}
