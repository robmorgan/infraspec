package dynamodb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to the resource,
// which can be a table or stream. When you attach a resource-based policy using this API,
// the policy application is eventually consistent.
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

	// Get resource policy from state if it exists
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var existingPolicyData map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicyData) == nil

	// If ExpectedRevisionId is provided, validate it matches
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if !policyExists {
			return s.errorResponse(400, "PolicyNotFoundException", "No policy exists for this resource"), nil
		}

		existingRevisionId := ""
		if revId, ok := existingPolicyData["RevisionId"].(string); ok {
			existingRevisionId = revId
		}

		if existingRevisionId != *input.ExpectedRevisionId {
			return s.errorResponse(400, "PolicyRevisionMismatchException", "Expected revision ID does not match current revision"), nil
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

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := PutResourcePolicyOutput{
		RevisionId: &revisionId,
	}

	return s.jsonResponse(200, response)
}
