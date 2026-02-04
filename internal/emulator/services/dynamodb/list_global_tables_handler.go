package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
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

		globalTableName, _ := globalTableDesc["GlobalTableName"].(string)

		// Filter by RegionName if specified: only include tables that have a replica in the given region
		if input.RegionName != nil && *input.RegionName != "" {
			regionFound := false
			if replicas, ok := globalTableDesc["ReplicationGroup"].([]interface{}); ok {
				for _, replica := range replicas {
					if replicaMap, ok := replica.(map[string]interface{}); ok {
						if regionName, ok := replicaMap["RegionName"].(string); ok && regionName == *input.RegionName {
							regionFound = true
							break
						}
					}
				}
			}
			if !regionFound {
				continue
			}
		}

		// Build replication group for summary
		replicationGroup := []interface{}{}
		if replicas, ok := globalTableDesc["ReplicationGroup"].([]interface{}); ok {
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

		globalTables = append(globalTables, map[string]interface{}{
			"GlobalTableName":  globalTableName,
			"ReplicationGroup": replicationGroup,
		})
	}

	// Apply pagination using ExclusiveStartGlobalTableName
	startIndex := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		startName := *input.ExclusiveStartGlobalTableName
		for i, gt := range globalTables {
			if gtMap, ok := gt.(map[string]interface{}); ok {
				if name, ok := gtMap["GlobalTableName"].(string); ok {
					if strings.Compare(name, startName) > 0 {
						startIndex = i
						break
					}
					// If we reach the end without finding a name > startName, start past the end
					if i == len(globalTables)-1 {
						startIndex = len(globalTables)
					}
				}
			}
		}
	}

	limit := 100
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
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
