package iam

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// ============================================================================
// Access Key Operations
// ============================================================================

func (s *IAMService) createAccessKey(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Check if user already has 2 access keys (AWS limit)
	existingKeys := s.listAccessKeysForUser(userName)
	if len(existingKeys) >= 2 {
		return s.errorResponse(409, "LimitExceeded", "Cannot exceed quota for AccessKeysPerUser: 2"), nil
	}

	// Generate access key ID and secret
	accessKeyId := generateAccessKeyId()
	secretAccessKey := generateSecretAccessKey()
	now := time.Now().UTC()

	accessKey := AccessKeyData{
		UserName:        userName,
		AccessKeyId:     accessKeyId,
		SecretAccessKey: secretAccessKey,
		Status:          "Active",
		CreateDate:      now,
	}

	// Store access key
	keyStateKey := fmt.Sprintf("iam:access-key:%s:%s", userName, accessKeyId)
	if err := s.state.Set(keyStateKey, &accessKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store access key"), nil
	}

	// Add relationship in graph: access-key -> user
	if err := s.addRelationship("access-key", accessKeyId, "user", userName, graph.RelContains); err != nil {
		// Non-critical, just log
	}

	result := CreateAccessKeyResult{
		AccessKey: XMLAccessKey{
			UserName:        userName,
			AccessKeyId:     accessKeyId,
			Status:          "Active",
			SecretAccessKey: secretAccessKey,
			CreateDate:      now,
		},
	}
	return s.successResponse("CreateAccessKey", result)
}

