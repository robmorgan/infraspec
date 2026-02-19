package lambda

import (
	"context"
	"fmt"
	"net/http"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// ============================================================================
// Phase 2: Invoke (Mock) - See invoke_handler.go
// ============================================================================

// ============================================================================
// Phase 3: Versions & Aliases
// ============================================================================

// handlePublishVersion handles the PublishVersion API
// POST /2015-03-31/functions/{FunctionName}/versions
func (s *LambdaService) handlePublishVersion(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input PublishVersionInput
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

	// Check code SHA if provided
	if input.CodeSha256 != "" && input.CodeSha256 != function.CodeSha256 {
		return s.errorResponse(http.StatusBadRequest, "CodeStorageExceededException",
			"The code SHA256 provided does not match"), nil
	}

	// Create new version
	version := fmt.Sprintf("%d", function.NextVersionNumber)
	storedVersion := &StoredVersion{
		Version:      version,
		Description:  input.Description,
		CodeSha256:   function.CodeSha256,
		CodeSize:     function.CodeSize,
		RevisionId:   generateRevisionId(),
		FunctionArn:  generateVersionArn(functionName, version),
		LastModified: now(),
	}

	if function.PublishedVersions == nil {
		function.PublishedVersions = make(map[string]*StoredVersion)
	}
	function.PublishedVersions[version] = storedVersion
	function.NextVersionNumber++

	// Save the function
	if err := s.state.Set(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to publish version"), nil
	}

	response := s.buildVersionConfigurationResponse(&function, storedVersion)
	return s.successResponse(http.StatusCreated, response)
}

// handleCreateAlias handles the CreateAlias API
// POST /2015-03-31/functions/{FunctionName}/aliases
func (s *LambdaService) handleCreateAlias(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input CreateAliasInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	if err := validateAliasName(input.Name); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	if input.FunctionVersion == "" {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"FunctionVersion is required"), nil
	}

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Verify the version exists
	if input.FunctionVersion != "$LATEST" {
		if _, exists := function.PublishedVersions[input.FunctionVersion]; !exists {
			return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
				fmt.Sprintf("Version not found: %s", input.FunctionVersion)), nil
		}
	}

	// Check if alias already exists
	aliasKey := fmt.Sprintf("lambda:aliases:%s:%s", functionName, input.Name)
	var existing StoredAlias
	if err := s.state.Get(aliasKey, &existing); err == nil {
		return s.errorResponse(http.StatusConflict, "ResourceConflictException",
			fmt.Sprintf("Alias already exists: %s", input.Name)), nil
	}

	// Create the alias
	alias := &StoredAlias{
		Name:            input.Name,
		FunctionName:    functionName,
		FunctionVersion: input.FunctionVersion,
		Description:     input.Description,
		AliasArn:        generateAliasArn(functionName, input.Name),
		RevisionId:      generateRevisionId(),
		RoutingConfig:   input.RoutingConfig,
	}

	// Save the alias
	if err := s.state.Set(aliasKey, alias); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to create alias"), nil
	}

	response := s.buildAliasResponse(alias)
	return s.successResponse(http.StatusCreated, response)
}

// handleGetAlias handles the GetAlias API
func (s *LambdaService) handleGetAlias(ctx context.Context, functionName, aliasName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	aliasKey := fmt.Sprintf("lambda:aliases:%s:%s", functionName, aliasName)
	var alias StoredAlias
	if err := s.state.Get(aliasKey, &alias); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Alias not found: %s", aliasName)), nil
	}

	response := s.buildAliasResponse(&alias)
	return s.successResponse(http.StatusOK, response)
}

// handleUpdateAlias handles the UpdateAlias API
func (s *LambdaService) handleUpdateAlias(ctx context.Context, functionName, aliasName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input UpdateAliasInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	aliasKey := fmt.Sprintf("lambda:aliases:%s:%s", functionName, aliasName)
	var alias StoredAlias
	if err := s.state.Get(aliasKey, &alias); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Alias not found: %s", aliasName)), nil
	}

	// Check revision ID
	if input.RevisionId != "" && input.RevisionId != alias.RevisionId {
		return s.errorResponse(http.StatusPreconditionFailed, "PreconditionFailedException",
			"The Revision Id provided does not match"), nil
	}

	// Update fields
	if input.FunctionVersion != "" {
		alias.FunctionVersion = input.FunctionVersion
	}
	if input.Description != "" {
		alias.Description = input.Description
	}
	if input.RoutingConfig != nil {
		alias.RoutingConfig = input.RoutingConfig
	}
	alias.RevisionId = generateRevisionId()

	// Save
	if err := s.state.Set(aliasKey, &alias); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update alias"), nil
	}

	response := s.buildAliasResponse(&alias)
	return s.successResponse(http.StatusOK, response)
}

