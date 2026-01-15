package dynamodb

import (
	"context"
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
	// ARN format: arn:aws:dynamodb:region:account-id:table/tablename
	// or: arn:aws:dynamodb:region:account-id:table/tablename/stream/streamlabel
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if this is a table or stream
	resourceType := "table"
	resourceName := parts[1]
	if len(parts) >= 4 && parts[2] == "stream" {
		resourceType = "stream"
		resourceName = fmt.Sprintf("%s/stream/%s", parts[1], parts[3])
	}

	// Verify the resource exists
	tableKey := fmt.Sprintf("dynamodb:table:%s", parts[1])
	var tableDesc map[string]interface{}
	if err := s.state.Get(tableKey, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: %s", resourceArn)), nil
	}

	// Check if there's an existing policy and validate revision ID if provided
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s:%s", resourceType, resourceName)
	var existingPolicy map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicy) == nil

	if policyExists && input.ExpectedRevisionId != nil {
		// Validate the expected revision ID matches
		if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyNotFoundException", "The policy revision ID does not match the expected revision ID"), nil
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

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := PutResourcePolicyOutput{
		RevisionId: &revisionId,
	}

	return s.jsonResponse(200, response)
}
