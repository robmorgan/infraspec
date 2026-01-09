package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to the resource, which can be a table or stream.
func (s *DynamoDBService) putResourcePolicy(ctx context.Context, input *PutResourcePolicyInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.ResourceArn == nil || *input.ResourceArn == "" {
		return s.errorResponse(400, "ValidationException", "ResourceArn is required"), nil
	}

	if input.Policy == nil || *input.Policy == "" {
		return s.errorResponse(400, "ValidationException", "Policy is required"), nil
	}

	resourceArn := *input.ResourceArn

	// Extract resource identifier from ARN
	// ARN format: arn:aws:dynamodb:region:account:table/tablename or arn:aws:dynamodb:region:account:table/tablename/stream/...
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	resourceType := "table"
	resourceName := parts[1]

	// Check if it's a stream ARN
	if len(parts) > 2 && parts[2] == "stream" {
		resourceType = "stream"
	}

	// Verify the resource exists (for table resources)
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: %s", resourceArn)), nil
		}
	}

	// Get the state key for the resource policy
	policyKey := fmt.Sprintf("dynamodb:resource_policy:%s", resourceArn)

	// Check if a policy already exists and handle ExpectedRevisionId
	var existingPolicy map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicy) == nil

	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if !policyExists {
			return s.errorResponse(400, "PolicyNotFoundException", "The resource policy was not found"), nil
		}

		// Verify the revision ID matches
		if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyRevisionMismatchException", "The policy revision ID does not match"), nil
			}
		}
	}

	// Generate a new revision ID
	revisionId := uuid.New().String()

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      *input.Policy,
		"RevisionId":  revisionId,
	}

	if input.ConfirmRemoveSelfResourceAccess != nil {
		policyData["ConfirmRemoveSelfResourceAccess"] = *input.ConfirmRemoveSelfResourceAccess
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
