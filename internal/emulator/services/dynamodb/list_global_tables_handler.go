package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This documentation is for version 2017.11.29 (Legacy) of global tables,
// which should be avoided for new global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global tables from state
	keys, err := s.state.List("dynamodb:globaltable:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	globalTables := []interface{}{}

	for _, key := range keys {
		var globalTableData map[string]interface{}
		if err := s.state.Get(key, &globalTableData); err != nil {
			continue
		}

		// Filter by region if specified
		if input.RegionName != nil && *input.RegionName != "" {
			// Check if this global table has a replica in the specified region
			if replicas, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
				hasRegion := false
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
				if !hasRegion {
					continue
				}
			} else {
				continue
			}
		}

		// Build global table entry
		entry := map[string]interface{}{}

		if globalTableName, ok := globalTableData["GlobalTableName"].(string); ok {
			entry["GlobalTableName"] = globalTableName
		}

		if replicationGroup, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
			entry["ReplicationGroup"] = replicationGroup
		}

		globalTables = append(globalTables, entry)
	}

	// Apply pagination if specified
	limit := 10 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		// Find the index of the exclusive start global table
		for i, entry := range globalTables {
			if entryMap, ok := entry.(map[string]interface{}); ok {
				if name, ok := entryMap["GlobalTableName"].(string); ok {
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

	paginatedTables := []interface{}{}
	if startIndex < len(globalTables) {
		paginatedTables = globalTables[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"GlobalTables": paginatedTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTables) {
		if lastTable, ok := paginatedTables[len(paginatedTables)-1].(map[string]interface{}); ok {
			if lastName, ok := lastTable["GlobalTableName"].(string); ok {
				response["LastEvaluatedGlobalTableName"] = lastName
			}
		}
	}

	return s.jsonResponse(200, response)
}

// Helper function to extract global table name from key
func extractGlobalTableNameFromKey(key string) string {
	// Key format: "dynamodb:globaltable:tablename"
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
