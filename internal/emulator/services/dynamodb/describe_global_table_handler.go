package dynamodb

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeGlobalTable returns information about the specified global table.
// This is for the legacy (2017.11.29) version of global tables.
func (s *DynamoDBService) describeGlobalTable(ctx context.Context, input *DescribeGlobalTableInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.GlobalTableName == nil || *input.GlobalTableName == "" {
		return s.errorResponse(400, "ValidationException", "GlobalTableName is required"), nil
	}

	globalTableName := *input.GlobalTableName

	// Check if global table exists in state
	stateKey := fmt.Sprintf("dynamodb:globaltable:%s", globalTableName)
	var globalTableDesc map[string]interface{}
	if err := s.state.Get(stateKey, &globalTableDesc); err != nil {
		return s.errorResponse(400, "GlobalTableNotFoundException", fmt.Sprintf("Global table not found: %s", globalTableName)), nil
	}

	// Return global table description
	response := map[string]interface{}{
		"GlobalTableDescription": globalTableDesc,
	}

	return s.jsonResponse(200, response)
}
