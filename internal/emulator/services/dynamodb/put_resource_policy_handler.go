package dynamodb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to the resource,
// which can be a table or stream.
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

	// Verify the resource exists (extract table name from ARN)
	// ARN format: arn:aws:dynamodb:us-east-1:000000000000:table/tablename
	// or arn:aws:dynamodb:us-east-1:000000000000:table/tablename/stream/streamlabel
	tableName, err := extractTableNameFromArn(resourceArn)
	if err != nil {
		return s.errorResponse(400, "ValidationException", fmt.Sprintf("Invalid ResourceArn: %s", err.Error())), nil
	}

	// Check if table exists
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(tableKey, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Table not found: %s", tableName)), nil
	}

	// Get existing policy if it exists
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	existingRevisionId := ""

	if err := s.state.Get(policyKey, &existingPolicy); err == nil {
		// Policy exists, check revision ID if provided
		if revId, ok := existingPolicy["RevisionId"].(string); ok {
			existingRevisionId = revId
		}

		// If ExpectedRevisionId is provided, verify it matches
		if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyNotFoundException", "Policy revision ID does not match"), nil
			}
		}
	}

	// Generate new revision ID
	newRevisionId := uuid.New().String()

	// Create or update the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  newRevisionId,
	}

	// Store the policy
	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": newRevisionId,
	}

	return s.jsonResponse(200, response)
}

// Helper function to extract table name from ARN
func extractTableNameFromArn(arn string) (string, error) {
	// ARN format: arn:aws:dynamodb:region:account:table/tablename
	// or arn:aws:dynamodb:region:account:table/tablename/stream/streamlabel

	// Simple parsing - split by / and get the table name
	parts := splitArnParts(arn)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid ARN format")
	}

	// Find "table/" in the ARN
	for i, part := range parts {
		if part == "table" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	return "", fmt.Errorf("table name not found in ARN")
}

// Helper function to split ARN into parts
func splitArnParts(arn string) []string {
	parts := []string{}
	current := ""

	for _, char := range arn {
		if char == ':' || char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
