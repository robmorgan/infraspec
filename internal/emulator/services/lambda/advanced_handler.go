package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// parseLastModifiedToTimestamp converts an RFC3339 date string to Unix timestamp
func parseLastModifiedToTimestamp(lastModified string) float64 {
	if t, err := time.Parse(time.RFC3339, lastModified); err == nil {
		return float64(t.Unix())
	}
	return float64(time.Now().Unix())
}

// ============================================================================
// Provisioned Concurrency Handlers
// ============================================================================

// handlePutProvisionedConcurrencyConfig handles PutProvisionedConcurrencyConfig API
// PUT /functions/{FunctionName}/provisioned-concurrency?Qualifier={Qualifier}
func (s *LambdaService) handlePutProvisionedConcurrencyConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query string
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"Qualifier is required"), nil
	}

	// Parse input
	var input PutProvisionedConcurrencyConfigInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			fmt.Sprintf("Invalid request body: %v", err)), nil
	}

	if input.ProvisionedConcurrentExecutions < 1 {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"ProvisionedConcurrentExecutions must be at least 1"), nil
	}

	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Validate qualifier (must be a version number or alias, not $LATEST)
	if qualifier == "$LATEST" {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"Provisioned concurrency cannot be configured on $LATEST"), nil
	}

	// Create provisioned concurrency config
	config := StoredProvisionedConcurrencyConfig{
		FunctionArn:                              fmt.Sprintf("%s:%s", function.FunctionArn, qualifier),
		Qualifier:                                qualifier,
		RequestedProvisionedConcurrentExecutions: input.ProvisionedConcurrentExecutions,
		AvailableProvisionedConcurrentExecutions: input.ProvisionedConcurrentExecutions, // Mock: immediately available
		AllocatedProvisionedConcurrentExecutions: input.ProvisionedConcurrentExecutions,
		Status:                                   "READY", // Mock: immediately ready
		LastModified:                             now(),
	}

	// Store the config
	pcKey := fmt.Sprintf("lambda:provisioned-concurrency:%s:%s", functionName, qualifier)
	if err := s.state.Set(pcKey, config); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to save provisioned concurrency config"), nil
	}

	response := map[string]interface{}{
		"RequestedProvisionedConcurrentExecutions": config.RequestedProvisionedConcurrentExecutions,
		"AvailableProvisionedConcurrentExecutions": config.AvailableProvisionedConcurrentExecutions,
		"AllocatedProvisionedConcurrentExecutions": config.AllocatedProvisionedConcurrentExecutions,
		"Status":       config.Status,
		"LastModified": config.LastModified, // Provisioned concurrency uses ISO 8601 string format
	}
	return s.successResponse(http.StatusAccepted, response)
}

// handleGetProvisionedConcurrencyConfig handles GetProvisionedConcurrencyConfig API
// GET /functions/{FunctionName}/provisioned-concurrency?Qualifier={Qualifier}
func (s *LambdaService) handleGetProvisionedConcurrencyConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query string
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"Qualifier is required"), nil
	}

	// Get the config
	pcKey := fmt.Sprintf("lambda:provisioned-concurrency:%s:%s", functionName, qualifier)
	var config StoredProvisionedConcurrencyConfig
	if err := s.state.Get(pcKey, &config); err != nil {
		return s.errorResponse(http.StatusNotFound, "ProvisionedConcurrencyConfigNotFoundException",
			fmt.Sprintf("No provisioned concurrency configuration found for function: %s:%s", functionName, qualifier)), nil
	}

	response := map[string]interface{}{
		"RequestedProvisionedConcurrentExecutions": config.RequestedProvisionedConcurrentExecutions,
		"AvailableProvisionedConcurrentExecutions": config.AvailableProvisionedConcurrentExecutions,
		"AllocatedProvisionedConcurrentExecutions": config.AllocatedProvisionedConcurrentExecutions,
		"Status":       config.Status,
		"LastModified": config.LastModified, // Provisioned concurrency uses ISO 8601 string format
	}
	if config.StatusReason != "" {
		response["StatusReason"] = config.StatusReason
	}
	return s.successResponse(http.StatusOK, response)
}

