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
	policy := *input.Policy

	// Extract table name from ARN
	// ARN format: arn:aws:dynamodb:us-east-1:000000000000:table/tablename or arn:aws:dynamodb:us-east-1:000000000000:table/tablename/stream/...
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Check if it's a table or stream ARN
	isStream := strings.Contains(resourceArn, "/stream/")
	var tableName string
	if isStream {
		// For stream ARNs, table name is after "table/" and before "/stream/"
		tableName = parts[len(parts)-3]
	} else {
		// For table ARNs, table name is the last part
		tableName = parts[len(parts)-1]
	}

	// Verify table exists
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(tableKey, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException",
			fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Build state key for resource policy
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)

	// Check for ExpectedRevisionId if provided
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		var existingPolicy map[string]interface{}
		if err := s.state.Get(policyKey, &existingPolicy); err == nil {
			// Policy exists, check revision ID
			if revisionId, ok := existingPolicy["RevisionId"].(string); ok {
				if revisionId != *input.ExpectedRevisionId {
					return s.errorResponse(400, "PolicyNotFoundException",
						"The resource policy with the revision id provided was not found"), nil
				}
			}
		} else {
			// Policy doesn't exist but ExpectedRevisionId was provided
			return s.errorResponse(400, "PolicyNotFoundException",
				"The resource policy with the revision id provided was not found"), nil
		}
	}

	// Generate new revision ID
	revisionId := uuid.New().String()

	// Store resource policy
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
