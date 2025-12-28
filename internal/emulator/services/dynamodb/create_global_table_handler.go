package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) createGlobalTable(ctx context.Context, input *CreateGlobalTableInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.GlobalTableName == nil || *input.GlobalTableName == "" {
		return s.errorResponse(400, "ValidationException", "GlobalTableName is required"), nil
	}

	if len(input.ReplicationGroup) == 0 {
		return s.errorResponse(400, "ValidationException", "ReplicationGroup is required and must contain at least one region"), nil
	}

	globalTableName := *input.GlobalTableName

	// Check if global table already exists
	globalTableKey := fmt.Sprintf("dynamodb:globaltable:%s", globalTableName)
	if s.state.Exists(globalTableKey) {
		return s.errorResponse(400, "GlobalTableAlreadyExistsException",
			fmt.Sprintf("Global Table already exists: %s", globalTableName)), nil
	}

	// Build replicas from replication group
	now := time.Now().Unix()
	replicas := []map[string]interface{}{}

	for _, replica := range input.ReplicationGroup {
		if replica.RegionName != nil {
			regionName := *replica.RegionName
			replicaEntry := map[string]interface{}{
				"RegionName":      regionName,
				"ReplicaStatus":   "ACTIVE",
				"ReplicaTableArn": fmt.Sprintf("arn:aws:dynamodb:%s:000000000000:table/%s", regionName, globalTableName),
			}
			replicas = append(replicas, replicaEntry)
		}
	}

	// Create global table description
	globalTableDesc := map[string]interface{}{
		"GlobalTableName":   globalTableName,
		"GlobalTableArn":    fmt.Sprintf("arn:aws:dynamodb::000000000000:global-table/%s", globalTableName),
		"GlobalTableStatus": "ACTIVE",
		"CreationDateTime":  float64(now),
		"ReplicationGroup":  replicas,
	}

	// Store global table in state
	if err := s.state.Set(globalTableKey, globalTableDesc); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to create global table"), nil
	}

	// Return response
	response := map[string]interface{}{
		"GlobalTableDescription": globalTableDesc,
	}

	return s.jsonResponse(200, response)
}
