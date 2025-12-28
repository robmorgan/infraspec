package lambda

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// Mock response configuration via function tags:
//
//	mock:response     - Custom JSON response body (base64 encoded)
//	mock:statusCode   - HTTP status code for mock response (default: 200)
//	mock:error        - Simulate function error (Handled, Unhandled)
//	mock:errorMessage - Custom error message
//	mock:echo         - If "true", echo the request payload as response
//
// Environment variable configuration:
//
//	MOCK_RESPONSE     - Default mock response for all functions
//	MOCK_STATUS_CODE  - Default status code

// handleInvoke handles the Invoke API
// POST /2015-03-31/functions/{FunctionName}/invocations
func (s *LambdaService) handleInvoke(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get invocation type from header
	invocationType := req.Headers["X-Amz-Invocation-Type"]
	if invocationType == "" {
		invocationType = InvocationTypeRequestResponse
	}

	// Get log type from header
	logType := req.Headers["X-Amz-Log-Type"]

	// Get client context if provided
	clientContext := req.Headers["X-Amz-Client-Context"]

	// Get qualifier from query params
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")
	if qualifier == "" {
		qualifier = "$LATEST"
	}

	// Parse function name (could be ARN or name)
	functionName = parseFunctionName(functionName)

	// Load the function to verify it exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Handle qualifier (version or alias)
	executedVersion := "$LATEST"
	if qualifier != "" && qualifier != "$LATEST" {
		// Check if it's a version
		if _, exists := function.PublishedVersions[qualifier]; exists {
			executedVersion = qualifier
		} else {
			// Check if it's an alias
			aliasKey := fmt.Sprintf("lambda:aliases:%s:%s", functionName, qualifier)
			var alias StoredAlias
			if err := s.state.Get(aliasKey, &alias); err != nil {
				return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
					fmt.Sprintf("Function version or alias not found: %s:%s", functionName, qualifier)), nil
			}
			executedVersion = alias.FunctionVersion
		}
	}

	// Check function state
	if function.State != StateActive {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			fmt.Sprintf("The function is in %s state. It must be in Active state to be invoked.", function.State)), nil
	}

	// For DryRun, just return 204 to indicate the function can be invoked
	if invocationType == InvocationTypeDryRun {
		return &emulator.AWSResponse{
			StatusCode: http.StatusNoContent,
			Headers: map[string]string{
				"Content-Type":           "application/json",
				"X-Amz-Executed-Version": executedVersion,
			},
			Body: nil,
		}, nil
	}

	// For Event (async), return 202 Accepted immediately
	if invocationType == InvocationTypeEvent {
		return &emulator.AWSResponse{
			StatusCode: http.StatusAccepted,
			Headers: map[string]string{
				"Content-Type":           "application/json",
				"X-Amz-Executed-Version": executedVersion,
			},
			Body: nil,
		}, nil
	}

	// For RequestResponse, build a mock response
	responseBody, functionError := s.buildMockInvokeResponse(&function, req.Body, clientContext)

	// Build response headers
	headers := map[string]string{
		"Content-Type":           "application/json",
		"X-Amz-Executed-Version": executedVersion,
	}

	// Add function error header if simulating an error
	if functionError != "" {
		headers["X-Amz-Function-Error"] = functionError
	}

	// Add log result if requested
	if logType == "Tail" {
		logResult := s.buildMockLogResult(&function, functionError)
		headers["X-Amz-Log-Result"] = base64.StdEncoding.EncodeToString([]byte(logResult))
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       responseBody,
	}, nil
}

// handleInvokeAsync handles the deprecated InvokeAsync API
// POST /2014-11-13/functions/{FunctionName}/invoke-async
func (s *LambdaService) handleInvokeAsync(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Parse function name
	functionName = parseFunctionName(functionName)

	// Load the function to verify it exists
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Check function state
	if function.State != StateActive {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			fmt.Sprintf("The function is in %s state", function.State)), nil
	}

	// InvokeAsync always returns 202 Accepted
	response := map[string]interface{}{
		"Status": 202,
	}

	body, _ := json.Marshal(response)
	return &emulator.AWSResponse{
		StatusCode: http.StatusAccepted,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}, nil
}

