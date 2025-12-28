package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

func TestCreateBackup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table first
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableId":   uuid.New().String(),
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Create a backup
	input := &CreateBackupInput{
		TableName:  stringPtr(tableName),
		BackupName: stringPtr("test-backup"),
	}

	resp, err := service.createBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseData map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseData)
	require.NoError(t, err)

	// Verify response structure
	backupDetails, ok := responseData["BackupDetails"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-backup", backupDetails["BackupName"])
	assert.Equal(t, "AVAILABLE", backupDetails["BackupStatus"])
	assert.Equal(t, "USER", backupDetails["BackupType"])
	assert.NotEmpty(t, backupDetails["BackupArn"])
	assert.NotNil(t, backupDetails["BackupCreationDateTime"])

	// Verify backup was stored in state
	backupKey := fmt.Sprintf("dynamodb:backup:%s:test-backup", tableName)
	assert.True(t, state.Exists(backupKey))
}

func TestCreateBackup_MultipleBackups(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a table
	tableName := "test-table"
	tableKey := fmt.Sprintf("dynamodb:table:%s", tableName)
	tableDesc := map[string]interface{}{
		"TableName": tableName,
		"TableId":   uuid.New().String(),
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
	}
	err := state.Set(tableKey, tableDesc)
	require.NoError(t, err)

	// Create first backup
	input1 := &CreateBackupInput{
		TableName:  stringPtr(tableName),
		BackupName: stringPtr("backup-1"),
	}
	resp, err := service.createBackup(context.Background(), input1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Create second backup
	input2 := &CreateBackupInput{
		TableName:  stringPtr(tableName),
		BackupName: stringPtr("backup-2"),
	}
	resp, err = service.createBackup(context.Background(), input2)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify both backups exist
	backupKey1 := fmt.Sprintf("dynamodb:backup:%s:backup-1", tableName)
	backupKey2 := fmt.Sprintf("dynamodb:backup:%s:backup-2", tableName)
	assert.True(t, state.Exists(backupKey1))
	assert.True(t, state.Exists(backupKey2))
}

func TestCreateBackup_MissingTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &CreateBackupInput{
		BackupName: stringPtr("test-backup"),
	}

	resp, err := service.createBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "TableName")
}

func TestCreateBackup_MissingBackupName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &CreateBackupInput{
		TableName: stringPtr("test-table"),
	}

	resp, err := service.createBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", errorData["__type"])
	assert.Contains(t, errorData["message"], "BackupName")
}

func TestCreateBackup_TableNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &CreateBackupInput{
		TableName:  stringPtr("nonexistent-table"),
		BackupName: stringPtr("test-backup"),
	}

	resp, err := service.createBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var errorData map[string]interface{}
	err = json.Unmarshal(resp.Body, &errorData)
	require.NoError(t, err)
	assert.Equal(t, "ResourceNotFoundException", errorData["__type"])
	assert.Contains(t, errorData["message"], "not found")
}