// handleDeleteAlias handles the DeleteAlias API
func (s *LambdaService) handleDeleteAlias(ctx context.Context, functionName, aliasName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	aliasKey := fmt.Sprintf("lambda:aliases:%s:%s", functionName, aliasName)
	var alias StoredAlias
	if err := s.state.Get(aliasKey, &alias); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Alias not found: %s", aliasName)), nil
	}

	if err := s.state.Delete(aliasKey); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to delete alias"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       nil,
	}, nil
}

// handleListAliases handles the ListAliases API
func (s *LambdaService) handleListAliases(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// List all aliases for this function
	prefix := fmt.Sprintf("lambda:aliases:%s:", functionName)
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to list aliases"), nil
	}

	aliases := make([]map[string]interface{}, 0)
	for _, key := range keys {
		var alias StoredAlias
		if err := s.state.Get(key, &alias); err == nil {
			aliases = append(aliases, s.buildAliasResponse(&alias))
		}
	}

	response := map[string]interface{}{
		"Aliases": aliases,
	}

	return s.successResponse(http.StatusOK, response)
}

// ============================================================================
// Phase 4: Function URLs
// ============================================================================

// handleCreateFunctionUrlConfig handles CreateFunctionUrlConfig API
func (s *LambdaService) handleCreateFunctionUrlConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input CreateFunctionUrlConfigInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	// Validate auth type
	if input.AuthType != AuthTypeNone && input.AuthType != AuthTypeIAM {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"AuthType must be NONE or AWS_IAM"), nil
	}

	// Load function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Check if URL already exists
	urlKey := fmt.Sprintf("lambda:function-urls:%s", functionName)
	var existing StoredFunctionUrl
	if err := s.state.Get(urlKey, &existing); err == nil {
		return s.errorResponse(http.StatusConflict, "ResourceConflictException",
			"Function URL already exists"), nil
	}

	// Create URL config
	urlConfig := &StoredFunctionUrl{
		FunctionName:     functionName,
		FunctionArn:      function.FunctionArn,
		FunctionUrl:      generateFunctionUrl(functionName),
		AuthType:         input.AuthType,
		Cors:             input.Cors,
		InvokeMode:       coalesce(input.InvokeMode, InvokeModeBuffered),
		CreationTime:     now(),
		LastModifiedTime: now(),
	}

	// Save
	if err := s.state.Set(urlKey, urlConfig); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to create function URL"), nil
	}

	response := s.buildFunctionUrlResponse(urlConfig)
	return s.successResponse(http.StatusCreated, response)
}

// handleGetFunctionUrlConfig handles GetFunctionUrlConfig API
func (s *LambdaService) handleGetFunctionUrlConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	urlKey := fmt.Sprintf("lambda:function-urls:%s", functionName)
	var urlConfig StoredFunctionUrl
	if err := s.state.Get(urlKey, &urlConfig); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"Function URL not found"), nil
	}

	response := s.buildFunctionUrlResponse(&urlConfig)
	return s.successResponse(http.StatusOK, response)
}

// handleUpdateFunctionUrlConfig handles UpdateFunctionUrlConfig API
func (s *LambdaService) handleUpdateFunctionUrlConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input UpdateFunctionUrlConfigInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	urlKey := fmt.Sprintf("lambda:function-urls:%s", functionName)
	var urlConfig StoredFunctionUrl
	if err := s.state.Get(urlKey, &urlConfig); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"Function URL not found"), nil
	}

	// Update fields
	if input.AuthType != "" {
		urlConfig.AuthType = input.AuthType
	}
	if input.Cors != nil {
		urlConfig.Cors = input.Cors
	}
	if input.InvokeMode != "" {
		urlConfig.InvokeMode = input.InvokeMode
	}
	urlConfig.LastModifiedTime = now()

	// Save
	if err := s.state.Set(urlKey, &urlConfig); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update function URL"), nil
	}

	response := s.buildFunctionUrlResponse(&urlConfig)
	return s.successResponse(http.StatusOK, response)
}

