package dynamodb

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// describeGlobalTableSettings describes Region-specific settings for a global table.
// This is for the legacy (2017.11.29) version of global tables.
func (s *DynamoDBService) describeGlobalTableSettings(ctx context.Context, input *DescribeGlobalTableSettingsInput) (*emulator.AWSResponse, error) {
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

	// Build global table settings response
	// Include the global table name and replica settings
	response := map[string]interface{}{
		"GlobalTableName": globalTableName,
	}

	// Add replica settings if they exist
	if replicationGroup, ok := globalTableDesc["ReplicationGroup"].([]interface{}); ok {
		replicaSettings := []map[string]interface{}{}
		for _, replica := range replicationGroup {
			if replicaMap, ok := replica.(map[string]interface{}); ok {
				replicaSetting := map[string]interface{}{}
				if regionName, ok := replicaMap["RegionName"].(string); ok {
					replicaSetting["RegionName"] = regionName
				}
				// Add default replica settings
				replicaSetting["ReplicaStatus"] = "ACTIVE"
				replicaSettings = append(replicaSettings, replicaSetting)
			}
		}
		response["ReplicaSettings"] = replicaSettings
	}

	return s.jsonResponse(200, response)
}
