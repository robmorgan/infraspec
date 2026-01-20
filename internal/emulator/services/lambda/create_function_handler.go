package lambda

import (
	"context"
	"fmt"
	"net/http"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// handleCreateFunction handles the CreateFunction API
// POST /2015-03-31/functions
func (s *LambdaService) handleCreateFunction(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input CreateFunctionInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	// Validate required fields
	if err := validateFunctionName(input.FunctionName); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	if err := validateRole(input.Role); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	// Determine package type
	packageType := input.PackageType
	if packageType == "" {
		packageType = PackageTypeZip
	}
	if err := validatePackageType(packageType); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	// Validate runtime (required for Zip, not for Image)
	if packageType == PackageTypeZip {
		if input.Runtime == "" {
			return s.errorResponse(http.StatusBadRequest, "ValidationException",
				"Runtime is required for Zip package type"), nil
		}
		if err := validateRuntime(input.Runtime); err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
		}
		if input.Handler == "" {
			return s.errorResponse(http.StatusBadRequest, "ValidationException",
				"Handler is required for Zip package type"), nil
		}
	}

	if err := validateCode(input.Code, packageType); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	if err := validateTimeout(input.Timeout); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	if err := validateMemorySize(input.MemorySize); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	if err := validateArchitectures(input.Architectures); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	// Check if function already exists
	stateKey := fmt.Sprintf("lambda:functions:%s", input.FunctionName)
	var existing StoredFunction
	if err := s.state.Get(stateKey, &existing); err == nil {
		return s.errorResponse(http.StatusConflict, "ResourceConflictException",
			fmt.Sprintf("Function already exists: %s", input.FunctionName)), nil
	}

	// Set defaults
	timeout := coalesceInt32(DefaultTimeout, input.Timeout)
	memorySize := coalesceInt32(DefaultMemorySize, input.MemorySize)
	architectures := input.Architectures
	if len(architectures) == 0 {
		architectures = getDefaultArchitectures()
	}

	// Generate code hash and estimate size
	codeSha256 := generateCodeSha256(input.Code)
	codeSize := estimateCodeSize(input.Code)

	// Build the stored function
	function := &StoredFunction{
		FunctionName:      input.FunctionName,
		FunctionArn:       generateFunctionArn(input.FunctionName),
		Runtime:           input.Runtime,
		Role:              input.Role,
		Handler:           input.Handler,
		Description:       input.Description,
		Timeout:           timeout,
		MemorySize:        memorySize,
		CodeSha256:        codeSha256,
		CodeSize:          codeSize,
		Version:           "$LATEST",
		State:             StateActive,
		LastModified:      now(),
		LastUpdateStatus:  "Successful",
		PackageType:       packageType,
		Architectures:     architectures,
		RevisionId:        generateRevisionId(),
		Tags:              input.Tags,
		Environment:       input.Environment,
		VpcConfig:         input.VpcConfig,
		DeadLetterConfig:  input.DeadLetterConfig,
		TracingConfig:     input.TracingConfig,
		EphemeralStorage:  input.EphemeralStorage,
		Layers:            input.Layers,
		KMSKeyArn:         input.KMSKeyArn,
		FileSystemConfigs: input.FileSystemConfigs,
		LoggingConfig:     input.LoggingConfig,
		Code:              input.Code,
		PublishedVersions: make(map[string]*StoredVersion),
		NextVersionNumber: 1,
	}

	// Set default ephemeral storage if not specified
	if function.EphemeralStorage == nil {
		function.EphemeralStorage = &StoredEphemeralStorage{Size: 512}
	}

	// Set default tracing config if not specified
	if function.TracingConfig == nil {
		function.TracingConfig = &StoredTracingConfig{Mode: "PassThrough"}
	}

	// Set default logging config if not specified
	if function.LoggingConfig == nil {
		function.LoggingConfig = &StoredLoggingConfig{
			LogFormat: "Text",
			LogGroup:  fmt.Sprintf("/aws/lambda/%s", input.FunctionName),
		}
	}

	// Handle container image configuration
	if packageType == PackageTypeImage && input.Code.ImageUri != "" {
		function.ImageUri = input.Code.ImageUri
		if input.ImageConfig != nil {
			function.ImageConfigResponse = &StoredImageConfigResponse{
				ImageConfig: input.ImageConfig,
			}
		}
	}

	// Store the function
	if err := s.state.Set(stateKey, function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to create function"), nil
	}

	// Publish version if requested
	if input.Publish {
		version := fmt.Sprintf("%d", function.NextVersionNumber)
		storedVersion := &StoredVersion{
			Version:      version,
			Description:  input.Description,
			CodeSha256:   codeSha256,
			CodeSize:     codeSize,
			RevisionId:   generateRevisionId(),
			FunctionArn:  generateVersionArn(input.FunctionName, version),
			LastModified: now(),
		}
		function.PublishedVersions[version] = storedVersion
		function.NextVersionNumber++
		function.Version = version

		// Update stored function
		if err := s.state.Set(stateKey, function); err != nil {
			return s.errorResponse(http.StatusInternalServerError, "ServiceException",
				"Failed to publish version"), nil
		}
	}

	// Build the response (matches FunctionConfiguration)
	response := s.buildFunctionConfigurationResponse(function)

	return s.successResponse(http.StatusCreated, response)
}

