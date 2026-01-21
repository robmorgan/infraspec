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
	// ARN format: arn:aws:dynamodb:region:account:table/tablename or arn:aws:dynamodb:region:account:table/tablename/stream/label
	var resourceKey string
	if strings.Contains(resourceArn, "/stream/") {
		// This is a stream ARN
		parts := strings.Split(resourceArn, "/")
		if len(parts) >= 2 {
			tableName := parts[1]
			resourceKey = fmt.Sprintf("dynamodb:table:%s", tableName)
		} else {
			return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
		}
	} else if strings.Contains(resourceArn, "/table/") {
		// This is a table ARN
		parts := strings.Split(resourceArn, "/")
		if len(parts) >= 2 {
			tableName := parts[len(parts)-1]
			resourceKey = fmt.Sprintf("dynamodb:table:%s", tableName)
		} else {
			return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
		}
	} else {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Verify the resource exists
	var resourceData map[string]interface{}
	if err := s.state.Get(resourceKey, &resourceData); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: %s", resourceArn)), nil
	}

	// Check if ExpectedRevisionId is provided and matches
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
		var existingPolicy map[string]interface{}
		if err := s.state.Get(policyKey, &existingPolicy); err == nil {
			// Policy exists, check revision ID
			if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
				if existingRevisionId != *input.ExpectedRevisionId {
					return s.errorResponse(400, "PolicyNotFoundException", "The resource policy with the revision ID provided was not found"), nil
				}
			}
		} else {
			// Policy doesn't exist but revision ID was expected
			return s.errorResponse(400, "PolicyNotFoundException", "The resource policy with the revision ID provided was not found"), nil
		}
	}

	// Generate a new revision ID
	revisionId := uuid.New().String()

	// Store the policy
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
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
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
