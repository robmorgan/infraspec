package lambda

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// handleListTags handles the ListTags API
// GET /2015-03-31/tags/{ARN}
func (s *LambdaService) handleListTags(ctx context.Context, resourceArn string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// URL decode the ARN
	decodedArn, err := url.PathUnescape(resourceArn)
	if err != nil {
		decodedArn = resourceArn
	}

	// Parse the ARN to get function name
	functionName := parseFunctionNameFromArn(decodedArn)
	if functionName == "" {
		// Try treating the ARN as a function name directly
		functionName = decodedArn
	}

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Resource not found: %s", resourceArn)), nil
	}

	// Return tags
	tags := function.Tags
	if tags == nil {
		tags = make(map[string]string)
	}

	response := map[string]interface{}{
		"Tags": tags,
	}

	return s.successResponse(http.StatusOK, response)
}

// handleTagResource handles the TagResource API
// POST /2015-03-31/tags/{ARN}
func (s *LambdaService) handleTagResource(ctx context.Context, resourceArn string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// URL decode the ARN
	decodedArn, err := url.PathUnescape(resourceArn)
	if err != nil {
		decodedArn = resourceArn
	}

	var input TagResourceInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	if len(input.Tags) == 0 {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"Tags is required"), nil
	}

	// Parse the ARN to get function name
	functionName := parseFunctionNameFromArn(decodedArn)
	if functionName == "" {
		functionName = decodedArn
	}

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Resource not found: %s", resourceArn)), nil
	}

	// Merge tags
	if function.Tags == nil {
		function.Tags = make(map[string]string)
	}
	for key, value := range input.Tags {
		function.Tags[key] = value
	}

	// Save the function
	if err := s.state.Set(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update tags"), nil
	}

	// Return 204 No Content for successful tagging
	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: nil,
	}, nil
}

// handleUntagResource handles the UntagResource API
// DELETE /2015-03-31/tags/{ARN}?tagKeys=key1&tagKeys=key2
func (s *LambdaService) handleUntagResource(ctx context.Context, resourceArn string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// URL decode the ARN
	decodedArn, err := url.PathUnescape(resourceArn)
	if err != nil {
		decodedArn = resourceArn
	}

	// Get tag keys from query parameters
	queryParams := parseQueryParams(req.Path)
	tagKeys := queryParams["tagKeys"]
	if len(tagKeys) == 0 {
		// Try comma-separated format
		tagKeysStr := queryParams.Get("tagKeys")
		if tagKeysStr != "" {
			tagKeys = strings.Split(tagKeysStr, ",")
		}
	}

	if len(tagKeys) == 0 {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"tagKeys is required"), nil
	}

	// Parse the ARN to get function name
	functionName := parseFunctionNameFromArn(decodedArn)
	if functionName == "" {
		functionName = decodedArn
	}

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Resource not found: %s", resourceArn)), nil
	}

	// Remove tags
	if function.Tags != nil {
		for _, key := range tagKeys {
			delete(function.Tags, key)
		}
	}

	// Save the function
	if err := s.state.Set(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update tags"), nil
	}

	// Return 204 No Content for successful untagging
	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: nil,
	}, nil
}
