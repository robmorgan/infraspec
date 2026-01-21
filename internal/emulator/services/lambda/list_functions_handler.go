package lambda

import (
	"context"
	"net/http"
	"sort"
	"strconv"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// handleListFunctions handles the ListFunctions API
// GET /2015-03-31/functions
func (s *LambdaService) handleListFunctions(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get optional query parameters
	queryParams := parseQueryParams(req.Path)
	masterRegion := queryParams.Get("MasterRegion")
	functionVersion := queryParams.Get("FunctionVersion")
	marker := queryParams.Get("Marker")
	maxItemsStr := queryParams.Get("MaxItems")

	// Parse MaxItems (default 50, max 10000)
	maxItems := 50
	if maxItemsStr != "" {
		if parsed, err := strconv.Atoi(maxItemsStr); err == nil && parsed > 0 {
			if parsed > 10000 {
				maxItems = 10000
			} else {
				maxItems = parsed
			}
		}
	}

	// List all functions from state
	keys, err := s.state.List("lambda:functions:")
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to list functions"), nil
	}

	// Load all functions
	var functions []StoredFunction
	for _, key := range keys {
		var fn StoredFunction
		if err := s.state.Get(key, &fn); err == nil {
			// Apply filters
			if masterRegion != "" && masterRegion != "ALL" {
				// Filter by master region (for Lambda@Edge)
				// For mock, we don't really have regions, so skip this filter
			}
			functions = append(functions, fn)
		}
	}

	// Sort functions by name for consistent ordering
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].FunctionName < functions[j].FunctionName
	})

	// Apply pagination
	startIndex := 0
	if marker != "" {
		// Find the position after the marker
		for i, fn := range functions {
			if fn.FunctionName == marker {
				startIndex = i + 1
				break
			}
		}
	}

	// Slice to get the page
	var pageFunctions []StoredFunction
	var nextMarker string
	if startIndex < len(functions) {
		endIndex := startIndex + maxItems
		if endIndex > len(functions) {
			endIndex = len(functions)
		} else {
			// There are more functions, set next marker
			nextMarker = functions[endIndex-1].FunctionName
		}
		pageFunctions = functions[startIndex:endIndex]
	}

	// Build response
	functionConfigs := make([]map[string]interface{}, len(pageFunctions))
	for i, fn := range pageFunctions {
		// Check if we should return a specific version (ALL) or just $LATEST
		if functionVersion == "ALL" {
			// Include published versions
			// TODO: Full implementation would flatten all versions into the list
			_ = s.buildAllVersionConfigurations(&fn)
			// For simplicity, just add the $LATEST for now
			functionConfigs[i] = s.buildFunctionConfigurationResponse(&fn)
		} else {
			functionConfigs[i] = s.buildFunctionConfigurationResponse(&fn)
		}
	}

	response := map[string]interface{}{
		"Functions": functionConfigs,
	}

	if nextMarker != "" {
		response["NextMarker"] = nextMarker
	}

	return s.successResponse(http.StatusOK, response)
}

// handleListVersionsByFunction handles the ListVersionsByFunction API
// GET /2015-03-31/functions/{FunctionName}/versions
func (s *LambdaService) handleListVersionsByFunction(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get optional query parameters
	queryParams := parseQueryParams(req.Path)
	marker := queryParams.Get("Marker")
	maxItemsStr := queryParams.Get("MaxItems")

	// Parse MaxItems
	maxItems := 50
	if maxItemsStr != "" {
		if parsed, err := strconv.Atoi(maxItemsStr); err == nil && parsed > 0 {
			if parsed > 10000 {
				maxItems = 10000
			} else {
				maxItems = parsed
			}
		}
	}

	// Load the function
	stateKey := "lambda:functions:" + functionName
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"Function not found: "+functionName), nil
	}

	// Collect all versions (including $LATEST)
	type versionEntry struct {
		version string
		config  map[string]interface{}
	}
	var versions []versionEntry

	// Add $LATEST
	latestConfig := s.buildFunctionConfigurationResponse(&function)
	versions = append(versions, versionEntry{
		version: "$LATEST",
		config:  latestConfig,
	})

	// Add published versions
	for versionNum, ver := range function.PublishedVersions {
		config := s.buildVersionConfigurationResponse(&function, ver)
		versions = append(versions, versionEntry{
			version: versionNum,
			config:  config,
		})
	}

	// Sort by version (numeric comparison for numbers, $LATEST first)
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].version == "$LATEST" {
			return true
		}
		if versions[j].version == "$LATEST" {
			return false
		}
		// Try numeric comparison
		vi, erri := strconv.Atoi(versions[i].version)
		vj, errj := strconv.Atoi(versions[j].version)
		if erri == nil && errj == nil {
			return vi < vj
		}
		return versions[i].version < versions[j].version
	})

	// Apply pagination
	startIndex := 0
	if marker != "" {
		for i, v := range versions {
			if v.version == marker {
				startIndex = i + 1
				break
			}
		}
	}

	var pageVersions []versionEntry
	var nextMarker string
	if startIndex < len(versions) {
		endIndex := startIndex + maxItems
		if endIndex > len(versions) {
			endIndex = len(versions)
		} else {
			nextMarker = versions[endIndex-1].version
		}
		pageVersions = versions[startIndex:endIndex]
	}

	// Build response
	versionConfigs := make([]map[string]interface{}, len(pageVersions))
	for i, v := range pageVersions {
		versionConfigs[i] = v.config
	}

	response := map[string]interface{}{
		"Versions": versionConfigs,
	}

	if nextMarker != "" {
		response["NextMarker"] = nextMarker
	}

	return s.successResponse(http.StatusOK, response)
}

// buildAllVersionConfigurations builds configurations for all versions including $LATEST
func (s *LambdaService) buildAllVersionConfigurations(fn *StoredFunction) []map[string]interface{} {
	configs := []map[string]interface{}{
		s.buildFunctionConfigurationResponse(fn),
	}

	// Sort version numbers
	var versionNums []string
	for v := range fn.PublishedVersions {
		versionNums = append(versionNums, v)
	}
	sort.Slice(versionNums, func(i, j int) bool {
		vi, _ := strconv.Atoi(versionNums[i])
		vj, _ := strconv.Atoi(versionNums[j])
		return vi < vj
	})

	for _, vNum := range versionNums {
		if ver, ok := fn.PublishedVersions[vNum]; ok {
			configs = append(configs, s.buildVersionConfigurationResponse(fn, ver))
		}
	}

	return configs
}

// Note: parseFunctionName is defined in invoke_handler.go
