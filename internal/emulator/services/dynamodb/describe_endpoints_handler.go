package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeEndpoints returns the regional endpoint information.
// This is a simple operation that doesn't require any parameters and always returns
// the same endpoint information for the emulator.
// Note: DescribeEndpoints has no input parameters in the AWS API.
func (s *DynamoDBService) describeEndpoints(ctx context.Context) (*emulator.AWSResponse, error) {
	// DescribeEndpoints returns endpoint information for the DynamoDB service
	// For the emulator, we return a static endpoint
	response := map[string]interface{}{
		"Endpoints": []map[string]interface{}{
			{
				"Address":              "localhost:3687",
				"CachePeriodInMinutes": 1440,
			},
		},
	}

	return s.jsonResponse(200, response)
}
