package lambda

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// Event source mapping states
const (
	EventSourceStateCreating  = "Creating"
	EventSourceStateEnabling  = "Enabling"
	EventSourceStateEnabled   = "Enabled"
	EventSourceStateDisabling = "Disabling"
	EventSourceStateDisabled  = "Disabled"
	EventSourceStateUpdating  = "Updating"
	EventSourceStateDeleting  = "Deleting"
)

// handleCreateEventSourceMapping handles CreateEventSourceMapping API
// POST /2015-03-31/event-source-mappings
func (s *LambdaService) handleCreateEventSourceMapping(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input CreateEventSourceMappingInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	// Validate required fields
	if input.FunctionName == "" {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"FunctionName is required"), nil
	}

	// Parse function name (could be ARN or name)
	functionName := parseFunctionName(input.FunctionName)

	// Verify function exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Validate event source - either EventSourceArn or SelfManagedEventSource required
	if input.EventSourceArn == "" && input.SelfManagedEventSource == nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"Either EventSourceArn or SelfManagedEventSource is required"), nil
	}

	// Validate starting position for stream sources
	if input.EventSourceArn != "" && isStreamSource(input.EventSourceArn) {
		if input.StartingPosition == "" {
			return s.errorResponse(http.StatusBadRequest, "ValidationException",
				"StartingPosition is required for stream event sources"), nil
		}
		validPositions := map[string]bool{"TRIM_HORIZON": true, "LATEST": true, "AT_TIMESTAMP": true}
		if !validPositions[input.StartingPosition] {
			return s.errorResponse(http.StatusBadRequest, "ValidationException",
				"StartingPosition must be TRIM_HORIZON, LATEST, or AT_TIMESTAMP"), nil
		}
	}

	// Set defaults
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	batchSize := getDefaultBatchSize(input.EventSourceArn)
	if input.BatchSize != nil {
		batchSize = *input.BatchSize
	}

	// Generate UUID
	mappingUUID := uuid.New().String()

	// Determine initial state
	state := EventSourceStateEnabled
	if !enabled {
		state = EventSourceStateDisabled
	}

	// Create the mapping
	mapping := &StoredEventSourceMapping{
		UUID:                                mappingUUID,
		EventSourceArn:                      input.EventSourceArn,
		FunctionArn:                         function.FunctionArn,
		FunctionName:                        functionName,
		State:                               state,
		StateTransitionReason:               "User action",
		LastModified:                        now(),
		BatchSize:                           &batchSize,
		MaximumBatchingWindowInSeconds:      input.MaximumBatchingWindowInSeconds,
		ParallelizationFactor:               input.ParallelizationFactor,
		StartingPosition:                    input.StartingPosition,
		StartingPositionTimestamp:           input.StartingPositionTimestamp,
		MaximumRecordAgeInSeconds:           input.MaximumRecordAgeInSeconds,
		BisectBatchOnFunctionError:          input.BisectBatchOnFunctionError,
		MaximumRetryAttempts:                input.MaximumRetryAttempts,
		TumblingWindowInSeconds:             input.TumblingWindowInSeconds,
		Enabled:                             &enabled,
		FilterCriteria:                      input.FilterCriteria,
		DestinationConfig:                   input.DestinationConfig,
		Queues:                              input.Queues,
		SourceAccessConfigurations:          convertSourceAccessConfigs(input.SourceAccessConfigurations),
		SelfManagedEventSource:              input.SelfManagedEventSource,
		FunctionResponseTypes:               input.FunctionResponseTypes,
		AmazonManagedKafkaEventSourceConfig: input.AmazonManagedKafkaEventSourceConfig,
		SelfManagedKafkaEventSourceConfig:   input.SelfManagedKafkaEventSourceConfig,
		ScalingConfig:                       input.ScalingConfig,
		DocumentDBEventSourceConfig:         input.DocumentDBEventSourceConfig,
	}

	// Save the mapping
	mappingKey := fmt.Sprintf("lambda:event-source-mappings:%s", mappingUUID)
	if err := s.state.Set(mappingKey, mapping); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to create event source mapping"), nil
	}

	response := s.buildEventSourceMappingResponse(mapping)
	return s.successResponse(http.StatusAccepted, response)
}

