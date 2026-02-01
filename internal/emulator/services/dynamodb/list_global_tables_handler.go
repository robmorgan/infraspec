package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This documentation is for version 2017.11.29 (Legacy) of global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global tables from state
	keys, err := s.state.List("dynamodb:globaltable:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	globalTableSummaries := []interface{}{}

	for _, key := range keys {
		var globalTableData map[string]interface{}
		if err := s.state.Get(key, &globalTableData); err != nil {
			continue
		}

		// Filter by region name if specified
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

		// Build global table summary
		summary := map[string]interface{}{}

		if globalTableName, ok := globalTableData["GlobalTableName"].(string); ok {
			summary["GlobalTableName"] = globalTableName
		}
		if replicationGroup, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
			summary["ReplicationGroup"] = replicationGroup
		}

		globalTableSummaries = append(globalTableSummaries, summary)
	}

	// Apply pagination if specified
	limit := 100 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		// Find the index of the exclusive start global table
		for i, summary := range globalTableSummaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				if name, ok := summaryMap["GlobalTableName"].(string); ok {
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
	if endIndex > len(globalTableSummaries) {
		endIndex = len(globalTableSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(globalTableSummaries) {
		paginatedSummaries = globalTableSummaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"GlobalTables": paginatedSummaries,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTableSummaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if lastName, ok := lastSummary["GlobalTableName"].(string); ok {
				response["LastEvaluatedGlobalTableName"] = lastName
			}
		}
	}

	return s.jsonResponse(200, response)
}
