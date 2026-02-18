package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This is for version 2017.11.29 (Legacy) of global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global tables from state
	keys, err := s.state.List("dynamodb:globaltable:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	globalTables := []interface{}{}

	for _, key := range keys {
		var globalTableDesc map[string]interface{}
		if err := s.state.Get(key, &globalTableDesc); err != nil {
			continue
		}

		// Filter by RegionName if specified
		if input.RegionName != nil && *input.RegionName != "" {
			regionName := *input.RegionName
			found := false
			if replicas, ok := globalTableDesc["ReplicationGroup"].([]interface{}); ok {
				for _, replica := range replicas {
					if replicaMap, ok := replica.(map[string]interface{}); ok {
						if rn, ok := replicaMap["RegionName"].(string); ok && rn == regionName {
							found = true
							break
						}
					}
				}
			}
			if !found {
				continue
			}
		}

		// Build global table summary
		globalTableName, _ := globalTableDesc["GlobalTableName"].(string)
		summary := map[string]interface{}{
			"GlobalTableName": globalTableName,
		}

		// Include replication group
		if replicationGroup, ok := globalTableDesc["ReplicationGroup"]; ok {
			summary["ReplicationGroup"] = replicationGroup
		} else {
			summary["ReplicationGroup"] = []interface{}{}
		}

		globalTables = append(globalTables, summary)
	}

	// Apply Limit and ExclusiveStartGlobalTableName for pagination
	limit := 100
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		startName := *input.ExclusiveStartGlobalTableName
		for i, table := range globalTables {
			if tableMap, ok := table.(map[string]interface{}); ok {
				if name, ok := tableMap["GlobalTableName"].(string); ok {
					if name == startName {
						startIndex = i + 1
						break
					}
				}
			}
		}
	}

	endIndex := startIndex + limit
	if endIndex > len(globalTables) {
		endIndex = len(globalTables)
	}

	paginatedTables := []interface{}{}
	if startIndex < len(globalTables) {
		paginatedTables = globalTables[startIndex:endIndex]
	}

	response := map[string]interface{}{
		"GlobalTables": paginatedTables,
	}

	// Set LastEvaluatedGlobalTableName if there are more results
	if endIndex < len(globalTables) {
		if lastTable, ok := paginatedTables[len(paginatedTables)-1].(map[string]interface{}); ok {
			if lastName, ok := lastTable["GlobalTableName"].(string); ok {
				response["LastEvaluatedGlobalTableName"] = lastName
			}
		}
	}

	return s.jsonResponse(200, response)
}