// handleGetEventSourceMapping handles GetEventSourceMapping API
// GET /2015-03-31/event-source-mappings/{UUID}
func (s *LambdaService) handleGetEventSourceMapping(ctx context.Context, mappingUUID string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	mappingKey := fmt.Sprintf("lambda:event-source-mappings:%s", mappingUUID)
	var mapping StoredEventSourceMapping
	if err := s.state.Get(mappingKey, &mapping); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Event source mapping not found: %s", mappingUUID)), nil
	}

	response := s.buildEventSourceMappingResponse(&mapping)
	return s.successResponse(http.StatusOK, response)
}

// handleUpdateEventSourceMapping handles UpdateEventSourceMapping API
// PUT /2015-03-31/event-source-mappings/{UUID}
func (s *LambdaService) handleUpdateEventSourceMapping(ctx context.Context, mappingUUID string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input UpdateEventSourceMappingInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	mappingKey := fmt.Sprintf("lambda:event-source-mappings:%s", mappingUUID)
	var mapping StoredEventSourceMapping
	if err := s.state.Get(mappingKey, &mapping); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Event source mapping not found: %s", mappingUUID)), nil
	}

	// Update fields if provided
	if input.Enabled != nil {
		mapping.Enabled = input.Enabled
		if *input.Enabled {
			mapping.State = EventSourceStateEnabled
		} else {
			mapping.State = EventSourceStateDisabled
		}
	}
	if input.BatchSize != nil {
		mapping.BatchSize = input.BatchSize
	}
	if input.MaximumBatchingWindowInSeconds != nil {
		mapping.MaximumBatchingWindowInSeconds = input.MaximumBatchingWindowInSeconds
	}
	if input.ParallelizationFactor != nil {
		mapping.ParallelizationFactor = input.ParallelizationFactor
	}
	if input.MaximumRecordAgeInSeconds != nil {
		mapping.MaximumRecordAgeInSeconds = input.MaximumRecordAgeInSeconds
	}
	if input.BisectBatchOnFunctionError != nil {
		mapping.BisectBatchOnFunctionError = input.BisectBatchOnFunctionError
	}
	if input.MaximumRetryAttempts != nil {
		mapping.MaximumRetryAttempts = input.MaximumRetryAttempts
	}
	if input.TumblingWindowInSeconds != nil {
		mapping.TumblingWindowInSeconds = input.TumblingWindowInSeconds
	}
	if input.FilterCriteria != nil {
		mapping.FilterCriteria = input.FilterCriteria
	}
	if input.DestinationConfig != nil {
		mapping.DestinationConfig = input.DestinationConfig
	}
	if len(input.SourceAccessConfigurations) > 0 {
		mapping.SourceAccessConfigurations = convertSourceAccessConfigs(input.SourceAccessConfigurations)
	}
	if input.FunctionName != "" {
		// Update to point to different function
		functionName := parseFunctionName(input.FunctionName)
		stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
		var function StoredFunction
		if err := s.state.Get(stateKey, &function); err != nil {
			return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
				fmt.Sprintf("Function not found: %s", functionName)), nil
		}
		mapping.FunctionName = functionName
		mapping.FunctionArn = function.FunctionArn
	}
	if len(input.FunctionResponseTypes) > 0 {
		mapping.FunctionResponseTypes = input.FunctionResponseTypes
	}
	if input.ScalingConfig != nil {
		mapping.ScalingConfig = input.ScalingConfig
	}
	if input.DocumentDBEventSourceConfig != nil {
		mapping.DocumentDBEventSourceConfig = input.DocumentDBEventSourceConfig
	}

	mapping.LastModified = now()
	mapping.StateTransitionReason = "User action"

	// Save updated mapping
	if err := s.state.Set(mappingKey, &mapping); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update event source mapping"), nil
	}

	response := s.buildEventSourceMappingResponse(&mapping)
	return s.successResponse(http.StatusAccepted, response)
}

