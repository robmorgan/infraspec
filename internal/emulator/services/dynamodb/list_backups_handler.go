package dynamodb

import (
	"context"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// listBackups returns a list of DynamoDB backups associated with an AWS account
// that weren't made with AWS Backup. To list backups for a given table, specify TableName.
// ListBackups returns a paginated list of backup summaries.
func (s *DynamoDBService) listBackups(ctx context.Context, input *ListBackupsInput) (*emulator.AWSResponse, error) {
	// List all backups from state
	keys, err := s.state.List("dynamodb:backup:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list backups"), nil
	}

	backupSummaries := []interface{}{}

	for _, key := range keys {
		var backupData map[string]interface{}
		if err := s.state.Get(key, &backupData); err != nil {
			continue
		}

		// Filter by table name if specified
		if input.TableName != nil && *input.TableName != "" {
			if tableNameInBackup, ok := backupData["TableName"].(string); ok {
				if tableNameInBackup != *input.TableName {
					continue
				}
			} else {
				continue
			}
		}

		// Filter by backup type if specified
		if input.BackupType != "" {
			if backupDetails, ok := backupData["BackupDetails"].(map[string]interface{}); ok {
				if backupType, ok := backupDetails["BackupType"].(string); ok {
					if string(input.BackupType) != backupType {
						continue
					}
				}
			}
		}

		// Build backup summary
		if backupDetails, ok := backupData["BackupDetails"].(map[string]interface{}); ok {
			summary := map[string]interface{}{
				"BackupArn":              backupDetails["BackupArn"],
				"BackupName":             backupDetails["BackupName"],
				"BackupStatus":           backupDetails["BackupStatus"],
				"BackupType":             backupDetails["BackupType"],
				"BackupSizeBytes":        backupDetails["BackupSizeBytes"],
				"BackupCreationDateTime": backupDetails["BackupCreationDateTime"],
			}

			// Add table info
			if tableName, ok := backupData["TableName"].(string); ok {
				summary["TableName"] = tableName
			}
			if tableId, ok := backupData["TableId"].(string); ok {
				summary["TableId"] = tableId
			}
			if tableArn, ok := backupData["TableArn"].(string); ok {
				summary["TableArn"] = tableArn
			}

			// Add optional expiry date if present
			if expiryDate, ok := backupDetails["BackupExpiryDateTime"]; ok {
				summary["BackupExpiryDateTime"] = expiryDate
			}

			backupSummaries = append(backupSummaries, summary)
		}
	}

	// Apply pagination if specified
	limit := 100 // Default limit
	if input.Limit != nil && *input.Limit > 0 {
		limit = int(*input.Limit)
	}

	startIndex := 0
	if input.ExclusiveStartBackupArn != nil && *input.ExclusiveStartBackupArn != "" {
		// Find the index of the exclusive start backup
		for i, summary := range backupSummaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				if arn, ok := summaryMap["BackupArn"].(string); ok {
					if arn == *input.ExclusiveStartBackupArn {
						startIndex = i + 1
						break
					}
				}
			}
		}
	}

	// Apply pagination
	endIndex := startIndex + limit
	if endIndex > len(backupSummaries) {
		endIndex = len(backupSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(backupSummaries) {
		paginatedSummaries = backupSummaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"BackupSummaries": paginatedSummaries,
	}

	// Add LastEvaluatedBackupArn if there are more results
	if endIndex < len(backupSummaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if lastArn, ok := lastSummary["BackupArn"].(string); ok {
				response["LastEvaluatedBackupArn"] = lastArn
			}
		}
	}

	return s.jsonResponse(200, response)
}

// Helper function to extract table name from backup key
func extractTableNameFromBackupKey(key string) string {
	// Key format: "dynamodb:backup:tablename:backupname"
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
