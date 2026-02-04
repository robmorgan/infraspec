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
		var importData map[string]interface{}
		if err := s.state.Get(key, &importData); err != nil {
			continue
		}

		// Filter by TableArn if specified
		if input.TableArn != nil && *input.TableArn != "" {
			if importTableArn, ok := importData["TableArn"].(string); ok {
				if importTableArn != *input.TableArn {
					continue
				}
			} else {
				continue
			}
		}

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
		if inputFormat, ok := importData["InputFormat"].(string); ok {
			summary["InputFormat"] = inputFormat
		}
		if startTime, ok := importData["StartTime"]; ok {
			summary["StartTime"] = startTime
		}
		if endTime, ok := importData["EndTime"]; ok {
			summary["EndTime"] = endTime
		}
		if s3Source, ok := importData["S3BucketSource"]; ok {
			summary["S3BucketSource"] = s3Source
		}
		if logGroupArn, ok := importData["CloudWatchLogGroupArn"].(string); ok {
			summary["CloudWatchLogGroupArn"] = logGroupArn
		}

		importSummaries = append(importSummaries, summary)
	}

	// Apply pagination
	pageSize := 100
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

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

	// Add NextToken if there are more results
	if endIndex < len(importSummaries) {
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if lastArn, ok := lastSummary["ImportArn"].(string); ok {
				response["NextToken"] = lastArn
			}
		}
	}

	return s.jsonResponse(200, response)
}