// handleDeleteEventSourceMapping handles DeleteEventSourceMapping API
// DELETE /2015-03-31/event-source-mappings/{UUID}
func (s *LambdaService) handleDeleteEventSourceMapping(ctx context.Context, mappingUUID string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	mappingKey := fmt.Sprintf("lambda:event-source-mappings:%s", mappingUUID)
	var mapping StoredEventSourceMapping
	if err := s.state.Get(mappingKey, &mapping); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Event source mapping not found: %s", mappingUUID)), nil
	}

	// Mark as deleting and return the current state
	mapping.State = EventSourceStateDeleting
	mapping.StateTransitionReason = "User action"
	mapping.LastModified = now()

	// Delete the mapping
	if err := s.state.Delete(mappingKey); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to delete event source mapping"), nil
	}

	response := s.buildEventSourceMappingResponse(&mapping)
	return s.successResponse(http.StatusAccepted, response)
}

// handleListEventSourceMappings handles ListEventSourceMappings API
// GET /2015-03-31/event-source-mappings
func (s *LambdaService) handleListEventSourceMappings(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get query parameters
	queryParams := parseQueryParams(req.Path)
	eventSourceArn := queryParams.Get("EventSourceArn")
	functionName := queryParams.Get("FunctionName")
	marker := queryParams.Get("Marker")
	maxItemsStr := queryParams.Get("MaxItems")

	// Parse MaxItems
	maxItems := 100
	if maxItemsStr != "" {
		if parsed, err := parseMaxItems(maxItemsStr); err == nil {
			maxItems = parsed
		}
	}

	// Parse function name if provided
	if functionName != "" {
		functionName = parseFunctionName(functionName)
	}

	// List all event source mappings
	prefix := "lambda:event-source-mappings:"
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to list event source mappings"), nil
	}

	// Load and filter mappings
	var mappings []StoredEventSourceMapping
	for _, key := range keys {
		var mapping StoredEventSourceMapping
		if err := s.state.Get(key, &mapping); err == nil {
			// Apply filters
			if eventSourceArn != "" && mapping.EventSourceArn != eventSourceArn {
				continue
			}
			if functionName != "" && mapping.FunctionName != functionName {
				continue
			}
			mappings = append(mappings, mapping)
		}
	}

	// Sort by UUID for consistent ordering
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].UUID < mappings[j].UUID
	})

	// Apply pagination
	startIndex := 0
	if marker != "" {
		for i, m := range mappings {
			if m.UUID == marker {
				startIndex = i + 1
				break
			}
		}
	}

	var pageMappings []StoredEventSourceMapping
	var nextMarker string
	if startIndex < len(mappings) {
		endIndex := startIndex + maxItems
		if endIndex > len(mappings) {
			endIndex = len(mappings)
		} else {
			nextMarker = mappings[endIndex-1].UUID
		}
		pageMappings = mappings[startIndex:endIndex]
	}

	// Build response
	mappingResponses := make([]map[string]interface{}, len(pageMappings))
	for i, m := range pageMappings {
		mappingResponses[i] = s.buildEventSourceMappingResponse(&m)
	}

	response := map[string]interface{}{
		"EventSourceMappings": mappingResponses,
	}
	if nextMarker != "" {
		response["NextMarker"] = nextMarker
	}

	return s.successResponse(http.StatusOK, response)
}

