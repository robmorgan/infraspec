package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *IAMService) createRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	assumeRolePolicyDocument := getStringValue(params, "AssumeRolePolicyDocument")
	if assumeRolePolicyDocument == "" {
		return s.errorResponse(400, "InvalidInput", "AssumeRolePolicyDocument is required"), nil
	}

	// Validate the policy document is valid JSON
	if !json.Valid([]byte(assumeRolePolicyDocument)) {
		return s.errorResponse(400, "MalformedPolicyDocument", "AssumeRolePolicyDocument is not valid JSON"), nil
	}

	// Check if role already exists
	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	if s.state.Exists(stateKey) {
		return s.errorResponse(409, "EntityAlreadyExists", fmt.Sprintf("Role with name %s already exists.", roleName)), nil
	}

	path := getStringValue(params, "Path")
	if path == "" {
		path = "/"
	}

	description := getStringValue(params, "Description")
	maxSessionDuration := getInt32Value(params, "MaxSessionDuration", defaultMaxSessionDur)

	role := XMLRole{
		RoleName:                 roleName,
		RoleId:                   generateIAMId("AROA"),
		Arn:                      fmt.Sprintf("arn:aws:iam::%s:role%s%s", defaultAccountID, path, roleName),
		Path:                     path,
		AssumeRolePolicyDocument: assumeRolePolicyDocument,
		Description:              description,
		MaxSessionDuration:       maxSessionDuration,
		CreateDate:               time.Now().UTC(),
		Tags:                     s.parseTags(params),
	}

	if err := s.state.Set(stateKey, &role); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store role"), nil
	}

	// Initialize empty attachments
	attachKey := fmt.Sprintf("iam:role-policies:%s", roleName)
	if err := s.state.Set(attachKey, &RoleAttachments{PolicyArns: []string{}}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to initialize role attachments"), nil
	}

	// Register role in the relationship graph
	s.registerResource("role", roleName, map[string]string{
		"arn":  role.Arn,
		"path": role.Path,
	})

	result := CreateRoleResult{Role: role}
	return s.successResponse("CreateRole", result)
}

func (s *IAMService) getRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	var role XMLRole
	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	if err := s.state.Get(stateKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	result := GetRoleResult{Role: role}
	return s.successResponse("GetRole", result)
}

func (s *IAMService) deleteRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Unregister from graph (validates no dependents via graph relationships)
	// The graph tracks policy attachments and instance profile membership as edges
	attachKey := fmt.Sprintf("iam:role-policies:%s", roleName)
	if err := s.unregisterResource("role", roleName); err != nil {
		return s.errorResponse(409, "DeleteConflict", fmt.Sprintf("Cannot delete role: %v", err)), nil
	}

	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete role"), nil
	}

	// Clean up attachments
	s.state.Delete(attachKey)

	return s.successResponse("DeleteRole", EmptyResult{})
}

func (s *IAMService) updateAssumeRolePolicy(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	policyDocument := getStringValue(params, "PolicyDocument")
	if policyDocument == "" {
		return s.errorResponse(400, "InvalidInput", "PolicyDocument is required"), nil
	}

	if !json.Valid([]byte(policyDocument)) {
		return s.errorResponse(400, "MalformedPolicyDocument", "PolicyDocument is not valid JSON"), nil
	}

	var role XMLRole
	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	if err := s.state.Get(stateKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	role.AssumeRolePolicyDocument = policyDocument

	if err := s.state.Set(stateKey, &role); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update role"), nil
	}

	return s.successResponse("UpdateAssumeRolePolicy", EmptyResult{})
}

func (s *IAMService) listRoles(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	pathPrefix := getStringValue(params, "PathPrefix")

	keys, err := s.state.List("iam:role:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list roles"), nil
	}

	var roles []XMLRoleListItem
	for _, key := range keys {
		var role XMLRole
		if err := s.state.Get(key, &role); err == nil {
			if pathPrefix == "" || strings.HasPrefix(role.Path, pathPrefix) {
				roles = append(roles, roleToListItem(role))
			}
		}
	}

	result := ListRolesResult{
		Roles:       roles,
		IsTruncated: false,
	}
	return s.successResponse("ListRoles", result)
}

