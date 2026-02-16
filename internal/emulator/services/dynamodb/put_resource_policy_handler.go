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
	// ARN format: arn:aws:dynamodb:region:account:table/tablename
	// or: arn:aws:dynamodb:region:account:table/tablename/stream/streamlabel
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine resource type and name
	resourceType := ""
	resourceName := ""
	if strings.Contains(resourceArn, "/stream/") {
		resourceType = "stream"
		// For streams, the resource name is table/stream/label
		resourceName = strings.Join(parts[1:], "/")
	} else {
		resourceType = "table"
		resourceName = parts[len(parts)-1]
	}

	// For tables, verify the table exists
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", resourceName)), nil
		}
	}

	// Check if a policy already exists for this resource
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicy) == nil

	// If ExpectedRevisionId is provided, validate it matches the current revision
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if policyExists {
			if currentRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
				if currentRevisionId != *input.ExpectedRevisionId {
					return s.errorResponse(400, "PolicyNotFoundException", "The policy revision ID does not match"), nil
				}
			}
		} else {
			return s.errorResponse(400, "PolicyNotFoundException", "No policy exists for this resource"), nil
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

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
