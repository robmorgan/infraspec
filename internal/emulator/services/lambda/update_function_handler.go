package lambda

import (
	"context"
	"fmt"
	"net/http"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// handleUpdateFunctionCode handles the UpdateFunctionCode API
// PUT /2015-03-31/functions/{FunctionName}/code
func (s *LambdaService) handleUpdateFunctionCode(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input UpdateFunctionCodeInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Check revision ID if provided
	if input.RevisionId != "" && input.RevisionId != function.RevisionId {
		return s.errorResponse(http.StatusPreconditionFailed, "PreconditionFailedException",
			"The Revision Id provided does not match the latest Revision Id"), nil
	}

	// DryRun just validates without making changes
	if input.DryRun {
		response := s.buildFunctionConfigurationResponse(&function)
		return s.successResponse(http.StatusOK, response)
	}

	// Validate code input
	hasZip := input.ZipFile != ""
	hasS3 := input.S3Bucket != "" && input.S3Key != ""
	hasImage := input.ImageUri != ""

	if !hasZip && !hasS3 && !hasImage {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"You must specify either ZipFile, S3Bucket/S3Key, or ImageUri"), nil
	}

	// Count how many code sources are specified
	codeSourceCount := 0
	if hasZip {
		codeSourceCount++
	}
	if hasS3 {
		codeSourceCount++
	}
	if hasImage {
		codeSourceCount++
	}
	if codeSourceCount > 1 {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"You can only specify one of ZipFile, S3Bucket/S3Key, or ImageUri"), nil
	}

	// Update the code
	newCode := &FunctionCode{
		ZipFile:         input.ZipFile,
		S3Bucket:        input.S3Bucket,
		S3Key:           input.S3Key,
		S3ObjectVersion: input.S3ObjectVersion,
		ImageUri:        input.ImageUri,
	}

	// Update architectures if provided
	if len(input.Architectures) > 0 {
		if err := validateArchitectures(input.Architectures); err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
		}
		function.Architectures = input.Architectures
	}

	// Update function code
	function.Code = newCode
	function.CodeSha256 = generateCodeSha256(newCode)
	function.CodeSize = estimateCodeSize(newCode)
	function.LastModified = now()
	function.RevisionId = generateRevisionId()
	function.LastUpdateStatus = "Successful"

	if hasImage {
		function.ImageUri = input.ImageUri
		function.PackageType = PackageTypeImage
	}

	// Publish new version if requested
	if input.Publish {
		version := fmt.Sprintf("%d", function.NextVersionNumber)
		storedVersion := &StoredVersion{
			Version:      version,
			CodeSha256:   function.CodeSha256,
			CodeSize:     function.CodeSize,
			RevisionId:   generateRevisionId(),
			FunctionArn:  generateVersionArn(functionName, version),
			LastModified: now(),
		}
		function.PublishedVersions[version] = storedVersion
		function.NextVersionNumber++
		function.Version = version
	}

	// Save the function
	if err := s.state.Set(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update function code"), nil
	}

	response := s.buildFunctionConfigurationResponse(&function)
	return s.successResponse(http.StatusOK, response)
}

// handleUpdateFunctionConfiguration handles the UpdateFunctionConfiguration API
// PUT /2015-03-31/functions/{FunctionName}/configuration
func (s *LambdaService) handleUpdateFunctionConfiguration(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input UpdateFunctionConfigurationInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Check revision ID if provided
	if input.RevisionId != "" && input.RevisionId != function.RevisionId {
		return s.errorResponse(http.StatusPreconditionFailed, "PreconditionFailedException",
			"The Revision Id provided does not match the latest Revision Id"), nil
	}

	// Validate and apply updates
	if input.Runtime != "" {
		if err := validateRuntime(input.Runtime); err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
		}
		function.Runtime = input.Runtime
	}

	if input.Role != "" {
		if err := validateRole(input.Role); err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
		}
		function.Role = input.Role
	}

	if input.Handler != "" {
		function.Handler = input.Handler
	}

	if input.Description != "" {
		function.Description = input.Description
	}

	if input.Timeout != nil {
		if err := validateTimeout(input.Timeout); err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
		}
		function.Timeout = *input.Timeout
	}

	if input.MemorySize != nil {
		if err := validateMemorySize(input.MemorySize); err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
		}
		function.MemorySize = *input.MemorySize
	}

	if input.VpcConfig != nil {
		function.VpcConfig = input.VpcConfig
	}

	if input.DeadLetterConfig != nil {
		function.DeadLetterConfig = input.DeadLetterConfig
	}

	if input.Environment != nil {
		function.Environment = input.Environment
	}

	if input.KMSKeyArn != "" {
		function.KMSKeyArn = input.KMSKeyArn
	}

	if input.TracingConfig != nil {
		function.TracingConfig = input.TracingConfig
	}

	if len(input.Layers) > 0 {
		function.Layers = input.Layers
	}

	if len(input.FileSystemConfigs) > 0 {
		function.FileSystemConfigs = input.FileSystemConfigs
	}

	if input.ImageConfig != nil {
		function.ImageConfigResponse = &StoredImageConfigResponse{
			ImageConfig: input.ImageConfig,
		}
	}

	if input.EphemeralStorage != nil {
		function.EphemeralStorage = input.EphemeralStorage
	}

	if input.LoggingConfig != nil {
		function.LoggingConfig = input.LoggingConfig
	}

	// Update metadata
	function.LastModified = now()
	function.RevisionId = generateRevisionId()
	function.LastUpdateStatus = "Successful"

	// Save the function
	if err := s.state.Set(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update function configuration"), nil
	}

	response := s.buildFunctionConfigurationResponse(&function)
	return s.successResponse(http.StatusOK, response)
}
