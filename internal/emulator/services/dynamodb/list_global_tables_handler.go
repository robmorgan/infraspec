package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This is for version 2017.11.29 (Legacy) of global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// Optional region name filter
	var regionName string
	if input.RegionName != nil {
		regionName = *input.RegionName
	}

	// List all global tables from state
	keys, err := s.state.List("dynamodb:global-table:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	var globalTables []interface{}

	for _, key := range keys {
		var globalTableData map[string]interface{}
		if err := s.state.Get(key, &globalTableData); err != nil {
			continue
		}

		// Filter by region if specified
		if regionName != "" {
			// Check if the global table has a replica in the specified region
			if replicas, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
				hasRegion := false
				for _, replica := range replicas {
					if replicaMap, ok := replica.(map[string]interface{}); ok {
						if replicaRegion, ok := replicaMap["RegionName"].(string); ok {
							if replicaRegion == regionName {
								hasRegion = true
								break
							}
						}
					}
				}
				if !hasRegion {
					continue
				}
			} else {
				continue
			}
		}

		// Build global table summary
		globalTable := map[string]interface{}{}

		if globalTableName, ok := globalTableData["GlobalTableName"].(string); ok {
			globalTable["GlobalTableName"] = globalTableName
		}

		if replicationGroup, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
			globalTable["ReplicationGroup"] = replicationGroup
		}

		globalTables = append(globalTables, globalTable)
	}

	// Apply pagination if specified
	limit := 100 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		// Find the index of the exclusive start global table
		for i, gt := range globalTables {
			if gtMap, ok := gt.(map[string]interface{}); ok {
				if name, ok := gtMap["GlobalTableName"].(string); ok {
					if name == *input.ExclusiveStartGlobalTableName {
						startIndex = i + 1
						break
					}
				}
			}
		}
	}

	// Apply pagination
	endIndex := startIndex + limit
	if endIndex > len(globalTables) {
		endIndex = len(globalTables)
	}

	paginatedGlobalTables := []interface{}{}
	if startIndex < len(globalTables) {
		paginatedGlobalTables = globalTables[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"GlobalTables": paginatedGlobalTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTables) {
		if lastGT, ok := paginatedGlobalTables[len(paginatedGlobalTables)-1].(map[string]interface{}); ok {
			if lastName, ok := lastGT["GlobalTableName"].(string); ok {
				response["LastEvaluatedGlobalTableName"] = lastName
			}
		}
	}

	return s.jsonResponse(200, response)
}

// Helper function to extract global table name from key
func extractGlobalTableNameFromKey(key string) string {
	// Key format: "dynamodb:global-table:tablename"
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		return strings.Join(parts[2:], ":")
	}
	return ""
}
