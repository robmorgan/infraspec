package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to a DynamoDB resource (table or stream).
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

	// Extract resource type and name from ARN
	// ARN format: arn:aws:dynamodb:region:account-id:table/table-name
	// or: arn:aws:dynamodb:region:account-id:table/table-name/stream/stream-label
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if this is a table or stream
	resourceType := ""
	resourceName := ""
	if strings.Contains(resourceArn, "/stream/") {
		resourceType = "stream"
		// Extract stream label
		resourceName = parts[len(parts)-1]
	} else {
		resourceType = "table"
		resourceName = parts[len(parts)-1]
	}

	// Check if the resource exists (for tables)
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: %s", resourceArn)), nil
		}
	}

	// Get existing policy to check revision ID
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	existingRevisionId := ""

	if err := s.state.Get(policyKey, &existingPolicy); err == nil {
		// Policy exists, check revision ID if provided
		if revId, ok := existingPolicy["RevisionId"].(string); ok {
			existingRevisionId = revId
		}

		// If ExpectedRevisionId is provided, validate it matches
		if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "PolicyNotFoundException", "The resource policy with the revision ID provided was not found"), nil
			}
		}
	}

	// Generate new revision ID
	newRevisionId := uuid.New().String()

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn":                     resourceArn,
		"Policy":                          policy,
		"RevisionId":                      newRevisionId,
		"ConfirmRemoveSelfResourceAccess": false,
	}

	if input.ConfirmRemoveSelfResourceAccess != nil {
		policyData["ConfirmRemoveSelfResourceAccess"] = *input.ConfirmRemoveSelfResourceAccess
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store policy"), nil
	}

	// Build response
	response := PutResourcePolicyOutput{
		RevisionId: &newRevisionId,
	}

	return s.jsonResponse(200, response)
}