// buildEventSourceMappingResponse builds the response for an event source mapping
func (s *LambdaService) buildEventSourceMappingResponse(mapping *StoredEventSourceMapping) map[string]interface{} {
	// AWS SDK expects LastModified as a Unix timestamp (float64), not a string
	var lastModified float64
	if t, err := time.Parse(time.RFC3339, mapping.LastModified); err == nil {
		lastModified = float64(t.Unix())
	} else {
		lastModified = float64(time.Now().Unix())
	}

	response := map[string]interface{}{
		"UUID":                  mapping.UUID,
		"FunctionArn":           mapping.FunctionArn,
		"State":                 mapping.State,
		"StateTransitionReason": mapping.StateTransitionReason,
		"LastModified":          lastModified,
	}

	if mapping.EventSourceArn != "" {
		response["EventSourceArn"] = mapping.EventSourceArn
	}
	if mapping.LastProcessingResult != "" {
		response["LastProcessingResult"] = mapping.LastProcessingResult
	}
	if mapping.BatchSize != nil {
		response["BatchSize"] = *mapping.BatchSize
	}
	if mapping.MaximumBatchingWindowInSeconds != nil {
		response["MaximumBatchingWindowInSeconds"] = *mapping.MaximumBatchingWindowInSeconds
	}
	if mapping.ParallelizationFactor != nil {
		response["ParallelizationFactor"] = *mapping.ParallelizationFactor
	}
	if mapping.StartingPosition != "" {
		response["StartingPosition"] = mapping.StartingPosition
	}
	if mapping.StartingPositionTimestamp != "" {
		response["StartingPositionTimestamp"] = mapping.StartingPositionTimestamp
	}
	if mapping.MaximumRecordAgeInSeconds != nil {
		response["MaximumRecordAgeInSeconds"] = *mapping.MaximumRecordAgeInSeconds
	}
	if mapping.BisectBatchOnFunctionError != nil {
		response["BisectBatchOnFunctionError"] = *mapping.BisectBatchOnFunctionError
	}
	if mapping.MaximumRetryAttempts != nil {
		response["MaximumRetryAttempts"] = *mapping.MaximumRetryAttempts
	}
	if mapping.TumblingWindowInSeconds != nil {
		response["TumblingWindowInSeconds"] = *mapping.TumblingWindowInSeconds
	}
	if mapping.FilterCriteria != nil && len(mapping.FilterCriteria.Filters) > 0 {
		response["FilterCriteria"] = mapping.FilterCriteria
	}
	if mapping.DestinationConfig != nil {
		response["DestinationConfig"] = mapping.DestinationConfig
	}
	if len(mapping.Queues) > 0 {
		response["Queues"] = mapping.Queues
	}
	if len(mapping.SourceAccessConfigurations) > 0 {
		response["SourceAccessConfigurations"] = mapping.SourceAccessConfigurations
	}
	if mapping.SelfManagedEventSource != nil {
		response["SelfManagedEventSource"] = mapping.SelfManagedEventSource
	}
	if len(mapping.FunctionResponseTypes) > 0 {
		response["FunctionResponseTypes"] = mapping.FunctionResponseTypes
	}
	if mapping.AmazonManagedKafkaEventSourceConfig != nil {
		response["AmazonManagedKafkaEventSourceConfig"] = mapping.AmazonManagedKafkaEventSourceConfig
	}
	if mapping.SelfManagedKafkaEventSourceConfig != nil {
		response["SelfManagedKafkaEventSourceConfig"] = mapping.SelfManagedKafkaEventSourceConfig
	}
	if mapping.ScalingConfig != nil {
		response["ScalingConfig"] = mapping.ScalingConfig
	}
	if mapping.DocumentDBEventSourceConfig != nil {
		response["DocumentDBEventSourceConfig"] = mapping.DocumentDBEventSourceConfig
	}

	return response
}

// isStreamSource checks if the event source ARN is a stream source (Kinesis, DynamoDB)
func isStreamSource(arn string) bool {
	return strings.Contains(arn, ":kinesis:") ||
		strings.Contains(arn, ":dynamodb:") && strings.Contains(arn, "/stream/")
}

// getDefaultBatchSize returns the default batch size for an event source type
func getDefaultBatchSize(eventSourceArn string) int32 {
	if strings.Contains(eventSourceArn, ":sqs:") {
		return 10
	}
	if strings.Contains(eventSourceArn, ":kinesis:") {
		return 100
	}
	if strings.Contains(eventSourceArn, ":dynamodb:") {
		return 100
	}
	if strings.Contains(eventSourceArn, ":kafka:") {
		return 100
	}
	return 10 // Default
}

// parseMaxItems parses the MaxItems query parameter
func parseMaxItems(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, err
	}
	if result < 1 {
		return 1, nil
	}
	if result > 10000 {
		return 10000, nil
	}
	return result, nil
}

// convertSourceAccessConfigs converts input configs to stored configs
func convertSourceAccessConfigs(configs []SourceAccessConfiguration) []SourceAccessConfiguration {
	// Since types are now the same, just return as-is
	return configs
}
