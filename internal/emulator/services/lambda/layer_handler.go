package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// handlePublishLayerVersion handles PublishLayerVersion API
// POST /2018-10-31/layers/{LayerName}/versions
func (s *LambdaService) handlePublishLayerVersion(ctx context.Context, layerName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input PublishLayerVersionInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	// Validate layer name
	if err := validateLayerName(layerName); err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException", err.Error()), nil
	}

	// Validate content
	if input.Content == nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"Content is required"), nil
	}
	if input.Content.ZipFile == "" && (input.Content.S3Bucket == "" || input.Content.S3Key == "") {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"Either ZipFile or S3Bucket/S3Key is required"), nil
	}

	// Load or create layer
	layerKey := fmt.Sprintf("lambda:layers:%s", layerName)
	var layer StoredLayer
	if err := s.state.Get(layerKey, &layer); err != nil {
		// Create new layer
		layer = StoredLayer{
			LayerName:           layerName,
			LayerArn:            generateLayerBaseArn(layerName),
			LatestVersionNumber: 0,
			Versions:            make(map[int64]*StoredLayerVersion),
		}
	}

	// Increment version
	layer.LatestVersionNumber++
	version := layer.LatestVersionNumber

	// Create layer version
	layerVersion := &StoredLayerVersion{
		LayerVersionArn:         generateLayerArn(layerName, version),
		Version:                 version,
		Description:             input.Description,
		CreatedDate:             now(),
		CompatibleRuntimes:      input.CompatibleRuntimes,
		CompatibleArchitectures: input.CompatibleArchitectures,
		LicenseInfo:             input.LicenseInfo,
		CodeSha256:              generateLayerCodeSha256(input.Content),
		CodeSize:                estimateLayerCodeSize(input.Content),
		Content:                 input.Content,
	}

	// Store version
	layer.Versions[version] = layerVersion

	// Save layer
	if err := s.state.Set(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to publish layer version"), nil
	}

	response := s.buildLayerVersionResponse(layerVersion, layerName)
	return s.successResponse(http.StatusCreated, response)
}

// handleGetLayerVersion handles GetLayerVersion API
// GET /2018-10-31/layers/{LayerName}/versions/{VersionNumber}
func (s *LambdaService) handleGetLayerVersion(ctx context.Context, layerName string, versionNumber int64, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	layerKey := fmt.Sprintf("lambda:layers:%s", layerName)
	var layer StoredLayer
	if err := s.state.Get(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer not found: %s", layerName)), nil
	}

	version, exists := layer.Versions[versionNumber]
	if !exists {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer version not found: %s:%d", layerName, versionNumber)), nil
	}

	response := s.buildLayerVersionDetailResponse(version, layerName)
	return s.successResponse(http.StatusOK, response)
}

// handleGetLayerVersionByArn handles GetLayerVersionByArn API
// GET /2018-10-31/layers?Arn={Arn}
func (s *LambdaService) handleGetLayerVersionByArn(ctx context.Context, arn string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Parse ARN to extract layer name and version
	layerName, versionNumber, err := parseLayerVersionArn(arn)
	if err != nil {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			fmt.Sprintf("Invalid layer ARN: %s", arn)), nil
	}

	return s.handleGetLayerVersion(ctx, layerName, versionNumber, req)
}

// handleDeleteLayerVersion handles DeleteLayerVersion API
// DELETE /2018-10-31/layers/{LayerName}/versions/{VersionNumber}
func (s *LambdaService) handleDeleteLayerVersion(ctx context.Context, layerName string, versionNumber int64, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	layerKey := fmt.Sprintf("lambda:layers:%s", layerName)
	var layer StoredLayer
	if err := s.state.Get(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer not found: %s", layerName)), nil
	}

	if _, exists := layer.Versions[versionNumber]; !exists {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer version not found: %s:%d", layerName, versionNumber)), nil
	}

	// Delete the version
	delete(layer.Versions, versionNumber)

	// If no versions left, delete the layer entirely
	if len(layer.Versions) == 0 {
		if err := s.state.Delete(layerKey); err != nil {
			return s.errorResponse(http.StatusInternalServerError, "ServiceException",
				"Failed to delete layer"), nil
		}
	} else {
		// Save updated layer
		if err := s.state.Set(layerKey, &layer); err != nil {
			return s.errorResponse(http.StatusInternalServerError, "ServiceException",
				"Failed to update layer"), nil
		}
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       nil,
	}, nil
}

