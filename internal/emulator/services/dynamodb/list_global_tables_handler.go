package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This is for the legacy (2017.11.29) version of global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global table keys from state
	keys, err := s.state.List("dynamodb:globaltable:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	// Collect all global tables, applying region filter
	allTables := []map[string]interface{}{}

	for _, key := range keys {
		var globalTableDesc map[string]interface{}
		if err := s.state.Get(key, &globalTableDesc); err != nil {
			continue
		}

		// Filter by region name if specified
		if input.RegionName != nil && *input.RegionName != "" {
			hasRegion := false
			if replicas, ok := globalTableDesc["ReplicationGroup"].([]interface{}); ok {
				for _, replica := range replicas {
					if replicaMap, ok := replica.(map[string]interface{}); ok {
						if regionName, ok := replicaMap["RegionName"].(string); ok {
							if regionName == *input.RegionName {
								hasRegion = true
								break
							}
						}
					}
				}
			}
			if !hasRegion {
				continue
			}
		}

		// Extract global table name from key
		// Key format: "dynamodb:globaltable:<globalTableName>"
		parts := strings.SplitN(key, ":", 3)
		if len(parts) < 3 {
			continue
		}
		globalTableName := parts[2]

		allTables = append(allTables, map[string]interface{}{
			"GlobalTableName":  globalTableName,
			"ReplicationGroup": globalTableDesc["ReplicationGroup"],
		})
	}

	// Apply pagination using ExclusiveStartGlobalTableName
	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		for i, t := range allTables {
			if name, ok := t["GlobalTableName"].(string); ok && name == *input.ExclusiveStartGlobalTableName {
				startIndex = i + 1
				break
			}
		}
	}

	// Apply limit
	limit := len(allTables)
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	endIndex := startIndex + limit
	if endIndex > len(allTables) {
		endIndex = len(allTables)
	}

	var paginatedTables []interface{}
	if startIndex < len(allTables) {
		for _, t := range allTables[startIndex:endIndex] {
			paginatedTables = append(paginatedTables, t)
		}
	}
	if paginatedTables == nil {
		paginatedTables = []interface{}{}
	}

	response := map[string]interface{}{
		"GlobalTables": paginatedTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(allTables) && len(paginatedTables) > 0 {
		if lastTable, ok := paginatedTables[len(paginatedTables)-1].(map[string]interface{}); ok {
			if name, ok := lastTable["GlobalTableName"].(string); ok {
				response["LastEvaluatedGlobalTableName"] = name
			}
		}
	}

	return s.jsonResponse(200, response)
}