// buildFunctionConfigurationResponse builds the response for function configuration
func (s *LambdaService) buildFunctionConfigurationResponse(fn *StoredFunction) map[string]interface{} {
	response := map[string]interface{}{
		"FunctionName":  fn.FunctionName,
		"FunctionArn":   fn.FunctionArn,
		"Role":          fn.Role,
		"CodeSize":      fn.CodeSize,
		"CodeSha256":    fn.CodeSha256,
		"Timeout":       fn.Timeout,
		"MemorySize":    fn.MemorySize,
		"LastModified":  fn.LastModified,
		"Version":       fn.Version,
		"State":         fn.State,
		"PackageType":   fn.PackageType,
		"RevisionId":    fn.RevisionId,
		"Architectures": fn.Architectures,
	}

	// Add optional fields if present
	if fn.Runtime != "" {
		response["Runtime"] = fn.Runtime
	}
	if fn.Handler != "" {
		response["Handler"] = fn.Handler
	}
	if fn.Description != "" {
		response["Description"] = fn.Description
	}
	if fn.StateReason != "" {
		response["StateReason"] = fn.StateReason
	}
	if fn.StateReasonCode != "" {
		response["StateReasonCode"] = fn.StateReasonCode
	}
	if fn.LastUpdateStatus != "" {
		response["LastUpdateStatus"] = fn.LastUpdateStatus
	}
	if fn.KMSKeyArn != "" {
		response["KMSKeyArn"] = fn.KMSKeyArn
	}
	if fn.ImageUri != "" {
		response["ImageUri"] = fn.ImageUri
	}

	// Add nested objects
	if fn.Environment != nil && len(fn.Environment.Variables) > 0 {
		response["Environment"] = map[string]interface{}{
			"Variables": fn.Environment.Variables,
		}
	}

	if fn.VpcConfig != nil && (len(fn.VpcConfig.SubnetIds) > 0 || len(fn.VpcConfig.SecurityGroupIds) > 0) {
		response["VpcConfig"] = map[string]interface{}{
			"SubnetIds":        fn.VpcConfig.SubnetIds,
			"SecurityGroupIds": fn.VpcConfig.SecurityGroupIds,
			"VpcId":            fn.VpcConfig.VpcId,
		}
	}

	if fn.DeadLetterConfig != nil && fn.DeadLetterConfig.TargetArn != "" {
		response["DeadLetterConfig"] = map[string]interface{}{
			"TargetArn": fn.DeadLetterConfig.TargetArn,
		}
	}

	if fn.TracingConfig != nil {
		response["TracingConfig"] = map[string]interface{}{
			"Mode": fn.TracingConfig.Mode,
		}
	}

	if fn.EphemeralStorage != nil {
		response["EphemeralStorage"] = map[string]interface{}{
			"Size": fn.EphemeralStorage.Size,
		}
	}

	if len(fn.Layers) > 0 {
		layers := make([]map[string]interface{}, len(fn.Layers))
		for i, layerArn := range fn.Layers {
			layers[i] = map[string]interface{}{
				"Arn": layerArn,
			}
		}
		response["Layers"] = layers
	}

	if len(fn.FileSystemConfigs) > 0 {
		fsConfigs := make([]map[string]interface{}, len(fn.FileSystemConfigs))
		for i, fsc := range fn.FileSystemConfigs {
			fsConfigs[i] = map[string]interface{}{
				"Arn":            fsc.Arn,
				"LocalMountPath": fsc.LocalMountPath,
			}
		}
		response["FileSystemConfigs"] = fsConfigs
	}

	if fn.ImageConfigResponse != nil {
		response["ImageConfigResponse"] = fn.ImageConfigResponse
	}

	if fn.LoggingConfig != nil {
		logConfig := map[string]interface{}{}
		if fn.LoggingConfig.LogFormat != "" {
			logConfig["LogFormat"] = fn.LoggingConfig.LogFormat
		}
		if fn.LoggingConfig.ApplicationLogLevel != "" {
			logConfig["ApplicationLogLevel"] = fn.LoggingConfig.ApplicationLogLevel
		}
		if fn.LoggingConfig.SystemLogLevel != "" {
			logConfig["SystemLogLevel"] = fn.LoggingConfig.SystemLogLevel
		}
		if fn.LoggingConfig.LogGroup != "" {
			logConfig["LogGroup"] = fn.LoggingConfig.LogGroup
		}
		if len(logConfig) > 0 {
			response["LoggingConfig"] = logConfig
		}
	}

	return response
}
