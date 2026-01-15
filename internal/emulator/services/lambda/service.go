package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// LambdaService implements the AWS Lambda API emulator.
// It uses REST-JSON protocol with path-based routing.
type LambdaService struct {
	state           emulator.StateManager
	validator       emulator.Validator
	resourceManager *graph.ResourceManager
}

// NewLambdaService creates a new Lambda service instance.
func NewLambdaService(state emulator.StateManager, validator emulator.Validator) *LambdaService {
	return &LambdaService{
		state:           state,
		validator:       validator,
		resourceManager: nil,
	}
}

// NewLambdaServiceWithGraph creates a new Lambda service instance with graph support.
func NewLambdaServiceWithGraph(state emulator.StateManager, validator emulator.Validator, rm *graph.ResourceManager) *LambdaService {
	return &LambdaService{
		state:           state,
		validator:       validator,
		resourceManager: rm,
	}
}

// ServiceName returns the service identifier.
func (s *LambdaService) ServiceName() string {
	return "lambda"
}

// HandleRequest routes Lambda API requests based on HTTP method and path.
// Lambda uses REST-JSON protocol with path-based routing:
//   - POST   /2015-03-31/functions                              -> CreateFunction
//   - GET    /2015-03-31/functions                              -> ListFunctions
//   - GET    /2015-03-31/functions/{name}                       -> GetFunction
//   - GET    /2015-03-31/functions/{name}/configuration         -> GetFunctionConfiguration
//   - PUT    /2015-03-31/functions/{name}/code                  -> UpdateFunctionCode
//   - PUT    /2015-03-31/functions/{name}/configuration         -> UpdateFunctionConfiguration
//   - DELETE /2015-03-31/functions/{name}                       -> DeleteFunction
//   - GET    /2015-03-31/tags/{arn}                             -> ListTags
//   - POST   /2015-03-31/tags/{arn}                             -> TagResource
//   - DELETE /2015-03-31/tags/{arn}                             -> UntagResource
//   - POST   /2015-03-31/functions/{name}/invocations           -> Invoke
//   - POST   /2014-11-13/functions/{name}/invoke-async           -> InvokeAsync (deprecated)
func (s *LambdaService) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	path := req.Path
	method := req.Method

	// Normalize path - remove leading slash and query string
	path = getPathWithoutQuery(path)
	path = strings.TrimPrefix(path, "/")

	// Handle different Lambda API versions:
	// - 2014-11-13: deprecated (InvokeAsync)
	// - 2015-03-31: main Lambda API (functions, aliases, versions)
	// - 2017-10-31: concurrency API
	// - 2018-10-31: layers API
	// - 2019-09-25: event invoke config API
	// - 2019-09-30: provisioned concurrency API
	// - 2020-06-30: code signing config API
	// - 2021-10-31: function URLs API
	apiVersionPrefixes := []string{
		"2014-11-13/",
		"2017-10-31/",
		"2018-10-31/",
		"2019-09-25/",
		"2019-09-30/",
		"2020-06-30/",
		"2021-10-31/",
		"2015-03-31/", // Default/main version - must be last
	}
	for _, prefix := range apiVersionPrefixes {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			break
		}
	}

	// Parse the path to determine the operation
	parts := strings.Split(path, "/")

	// Route based on path pattern
	switch {
	// Account settings endpoint
	case path == "account-settings":
		if method == http.MethodGet {
			return s.handleGetAccountSettings(ctx, req)
		}
		return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
			fmt.Sprintf("Method %s not allowed on /account-settings", method)), nil

	// Functions endpoints
	case strings.HasPrefix(path, "functions"):
		return s.routeFunctionsEndpoint(ctx, req, method, parts)

	// Tags endpoints
	case strings.HasPrefix(path, "tags"):
		return s.routeTagsEndpoint(ctx, req, method, parts)

	// Event source mappings
	case strings.HasPrefix(path, "event-source-mappings"):
		return s.routeEventSourceMappingsEndpoint(ctx, req, method, parts)

	// Layers endpoints
	case strings.HasPrefix(path, "layers"):
		return s.routeLayersEndpoint(ctx, req, method, parts)

	default:
		return s.errorResponse(http.StatusBadRequest, "InvalidAction",
			fmt.Sprintf("Unknown path: %s", path)), nil
	}
}

