package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to a DynamoDB resource (table or stream).
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
	var resourceKey string
	if strings.Contains(resourceArn, "/stream/") {
		// Stream ARN
		parts := strings.Split(resourceArn, "/")
		if len(parts) >= 4 {
			streamLabel := parts[len(parts)-1]
			resourceKey = fmt.Sprintf("dynamodb:stream-policy:%s", streamLabel)
		} else {
			return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
		}
	} else {
		// Table ARN
		parts := strings.Split(resourceArn, "/")
		if len(parts) >= 2 {
			tableName := parts[len(parts)-1]
			resourceKey = fmt.Sprintf("dynamodb:table-policy:%s", tableName)

			// Verify table exists
			tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
			var tableDesc map[string]interface{}
			if err := s.state.Get(tableKey, &tableDesc); err != nil {
				return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Table not found: %s", tableName)), nil
			}
		} else {
			return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
		}
	}

	// Check if a policy already exists and validate ExpectedRevisionId if provided
	var existingPolicy map[string]interface{}
	if err := s.state.Get(resourceKey, &existingPolicy); err == nil {
		// Policy exists, check ExpectedRevisionId if provided
		if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
			if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
				if existingRevisionId != *input.ExpectedRevisionId {
					return s.errorResponse(400, "PolicyNotFoundException", "The policy revision ID does not match"), nil
				}
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

	if err := s.state.Set(resourceKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := PutResourcePolicyOutput{
		RevisionId: &revisionId,
	}

	return s.jsonResponse(200, response)
}