// handleListLayers handles ListLayers API
// GET /2018-10-31/layers
func (s *LambdaService) handleListLayers(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	queryParams := parseQueryParams(req.Path)
	compatibleRuntime := queryParams.Get("CompatibleRuntime")
	compatibleArchitecture := queryParams.Get("CompatibleArchitecture")
	marker := queryParams.Get("Marker")
	maxItemsStr := queryParams.Get("MaxItems")

	maxItems := 50
	if maxItemsStr != "" {
		if parsed, err := strconv.Atoi(maxItemsStr); err == nil && parsed > 0 {
			if parsed > 50 {
				maxItems = 50
			} else {
				maxItems = parsed
			}
		}
	}

	// List all layers
	prefix := "lambda:layers:"
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to list layers"), nil
	}

	// Load layers and get latest versions
	type layerSummary struct {
		LayerName     string
		LayerArn      string
		LatestVersion *StoredLayerVersion
	}
	var layers []layerSummary

	for _, key := range keys {
		var layer StoredLayer
		if err := s.state.Get(key, &layer); err == nil {
			// Find latest matching version
			var latestVersion *StoredLayerVersion
			for _, v := range layer.Versions {
				// Apply filters
				if compatibleRuntime != "" && !containsString(v.CompatibleRuntimes, compatibleRuntime) {
					continue
				}
				if compatibleArchitecture != "" && !containsString(v.CompatibleArchitectures, compatibleArchitecture) {
					continue
				}
				if latestVersion == nil || v.Version > latestVersion.Version {
					latestVersion = v
				}
			}
			if latestVersion != nil {
				layers = append(layers, layerSummary{
					LayerName:     layer.LayerName,
					LayerArn:      layer.LayerArn,
					LatestVersion: latestVersion,
				})
			}
		}
	}

	// Sort by name
	sort.Slice(layers, func(i, j int) bool {
		return layers[i].LayerName < layers[j].LayerName
	})

	// Apply pagination
	startIndex := 0
	if marker != "" {
		for i, l := range layers {
			if l.LayerName == marker {
				startIndex = i + 1
				break
			}
		}
	}

	var pageLayers []layerSummary
	var nextMarker string
	if startIndex < len(layers) {
		endIndex := startIndex + maxItems
		if endIndex > len(layers) {
			endIndex = len(layers)
		} else {
			nextMarker = layers[endIndex-1].LayerName
		}
		pageLayers = layers[startIndex:endIndex]
	}

	// Build response
	layerResponses := make([]map[string]interface{}, len(pageLayers))
	for i, l := range pageLayers {
		layerResponses[i] = map[string]interface{}{
			"LayerName":             l.LayerName,
			"LayerArn":              l.LayerArn,
			"LatestMatchingVersion": s.buildLayerVersionSummary(l.LatestVersion),
		}
	}

	response := map[string]interface{}{
		"Layers": layerResponses,
	}
	if nextMarker != "" {
		response["NextMarker"] = nextMarker
	}

	return s.successResponse(http.StatusOK, response)
}

// handleListLayerVersions handles ListLayerVersions API
// GET /2018-10-31/layers/{LayerName}/versions
func (s *LambdaService) handleListLayerVersions(ctx context.Context, layerName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	queryParams := parseQueryParams(req.Path)
	compatibleRuntime := queryParams.Get("CompatibleRuntime")
	compatibleArchitecture := queryParams.Get("CompatibleArchitecture")
	marker := queryParams.Get("Marker")
	maxItemsStr := queryParams.Get("MaxItems")

	maxItems := 50
	if maxItemsStr != "" {
		if parsed, err := strconv.Atoi(maxItemsStr); err == nil && parsed > 0 {
			if parsed > 50 {
				maxItems = 50
			} else {
				maxItems = parsed
			}
		}
	}

	layerKey := fmt.Sprintf("lambda:layers:%s", layerName)
	var layer StoredLayer
	if err := s.state.Get(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer not found: %s", layerName)), nil
	}

	// Collect and filter versions
	var versions []*StoredLayerVersion
	for _, v := range layer.Versions {
		// Apply filters
		if compatibleRuntime != "" && !containsString(v.CompatibleRuntimes, compatibleRuntime) {
			continue
		}
		if compatibleArchitecture != "" && !containsString(v.CompatibleArchitectures, compatibleArchitecture) {
			continue
		}
		versions = append(versions, v)
	}

	// Sort by version descending (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	// Apply pagination
	startIndex := 0
	if marker != "" {
		markerVersion, _ := strconv.ParseInt(marker, 10, 64)
		for i, v := range versions {
			if v.Version == markerVersion {
				startIndex = i + 1
				break
			}
		}
	}

	var pageVersions []*StoredLayerVersion
	var nextMarker string
	if startIndex < len(versions) {
		endIndex := startIndex + maxItems
		if endIndex > len(versions) {
			endIndex = len(versions)
		} else {
			nextMarker = strconv.FormatInt(versions[endIndex-1].Version, 10)
		}
		pageVersions = versions[startIndex:endIndex]
	}

	// Build response
	versionResponses := make([]map[string]interface{}, len(pageVersions))
	for i, v := range pageVersions {
		versionResponses[i] = s.buildLayerVersionSummary(v)
	}

	response := map[string]interface{}{
		"LayerVersions": versionResponses,
	}
	if nextMarker != "" {
		response["NextMarker"] = nextMarker
	}

	return s.successResponse(http.StatusOK, response)
}

