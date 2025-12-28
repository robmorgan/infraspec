package dynamodb

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// getResourcePolicy returns the resource-based policy document attached to a resource.
// The resource can be a table or stream, and the policy is returned in JSON format.
// This operation follows an eventually consistent model.
func (s *DynamoDBService) getResourcePolicy(ctx context.Context, input *GetResourcePolicyInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.ResourceArn == nil || *input.ResourceArn == "" {
		return s.errorResponse(400, "ValidationException", "ResourceArn is required"), nil
	}

	resourceArn := *input.ResourceArn

	// Get resource policy from state
	// We store policies by ARN for tables and streams
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceArn)
	var policyData map[string]interface{}
	if err := s.state.Get(policyKey, &policyData); err != nil {
		// If no policy exists, return an empty policy response
		// AWS returns a successful response with no policy if none is attached
		response := map[string]interface{}{
			"ResourceArn": resourceArn,
		}
		return s.jsonResponse(200, response)
	}

	// Return the policy
	response := map[string]interface{}{
		"Policy":      policyData["Policy"],
		"ResourceArn": resourceArn,
	}

	// Include RevisionId if present
	if revisionId, ok := policyData["RevisionId"].(string); ok {
		response["RevisionId"] = revisionId
	}

	return s.jsonResponse(200, response)
}
