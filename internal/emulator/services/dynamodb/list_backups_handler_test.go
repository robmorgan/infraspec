package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListBackups_EmptyList(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "DynamoDB_20120810.ListBackups",
		},
		Body:   []byte("{}"),
		Action: "ListBackups",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-amz-json-1.0", resp.Headers["Content-Type"])

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, responseBody, "BackupSummaries")
	summaries, ok := responseBody["BackupSummaries"].([]interface{})
	require.True(t, ok, "BackupSummaries should be an array")
	assert.Empty(t, summaries, "Should have no backups initially")
}

func TestListBackups_WithBackups(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create test backups
	tableName := "test-table"
	now := time.Now().Unix()

	backup1Key := fmt.Sprintf("dynamodb:backup:%s:backup1", tableName)
	backup1Data := map[string]interface{}{
		"TableName": tableName,
		"TableId":   "table-id-123",
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"BackupDetails": map[string]interface{}{
			"BackupArn":              fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/backup/backup1", tableName),
			"BackupName":             "backup1",
			"BackupStatus":           "AVAILABLE",
			"BackupType":             "USER",
			"BackupSizeBytes":        1024,
			"BackupCreationDateTime": float64(now),
		},
	}
	err := state.Set(backup1Key, backup1Data)
	require.NoError(t, err)

	backup2Key := fmt.Sprintf("dynamodb:backup:%s:backup2", tableName)
	backup2Data := map[string]interface{}{
		"TableName": tableName,
		"TableId":   "table-id-123",
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"BackupDetails": map[string]interface{}{
			"BackupArn":              fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/backup/backup2", tableName),
			"BackupName":             "backup2",
			"BackupStatus":           "AVAILABLE",
			"BackupType":             "USER",
			"BackupSizeBytes":        2048,
			"BackupCreationDateTime": float64(now + 100),
		},
	}
	err = state.Set(backup2Key, backup2Data)
	require.NoError(t, err)

	// List all backups
	input := &ListBackupsInput{}

	resp, err := service.listBackups(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	// Verify response structure
	summaries, ok := responseBody["BackupSummaries"].([]interface{})
	require.True(t, ok, "BackupSummaries should be an array")
	assert.Len(t, summaries, 2, "Should have two backups")

	// Verify backup summaries contain expected fields
	for _, summary := range summaries {
		summaryMap, ok := summary.(map[string]interface{})
		require.True(t, ok, "Each summary should be an object")
		assert.Contains(t, summaryMap, "BackupArn")
		assert.Contains(t, summaryMap, "BackupName")
		assert.Contains(t, summaryMap, "BackupStatus")
		assert.Contains(t, summaryMap, "BackupType")
		assert.Contains(t, summaryMap, "TableName")
		assert.Equal(t, tableName, summaryMap["TableName"])
	}
}

func TestListBackups_FilterByTableName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create backups for different tables
	now := time.Now().Unix()

	table1Name := "table1"
	backup1Key := fmt.Sprintf("dynamodb:backup:%s:backup1", table1Name)
	backup1Data := map[string]interface{}{
		"TableName": table1Name,
		"TableId":   "table-id-1",
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table1Name),
		"BackupDetails": map[string]interface{}{
			"BackupArn":              fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/backup/backup1", table1Name),
			"BackupName":             "backup1",
			"BackupStatus":           "AVAILABLE",
			"BackupType":             "USER",
			"BackupSizeBytes":        1024,
			"BackupCreationDateTime": float64(now),
		},
	}
	err := state.Set(backup1Key, backup1Data)
	require.NoError(t, err)

	table2Name := "table2"
	backup2Key := fmt.Sprintf("dynamodb:backup:%s:backup2", table2Name)
	backup2Data := map[string]interface{}{
		"TableName": table2Name,
		"TableId":   "table-id-2",
		"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", table2Name),
		"BackupDetails": map[string]interface{}{
			"BackupArn":              fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/backup/backup2", table2Name),
			"BackupName":             "backup2",
			"BackupStatus":           "AVAILABLE",
			"BackupType":             "USER",
			"BackupSizeBytes":        2048,
			"BackupCreationDateTime": float64(now + 100),
		},
	}
	err = state.Set(backup2Key, backup2Data)
	require.NoError(t, err)

	// List backups for table1 only
	input := &ListBackupsInput{
		TableName: &table1Name,
	}

	resp, err := service.listBackups(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["BackupSummaries"].([]interface{})
	require.True(t, ok, "BackupSummaries should be an array")
	assert.Len(t, summaries, 1, "Should have only one backup for table1")

	summaryMap, ok := summaries[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, table1Name, summaryMap["TableName"])
}

func TestListBackups_Pagination(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// Create multiple backups
	tableName := "test-table-paginated"
	now := time.Now().Unix()

	for i := 1; i <= 5; i++ {
		backupKey := fmt.Sprintf("dynamodb:backup:%s:backup%d", tableName, i)
		backupData := map[string]interface{}{
			"TableName": tableName,
			"TableId":   "table-id-123",
			"TableArn":  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
			"BackupDetails": map[string]interface{}{
				"BackupArn":              fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/backup/backup%d", tableName, i),
				"BackupName":             fmt.Sprintf("backup%d", i),
				"BackupStatus":           "AVAILABLE",
				"BackupType":             "USER",
				"BackupSizeBytes":        1024 * i,
				"BackupCreationDateTime": float64(now + int64(i*100)),
			},
		}
		err := state.Set(backupKey, backupData)
		require.NoError(t, err)
	}

	// List backups with limit
	limit := int32(2)
	input := &ListBackupsInput{
		Limit: &limit,
	}

	resp, err := service.listBackups(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	summaries, ok := responseBody["BackupSummaries"].([]interface{})
	require.True(t, ok, "BackupSummaries should be an array")
	assert.Len(t, summaries, 2, "Should have only 2 backups due to limit")

	// Should have LastEvaluatedBackupArn for pagination
	assert.Contains(t, responseBody, "LastEvaluatedBackupArn")
}

func TestListBackups_NoParameters(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewDynamoDBService(state, validator)

	// ListBackups should work with no parameters
	input := &ListBackupsInput{}

	resp, err := service.listBackups(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body, &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "BackupSummaries")
}
