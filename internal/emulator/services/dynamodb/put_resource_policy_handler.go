package dynamodb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

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
	// ARN format: arn:aws:dynamodb:region:account:table/tablename or arn:aws:dynamodb:region:account:table/tablename/stream/...
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}

	// Determine if this is a table or stream
	resourceType := "table"
	resourceName := parts[1]

	if len(parts) >= 3 && parts[2] == "stream" {
		resourceType = "stream"
		resourceName = parts[1] + "/stream/" + strings.Join(parts[3:], "/")
	}

	// Verify the resource exists
	if resourceType == "table" {
		tableKey := fmt.Sprintf("dynamodb:table:%s", resourceName)
		var tableDesc map[string]interface{}
		if err := s.state.Get(tableKey, &tableDesc); err != nil {
			return s.errorResponse(400, "ResourceNotFoundException",
				fmt.Sprintf("Requested resource not found: %s", resourceArn)), nil
		}
	}

	// Check if a policy already exists
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var existingPolicy map[string]interface{}
	policyExists := s.state.Get(policyKey, &existingPolicy) == nil

	// If ExpectedRevisionId is provided, verify it matches
	if input.ExpectedRevisionId != nil && *input.ExpectedRevisionId != "" {
		if !policyExists {
			return s.errorResponse(400, "PolicyNotFoundException",
				"The resource policy does not exist"), nil
		}

		existingRevisionId, _ := existingPolicy["RevisionId"].(string)
		if existingRevisionId != *input.ExpectedRevisionId {
			return s.errorResponse(400, "PolicyRevisionMismatchException",
				"The policy revision ID does not match"), nil
		}
	}

	// Generate a new revision ID based on the policy content
	hash := sha256.Sum256([]byte(policy))
	revisionId := hex.EncodeToString(hash[:])[:16]

	// Store the policy
	policyData := map[string]interface{}{
		"ResourceArn": resourceArn,
		"Policy":      policy,
		"RevisionId":  revisionId,
	}

	if err := s.state.Set(policyKey, policyData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to store resource policy"), nil
	}

	// Return response with revision ID
	response := map[string]interface{}{
		"RevisionId": revisionId,
	}

	return s.jsonResponse(200, response)
}
