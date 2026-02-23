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

	// State key for resource policy uses the full ARN (consistent with GetResourcePolicy)
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)

	// Check ExpectedRevisionId if provided
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		var existingPolicy map[string]interface{}
		if err := s.state.Get(policyKey, &existingPolicy); err != nil {
			return s.errorResponse(400, "PolicyNotFoundException",
				"The specified resource does not have a resource-based policy"), nil
		}
		existingRevisionId, _ := existingPolicy["RevisionId"].(string)
		if existingRevisionId != *input.ExpectedRevisionId {
			return s.errorResponse(400, "PolicyRevisionIdMismatchException",
				"The provided revision ID does not match the current revision ID"), nil
		}
	}

	// Generate a new revision ID
	revisionId := uuid.New().String()

	// Store the policy
	policyData := map[string]interface{}{
		"Policy":      policy,
		"ResourceArn": resourceArn,
		"RevisionId":  revisionId,
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