// buildMockInvokeResponse builds the mock response body based on function configuration
func (s *LambdaService) buildMockInvokeResponse(fn *StoredFunction, payload []byte, clientContext string) ([]byte, string) {
	tags := fn.Tags
	if tags == nil {
		tags = make(map[string]string)
	}

	// Check for error simulation
	if errorType, ok := tags["mock:error"]; ok {
		errorMessage := tags["mock:errorMessage"]
		if errorMessage == "" {
			errorMessage = "Simulated Lambda error"
		}

		errorResponse := map[string]interface{}{
			"errorType":    "Error",
			"errorMessage": errorMessage,
		}

		if errorType == "Unhandled" {
			errorResponse["stackTrace"] = []string{
				"at Object.<anonymous> (/var/task/index.js:1:1)",
				"at Module._compile (internal/modules/cjs/loader.js:1085:14)",
			}
		}

		body, _ := json.Marshal(errorResponse)
		return body, errorType
	}

	// Check for echo mode
	if echo, ok := tags["mock:echo"]; ok && strings.ToLower(echo) == "true" {
		if len(payload) > 0 {
			return payload, ""
		}
	}

	// Check for custom response in tags
	if customResponse, ok := tags["mock:response"]; ok {
		// Try base64 decode first
		decoded, err := base64.StdEncoding.DecodeString(customResponse)
		if err == nil {
			return decoded, ""
		}
		// If not base64, use as-is
		return []byte(customResponse), ""
	}

	// Check environment variables for mock response
	if fn.Environment != nil && fn.Environment.Variables != nil {
		if mockResponse, ok := fn.Environment.Variables["MOCK_RESPONSE"]; ok {
			return []byte(mockResponse), ""
		}
	}

	// Default mock response - typical Lambda response format
	defaultResponse := map[string]interface{}{
		"statusCode": 200,
		"headers": map[string]string{
			"Content-Type": "application/json",
		},
		"body": `{"message":"Hello from Lambda!","input":` + string(payload) + `}`,
	}

	// If payload is valid JSON, include it in the response
	if len(payload) > 0 {
		var inputData interface{}
		if json.Unmarshal(payload, &inputData) == nil {
			defaultResponse["body"] = map[string]interface{}{
				"message": "Hello from Lambda!",
				"input":   inputData,
			}
		}
	}

	body, _ := json.Marshal(defaultResponse)
	return body, ""
}

// buildMockLogResult builds a mock CloudWatch log result
func (s *LambdaService) buildMockLogResult(fn *StoredFunction, functionError string) string {
	requestId := generateRevisionId() // Reuse as mock request ID
	duration := "1.23"
	billedDuration := "2"
	memorySize := fmt.Sprintf("%d", fn.MemorySize)
	maxMemoryUsed := "64"

	var logLines []string
	logLines = append(logLines, fmt.Sprintf("START RequestId: %s Version: $LATEST", requestId))

	if functionError != "" {
		logLines = append(logLines, fmt.Sprintf("ERROR: %s error simulated", functionError))
	} else {
		logLines = append(logLines, "INFO: Function executed successfully")
	}

	logLines = append(logLines, fmt.Sprintf("END RequestId: %s", requestId))
	logLines = append(logLines, fmt.Sprintf(
		"REPORT RequestId: %s\tDuration: %s ms\tBilled Duration: %s ms\tMemory Size: %s MB\tMax Memory Used: %s MB",
		requestId, duration, billedDuration, memorySize, maxMemoryUsed,
	))

	return strings.Join(logLines, "\n")
}

// parseFunctionName extracts the function name from a name or ARN
func parseFunctionName(nameOrArn string) string {
	// Handle ARN format: arn:aws:lambda:region:account:function:name[:qualifier]
	if strings.HasPrefix(nameOrArn, "arn:aws:lambda:") {
		parts := strings.Split(nameOrArn, ":")
		if len(parts) >= 7 && parts[5] == "function" {
			return parts[6]
		}
	}
	return nameOrArn
}
