package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This is for version 2017.11.29 (Legacy) of global tables.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global table keys from state
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

		globalTableName, _ := globalTableData["GlobalTableName"].(string)
		if globalTableName == "" {
			// Derive from state key: dynamodb:globaltable:<name>
			globalTableName = strings.TrimPrefix(key, "dynamodb:globaltable:")
		}

		// Filter by RegionName if specified
		if input.RegionName != nil && *input.RegionName != "" {
			regionName := *input.RegionName
			found := false
			if replicas, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
				for _, replica := range replicas {
					if replicaMap, ok := replica.(map[string]interface{}); ok {
						if region, ok := replicaMap["RegionName"].(string); ok && region == regionName {
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

		// Build ReplicationGroup
		replicationGroup := []interface{}{}
		if replicas, ok := globalTableData["ReplicationGroup"].([]interface{}); ok {
			for _, replica := range replicas {
				if replicaMap, ok := replica.(map[string]interface{}); ok {
					entry := map[string]interface{}{}
					if regionName, ok := replicaMap["RegionName"].(string); ok {
						entry["RegionName"] = regionName
					}
					replicationGroup = append(replicationGroup, entry)
				}
			}
		}

		tableEntry := map[string]interface{}{
			"GlobalTableName":  globalTableName,
			"ReplicationGroup": replicationGroup,
		}
		globalTables = append(globalTables, tableEntry)
	}

	// Apply ExclusiveStartGlobalTableName for pagination
	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		startName := *input.ExclusiveStartGlobalTableName
		for i, table := range globalTables {
			if tableMap, ok := table.(map[string]interface{}); ok {
				if name, ok := tableMap["GlobalTableName"].(string); ok && name == startName {
					startIndex = i + 1
					break
				}
			}
		}
	}

	// Apply Limit
	limit := len(globalTables)
	if input.Limit != nil && *input.Limit > 0 && int(*input.Limit) < limit-startIndex {
		limit = startIndex + int(*input.Limit)
	}

	endIndex := limit
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