// ============================================================================
// Role Update Operations
// ============================================================================

func (s *IAMService) updateRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	var role XMLRole
	if err := s.state.Get(stateKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Update description if provided
	if description, ok := params["Description"]; ok {
		if desc, ok := description.(string); ok {
			role.Description = desc
		}
	}

	// Update max session duration if provided
	if _, ok := params["MaxSessionDuration"]; ok {
		duration := getInt32Value(params, "MaxSessionDuration", role.MaxSessionDuration)
		if duration < 3600 || duration > 43200 {
			return s.errorResponse(400, "ValidationError", "MaxSessionDuration must be between 3600 and 43200 seconds"), nil
		}
		role.MaxSessionDuration = duration
	}

	if err := s.state.Set(stateKey, &role); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update role"), nil
	}

	return s.successResponse("UpdateRole", UpdateRoleResult{})
}

func (s *IAMService) updateRoleDescription(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	description := getStringValue(params, "Description")

	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	var role XMLRole
	if err := s.state.Get(stateKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	role.Description = description

	if err := s.state.Set(stateKey, &role); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update role description"), nil
	}

	result := GetRoleResult{Role: role}
	return s.successResponse("UpdateRoleDescription", result)
}

// ============================================================================
// Role Tag Operations
// ============================================================================

func (s *IAMService) tagRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	var role XMLRole
	if err := s.state.Get(stateKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	newTags := s.parseTags(params)
	if len(newTags) == 0 {
		return s.errorResponse(400, "ValidationError", "Tags is required"), nil
	}

	// Merge tags - new tags override existing ones with same key
	tagMap := make(map[string]string)
	for _, tag := range role.Tags {
		tagMap[tag.Key] = tag.Value
	}
	for _, tag := range newTags {
		tagMap[tag.Key] = tag.Value
	}

	// Convert back to slice
	role.Tags = make([]XMLTag, 0, len(tagMap))
	for k, v := range tagMap {
		role.Tags = append(role.Tags, XMLTag{Key: k, Value: v})
	}

	if err := s.state.Set(stateKey, &role); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update role tags"), nil
	}

	return s.successResponse("TagRole", EmptyResult{})
}

func (s *IAMService) untagRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	var role XMLRole
	if err := s.state.Get(stateKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Parse tag keys to remove
	tagKeysToRemove := s.parseRoleTagKeys(params)
	if len(tagKeysToRemove) == 0 {
		return s.errorResponse(400, "ValidationError", "TagKeys is required"), nil
	}

	// Remove specified tags
	removeSet := make(map[string]bool)
	for _, key := range tagKeysToRemove {
		removeSet[key] = true
	}

	newTags := make([]XMLTag, 0)
	for _, tag := range role.Tags {
		if !removeSet[tag.Key] {
			newTags = append(newTags, tag)
		}
	}
	role.Tags = newTags

	if err := s.state.Set(stateKey, &role); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update role tags"), nil
	}

	return s.successResponse("UntagRole", EmptyResult{})
}

func (s *IAMService) listRoleTags(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	var role XMLRole
	if err := s.state.Get(stateKey, &role); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	result := ListRoleTagsResult{
		Tags:        role.Tags,
		IsTruncated: false,
	}
	return s.successResponse("ListRoleTags", result)
}

// parseRoleTagKeys parses tag keys from request parameters for role operations
func (s *IAMService) parseRoleTagKeys(params map[string]interface{}) []string {
	var keys []string
	tagIndex := 1

	for {
		keyParam := fmt.Sprintf("TagKeys.member.%d", tagIndex)
		key, hasKey := params[keyParam].(string)

		if !hasKey {
			// Try TagKeys.TagKey.N format
			keyParam = fmt.Sprintf("TagKeys.TagKey.%d", tagIndex)
			key, hasKey = params[keyParam].(string)
		}

		if !hasKey {
			break
		}

		keys = append(keys, key)
		tagIndex++
	}

	return keys
}

// ============================================================================
// Service-Linked Role Operations
// ============================================================================

