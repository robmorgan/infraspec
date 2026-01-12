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
	// or: arn:aws:dynamodb:region:account:table/tablename/stream/timestamp
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if this is a table or stream ARN
	resourceType := "table"
	resourceId := parts[1]
	if len(parts) >= 4 && parts[2] == "stream" {
		resourceType = "stream"
		resourceId = parts[1] + "/stream/" + parts[3]
	}

	// For table ARNs, verify the table exists
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceId)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", resourceId)), nil
		}
	}

	// Build state key for resource policy
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s:%s", resourceType, resourceId)

	// Check if there's an existing policy and validate revision ID if provided
	var existingPolicyData map[string]interface{}
	if err := s.state.Get(policyKey, &existingPolicyData); err == nil {
		// Policy exists, check revision ID if provided
		if input.ExpectedRevisionId != nil {
			existingRevisionId, _ := existingPolicyData["RevisionId"].(string)
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyNotFoundException",
					"The policy revision ID does not match the expected revision ID"), nil
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