// handleDeleteProvisionedConcurrencyConfig handles DeleteProvisionedConcurrencyConfig API
// DELETE /functions/{FunctionName}/provisioned-concurrency?Qualifier={Qualifier}
func (s *LambdaService) handleDeleteProvisionedConcurrencyConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query string
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"Qualifier is required"), nil
	}

	// Check if config exists
	pcKey := fmt.Sprintf("lambda:provisioned-concurrency:%s:%s", functionName, qualifier)
	var config StoredProvisionedConcurrencyConfig
	if err := s.state.Get(pcKey, &config); err != nil {
		return s.errorResponse(http.StatusNotFound, "ProvisionedConcurrencyConfigNotFoundException",
			fmt.Sprintf("No provisioned concurrency configuration found for function: %s:%s", functionName, qualifier)), nil
	}

	// Delete the config
	if err := s.state.Delete(pcKey); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to delete provisioned concurrency config"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// handleListProvisionedConcurrencyConfigs handles ListProvisionedConcurrencyConfigs API
// GET /functions/{FunctionName}/provisioned-concurrency?List=ALL
func (s *LambdaService) handleListProvisionedConcurrencyConfigs(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// List all provisioned concurrency configs for this function
	prefix := fmt.Sprintf("lambda:provisioned-concurrency:%s:", functionName)
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to list provisioned concurrency configs"), nil
	}

	configs := []map[string]interface{}{}
	for _, key := range keys {
		var config StoredProvisionedConcurrencyConfig
		if err := s.state.Get(key, &config); err != nil {
			continue
		}
		item := map[string]interface{}{
			"FunctionArn": config.FunctionArn,
			"RequestedProvisionedConcurrentExecutions": config.RequestedProvisionedConcurrentExecutions,
			"AvailableProvisionedConcurrentExecutions": config.AvailableProvisionedConcurrentExecutions,
			"AllocatedProvisionedConcurrentExecutions": config.AllocatedProvisionedConcurrentExecutions,
			"Status":       config.Status,
			"LastModified": config.LastModified, // Provisioned concurrency uses ISO 8601 string format
		}
		if config.StatusReason != "" {
			item["StatusReason"] = config.StatusReason
		}
		configs = append(configs, item)
	}

	response := map[string]interface{}{
		"ProvisionedConcurrencyConfigs": configs,
	}
	return s.successResponse(http.StatusOK, response)
}

// ============================================================================
// Event Invoke Config Handlers
// ============================================================================

// handlePutFunctionEventInvokeConfig handles PutFunctionEventInvokeConfig API
// PUT /functions/{FunctionName}/event-invoke-config
func (s *LambdaService) handlePutFunctionEventInvokeConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query string (optional)
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		qualifier = "$LATEST"
	}

	// Parse input
	var input PutFunctionEventInvokeConfigInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			fmt.Sprintf("Invalid request body: %v", err)), nil
	}

	// Validate input
	if input.MaximumEventAgeInSeconds != nil {
		if *input.MaximumEventAgeInSeconds < 60 || *input.MaximumEventAgeInSeconds > 21600 {
			return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
				"MaximumEventAgeInSeconds must be between 60 and 21600"), nil
		}
	}
	if input.MaximumRetryAttempts != nil {
		if *input.MaximumRetryAttempts < 0 || *input.MaximumRetryAttempts > 2 {
			return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
				"MaximumRetryAttempts must be between 0 and 2"), nil
		}
	}

	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Create event invoke config
	config := StoredEventInvokeConfig{
		FunctionArn:              function.FunctionArn,
		Qualifier:                qualifier,
		MaximumEventAgeInSeconds: input.MaximumEventAgeInSeconds,
		MaximumRetryAttempts:     input.MaximumRetryAttempts,
		DestinationConfig:        input.DestinationConfig,
		LastModified:             now(),
	}

	// Store the config
	eicKey := fmt.Sprintf("lambda:event-invoke-config:%s:%s", functionName, qualifier)
	if err := s.state.Set(eicKey, config); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to save event invoke config"), nil
	}

	return s.buildEventInvokeConfigResponse(config)
}

