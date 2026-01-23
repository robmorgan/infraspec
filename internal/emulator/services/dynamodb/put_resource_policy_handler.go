package dynamodb

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	// Extract resource identifier from ARN
	// ARN format: arn:aws:dynamodb:us-east-1:000000000000:table/tablename
	// or: arn:aws:dynamodb:us-east-1:000000000000:table/tablename/stream/timestamp
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if this is a table or stream
	resourceType := "table"
	resourceName := parts[1]
	if len(parts) >= 4 && parts[2] == "stream" {
		resourceType = "stream"
		resourceName = parts[1] + "/stream/" + parts[3]
	}

	// For table resources, verify the table exists
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", resourceName)), nil
		}
	}

	// Generate or retrieve policy metadata
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
	var existingPolicy map[string]interface{}

	revisionId := uuid.New().String()

	// Check if there's an existing policy
	if err := s.state.Get(policyKey, &existingPolicy); err == nil {
		// Policy exists - check ExpectedRevisionId if provided
		if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
			if existingRevision, ok := existingPolicy["RevisionId"].(string); ok {
				if existingRevision != *input.ExpectedRevisionId {
					return s.errorResponse(400, "PolicyNotFoundException", "The policy revision ID does not match the expected value"), nil
				}
			}
		}
	}

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  revisionId,
		"CreatedAt":   time.Now().Unix(),
		"UpdatedAt":   time.Now().Unix(),
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