// handleDeleteFunctionUrlConfig handles DeleteFunctionUrlConfig API
func (s *LambdaService) handleDeleteFunctionUrlConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	urlKey := fmt.Sprintf("lambda:function-urls:%s", functionName)
	var urlConfig StoredFunctionUrl
	if err := s.state.Get(urlKey, &urlConfig); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"Function URL not found"), nil
	}

	if err := s.state.Delete(urlKey); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to delete function URL"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       nil,
	}, nil
}

// ============================================================================
// Phase 7: Permissions & Concurrency
// ============================================================================

// handleGetPolicy handles GetPolicy API
func (s *LambdaService) handleGetPolicy(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	if function.Policy == "" {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"No policy is associated with the function"), nil
	}

	response := map[string]interface{}{
		"Policy":     function.Policy,
		"RevisionId": function.RevisionId,
	}
	return s.successResponse(http.StatusOK, response)
}

// handleGetFunctionConcurrency handles GetFunctionConcurrency API
func (s *LambdaService) handleGetFunctionConcurrency(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	response := map[string]interface{}{}
	if function.ReservedConcurrentExecutions != nil {
		response["ReservedConcurrentExecutions"] = *function.ReservedConcurrentExecutions
	}

	return s.successResponse(http.StatusOK, response)
}

// handlePutFunctionConcurrency handles PutFunctionConcurrency API
func (s *LambdaService) handlePutFunctionConcurrency(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input PutFunctionConcurrencyInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	function.ReservedConcurrentExecutions = &input.ReservedConcurrentExecutions

	if err := s.state.Set(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update concurrency"), nil
	}

	response := map[string]interface{}{
		"ReservedConcurrentExecutions": input.ReservedConcurrentExecutions,
	}
	return s.successResponse(http.StatusOK, response)
}

// handleDeleteFunctionConcurrency handles DeleteFunctionConcurrency API
func (s *LambdaService) handleDeleteFunctionConcurrency(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	function.ReservedConcurrentExecutions = nil

	if err := s.state.Set(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to delete concurrency"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       nil,
	}, nil
}

// ============================================================================
// Response Builders
// ============================================================================

func (s *LambdaService) buildAliasResponse(alias *StoredAlias) map[string]interface{} {
	response := map[string]interface{}{
		"AliasArn":        alias.AliasArn,
		"Name":            alias.Name,
		"FunctionVersion": alias.FunctionVersion,
		"RevisionId":      alias.RevisionId,
	}
	if alias.Description != "" {
		response["Description"] = alias.Description
	}
	if alias.RoutingConfig != nil && len(alias.RoutingConfig.AdditionalVersionWeights) > 0 {
		response["RoutingConfig"] = map[string]interface{}{
			"AdditionalVersionWeights": alias.RoutingConfig.AdditionalVersionWeights,
		}
	}
	return response
}

func (s *LambdaService) buildFunctionUrlResponse(url *StoredFunctionUrl) map[string]interface{} {
	response := map[string]interface{}{
		"FunctionArn":      url.FunctionArn,
		"FunctionUrl":      url.FunctionUrl,
		"AuthType":         url.AuthType,
		"CreationTime":     url.CreationTime,
		"LastModifiedTime": url.LastModifiedTime,
	}
	if url.InvokeMode != "" {
		response["InvokeMode"] = url.InvokeMode
	}
	if url.Cors != nil {
		corsConfig := map[string]interface{}{}
		if url.Cors.AllowCredentials {
			corsConfig["AllowCredentials"] = url.Cors.AllowCredentials
		}
		if len(url.Cors.AllowHeaders) > 0 {
			corsConfig["AllowHeaders"] = url.Cors.AllowHeaders
		}
		if len(url.Cors.AllowMethods) > 0 {
			corsConfig["AllowMethods"] = url.Cors.AllowMethods
		}
		if len(url.Cors.AllowOrigins) > 0 {
			corsConfig["AllowOrigins"] = url.Cors.AllowOrigins
		}
		if len(url.Cors.ExposeHeaders) > 0 {
			corsConfig["ExposeHeaders"] = url.Cors.ExposeHeaders
		}
		if url.Cors.MaxAge > 0 {
			corsConfig["MaxAge"] = url.Cors.MaxAge
		}
		if len(corsConfig) > 0 {
			response["Cors"] = corsConfig
		}
	}
	return response
}
