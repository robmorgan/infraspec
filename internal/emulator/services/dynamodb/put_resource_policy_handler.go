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
	// ARN format: arn:aws:dynamodb:us-east-1:000000000000:table/tablename
	// or arn:aws:dynamodb:us-east-1:000000000000:table/tablename/stream/...
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}
	tableName := parts[1]

	// Verify table exists
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(tableKey, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Get existing policy to check revision
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	existingRevisionId := ""
	if err := s.state.Get(policyKey, &existingPolicy); err == nil {
		if revisionId, ok := existingPolicy["RevisionId"].(string); ok {
			existingRevisionId = revisionId
		}
	}

	// If ExpectedRevisionId is provided, verify it matches
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if existingRevisionId != *input.ExpectedRevisionId {
			return s.errorResponse(400, "PolicyNotFoundException", "The resource policy with the revision ID provided in the request could not be found."), nil
		}
	}

	// Generate new revision ID
	newRevisionId := uuid.New().String()

	// Store policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  newRevisionId,
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
