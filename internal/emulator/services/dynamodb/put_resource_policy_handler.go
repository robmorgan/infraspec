package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// putResourcePolicy attaches a resource-based policy document to the resource,
// which can be a table or stream. When you attach a resource-based policy using this API,
// the policy application is eventually consistent.
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

	// Validate policy is valid JSON
	var policyDoc map[string]interface{}
	if err := json.Unmarshal([]byte(policy), &policyDoc); err != nil {
		return s.errorResponse(400, "ValidationException", "Policy must be a valid JSON document"), nil
	}

	// Extract resource identifier from ARN
	// ARN format: arn:aws:dynamodb:region:account-id:table/table-name
	// or: arn:aws:dynamodb:region:account-id:table/table-name/stream/label
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	resourceType := "table"
	resourceName := parts[1]

	// Check if this is a stream ARN
	if len(parts) >= 4 && parts[2] == "stream" {
		resourceType = "stream"
	}

	// Verify the resource exists
	if resourceType == "table" {
		key := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(key, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", resourceName)), nil
		}
	}

	// Check if policy already exists and validate ExpectedRevisionId if provided
	policyKey := fmt.Sprintf("dynamodb:policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicy) == nil

	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if !policyExists {
			return s.errorResponse(400, "PolicyNotFoundException", "Policy not found for the specified resource"), nil
		}
		if existingRevisionId, ok := existingPolicy["RevisionId"].(string); ok {
			if existingRevisionId != *input.ExpectedRevisionId {
				return s.errorResponse(409, "PolicyRevisionIdMismatchException", "The policy revision ID does not match the expected revision ID"), nil
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
		return s.errorResponse(500, "InternalServerError", "Failed to store policy"), nil
	}

	// Build response
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
