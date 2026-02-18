package dynamodb

import (
	"context"

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
		var importDesc map[string]interface{}
		if err := s.state.Get(key, &importDesc); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			tableArn, _ := importDesc["TableArn"].(string)
			if tableArn != *input.TableArn {
				continue
			}
		}

		// Build import summary from import description
		summary := map[string]interface{}{}

		if importArn, ok := importDesc["ImportArn"].(string); ok {
			summary["ImportArn"] = importArn
		}
		if importStatus, ok := importDesc["ImportStatus"].(string); ok {
			summary["ImportStatus"] = importStatus
		}
		if tableArn, ok := importDesc["TableArn"].(string); ok {
			summary["TableArn"] = tableArn
		}
		if s3BucketSource, ok := importDesc["S3BucketSource"]; ok {
			summary["S3BucketSource"] = s3BucketSource
		}
		if cloudWatchLogGroupArn, ok := importDesc["CloudWatchLogGroupArn"].(string); ok {
			summary["CloudWatchLogGroupArn"] = cloudWatchLogGroupArn
		}
		if inputFormat, ok := importDesc["InputFormat"].(string); ok {
			summary["InputFormat"] = inputFormat
		}
		if startTime, ok := importDesc["StartTime"]; ok {
			summary["StartTime"] = startTime
		}
		if endTime, ok := importDesc["EndTime"]; ok {
			summary["EndTime"] = endTime
		}

		importSummaries = append(importSummaries, summary)
	}

	// Apply PageSize pagination
	pageSize := 25 // DynamoDB default for ListImports
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	// Handle NextToken pagination
	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		for i, summary := range importSummaries {
			if summaryMap, ok := summary.(map[string]interface{}); ok {
				if arn, ok := summaryMap["ImportArn"].(string); ok {
					if arn == *input.NextToken {
						startIndex = i + 1
						break
					}
				}
			}
		}
	}

	endIndex := startIndex + pageSize
	if endIndex > len(importSummaries) {
		endIndex = len(importSummaries)
	}

	paginatedSummaries := []interface{}{}
	if startIndex < len(importSummaries) {
		paginatedSummaries = importSummaries[startIndex:endIndex]
	}

	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	// Set NextToken if there are more results
	if endIndex < len(importSummaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if lastArn, ok := lastSummary["ImportArn"].(string); ok {
				response["NextToken"] = lastArn
			}
		}
	}

	return s.jsonResponse(200, response)
}