func (s *IAMService) createServiceLinkedRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	awsServiceName := getStringValue(params, "AWSServiceName")
	if awsServiceName == "" {
		return s.errorResponse(400, "InvalidInput", "AWSServiceName is required"), nil
	}

	// Generate role name from service name
	// e.g., elasticmapreduce.amazonaws.com -> AWSServiceRoleForEMR
	servicePart := strings.Split(awsServiceName, ".")[0]
	roleName := fmt.Sprintf("AWSServiceRoleFor%s", strings.Title(servicePart))

	// Check if role already exists
	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	if s.state.Exists(stateKey) {
		return s.errorResponse(400, "InvalidInput", fmt.Sprintf("Service role for %s already exists", awsServiceName)), nil
	}

	description := getStringValue(params, "Description")
	customSuffix := getStringValue(params, "CustomSuffix")
	if customSuffix != "" {
		roleName = fmt.Sprintf("%s_%s", roleName, customSuffix)
	}

	// Create the assume role policy document for the service
	assumeRolePolicyDocument := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"Service": "%s"},
			"Action": "sts:AssumeRole"
		}]
	}`, awsServiceName)

	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)

	role := XMLRole{
		RoleName:                 roleName,
		RoleId:                   generateIAMId("AROA"),
		Arn:                      fmt.Sprintf("arn:aws:iam::%s:role%s%s", defaultAccountID, path, roleName),
		Path:                     path,
		AssumeRolePolicyDocument: assumeRolePolicyDocument,
		Description:              description,
		MaxSessionDuration:       defaultMaxSessionDur,
		CreateDate:               time.Now().UTC(),
	}

	if err := s.state.Set(stateKey, &role); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store service-linked role"), nil
	}

	// Initialize empty attachments
	attachKey := fmt.Sprintf("iam:role-policies:%s", roleName)
	s.state.Set(attachKey, &RoleAttachments{PolicyArns: []string{}})

	// Register role in the relationship graph
	s.registerResource("role", roleName, map[string]string{
		"arn":              role.Arn,
		"path":             role.Path,
		"serviceLinked":    "true",
		"awsServiceName":   awsServiceName,
	})

	result := CreateServiceLinkedRoleResult{Role: role}
	return s.successResponse("CreateServiceLinkedRole", result)
}

func (s *IAMService) deleteServiceLinkedRole(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	roleName := getStringValue(params, "RoleName")
	if roleName == "" {
		return s.errorResponse(400, "InvalidInput", "RoleName is required"), nil
	}

	stateKey := fmt.Sprintf("iam:role:%s", roleName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The role with name %s cannot be found.", roleName)), nil
	}

	// Generate a deletion task ID
	taskId := generateIAMId("task")

	// Create a deletion task (in a real implementation, this would be async)
	task := ServiceLinkedRoleDeletionTask{
		TaskId:     taskId,
		RoleName:   roleName,
		Status:     "SUCCEEDED", // For emulation, we complete immediately
		CreateDate: time.Now().UTC(),
	}

	// Store the task
	taskKey := fmt.Sprintf("iam:slr-deletion-task:%s", taskId)
	s.state.Set(taskKey, &task)

	// Actually delete the role
	attachKey := fmt.Sprintf("iam:role-policies:%s", roleName)
	s.unregisterResource("role", roleName)
	s.state.Delete(stateKey)
	s.state.Delete(attachKey)

	result := DeleteServiceLinkedRoleResult{DeletionTaskId: taskId}
	return s.successResponse("DeleteServiceLinkedRole", result)
}

func (s *IAMService) getServiceLinkedRoleDeletionStatus(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	deletionTaskId := getStringValue(params, "DeletionTaskId")
	if deletionTaskId == "" {
		return s.errorResponse(400, "InvalidInput", "DeletionTaskId is required"), nil
	}

	taskKey := fmt.Sprintf("iam:slr-deletion-task:%s", deletionTaskId)
	var task ServiceLinkedRoleDeletionTask
	if err := s.state.Get(taskKey, &task); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("Deletion task %s not found.", deletionTaskId)), nil
	}

	result := GetServiceLinkedRoleDeletionStatusResult{
		Status: task.Status,
	}

	if task.Status == "FAILED" && task.Reason != "" {
		result.Reason = &RoleDeletionFailureReason{
			Reason: task.Reason,
		}
	}

	return s.successResponse("GetServiceLinkedRoleDeletionStatus", result)
}