// routeFunctionsEndpoint handles /functions/* routes
func (s *LambdaService) routeFunctionsEndpoint(ctx context.Context, req *emulator.AWSRequest, method string, parts []string) (*emulator.AWSResponse, error) {
	// parts[0] = "functions"
	// parts[1] = function name (optional)
	// parts[2] = sub-resource (optional): "configuration", "code", "invocations", etc.

	switch len(parts) {
	case 1: // /functions
		switch method {
		case http.MethodPost:
			return s.handleCreateFunction(ctx, req)
		case http.MethodGet:
			return s.handleListFunctions(ctx, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /functions", method)), nil
		}

	case 2: // /functions/{name}
		functionName := parts[1]
		switch method {
		case http.MethodGet:
			return s.handleGetFunction(ctx, functionName, req)
		case http.MethodDelete:
			return s.handleDeleteFunction(ctx, functionName, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /functions/{name}", method)), nil
		}

	case 3: // /functions/{name}/{sub-resource}
		functionName := parts[1]
		subResource := parts[2]

		switch subResource {
		case "configuration":
			switch method {
			case http.MethodGet:
				return s.handleGetFunctionConfiguration(ctx, functionName, req)
			case http.MethodPut:
				return s.handleUpdateFunctionConfiguration(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/configuration", method)), nil
			}

		case "code":
			switch method {
			case http.MethodPut:
				return s.handleUpdateFunctionCode(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/code", method)), nil
			}

		case "invocations":
			switch method {
			case http.MethodPost:
				return s.handleInvoke(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/invocations", method)), nil
			}

		case "invoke-async":
			// Deprecated InvokeAsync API (2014-11-13 version)
			switch method {
			case http.MethodPost:
				return s.handleInvokeAsync(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/invoke-async", method)), nil
			}

		case "versions":
			switch method {
			case http.MethodGet:
				return s.handleListVersionsByFunction(ctx, functionName, req)
			case http.MethodPost:
				return s.handlePublishVersion(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/versions", method)), nil
			}

		case "aliases":
			switch method {
			case http.MethodGet:
				return s.handleListAliases(ctx, functionName, req)
			case http.MethodPost:
				return s.handleCreateAlias(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/aliases", method)), nil
			}

		case "policy":
			switch method {
			case http.MethodGet:
				return s.handleGetPolicy(ctx, functionName, req)
			case http.MethodPost:
				return s.handleAddPermission(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/policy", method)), nil
			}

		case "concurrency":
			switch method {
			case http.MethodGet:
				return s.handleGetFunctionConcurrency(ctx, functionName, req)
			case http.MethodPut:
				return s.handlePutFunctionConcurrency(ctx, functionName, req)
			case http.MethodDelete:
				return s.handleDeleteFunctionConcurrency(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/concurrency", method)), nil
			}

		case "url":
			switch method {
			case http.MethodGet:
				return s.handleGetFunctionUrlConfig(ctx, functionName, req)
			case http.MethodPost:
				return s.handleCreateFunctionUrlConfig(ctx, functionName, req)
			case http.MethodPut:
				return s.handleUpdateFunctionUrlConfig(ctx, functionName, req)
			case http.MethodDelete:
				return s.handleDeleteFunctionUrlConfig(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/url", method)), nil
			}

		case "provisioned-concurrency":
			// Check for List=ALL query parameter
			queryParams := parseQueryParams(req.Path)
			if queryParams.Get("List") == "ALL" {
				return s.handleListProvisionedConcurrencyConfigs(ctx, functionName, req)
			}
			switch method {
			case http.MethodGet:
				return s.handleGetProvisionedConcurrencyConfig(ctx, functionName, req)
			case http.MethodPut:
				return s.handlePutProvisionedConcurrencyConfig(ctx, functionName, req)
			case http.MethodDelete:
				return s.handleDeleteProvisionedConcurrencyConfig(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/provisioned-concurrency", method)), nil
			}

		case "event-invoke-config":
			switch method {
			case http.MethodGet:
				return s.handleGetFunctionEventInvokeConfig(ctx, functionName, req)
			case http.MethodPut:
				return s.handlePutFunctionEventInvokeConfig(ctx, functionName, req)
			case http.MethodPost:
				return s.handleUpdateFunctionEventInvokeConfig(ctx, functionName, req)
			case http.MethodDelete:
				return s.handleDeleteFunctionEventInvokeConfig(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/event-invoke-config", method)), nil
			}

		case "code-signing-config":
			// Code signing config operations (stub - returns empty config)
			switch method {
			case http.MethodGet:
				return s.handleGetFunctionCodeSigningConfig(ctx, functionName, req)
			case http.MethodPut:
				return s.handlePutFunctionCodeSigningConfig(ctx, functionName, req)
			case http.MethodDelete:
				return s.handleDeleteFunctionCodeSigningConfig(ctx, functionName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/code-signing-config", method)), nil
			}

		default:
			return s.errorResponse(http.StatusBadRequest, "InvalidAction",
				fmt.Sprintf("Unknown sub-resource: %s", subResource)), nil
		}

	case 4: // /functions/{name}/aliases/{aliasName} or /functions/{name}/policy/{statementId}
		functionName := parts[1]
		subResource := parts[2]
		switch subResource {
		case "aliases":
			aliasName := parts[3]
			switch method {
			case http.MethodGet:
				return s.handleGetAlias(ctx, functionName, aliasName, req)
			case http.MethodPut:
				return s.handleUpdateAlias(ctx, functionName, aliasName, req)
			case http.MethodDelete:
				return s.handleDeleteAlias(ctx, functionName, aliasName, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/aliases/{alias}", method)), nil
			}
		case "policy":
			statementId := parts[3]
			switch method {
			case http.MethodDelete:
				return s.handleRemovePermission(ctx, functionName, statementId, req)
			default:
				return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
					fmt.Sprintf("Method %s not allowed on /functions/{name}/policy/{statementId}", method)), nil
			}
		case "event-invoke-config":
			// Handle /functions/{name}/event-invoke-config/list
			if parts[3] == "list" && method == http.MethodGet {
				return s.handleListFunctionEventInvokeConfigs(ctx, functionName, req)
			}
			return s.errorResponse(http.StatusBadRequest, "InvalidAction",
				fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
		default:
			return s.errorResponse(http.StatusBadRequest, "InvalidAction",
				fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
		}

	default:
		return s.errorResponse(http.StatusBadRequest, "InvalidAction",
			fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
	}
}

// routeTagsEndpoint handles /tags/* routes
func (s *LambdaService) routeTagsEndpoint(ctx context.Context, req *emulator.AWSRequest, method string, parts []string) (*emulator.AWSResponse, error) {
	// parts[0] = "tags"
	// parts[1+] = ARN (URL-encoded, may contain slashes)

	if len(parts) < 2 {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValue",
			"Resource ARN is required"), nil
	}

	// Reconstruct the ARN from remaining parts
	arn := strings.Join(parts[1:], "/")

	switch method {
	case http.MethodGet:
		return s.handleListTags(ctx, arn, req)
	case http.MethodPost:
		return s.handleTagResource(ctx, arn, req)
	case http.MethodDelete:
		return s.handleUntagResource(ctx, arn, req)
	default:
		return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
			fmt.Sprintf("Method %s not allowed on /tags", method)), nil
	}
}

// routeEventSourceMappingsEndpoint handles /event-source-mappings/* routes
func (s *LambdaService) routeEventSourceMappingsEndpoint(ctx context.Context, req *emulator.AWSRequest, method string, parts []string) (*emulator.AWSResponse, error) {
	// parts[0] = "event-source-mappings"
	// parts[1] = UUID (optional)

	switch len(parts) {
	case 1: // /event-source-mappings
		switch method {
		case http.MethodPost:
			return s.handleCreateEventSourceMapping(ctx, req)
		case http.MethodGet:
			return s.handleListEventSourceMappings(ctx, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /event-source-mappings", method)), nil
		}

	case 2: // /event-source-mappings/{UUID}
		mappingUUID := parts[1]
		switch method {
		case http.MethodGet:
			return s.handleGetEventSourceMapping(ctx, mappingUUID, req)
		case http.MethodPut:
			return s.handleUpdateEventSourceMapping(ctx, mappingUUID, req)
		case http.MethodDelete:
			return s.handleDeleteEventSourceMapping(ctx, mappingUUID, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /event-source-mappings/{UUID}", method)), nil
		}

	default:
		return s.errorResponse(http.StatusBadRequest, "InvalidAction",
			fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
	}
}

// routeLayersEndpoint handles /layers/* routes
func (s *LambdaService) routeLayersEndpoint(ctx context.Context, req *emulator.AWSRequest, method string, parts []string) (*emulator.AWSResponse, error) {
	// parts[0] = "layers"
	// parts[1] = layer name (optional)
	// parts[2] = "versions" (optional)
	// parts[3] = version number (optional)
	// parts[4] = "policy" (optional)
	// parts[5] = statement id (optional)

	// Handle GET /layers?Arn=... (GetLayerVersionByArn)
	if len(parts) == 1 && method == http.MethodGet {
		queryParams := parseQueryParams(req.Path)
		if arn := queryParams.Get("Arn"); arn != "" {
			return s.handleGetLayerVersionByArn(ctx, arn, req)
		}
		// Otherwise list layers
		return s.handleListLayers(ctx, req)
	}

	switch len(parts) {
	case 1: // /layers
		// GET handled above with Arn check
		return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
			fmt.Sprintf("Method %s not allowed on /layers", method)), nil

	case 2: // /layers/{LayerName} - not a valid endpoint by itself
		return s.errorResponse(http.StatusBadRequest, "InvalidAction",
			"Layer name must be followed by /versions"), nil

	case 3: // /layers/{LayerName}/versions
		layerName := parts[1]
		if parts[2] != "versions" {
			return s.errorResponse(http.StatusBadRequest, "InvalidAction",
				fmt.Sprintf("Unknown sub-resource: %s", parts[2])), nil
		}
		switch method {
		case http.MethodPost:
			return s.handlePublishLayerVersion(ctx, layerName, req)
		case http.MethodGet:
			return s.handleListLayerVersions(ctx, layerName, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /layers/{name}/versions", method)), nil
		}

	case 4: // /layers/{LayerName}/versions/{VersionNumber}
		layerName := parts[1]
		if parts[2] != "versions" {
			return s.errorResponse(http.StatusBadRequest, "InvalidAction",
				fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
		}
		versionNumber, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException",
				"Invalid version number"), nil
		}
		switch method {
		case http.MethodGet:
			return s.handleGetLayerVersion(ctx, layerName, versionNumber, req)
		case http.MethodDelete:
			return s.handleDeleteLayerVersion(ctx, layerName, versionNumber, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /layers/{name}/versions/{version}", method)), nil
		}

	case 5: // /layers/{LayerName}/versions/{VersionNumber}/policy
		layerName := parts[1]
		if parts[2] != "versions" || parts[4] != "policy" {
			return s.errorResponse(http.StatusBadRequest, "InvalidAction",
				fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
		}
		versionNumber, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException",
				"Invalid version number"), nil
		}
		switch method {
		case http.MethodGet:
			return s.handleGetLayerVersionPolicy(ctx, layerName, versionNumber, req)
		case http.MethodPost:
			return s.handleAddLayerVersionPermission(ctx, layerName, versionNumber, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /layers/{name}/versions/{version}/policy", method)), nil
		}

	case 6: // /layers/{LayerName}/versions/{VersionNumber}/policy/{StatementId}
		layerName := parts[1]
		if parts[2] != "versions" || parts[4] != "policy" {
			return s.errorResponse(http.StatusBadRequest, "InvalidAction",
				fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
		}
		versionNumber, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return s.errorResponse(http.StatusBadRequest, "ValidationException",
				"Invalid version number"), nil
		}
		statementId := parts[5]
		switch method {
		case http.MethodDelete:
			return s.handleRemoveLayerVersionPermission(ctx, layerName, versionNumber, statementId, req)
		default:
			return s.errorResponse(http.StatusMethodNotAllowed, "MethodNotAllowed",
				fmt.Sprintf("Method %s not allowed on /layers/{name}/versions/{version}/policy/{sid}", method)), nil
		}

	default:
		return s.errorResponse(http.StatusBadRequest, "InvalidAction",
			fmt.Sprintf("Unknown path: %s", strings.Join(parts, "/"))), nil
	}
}

// parseJSONBody parses the request body as JSON into the target struct
func (s *LambdaService) parseJSONBody(req *emulator.AWSRequest, target interface{}) error {
	if len(req.Body) == 0 {
		return nil
	}
	return json.Unmarshal(req.Body, target)
}

// successResponse builds a successful JSON response
func (s *LambdaService) successResponse(statusCode int, data interface{}) (*emulator.AWSResponse, error) {
	return emulator.BuildRESTJSONResponse(statusCode, data)
}

// errorResponse builds an error JSON response
func (s *LambdaService) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	return emulator.BuildRESTJSONErrorResponse(statusCode, code, message)
}
