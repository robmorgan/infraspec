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
	// ARN format: arn:aws:dynamodb:region:account-id:table/table-name
	// or: arn:aws:dynamodb:region:account-id:table/table-name/stream/stream-label
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if this is a table or stream
	resourceType := "table"
	resourceId := parts[1]
	if len(parts) >= 4 && parts[2] == "stream" {
		resourceType = "stream"
		resourceId = fmt.Sprintf("%s/stream/%s", parts[1], parts[3])
	}

	// For tables, verify the table exists
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceId)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Table not found: %s", resourceId)), nil
		}
	}

	// Check for existing policy and validate ExpectedRevisionId if provided
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s:%s", resourceType, resourceId)
	var existingPolicy map[string]interface{}
	_policyExists := false
	if err := s.state.Get(policyKey, &existingPolicy); err == nil {
		policyExists = true

		// If ExpectedRevisionId is provided, validate it matches
		if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
			if revisionId, ok := existingPolicy["RevisionId"].(string); ok {
				if revisionId != *input.ExpectedRevisionId {
					return s.errorResponse(400, "PolicyNotFoundException", "Policy revision does not match"), nil
				}
			}
		}
	}

	// Generate new revision ID
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
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
