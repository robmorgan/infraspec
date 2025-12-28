package lambda

import (
	"context"
	"fmt"
	"net/http"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// handleGetFunction handles the GetFunction API
// GET /2015-03-31/functions/{FunctionName}
func (s *LambdaService) handleGetFunction(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query params (optional - for specific version or alias)
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// If a qualifier is specified, handle version or alias lookup
	var configuration map[string]interface{}
	if qualifier != "" && qualifier != "$LATEST" {
		// Check if it's a version number
		if version, exists := function.PublishedVersions[qualifier]; exists {
			configuration = s.buildVersionConfigurationResponse(&function, version)
		} else {
			// Check if it's an alias
			aliasKey := fmt.Sprintf("lambda:aliases:%s:%s", functionName, qualifier)
			var alias StoredAlias
			if err := s.state.Get(aliasKey, &alias); err != nil {
				return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
					fmt.Sprintf("Function version or alias not found: %s:%s", functionName, qualifier)), nil
			}
			// Get the version the alias points to
			if version, exists := function.PublishedVersions[alias.FunctionVersion]; exists {
				configuration = s.buildVersionConfigurationResponse(&function, version)
			} else if alias.FunctionVersion == "$LATEST" {
				configuration = s.buildFunctionConfigurationResponse(&function)
			} else {
				return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
					fmt.Sprintf("Function version not found: %s:%s", functionName, alias.FunctionVersion)), nil
			}
		}
	} else {
		configuration = s.buildFunctionConfigurationResponse(&function)
	}

	// Build the GetFunction response which includes Code and Configuration
	response := map[string]interface{}{
		"Configuration": configuration,
		"Code":          s.buildCodeResponse(&function),
	}

	// Add tags if present
	if len(function.Tags) > 0 {
		response["Tags"] = function.Tags
	}

	return s.successResponse(http.StatusOK, response)
}

// handleGetFunctionConfiguration handles the GetFunctionConfiguration API
// GET /2015-03-31/functions/{FunctionName}/configuration
func (s *LambdaService) handleGetFunctionConfiguration(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query params (optional)
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// If a qualifier is specified, handle version or alias lookup
	var response map[string]interface{}
	if qualifier != "" && qualifier != "$LATEST" {
		// Check if it's a version number
		if version, exists := function.PublishedVersions[qualifier]; exists {
			response = s.buildVersionConfigurationResponse(&function, version)
		} else {
			// Check if it's an alias
			aliasKey := fmt.Sprintf("lambda:aliases:%s:%s", functionName, qualifier)
			var alias StoredAlias
			if err := s.state.Get(aliasKey, &alias); err != nil {
				return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
					fmt.Sprintf("Function version or alias not found: %s:%s", functionName, qualifier)), nil
			}
			// Get the version the alias points to
			if version, exists := function.PublishedVersions[alias.FunctionVersion]; exists {
				response = s.buildVersionConfigurationResponse(&function, version)
			} else if alias.FunctionVersion == "$LATEST" {
				response = s.buildFunctionConfigurationResponse(&function)
			} else {
				return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
					fmt.Sprintf("Function version not found: %s:%s", functionName, alias.FunctionVersion)), nil
			}
		}
	} else {
		response = s.buildFunctionConfigurationResponse(&function)
	}

	return s.successResponse(http.StatusOK, response)
}

// buildCodeResponse builds the Code section of GetFunction response
func (s *LambdaService) buildCodeResponse(fn *StoredFunction) map[string]interface{} {
	code := map[string]interface{}{
		"RepositoryType": "S3",
	}

	// For mock purposes, provide a fake S3 location
	if fn.Code != nil {
		if fn.Code.S3Bucket != "" {
			code["Location"] = fmt.Sprintf("https://awslambda-%s-tasks.s3.%s.amazonaws.com/snapshots/%s/%s",
				DefaultRegion, DefaultRegion, DefaultAccountID, fn.FunctionName)
		} else if fn.Code.ImageUri != "" {
			code["RepositoryType"] = "ECR"
			code["ImageUri"] = fn.Code.ImageUri
			code["ResolvedImageUri"] = fn.Code.ImageUri
		} else {
			// ZipFile case - provide mock S3 location
			code["Location"] = fmt.Sprintf("https://awslambda-%s-tasks.s3.%s.amazonaws.com/snapshots/%s/%s",
				DefaultRegion, DefaultRegion, DefaultAccountID, fn.FunctionName)
		}
	}

	return code
}

// buildVersionConfigurationResponse builds configuration for a published version
func (s *LambdaService) buildVersionConfigurationResponse(fn *StoredFunction, version *StoredVersion) map[string]interface{} {
	// Start with base configuration
	response := s.buildFunctionConfigurationResponse(fn)

	// Override version-specific fields
	response["Version"] = version.Version
	response["FunctionArn"] = version.FunctionArn
	response["CodeSha256"] = version.CodeSha256
	response["CodeSize"] = version.CodeSize
	response["RevisionId"] = version.RevisionId
	response["LastModified"] = version.LastModified

	if version.Description != "" {
		response["Description"] = version.Description
	}

	return response
}
