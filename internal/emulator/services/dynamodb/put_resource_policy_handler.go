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
	// ARN format: arn:aws:dynamodb:region:account:table/tablename
	// or arn:aws:dynamodb:region:account:table/tablename/stream/timestamp
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	resourceType := "table" // Default to table
	resourceName := parts[1]

	// Check if this is a stream ARN
	if len(parts) >= 3 && parts[2] == "stream" {
		resourceType = "stream"
	}

	// Verify the resource exists (table)
	if resourceType == "table" {
		key := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(key, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", resourceName)), nil
		}
	}

	// Get existing policy if any
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	s.state.Get(policyKey, &existingPolicy) // Ignore error if policy doesn't exist

	// Check ExpectedRevisionId if provided
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if existingPolicy == nil {
			return s.errorResponse(400, "PolicyNotFoundException",
				"No policy exists for the specified resource"), nil
		}

		if revisionId, ok := existingPolicy["RevisionId"].(string); ok {
			if revisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyRevisionMismatchException",
					"The policy revision ID does not match"), nil
			}
		}
	}

	// Generate new revision ID
	newRevisionId := uuid.New().String()

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  newRevisionId,
	}

	if input.ConfirmRemoveSelfResourceAccess != nil {
		policyData["ConfirmRemoveSelfResourceAccess"] = *input.ConfirmRemoveSelfResourceAccess
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": newRevisionId,
	}

	return s.jsonResponse(200, response)
}
