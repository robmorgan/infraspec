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

	// Extract resource identifier from ARN
	// ARN format: arn:aws:dynamodb:region:account:table/tablename
	// or: arn:aws:dynamodb:region:account:table/tablename/stream/streamlabel
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine resource type and name
	resourceType := "table"
	resourceName := parts[1]
	if len(parts) > 3 && parts[2] == "stream" {
		resourceType = "stream"
		resourceName = strings.Join(parts[1:], "/")
	}

	// For table resources, verify the table exists
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: Table: %s not found", resourceName)), nil
		}
	}

	// Check if there's an existing policy and validate ExpectedRevisionId if provided
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicy) == nil

	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if !policyExists {
			return s.errorResponse(400, "PolicyNotFoundException",
				"The resource policy was not found"), nil
		}
		if currentRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
			if currentRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(400, "ConditionalCheckFailedException",
					"The conditional request failed"), nil
			}
		}
	}

	// Generate new revision ID
	revisionId := uuid.New().String()

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      *input.Policy,
		"RevisionId":  revisionId,
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store policy"), nil
	}

	// Build response
	response := PutResourcePolicyOutput{
		RevisionId: &revisionId,
	}

	return s.jsonResponse(200, response)
}
