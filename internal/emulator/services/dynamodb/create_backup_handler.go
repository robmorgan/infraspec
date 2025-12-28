package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *DynamoDBService) createBackup(ctx context.Context, input *CreateBackupInput) (*emulator.AWSResponse, error) {
	// Validate required parameters
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}

	if input.BackupName == nil || *input.BackupName == "" {
		return s.errorResponse(400, "ValidationException", "BackupName is required"), nil
	}

	tableName := *input.TableName
	backupName := *input.BackupName

	// Verify table exists
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(tableKey, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException",
			fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Create backup details
	now := time.Now().Unix()
	backupArn := fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/backup/%s", tableName, uuid.New().String())

	backupDetails := map[string]interface{}{
		"BackupArn":              backupArn,
		"BackupName":             backupName,
		"BackupSizeBytes":        0,
		"BackupStatus":           "AVAILABLE",
		"BackupType":             "USER",
		"BackupCreationDateTime": float64(now),
	}

	// Store backup in state
	backupKey := fmt.Sprintf("dynamodb:backup:%s:%s", tableName, backupName)
	backupData := map[string]interface{}{
		"BackupDetails": backupDetails,
		"TableName":     tableName,
		"TableId":       tableDesc["TableId"],
		"TableArn":      tableDesc["TableArn"],
	}

	if err := s.state.Set(backupKey, backupData); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to create backup"), nil
	}

	// Return response
	response := map[string]interface{}{
		"BackupDetails": backupDetails,
	}

	return s.jsonResponse(200, response)
}
