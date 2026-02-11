package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This documentation is for version 2017.11.29 (Legacy) of global tables, which should be
// avoided for new global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global tables from state
	keys, err := s.state.List("dynamodb:global-table:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	globalTableNames := []string{}

	for _, key := range keys {
		var globalTableData map[string]interface{}
		if err := s.state.Get(key, &globalTableData); err != nil {
			continue
		}

		// Filter by RegionName if specified
		if input.RegionName != nil && *input.RegionName != "" {
			// Check if global table has a replica in the specified region
			replicas, ok := globalTableData["ReplicationGroup"].([]interface{})
			if !ok {
				continue
			}

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
		}

		// Extract global table name
		if globalTableName, ok := globalTableData["GlobalTableName"].(string); ok {
			globalTableNames = append(globalTableNames, globalTableName)
		}
	}

	// Sort for consistent ordering
	// Note: In production, we'd want to sort these alphabetically

	// Apply pagination
	limit := 100 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		// Find the index of the exclusive start table
		for i, name := range globalTableNames {
			if name == *input.ExclusiveStartGlobalTableName {
				startIndex = i + 1
				break
			}
		}
	}

	endIndex := startIndex + limit
	if endIndex > len(globalTableNames) {
		endIndex = len(globalTableNames)
	}

	paginatedNames := []string{}
	if startIndex < len(globalTableNames) {
		paginatedNames = globalTableNames[startIndex:endIndex]
	}

	// Build global table list
	globalTables := []interface{}{}
	for _, name := range paginatedNames {
		globalTables = append(globalTables, map[string]interface{}{
			"GlobalTableName": name,
			"ReplicationGroup": []interface{}{
				map[string]interface{}{
					"RegionName": "us-east-1",
				},
			},
		})
	}

	// Build response
	response := map[string]interface{}{
		"GlobalTables": globalTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTableNames) {
		response["LastEvaluatedGlobalTableName"] = paginatedNames[len(paginatedNames)-1]
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
