package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This implementation is for version 2017.11.29 (Legacy) of global tables.
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

		// Filter by region if specified
		if input.RegionName != nil && *input.RegionName != "" {
			// Check if the global table has a replica in the specified region
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

		// Extract global table name from key
		parts := strings.Split(key, ":")
		if len(parts) >= 3 {
			globalTableName := strings.Join(parts[2:], ":")
			globalTableNames = append(globalTableNames, globalTableName)
		}
	}

	// Apply pagination if specified
	limit := 100 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		// Find the index of the exclusive start global table
		for i, name := range globalTableNames {
			if name == *input.ExclusiveStartGlobalTableName {
				startIndex = i + 1
				break
			}
		}
	}

	// Apply pagination
	endIndex := startIndex + limit
	if endIndex > len(globalTableNames) {
		endIndex = len(globalTableNames)
	}

	paginatedNames := []string{}
	if startIndex < len(globalTableNames) {
		paginatedNames = globalTableNames[startIndex:endIndex]
	}

	// Build global tables list with replica information
	globalTables := []interface{}{}
	for _, name := range paginatedNames {
		key := "dynamodb:global-table:" + name
		var globalTableData map[string]interface{}
		if err := s.state.Get(key, &globalTableData); err != nil {
			continue
		}

		globalTable := map[string]interface{}{
			"GlobalTableName": name,
		}

		// Add replication group if present
		if replicas, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
			globalTable["ReplicationGroup"] = replicas
		}

		globalTables = append(globalTables, globalTable)
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
