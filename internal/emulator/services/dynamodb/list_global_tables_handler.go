package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This documentation is for version 2017.11.29 (Legacy) of global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global tables from state
	keys, err := s.state.List("dynamodb:global-table:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	globalTables := []GlobalTable{}

	for _, key := range keys {
		var globalTableData map[string]interface{}
		if err := s.state.Get(key, &globalTableData); err != nil {
			continue
		}

		// Filter by region name if specified
		if input.RegionName != nil && *input.RegionName != "" {
			// Check if this global table has a replica in the specified region
			if replicationGroup, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
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
			} else {
				continue
			}
		}

		// Build global table summary
		globalTable := GlobalTable{}
		if globalTableName, ok := globalTableData["GlobalTableName"].(string); ok {
			globalTable.GlobalTableName = &globalTableName
		}
		if replicationGroup, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
			replicas := []Replica{}
			for _, replica := range replicationGroup {
				if replicaMap, ok := replica.(map[string]interface{}); ok {
					r := Replica{}
					if regionName, ok := replicaMap["RegionName"].(string); ok {
						r.RegionName = &regionName
					}
					replicas = append(replicas, r)
				}
			}
			globalTable.ReplicationGroup = replicas
		}

		globalTables = append(globalTables, globalTable)
	}

	// Apply pagination
	limit := 100 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		// Find the index of the exclusive start global table
		for i, gt := range globalTables {
			if gt.GlobalTableName != nil && *gt.GlobalTableName == *input.ExclusiveStartGlobalTableName {
				startIndex = i + 1
				break
			}
		}
	}

	// Apply pagination
	endIndex := startIndex + limit
	if endIndex > len(globalTables) {
		endIndex = len(globalTables)
	}

	paginatedTables := []GlobalTable{}
	if startIndex < len(globalTables) {
		paginatedTables = globalTables[startIndex:endIndex]
	}

	// Build response
	response := ListGlobalTablesOutput{
		GlobalTables: paginatedTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTables) {
		if len(paginatedTables) > 0 {
			lastTable := paginatedTables[len(paginatedTables)-1]
			if lastTable.GlobalTableName != nil {
				response.LastEvaluatedGlobalTableName = lastTable.GlobalTableName
			}
		}
	}

	return s.jsonResponse(200, response)
}