// handleAddLayerVersionPermission handles AddLayerVersionPermission API
// POST /2018-10-31/layers/{LayerName}/versions/{VersionNumber}/policy
func (s *LambdaService) handleAddLayerVersionPermission(ctx context.Context, layerName string, versionNumber int64, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	var input AddLayerVersionPermissionInput
	if err := s.parseJSONBody(req, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidRequestContentException",
			fmt.Sprintf("Could not parse request body: %v", err)), nil
	}

	// Validate required fields
	if input.StatementId == "" {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"StatementId is required"), nil
	}
	if input.Action == "" {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"Action is required"), nil
	}
	if input.Principal == "" {
		return s.errorResponse(http.StatusBadRequest, "ValidationException",
			"Principal is required"), nil
	}

	layerKey := fmt.Sprintf("lambda:layers:%s", layerName)
	var layer StoredLayer
	if err := s.state.Get(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer not found: %s", layerName)), nil
	}

	version, exists := layer.Versions[versionNumber]
	if !exists {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer version not found: %s:%d", layerName, versionNumber)), nil
	}

	// Build policy statement
	statement := map[string]interface{}{
		"Sid":       input.StatementId,
		"Effect":    "Allow",
		"Principal": input.Principal,
		"Action":    input.Action,
		"Resource":  version.LayerVersionArn,
	}
	if input.OrganizationId != "" {
		statement["Condition"] = map[string]interface{}{
			"StringEquals": map[string]string{
				"aws:PrincipalOrgID": input.OrganizationId,
			},
		}
	}

	// Parse or create policy
	var policy map[string]interface{}
	if version.Policy != "" {
		json.Unmarshal([]byte(version.Policy), &policy)
	} else {
		policy = map[string]interface{}{
			"Version":   "2012-10-17",
			"Id":        "default",
			"Statement": []interface{}{},
		}
	}

	// Add statement
	statements := policy["Statement"].([]interface{})
	statements = append(statements, statement)
	policy["Statement"] = statements

	// Save policy
	policyBytes, _ := json.Marshal(policy)
	version.Policy = string(policyBytes)
	layer.Versions[versionNumber] = version

	if err := s.state.Set(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to add permission"), nil
	}

	response := map[string]interface{}{
		"Statement":  string(policyBytes),
		"RevisionId": generateRevisionId(),
	}
	return s.successResponse(http.StatusCreated, response)
}

// handleRemoveLayerVersionPermission handles RemoveLayerVersionPermission API
// DELETE /2018-10-31/layers/{LayerName}/versions/{VersionNumber}/policy/{StatementId}
func (s *LambdaService) handleRemoveLayerVersionPermission(ctx context.Context, layerName string, versionNumber int64, statementId string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	layerKey := fmt.Sprintf("lambda:layers:%s", layerName)
	var layer StoredLayer
	if err := s.state.Get(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer not found: %s", layerName)), nil
	}

	version, exists := layer.Versions[versionNumber]
	if !exists {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer version not found: %s:%d", layerName, versionNumber)), nil
	}

	if version.Policy == "" {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"No policy associated with this layer version"), nil
	}

	// Parse policy and remove statement
	var policy map[string]interface{}
	json.Unmarshal([]byte(version.Policy), &policy)

	statements := policy["Statement"].([]interface{})
	found := false
	newStatements := make([]interface{}, 0)
	for _, stmt := range statements {
		s := stmt.(map[string]interface{})
		if s["Sid"] == statementId {
			found = true
			continue
		}
		newStatements = append(newStatements, stmt)
	}

	if !found {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Statement not found: %s", statementId)), nil
	}

	policy["Statement"] = newStatements
	policyBytes, _ := json.Marshal(policy)
	version.Policy = string(policyBytes)
	layer.Versions[versionNumber] = version

	if err := s.state.Set(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to remove permission"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       nil,
	}, nil
}

