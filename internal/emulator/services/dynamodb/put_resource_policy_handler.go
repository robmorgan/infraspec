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

	// State key for resource policy
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)

	// Check if a policy already exists and validate ExpectedRevisionId
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		expectedRevisionId := *input.ExpectedRevisionId
		var existingPolicy map[string]interface{}
		if err := s.state.Get(policyKey, &existingPolicy); err != nil {
			// No existing policy, but an expected revision was provided
			return s.errorResponse(400, "PolicyNotFoundException",
				fmt.Sprintf("No resource-based policy found for resource %s", resourceArn)), nil
		}

		// Validate the revision ID matches
		currentRevisionId, _ := existingPolicy["RevisionId"].(string)
		if currentRevisionId != expectedRevisionId {
			return s.errorResponse(400, "TransactionConflictException",
				"The revision ID does not match the provided expected revision ID"), nil
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

	// Return the revision ID
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
