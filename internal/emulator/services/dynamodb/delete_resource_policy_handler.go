package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) deleteResourcePolicy(ctx context.Context, input *DeleteResourcePolicyInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.ResourceArn == nil || *input.ResourceArn == "" {
		return s.errorResponse(400, "ValidationException", "ResourceArn is required"), nil
	}

	resourceArn := *input.ResourceArn

	// Extract resource name from ARN
	// ARN format: arn:aws:dynamodb:us-east-1:000000000000:table/tablename
	// or: arn:aws:dynamodb:us-east-1:000000000000:table/tablename/stream/streamlabel
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}
	resourceName := parts[1]

	// State key for resource policy
	policyKey := fmt.Sprintf("dynamodb:resource-policy:%s", resourceName)

	// DeleteResourcePolicy is idempotent - it doesn't fail if the policy doesn't exist
	// Just delete if it exists
	s.state.Delete(policyKey)

	// Return empty response (successful deletion)
	response := map[string]interface{}{
		"RevisionId": "",
	}

	return s.jsonResponse(200, response)
}