// handleGetFunctionEventInvokeConfig handles GetFunctionEventInvokeConfig API
// GET /functions/{FunctionName}/event-invoke-config
func (s *LambdaService) handleGetFunctionEventInvokeConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query string (optional)
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		qualifier = "$LATEST"
	}

	// Get the config
	eicKey := fmt.Sprintf("lambda:event-invoke-config:%s:%s", functionName, qualifier)
	var config StoredEventInvokeConfig
	if err := s.state.Get(eicKey, &config); err != nil {
		return s.errorResponse(http.StatusNotFound, "EventInvokeConfigNotFoundException",
			fmt.Sprintf("No event invoke configuration found for function: %s:%s", functionName, qualifier)), nil
	}

	return s.buildEventInvokeConfigResponse(config)
}

// handleUpdateFunctionEventInvokeConfig handles UpdateFunctionEventInvokeConfig API
// POST /functions/{FunctionName}/event-invoke-config
func (s *LambdaService) handleUpdateFunctionEventInvokeConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query string (optional)
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		qualifier = "$LATEST"
	}

	// Parse input
	var input PutFunctionEventInvokeConfigInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			fmt.Sprintf("Invalid request body: %v", err)), nil
	}

	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Get existing config or create new one
	eicKey := fmt.Sprintf("lambda:event-invoke-config:%s:%s", functionName, qualifier)
	var config StoredEventInvokeConfig
	if err := s.state.Get(eicKey, &config); err != nil {
		// Create new config if doesn't exist
		config = StoredEventInvokeConfig{
			FunctionArn: function.FunctionArn,
			Qualifier:   qualifier,
		}
	}

	// Update fields if provided
	if input.MaximumEventAgeInSeconds != nil {
		if *input.MaximumEventAgeInSeconds < 60 || *input.MaximumEventAgeInSeconds > 21600 {
			return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
				"MaximumEventAgeInSeconds must be between 60 and 21600"), nil
		}
		config.MaximumEventAgeInSeconds = input.MaximumEventAgeInSeconds
	}
	if input.MaximumRetryAttempts != nil {
		if *input.MaximumRetryAttempts < 0 || *input.MaximumRetryAttempts > 2 {
			return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
				"MaximumRetryAttempts must be between 0 and 2"), nil
		}
		config.MaximumRetryAttempts = input.MaximumRetryAttempts
	}
	if input.DestinationConfig != nil {
		config.DestinationConfig = input.DestinationConfig
	}
	config.LastModified = now()

	// Store the config
	if err := s.state.Set(eicKey, config); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to save event invoke config"), nil
	}

	return s.buildEventInvokeConfigResponse(config)
}

// handleDeleteFunctionEventInvokeConfig handles DeleteFunctionEventInvokeConfig API
// DELETE /functions/{FunctionName}/event-invoke-config
func (s *LambdaService) handleDeleteFunctionEventInvokeConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query string (optional)
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		qualifier = "$LATEST"
	}

	// Check if config exists
	eicKey := fmt.Sprintf("lambda:event-invoke-config:%s:%s", functionName, qualifier)
	var config StoredEventInvokeConfig
	if err := s.state.Get(eicKey, &config); err != nil {
		return s.errorResponse(http.StatusNotFound, "EventInvokeConfigNotFoundException",
			fmt.Sprintf("No event invoke configuration found for function: %s:%s", functionName, qualifier)), nil
	}

	// Delete the config
	if err := s.state.Delete(eicKey); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to delete event invoke config"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// handleListFunctionEventInvokeConfigs handles ListFunctionEventInvokeConfigs API
// GET /functions/{FunctionName}/event-invoke-config/list
func (s *LambdaService) handleListFunctionEventInvokeConfigs(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// List all event invoke configs for this function
	prefix := fmt.Sprintf("lambda:event-invoke-config:%s:", functionName)
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to list event invoke configs"), nil
	}

	configs := []map[string]interface{}{}
	for _, key := range keys {
		var config StoredEventInvokeConfig
		if err := s.state.Get(key, &config); err != nil {
			continue
		}
		item := map[string]interface{}{
			"FunctionArn":  config.FunctionArn,
			"LastModified": parseLastModifiedToTimestamp(config.LastModified), // Event invoke config uses Unix timestamp
		}
		if config.MaximumEventAgeInSeconds != nil {
			item["MaximumEventAgeInSeconds"] = *config.MaximumEventAgeInSeconds
		}
		if config.MaximumRetryAttempts != nil {
			item["MaximumRetryAttempts"] = *config.MaximumRetryAttempts
		}
		if config.DestinationConfig != nil {
			item["DestinationConfig"] = config.DestinationConfig
		}
		configs = append(configs, item)
	}

	response := map[string]interface{}{
		"FunctionEventInvokeConfigs": configs,
	}
	return s.successResponse(http.StatusOK, response)
}

