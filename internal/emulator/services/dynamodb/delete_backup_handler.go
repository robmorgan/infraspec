package dynamodb

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) deleteBackup(ctx context.Context, input *DeleteBackupInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.BackupArn == nil || *input.BackupArn == "" {
		return s.errorResponse(400, "ValidationException", "BackupArn is required"), nil
	}

	backupArn := *input.BackupArn

	// Find the backup by ARN
	// List all backups and find the one with matching ARN
	keys, err := s.state.List("dynamodb:backup:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list backups"), nil
	}

	var backupKey string
	var backupData map[string]interface{}
	found := false

	for _, key := range keys {
		var data map[string]interface{}
		if err := s.state.Get(key, &data); err == nil {
			if details, ok := data["BackupDetails"].(map[string]interface{}); ok {
				if arn, ok := details["BackupArn"].(string); ok && arn == backupArn {
					backupKey = key
					backupData = data
					found = true
					break
				}
			}
		}
	}

	if !found {
		return s.errorResponse(400, "BackupNotFoundException",
			fmt.Sprintf("Backup not found: %s", backupArn)), nil
	}

	// Extract backup details for response
	backupDetails, _ := backupData["BackupDetails"].(map[string]interface{})

	// Update backup status to DELETED
	if backupDetails != nil {
		backupDetails["BackupStatus"] = "DELETED"
	}

	// Delete backup from state
	if err := s.state.Delete(backupKey); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to delete backup"), nil
	}

	// Return response with backup description
	response := map[string]interface{}{
		"BackupDescription": map[string]interface{}{
			"BackupDetails": backupDetails,
		},
	}

	return s.jsonResponse(200, response)
}
