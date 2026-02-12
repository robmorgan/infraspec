package dynamodb

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeTableReplicaAutoScaling describes auto scaling settings across replicas of a global table.
// This operation returns auto scaling configuration for the table and its global secondary indexes.
func (s *DynamoDBService) describeTableReplicaAutoScaling(ctx context.Context, input *DescribeTableReplicaAutoScalingInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}

	tableName := *input.TableName

	// Verify table exists
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(tableKey, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException",
			fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Get auto scaling settings for this table
	// In the emulator, we'll store auto scaling settings separately
	autoScalingKey := fmt.Sprintf("dynamodb:auto-scaling:%s", tableName)
	var autoScalingDesc map[string]interface{}
	if err := s.state.Get(autoScalingKey, &autoScalingDesc); err != nil {
		// If no auto scaling settings exist, return default/disabled configuration
		tableArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName)
		if arn, ok := tableDesc["TableArn"].(string); ok {
			tableArn = arn
		}

		autoScalingDesc = map[string]interface{}{
			"TableName": tableName,
			"TableArn":  tableArn,
			"Replicas":  []interface{}{},
		}
	}

	// Return response
	response := map[string]interface{}{
		"TableAutoScalingDescription": autoScalingDesc,
	}

	return s.jsonResponse(200, response)
}
