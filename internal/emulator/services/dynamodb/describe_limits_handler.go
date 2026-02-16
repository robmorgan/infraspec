package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeLimits returns the current provisioned-capacity quotas for the AWS account in a Region.
// This operation returns both Region-level quotas and any per-table quotas.
// In the emulator, we return static default values that match typical AWS account limits.
// Note: DescribeLimits has no input parameters in the AWS API.
func (s *DynamoDBService) describeLimits(ctx context.Context) (*emulator.AWSResponse, error) {
	// Return standard DynamoDB account limits
	// These are typical default limits for a new AWS account
	response := map[string]interface{}{
		"AccountMaxReadCapacityUnits":  80000,
		"AccountMaxWriteCapacityUnits": 80000,
		"TableMaxReadCapacityUnits":    40000,
		"TableMaxWriteCapacityUnits":   40000,
	}

	return s.jsonResponse(200, response)
}
