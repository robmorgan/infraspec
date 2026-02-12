package dynamodb

import (
	"context"
	"fmt"
	"strings"

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

	// Extract resource identifier from ARN
	// ARN format: arn:aws:dynamodb:region:account:table/tablename or arn:aws:dynamodb:region:account:table/tablename/stream/label
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if it's a table or stream
	isStream := len(parts) >= 4 && parts[2] == "stream"

	var resourceKey string
	if isStream {
		// Stream ARN: arn:aws:dynamodb:region:account:table/tablename/stream/label
		tableName := parts[1]
		streamLabel := parts[3]
		resourceKey = fmt.Sprintf("dynamodb:resource-policy:stream:%s:%s", tableName, streamLabel)
	} else {
		// Table ARN: arn:aws:dynamodb:region:account:table/tablename
		tableName := parts[1]

		// Verify table exists
		tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
		}

		resourceKey = fmt.Sprintf("dynamodb:resource-policy:table:%s", tableName)
	}

	// Check for expected revision ID if provided
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		var existingPolicy map[string]interface{}
		if err := s.state.Get(resourceKey, &existingPolicy); err == nil {
			// Policy exists, check revision ID
			if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
				if existingRevisionId != *input.ExpectedRevisionId {
					return s.errorResponse(400, "PolicyNotFoundException", "The policy revision ID does not match the expected value"), nil
				}
			}
		}
	}

	// Generate new revision ID
	revisionId := uuid.New().String()

	// Store policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  revisionId,
	}

	if input.ConfirmRemoveSelfResourceAccess != nil {
		policyData["ConfirmRemoveSelfResourceAccess"] = *input.ConfirmRemoveSelfResourceAccess
	}

	if err := s.state.Set(resourceKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := PutResourcePolicyOutput{
		RevisionId: &revisionId,
	}

	return s.jsonResponse(200, response)
}
