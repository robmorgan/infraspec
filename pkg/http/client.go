package http

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

type HttpRequest struct {
	Url      string
	Headers  map[string]string
	FormData map[string]string
	File     *File
}

type File struct {
	FieldName string
	FilePath  string
}

type HttpResponse struct {
	Status     string
	StatusCode int
	Body       []byte
}

func UploadFile(filename, fieldname, url string) (*HttpResponse, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a buffer to store the multipart data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create a form file field
	// TODO - only upload files if one was specified in options
	part, err := writer.CreateFormFile(fieldname, filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content to the form field
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add other form fields
	for key, value := range formData {
		err = writer.WriteField(key, value)
		if err != nil {
			return fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	// Close the multipart writer to finalize the form
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the content type with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and display response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", responseBody)

}

// UploadFile uploads a file using multipart/form-data
func (h *HTTPAsserter) UploadFile(ctx context.Context, opts HttpRequest) (*HttpResponse, error) {
	// Resolve file path relative to base directory if needed
	fullPath := filePath
	if !filepath.IsAbs(filePath) && h.baseDir != "" {
		fullPath = filepath.Join(h.baseDir, filePath)
	}

	// TODO - only upload a file if one was specified
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fullPath, err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add other form fields
	for key, value := range formData {
		err = writer.WriteField(key, value)
		if err != nil {
			return fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set additional headers
	for key, value := range headers {
		if strings.ToLower(key) != "content-type" {
			req.Header.Set(key, value)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP file upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("file upload failed with status %d: %s", resp.StatusCode, string(h.lastBody))
	}

	return &HttpResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       responseBody,
	}, nil
}
