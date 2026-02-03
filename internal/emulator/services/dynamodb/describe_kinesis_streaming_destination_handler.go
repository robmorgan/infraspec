package dynamodb

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeKinesisStreamingDestination returns information about the status of Kinesis streaming.
// This operation describes the current Kinesis Data Streams replication for a table.
func (s *DynamoDBService) describeKinesisStreamingDestination(ctx context.Context, input *DescribeKinesisStreamingDestinationInput) (*emulator.AWSResponse, error) {
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

	// Get Kinesis streaming destinations for this table
	// In the emulator, we'll store Kinesis streaming destinations separately
	kinesisKey := fmt.Sprintf("dynamodb:kinesis-destinations:%s", tableName)
	var kinesisDestinations []interface{}
	if err := s.state.Get(kinesisKey, &kinesisDestinations); err != nil {
		// If no destinations exist, return empty array
		kinesisDestinations = []interface{}{}
	}

	// Return response with table name and Kinesis streaming destinations
	response := map[string]interface{}{
		"TableName":                     tableName,
		"KinesisDataStreamDestinations": kinesisDestinations,
	}

	return s.jsonResponse(200, response)
}
