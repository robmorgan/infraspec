package dynamodb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to the resource, which can be a table or stream.
// When you attach a resource-based policy using this API, the policy application is eventually consistent.
func (s *DynamoDBService) putResourcePolicy(ctx context.Context, input *PutResourcePolicyInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.ResourceArn == nil || *input.ResourceArn == "" {
		return s.errorResponse(400, "ValidationException", "ResourceArn is required"), nil
	}
	if input.Policy == nil || *input.Policy == "" {
		return s.errorResponse(400, "ValidationException", "Policy is required"), nil
	}

	resourceArn := *input.ResourceArn
	policy := *input.Policy

	// Extract resource identifier from ARN
	// ARN format: arn:aws:dynamodb:region:account:table/tablename or arn:aws:dynamodb:region:account:table/tablename/stream/...
	resourceKey := extractResourceKeyFromArn(resourceArn)
	if resourceKey == "" {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Check if resource exists (table or stream)
	// For simplicity, we'll check if it's a table
	if strings.Contains(resourceArn, "/table/") {
		tableName := extractTableNameFromArn(resourceArn)
		tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
		var tableData map[string]interface{}
		if err := s.state.Get(tableKey, &tableData); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: %s", resourceArn)), nil
		}
	}

	// Get existing policy if any
	policyKey := fmt.Sprintf("dynamodb:resourcepolicy:%s", resourceKey)
	var existingPolicyData map[string]interface{}
	_ = s.state.Get(policyKey, &existingPolicyData)

	// Check ExpectedRevisionId if specified
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if existingPolicyData == nil {
			return s.errorResponse(400, "PolicyNotFoundException", "No existing policy found for the resource"), nil
		}
		if existingRevisionId, ok := existingPolicyData["RevisionId"].(string); ok {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyRevisionMismatchException", "Policy revision does not match"), nil
			}
		}
	}

	// Generate a new revision ID based on the policy content
	revisionId := generateRevisionId(policy)

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  revisionId,
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}

// extractResourceKeyFromArn extracts a unique resource key from the ARN
func extractResourceKeyFromArn(arn string) string {
	// ARN format: arn:aws:dynamodb:region:account:table/tablename or arn:aws:dynamodb:region:account:table/tablename/stream/...
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return ""
	}
	// Use the last part (table/tablename or table/tablename/stream/...) as the resource key
	resourcePart := parts[5]
	// Replace slashes with colons to make it a valid state key
	return strings.ReplaceAll(resourcePart, "/", ":")
}

// extractTableNameFromArn extracts the table name from the ARN
func extractTableNameFromArn(arn string) string {
	// ARN format: arn:aws:dynamodb:region:account:table/tablename
	parts := strings.Split(arn, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// generateRevisionId generates a unique revision ID based on the policy content
func generateRevisionId(policy string) string {
	// Use a hash of the policy content combined with a UUID for uniqueness
	hash := sha256.Sum256([]byte(policy + uuid.New().String()))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for brevity
}
