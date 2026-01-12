package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports lists completed imports within the past 90 days.
func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
	// Optional table ARN filter
	var tableArn string
	if input.TableArn != nil {
		tableArn = *input.TableArn
	}

	// List all imports from state
	keys, err := s.state.List("dynamodb:import:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list imports"), nil
	}

	var importSummaries []interface{}

	for _, key := range keys {
		var importData map[string]interface{}
		if err := s.state.Get(key, &importData); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if tableArn != "" {
			if importTableArn, ok := importData["TableArn"].(string); ok {
				if importTableArn != tableArn {
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

		if tableArnValue, ok := importData["TableArn"].(string); ok {
			summary["TableArn"] = tableArnValue
		}

		if s3BucketSource, ok := importData["S3BucketSource"].(map[string]interface{}); ok {
			summary["S3BucketSource"] = s3BucketSource
		}

		if cloudWatchLogGroupArn, ok := importData["CloudWatchLogGroupArn"].(string); ok {
			summary["CloudWatchLogGroupArn"] = cloudWatchLogGroupArn
		}

		if inputFormat, ok := importData["InputFormat"].(string); ok {
			summary["InputFormat"] = inputFormat
		}

		if startTime, ok := importData["StartTime"].(float64); ok {
			summary["StartTime"] = startTime
		}

		if endTime, ok := importData["EndTime"].(float64); ok {
			summary["EndTime"] = endTime
		}

		importSummaries = append(importSummaries, summary)
	}

	// Apply pagination if PageSize is specified
	pageSize := 25 // Default page size
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	// If NextToken is provided, decode it to get start index
	// For simplicity, we'll use a simple index-based pagination
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, NextToken would be decoded
		// For now, we'll keep it simple and not implement pagination offset
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
		response["NextToken"] = "next-page-token" // Simplified token
	}

	return s.jsonResponse(200, response)
}
