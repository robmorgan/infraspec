package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func TestDeleteBackup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a backup first
	backupArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/backup/test-backup-123"
	backupData := map[string]interface{}{
		"BackupDetails": map[string]interface{}{
			"BackupArn":    backupArn,
			"BackupName":   "test-backup",
			"BackupStatus": "AVAILABLE",
		},
		"TableName": "test-table",
	}
	state.Set("dynamodb:backup:test-table:test-backup", backupData)

	input := &DeleteBackupInput{
		BackupArn: strPtr(backupArn),
	}

	resp, err := service.deleteBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	backupDesc, ok := result["BackupDescription"].(map[string]interface{})
	require.True(t, ok)

	backupDetails, ok := backupDesc["BackupDetails"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "DELETED", backupDetails["BackupStatus"])
	assert.Equal(t, backupArn, backupDetails["BackupArn"])

	// Verify backup was deleted from state
	assert.False(t, state.Exists("dynamodb:backup:test-table:test-backup"))
}

func TestDeleteBackup_MissingBackupArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DeleteBackupInput{}

	resp, err := service.deleteBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "BackupArn is required")
}

func TestDeleteBackup_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DeleteBackupInput{
		BackupArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/backup/nonexistent"),
	}

	resp, err := service.deleteBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "BackupNotFoundException", result["__type"])
}
