package dynamodb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to a DynamoDB resource
// (table or stream). The policy application is eventually consistent.
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

	// State key for resource policy - keyed by full ARN for consistency with getResourcePolicy
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)

	// If ExpectedRevisionId is specified, verify it matches the current revision
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		var existingPolicy map[string]interface{}
		if err := s.state.Get(policyKey, &existingPolicy); err != nil {
			// No existing policy but ExpectedRevisionId was provided
			return s.errorResponse(400, "ConditionalCheckFailedException",
				"The conditional request failed: expected revision ID does not match"), nil
		}
		currentRevisionId, _ := existingPolicy["RevisionId"].(string)
		if currentRevisionId != *input.ExpectedRevisionId {
			return s.errorResponse(400, "ConditionalCheckFailedException",
				"The conditional request failed: expected revision ID does not match"), nil
		}
	}

	// Generate a new revision ID
	revisionId := uuid.New().String()

	// Store the policy
	policyData := map[string]interface{}{
		"Policy":     policy,
		"RevisionId": revisionId,
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