// buildEventInvokeConfigResponse builds the response for event invoke config operations
func (s *LambdaService) buildEventInvokeConfigResponse(config StoredEventInvokeConfig) (*emulator.AWSResponse, error) {
	response := map[string]interface{}{
		"FunctionArn":  config.FunctionArn,
		"LastModified": parseLastModifiedToTimestamp(config.LastModified), // Event invoke config uses Unix timestamp
	}
	if config.MaximumEventAgeInSeconds != nil {
		response["MaximumEventAgeInSeconds"] = *config.MaximumEventAgeInSeconds
	}
	if config.MaximumRetryAttempts != nil {
		response["MaximumRetryAttempts"] = *config.MaximumRetryAttempts
	}
	if config.DestinationConfig != nil {
		response["DestinationConfig"] = config.DestinationConfig
	}
	return s.successResponse(http.StatusOK, response)
}

// ============================================================================
// Account Settings Handler
// ============================================================================

// handleGetAccountSettings handles GetAccountSettings API
// GET /account-settings
func (s *LambdaService) handleGetAccountSettings(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Count functions
	functionKeys, err := s.state.List("lambda:functions:")
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to list functions"), nil
	}

	// Calculate total code size (mock estimate)
	var totalCodeSize int64 = 0
	for _, key := range functionKeys {
		var function StoredFunction
		if err := s.state.Get(key, &function); err == nil {
			totalCodeSize += function.CodeSize
		}
	}

	functionCount := int64(len(functionKeys))
	response := map[string]interface{}{
		"AccountLimit": &AccountLimit{
			TotalCodeSize:                  ptr(int64(80530636800)), // 75 GB
			CodeSizeUnzipped:               ptr(int64(262144000)),   // 250 MB
			CodeSizeZipped:                 ptr(int64(52428800)),    // 50 MB
			ConcurrentExecutions:           ptr(int32(1000)),
			UnreservedConcurrentExecutions: ptr(int32(1000)),
		},
		"AccountUsage": &AccountUsage{
			TotalCodeSize: &totalCodeSize,
			FunctionCount: &functionCount,
		},
	}

	return s.successResponse(http.StatusOK, response)
}

// ============================================================================
// Code Signing Config Handlers (Stubs)
// ============================================================================

// handleGetFunctionCodeSigningConfig handles GetFunctionCodeSigningConfig API
// GET /functions/{FunctionName}/code-signing-config
// Returns empty response indicating no code signing config (common for development)
func (s *LambdaService) handleGetFunctionCodeSigningConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Return empty response - no code signing config by default
	response := map[string]interface{}{
		"FunctionName": functionName,
	}
	return s.successResponse(http.StatusOK, response)
}

// handlePutFunctionCodeSigningConfig handles PutFunctionCodeSigningConfig API
// PUT /functions/{FunctionName}/code-signing-config
// Stub implementation - accepts but doesn't enforce code signing
func (s *LambdaService) handlePutFunctionCodeSigningConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Parse input to get CodeSigningConfigArn (but we don't store it)
	var input struct {
		CodeSigningConfigArn string `json:"CodeSigningConfigArn"`
	}
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			fmt.Sprintf("Invalid request body: %v", err)), nil
	}

	response := map[string]interface{}{
		"CodeSigningConfigArn": input.CodeSigningConfigArn,
		"FunctionName":         functionName,
	}
	return s.successResponse(http.StatusOK, response)
}

// handleDeleteFunctionCodeSigningConfig handles DeleteFunctionCodeSigningConfig API
// DELETE /functions/{FunctionName}/code-signing-config
func (s *LambdaService) handleDeleteFunctionCodeSigningConfig(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Check if function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
