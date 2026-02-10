package dynamodb

import (
	"context"
	"sort"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global tables
	prefix := "dynamodb:globaltable:"
	keys, err := s.state.List(prefix)
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	// Build list of global tables
	globalTables := []map[string]interface{}{}
	for _, key := range keys {
		var globalTableDesc map[string]interface{}
		if err := s.state.Get(key, &globalTableDesc); err != nil {
			continue
		}

		// Filter by region if specified
		if input.RegionName != nil && *input.RegionName != "" {
			replicationGroup, ok := globalTableDesc["ReplicationGroup"].([]interface{})
			if !ok {
				continue
			}

			// Check if any replica is in the specified region
			hasRegion := false
			for _, replica := range replicationGroup {
				if replicaMap, ok := replica.(map[string]interface{}); ok {
					if regionName, ok := replicaMap["RegionName"].(string); ok {
						if regionName == *input.RegionName {
							hasRegion = true
							break
						}
					}
				}
			}
			if !hasRegion {
				continue
			}
		}

		// Build global table entry
		globalTable := map[string]interface{}{}

		if globalTableName, ok := globalTableDesc["GlobalTableName"].(string); ok {
			globalTable["GlobalTableName"] = globalTableName
		}

		if replicationGroup, ok := globalTableDesc["ReplicationGroup"].([]interface{}); ok {
			globalTable["ReplicationGroup"] = replicationGroup
		}

		globalTables = append(globalTables, globalTable)
	}

	// Sort by GlobalTableName for consistent ordering
	sort.Slice(globalTables, func(i, j int) bool {
		name1, _ := globalTables[i]["GlobalTableName"].(string)
		name2, _ := globalTables[j]["GlobalTableName"].(string)
		return name1 < name2
	})

	// Apply pagination
	limit := 100 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	// Handle ExclusiveStartGlobalTableName for pagination
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		for i, gt := range globalTables {
			if name, ok := gt["GlobalTableName"].(string); ok {
				if name == *input.ExclusiveStartGlobalTableName {
					startIndex = i + 1
					break
				}
			}
		}
	}

	endIndex := startIndex + limit
	if endIndex > len(globalTables) {
		endIndex = len(globalTables)
	}

	paginatedGlobalTables := globalTables[startIndex:endIndex]

	response := map[string]interface{}{
		"GlobalTables": paginatedGlobalTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTables) {
		if name, ok := paginatedGlobalTables[len(paginatedGlobalTables)-1]["GlobalTableName"].(string); ok {
			response["LastEvaluatedGlobalTableName"] = name
		}
	}

	return s.jsonResponse(200, response)
}
