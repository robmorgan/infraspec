package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listGlobalTables lists all global tables that have a replica in the specified Region.
// This implementation supports the legacy 2017.11.29 global tables API.
func (s *DynamoDBService) listGlobalTables(ctx context.Context, input *ListGlobalTablesInput) (*emulator.AWSResponse, error) {
	// List all global table keys from state
	keys, err := s.state.List("dynamodb:globaltable:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list global tables"), nil
	}

	globalTables := []map[string]interface{}{}

	for _, key := range keys {
		var tableData map[string]interface{}
		if err := s.state.Get(key, &tableData); err != nil {
			continue
		}

		tableName, _ := tableData["GlobalTableName"].(string)
		if tableName == "" {
			// Fallback: extract from key
			parts := strings.Split(key, ":")
			if len(parts) >= 3 {
				tableName = strings.Join(parts[2:], ":")
			}
		}
		if tableName == "" {
			continue
		}

		// Build replica list
		replicationGroup := []map[string]interface{}{}
		if replicas, ok := tableData["ReplicationGroup"].([]interface{}); ok {
			for _, r := range replicas {
				if replica, ok := r.(map[string]interface{}); ok {
					replicationGroup = append(replicationGroup, replica)
				}
			}
		}

		entry := map[string]interface{}{
			"GlobalTableName":  tableName,
			"ReplicationGroup": replicationGroup,
		}
		globalTables = append(globalTables, entry)
	}

	// Apply pagination: ExclusiveStartGlobalTableName and Limit
	startIdx := 0
	if input.ExclusiveStartGlobalTableName != nil && *input.ExclusiveStartGlobalTableName != "" {
		for i, gt := range globalTables {
			if name, ok := gt["GlobalTableName"].(string); ok && name == *input.ExclusiveStartGlobalTableName {
				startIdx = i + 1
				break
			}
		}
	}

	limit := len(globalTables)
	if input.Limit != nil && int(*input.Limit) > 0 && int(*input.Limit) < limit {
		limit = int(*input.Limit)
	}

	endIdx := startIdx + limit
	if endIdx > len(globalTables) {
		endIdx = len(globalTables)
	}

	paginatedTables := []interface{}{}
	if startIdx < len(globalTables) {
		for _, gt := range globalTables[startIdx:endIdx] {
			paginatedTables = append(paginatedTables, gt)
		}
	}

	response := map[string]interface{}{
		"GlobalTables": paginatedTables,
	}

	// Add LastEvaluatedGlobalTableName if there are more results
	if endIdx < len(globalTables) && len(paginatedTables) > 0 {
		if last, ok := paginatedTables[len(paginatedTables)-1].(map[string]interface{}); ok {
			if lastName, ok := last["GlobalTableName"].(string); ok {
				response["LastEvaluatedGlobalTableName"] = lastName
			}
		}
	}

	return s.jsonResponse(200, response)
}
