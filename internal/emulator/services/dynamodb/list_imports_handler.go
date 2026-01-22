package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports lists completed imports within the past 90 days.
func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
	// List all imports from state
	keys, err := s.state.List("dynamodb:import:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list imports"), nil
	}

	importSummaries := []interface{}{}

	for _, key := range keys {
		var importData map[string]interface{}
		if err := s.state.Get(key, &importData); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			if tableArn, ok := importData["TableArn"].(string); ok {
				if tableArn != *input.TableArn {
					continue
				}
			} else {
				continue
			}
		}

		// Build import summary
		summary := map[string]interface{}{}

		if importArn, ok := importData["ImportArn"].(string); ok {
			summary["ImportArn"] = importArn
		}

		if importStatus, ok := importData["ImportStatus"].(string); ok {
			summary["ImportStatus"] = importStatus
		}

		if tableArn, ok := importData["TableArn"].(string); ok {
			summary["TableArn"] = tableArn
		}

		if cloudWatchLogGroupArn, ok := importData["CloudWatchLogGroupArn"].(string); ok {
			summary["CloudWatchLogGroupArn"] = cloudWatchLogGroupArn
		}

		if s3BucketSource, ok := importData["S3BucketSource"].(map[string]interface{}); ok {
			summary["S3BucketSource"] = s3BucketSource
		}

		if startTime, ok := importData["StartTime"]; ok {
			summary["StartTime"] = startTime
		}

		if endTime, ok := importData["EndTime"]; ok {
			summary["EndTime"] = endTime
		}

		importSummaries = append(importSummaries, summary)
	}

	// Apply pagination if specified
	pageSize := 25 // Default page size
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, NextToken would be decoded to determine start position
		// For simplicity, we'll start from the beginning
		startIndex = 0
	}

	// Apply pagination
	endIndex := startIndex + pageSize
	if endIndex > len(importSummaries) {
		endIndex = len(importSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(importSummaries) {
		paginatedSummaries = importSummaries[startIndex:endIndex]
	}

	// Build response
	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(importSummaries) {
		response["NextToken"] = "next-page-token"
	}

	return s.jsonResponse(200, response)
}

// Helper function to extract table name from import key
func extractTableNameFromImportKey(key string) string {
	// Key format: "dynamodb:import:tablename:importarn"
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
