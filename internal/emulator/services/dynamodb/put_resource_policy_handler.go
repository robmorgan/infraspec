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
	// ARN format: arn:aws:dynamodb:region:account-id:table/table-name or arn:aws:dynamodb:region:account-id:table/table-name/stream/stream-label
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if it's a table or stream
	var resourceType, resourceName string
	if len(parts) >= 4 && parts[2] == "stream" {
		// Stream ARN
		resourceType = "stream"
		resourceName = parts[1] + "/stream/" + parts[3]
	} else {
		// Table ARN
		resourceType = "table"
		resourceName = parts[len(parts)-1]
	}

	// Verify that the resource exists
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", resourceName)), nil
		}
	}

	// Check for expected revision ID if provided
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicy) == nil

	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if !policyExists {
			return s.errorResponse(400, "PolicyNotFoundException", "No policy exists for this resource"), nil
		}

		if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyRevisionMismatch", "The provided revision ID does not match the current policy revision"), nil
			}
		}
	}

	// Generate a new revision ID
	revisionId := uuid.New().String()

	// Store the resource policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  revisionId,
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
