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

	// Generate a state key for the resource policy
	// Extract resource identifier from ARN for the key
	// ARN format: arn:aws:dynamodb:region:account:table/tablename or arn:aws:dynamodb:region:account:table/tablename/stream/timestamp
	stateKey := fmt.Sprintf("dynamodb:resourcepolicy:%s", resourceArn)

	// Check if a policy already exists
	var existingPolicy map[string]interface{}
	policyExists := false
	if err := s.state.Get(stateKey, &existingPolicy); err == nil {
		policyExists = true
	}

	// Handle ExpectedRevisionId if provided
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if !policyExists {
			return s.errorResponse(400, "PolicyNotFoundException", "The resource policy does not exist"), nil
		}

		// Verify the revision ID matches
		if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyNotFoundException", "The provided revision ID does not match the current revision ID"), nil
			}
		}
	}

	// Generate a new revision ID
	revisionId := uuid.New().String()

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  revisionId,
	}

	if input.ConfirmRemoveSelfResourceAccess != nil {
		policyData["ConfirmRemoveSelfResourceAccess"] = *input.ConfirmRemoveSelfResourceAccess
	}

	if err := s.state.Set(stateKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
