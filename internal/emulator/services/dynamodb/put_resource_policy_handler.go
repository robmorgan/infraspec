package dynamodb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to a DynamoDB resource.
// The resource can be a table or stream. Policy application is eventually consistent.
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

	// Store the policy using the full ARN as part of the state key
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)

	// If ExpectedRevisionId is specified, verify it matches the current revision
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		var existingPolicy map[string]interface{}
		if err := s.state.Get(policyKey, &existingPolicy); err != nil {
			// No existing policy but revision ID was specified
			return s.errorResponse(400, "PolicyNotFoundException",
				fmt.Sprintf("No policy found for resource: %s", resourceArn)), nil
		}

		currentRevision, _ := existingPolicy["RevisionId"].(string)
		if currentRevision != *input.ExpectedRevisionId {
			return s.errorResponse(400, "TransactionCanceledException",
				fmt.Sprintf("Expected revision ID %s does not match current revision ID %s",
					*input.ExpectedRevisionId, currentRevision)), nil
		}
	}

	// Generate a new revision ID for the policy
	revisionId := uuid.New().String()

	// Store the policy with its revision ID
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