func (s *IAMService) deleteAccessKey(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	accessKeyId := getStringValue(params, "AccessKeyId")
	if accessKeyId == "" {
		return s.errorResponse(400, "ValidationError", "AccessKeyId is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Check if access key exists
	keyStateKey := fmt.Sprintf("iam:access-key:%s:%s", userName, accessKeyId)
	if !s.state.Exists(keyStateKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The Access Key with id %s cannot be found.", accessKeyId)), nil
	}

	// Delete the access key
	if err := s.state.Delete(keyStateKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete access key"), nil
	}

	// Remove relationship in graph
	s.removeRelationship("access-key", accessKeyId, "user", userName, graph.RelContains)

	return s.successResponse("DeleteAccessKey", EmptyResult{})
}

func (s *IAMService) updateAccessKey(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	accessKeyId := getStringValue(params, "AccessKeyId")
	if accessKeyId == "" {
		return s.errorResponse(400, "ValidationError", "AccessKeyId is required"), nil
	}

	status := getStringValue(params, "Status")
	if status == "" {
		return s.errorResponse(400, "ValidationError", "Status is required"), nil
	}

	// Validate status
	if status != "Active" && status != "Inactive" {
		return s.errorResponse(400, "ValidationError", "Status must be Active or Inactive"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get access key
	keyStateKey := fmt.Sprintf("iam:access-key:%s:%s", userName, accessKeyId)
	var accessKey AccessKeyData
	if err := s.state.Get(keyStateKey, &accessKey); err != nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The Access Key with id %s cannot be found.", accessKeyId)), nil
	}

	// Update status
	accessKey.Status = status
	if err := s.state.Set(keyStateKey, &accessKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update access key"), nil
	}

	return s.successResponse("UpdateAccessKey", EmptyResult{})
}

func (s *IAMService) listAccessKeys(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	userName := getStringValue(params, "UserName")
	if userName == "" {
		return s.errorResponse(400, "ValidationError", "UserName is required"), nil
	}

	// Verify user exists
	userKey := fmt.Sprintf("iam:user:%s", userName)
	if !s.state.Exists(userKey) {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The user with name %s cannot be found.", userName)), nil
	}

	// Get all access keys for the user
	accessKeys := s.listAccessKeysForUser(userName)

	var metadata []XMLAccessKeyMetadata
	for _, key := range accessKeys {
		metadata = append(metadata, XMLAccessKeyMetadata{
			UserName:    key.UserName,
			AccessKeyId: key.AccessKeyId,
			Status:      key.Status,
			CreateDate:  key.CreateDate,
		})
	}

	result := ListAccessKeysResult{
		AccessKeyMetadata: metadata,
		IsTruncated:       false,
	}
	return s.successResponse("ListAccessKeys", result)
}

func (s *IAMService) getAccessKeyLastUsed(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	accessKeyId := getStringValue(params, "AccessKeyId")
	if accessKeyId == "" {
		return s.errorResponse(400, "ValidationError", "AccessKeyId is required"), nil
	}

	// Find the access key across all users
	accessKey, userName := s.findAccessKeyById(accessKeyId)
	if accessKey == nil {
		return s.errorResponse(404, "NoSuchEntity", fmt.Sprintf("The Access Key with id %s cannot be found.", accessKeyId)), nil
	}

	result := GetAccessKeyLastUsedResult{
		UserName: userName,
		AccessKeyLastUsed: XMLAccessKeyLastUsed{
			LastUsedDate: accessKey.LastUsedDate,
			ServiceName:  accessKey.LastUsedService,
			Region:       accessKey.LastUsedRegion,
		},
	}
	return s.successResponse("GetAccessKeyLastUsed", result)
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateAccessKeyId generates an AWS-style access key ID (AKIA prefix + 16 chars)
func generateAccessKeyId() string {
	b := make([]byte, 12)
	rand.Read(b)
	encoded := base64.RawURLEncoding.EncodeToString(b)
	encoded = strings.ReplaceAll(encoded, "-", "X")
	encoded = strings.ReplaceAll(encoded, "_", "Y")
	encoded = strings.ToUpper(encoded)
	if len(encoded) > 16 {
		encoded = encoded[:16]
	}
	return "AKIA" + encoded
}

// generateSecretAccessKey generates an AWS-style secret access key (40 chars)
func generateSecretAccessKey() string {
	b := make([]byte, 30)
	rand.Read(b)
	encoded := base64.RawStdEncoding.EncodeToString(b)
	if len(encoded) > 40 {
		encoded = encoded[:40]
	}
	return encoded
}

// listAccessKeysForUser returns all access keys for a user
func (s *IAMService) listAccessKeysForUser(userName string) []AccessKeyData {
	prefix := fmt.Sprintf("iam:access-key:%s:", userName)
	keys, err := s.state.List(prefix)
	if err != nil {
		return []AccessKeyData{}
	}

	var accessKeys []AccessKeyData
	for _, key := range keys {
		var accessKey AccessKeyData
		if err := s.state.Get(key, &accessKey); err == nil {
			accessKeys = append(accessKeys, accessKey)
		}
	}
	return accessKeys
}

// findAccessKeyById finds an access key by ID across all users
func (s *IAMService) findAccessKeyById(accessKeyId string) (*AccessKeyData, string) {
	// List all access keys
	keys, err := s.state.List("iam:access-key:")
	if err != nil {
		return nil, ""
	}

	for _, key := range keys {
		var accessKey AccessKeyData
		if err := s.state.Get(key, &accessKey); err == nil {
			if accessKey.AccessKeyId == accessKeyId {
				return &accessKey, accessKey.UserName
			}
		}
	}
	return nil, ""
}

// countUserAccessKeys returns the number of access keys for a user
func (s *IAMService) countUserAccessKeys(userName string) int {
	return len(s.listAccessKeysForUser(userName))
}

// hasAccessKeys checks if a user has any access keys
func (s *IAMService) hasAccessKeys(userName string) bool {
	return s.countUserAccessKeys(userName) > 0
}

// deleteUserAccessKeys deletes all access keys for a user (used when deleting a user)
func (s *IAMService) deleteUserAccessKeys(userName string) {
	accessKeys := s.listAccessKeysForUser(userName)
	for _, key := range accessKeys {
		keyStateKey := fmt.Sprintf("iam:access-key:%s:%s", userName, key.AccessKeyId)
		s.state.Delete(keyStateKey)
		s.removeRelationship("access-key", key.AccessKeyId, "user", userName, graph.RelContains)
	}
}