// handleGetLayerVersionPolicy handles GetLayerVersionPolicy API
// GET /2018-10-31/layers/{LayerName}/versions/{VersionNumber}/policy
func (s *LambdaService) handleGetLayerVersionPolicy(ctx context.Context, layerName string, versionNumber int64, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	layerKey := fmt.Sprintf("lambda:layers:%s", layerName)
	var layer StoredLayer
	if err := s.state.Get(layerKey, &layer); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer not found: %s", layerName)), nil
	}

	version, exists := layer.Versions[versionNumber]
	if !exists {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Layer version not found: %s:%d", layerName, versionNumber)), nil
	}

	if version.Policy == "" {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"No policy associated with this layer version"), nil
	}

	response := map[string]interface{}{
		"Policy":     version.Policy,
		"RevisionId": generateRevisionId(),
	}
	return s.successResponse(http.StatusOK, response)
}

// Response builders

func (s *LambdaService) buildLayerVersionResponse(v *StoredLayerVersion, layerName string) map[string]interface{} {
	response := map[string]interface{}{
		"LayerArn":        generateLayerBaseArn(layerName),
		"LayerVersionArn": v.LayerVersionArn,
		"Version":         v.Version,
		"CreatedDate":     v.CreatedDate,
		"Content": map[string]interface{}{
			"CodeSha256": v.CodeSha256,
			"CodeSize":   v.CodeSize,
		},
	}
	if v.Description != "" {
		response["Description"] = v.Description
	}
	if len(v.CompatibleRuntimes) > 0 {
		response["CompatibleRuntimes"] = v.CompatibleRuntimes
	}
	if len(v.CompatibleArchitectures) > 0 {
		response["CompatibleArchitectures"] = v.CompatibleArchitectures
	}
	if v.LicenseInfo != "" {
		response["LicenseInfo"] = v.LicenseInfo
	}
	return response
}

func (s *LambdaService) buildLayerVersionDetailResponse(v *StoredLayerVersion, layerName string) map[string]interface{} {
	response := s.buildLayerVersionResponse(v, layerName)
	// Add location info for GetLayerVersion
	content := response["Content"].(map[string]interface{})
	content["Location"] = fmt.Sprintf("https://awslambda-us-east-1-layers.s3.amazonaws.com/snapshots/%s/%d", layerName, v.Version)
	return response
}

func (s *LambdaService) buildLayerVersionSummary(v *StoredLayerVersion) map[string]interface{} {
	response := map[string]interface{}{
		"LayerVersionArn": v.LayerVersionArn,
		"Version":         v.Version,
		"CreatedDate":     v.CreatedDate,
	}
	if v.Description != "" {
		response["Description"] = v.Description
	}
	if len(v.CompatibleRuntimes) > 0 {
		response["CompatibleRuntimes"] = v.CompatibleRuntimes
	}
	if len(v.CompatibleArchitectures) > 0 {
		response["CompatibleArchitectures"] = v.CompatibleArchitectures
	}
	if v.LicenseInfo != "" {
		response["LicenseInfo"] = v.LicenseInfo
	}
	return response
}

// Helper functions

func generateLayerBaseArn(layerName string) string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:layer:%s",
		DefaultRegion, DefaultAccountID, layerName)
}

func validateLayerName(name string) error {
	if name == "" {
		return fmt.Errorf("layer name is required")
	}
	if len(name) > 140 {
		return fmt.Errorf("layer name must be 140 characters or fewer")
	}
	return nil
}

func generateLayerCodeSha256(content *LayerContent) string {
	if content == nil {
		return ""
	}
	code := &FunctionCode{
		ZipFile:         content.ZipFile,
		S3Bucket:        content.S3Bucket,
		S3Key:           content.S3Key,
		S3ObjectVersion: content.S3ObjectVersion,
	}
	return generateCodeSha256(code)
}

func estimateLayerCodeSize(content *LayerContent) int64 {
	if content == nil {
		return 0
	}
	code := &FunctionCode{
		ZipFile:         content.ZipFile,
		S3Bucket:        content.S3Bucket,
		S3Key:           content.S3Key,
		S3ObjectVersion: content.S3ObjectVersion,
	}
	return estimateCodeSize(code)
}

func parseLayerVersionArn(arn string) (string, int64, error) {
	// ARN format: arn:aws:lambda:region:account:layer:name:version
	parts := splitArn(arn)
	if len(parts) < 8 || parts[5] != "layer" {
		return "", 0, fmt.Errorf("invalid layer ARN format")
	}
	layerName := parts[6]
	version, err := strconv.ParseInt(parts[7], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid version number in ARN")
	}
	return layerName, version, nil
}

func splitArn(arn string) []string {
	var parts []string
	current := ""
	for _, c := range arn {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
