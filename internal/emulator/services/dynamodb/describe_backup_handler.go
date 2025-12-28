package dynamodb

import (
	"context"
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) describeBackup(ctx context.Context, input *DescribeBackupInput) (*emulator.AWSResponse, error) {
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

	var backupData map[string]interface{}
	found := false

	for _, key := range keys {
		var data map[string]interface{}
		if err := s.state.Get(key, &data); err == nil {
			if details, ok := data["BackupDetails"].(map[string]interface{}); ok {
				if arn, ok := details["BackupArn"].(string); ok && arn == backupArn {
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

	// Build backup description
	backupDetails, _ := backupData["BackupDetails"].(map[string]interface{})
	tableName, _ := backupData["TableName"].(string)
	tableId, _ := backupData["TableId"].(string)
	tableArn, _ := backupData["TableArn"].(string)

	backupDescription := map[string]interface{}{
		"BackupDetails": backupDetails,
		"SourceTableDetails": map[string]interface{}{
			"TableName": tableName,
			"TableId":   tableId,
			"TableArn":  tableArn,
		},
	}

	// Return response
	response := map[string]interface{}{
		"BackupDescription": backupDescription,
	}

	return s.jsonResponse(200, response)
}
