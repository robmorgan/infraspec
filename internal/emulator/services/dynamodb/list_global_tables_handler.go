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
		var gtData map[string]interface{}
		if err := s.state.Get(key, &gtData); err != nil {
			continue
		}

		// Filter by region if specified
		if input.RegionName != nil && *input.RegionName != "" {
			// Check if the global table has a replica in the specified region
			if replicas, ok := gtData["ReplicationGroup"].([]interface{}); ok {
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

		// Build global table
		gt := GlobalTable{}

		if globalTableName, ok := gtData["GlobalTableName"].(string); ok {
			gt.GlobalTableName = &globalTableName
		}

		// Add replication group
		if replicas, ok := gtData["ReplicationGroup"].([]interface{}); ok {
			replicationGroup := make([]Replica, 0, len(replicas))
			for _, replica := range replicas {
				if replicaMap, ok := replica.(map[string]interface{}); ok {
					r := Replica{}
					if regionName, ok := replicaMap["RegionName"].(string); ok {
						r.RegionName = &regionName
					}
					replicationGroup = append(replicationGroup, r)
				}
			}
			gt.ReplicationGroup = replicationGroup
		}

		globalTables = append(globalTables, gt)
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

	endIndex := startIndex + limit
	if endIndex > len(globalTables) {
		endIndex = len(globalTables)
	}

	paginatedTables := []GlobalTable{}
	if startIndex < len(globalTables) {
		paginatedTables = globalTables[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"GlobalTables": paginatedTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTables) {
		if paginatedTables[len(paginatedTables)-1].GlobalTableName != nil {
			response["LastEvaluatedGlobalTableName"] = *paginatedTables[len(paginatedTables)-1].GlobalTableName
		}
	}

	return s.jsonResponse(200, response)
}
