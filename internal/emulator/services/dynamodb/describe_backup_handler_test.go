package dynamodb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeBackup_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create a backup first
	backupArn := "arn:aws:dynamodb:us-east-1:000000000000:table/test-table/backup/test-backup-123"
	backupData := map[string]interface{}{
		"BackupDetails": map[string]interface{}{
			"BackupArn":              backupArn,
			"BackupName":             "test-backup",
			"BackupStatus":           "AVAILABLE",
			"BackupType":             "USER",
			"BackupSizeBytes":        1024,
			"BackupCreationDateTime": float64(1234567890),
		},
		"TableName": "test-table",
		"TableId":   "table-id-123",
		"TableArn":  "arn:aws:dynamodb:us-east-1:000000000000:table/test-table",
	}
	state.Set("dynamodb:backup:test-table:test-backup", backupData)

	input := &DescribeBackupInput{
		BackupArn: strPtr(backupArn),
	}

	resp, err := service.describeBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)

	backupDesc, ok := result["BackupDescription"].(map[string]interface{})
	require.True(t, ok)

	// Verify backup details
	backupDetails, ok := backupDesc["BackupDetails"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, backupArn, backupDetails["BackupArn"])
	assert.Equal(t, "test-backup", backupDetails["BackupName"])
	assert.Equal(t, "AVAILABLE", backupDetails["BackupStatus"])

	// Verify source table details
	sourceTableDetails, ok := backupDesc["SourceTableDetails"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-table", sourceTableDetails["TableName"])
	assert.Equal(t, "table-id-123", sourceTableDetails["TableId"])
}

func TestDescribeBackup_MissingBackupArn(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DescribeBackupInput{}

	resp, err := service.describeBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "ValidationException", result["__type"])
	assert.Contains(t, result["message"], "BackupArn is required")
}

func TestDescribeBackup_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	input := &DescribeBackupInput{
		BackupArn: strPtr("arn:aws:dynamodb:us-east-1:000000000000:table/test-table/backup/nonexistent"),
	}

	resp, err := service.describeBackup(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body, &result)
	require.NoError(t, err)
	assert.Equal(t, "BackupNotFoundException", result["__type"])
}
