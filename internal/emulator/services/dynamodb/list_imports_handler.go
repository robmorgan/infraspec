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

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			if tableArnInImport, ok := importData["TableArn"].(string); ok {
				if tableArnInImport != *input.TableArn {
					continue
				}
			} else {
				continue
			}
		}

		// Build import summary
		summary := map[string]interface{}{}

		// Add optional fields if present
		if importArn, ok := importData["ImportArn"]; ok {
			summary["ImportArn"] = importArn
		}
		if importStatus, ok := importData["ImportStatus"]; ok {
			summary["ImportStatus"] = importStatus
		}
		if tableArn, ok := importData["TableArn"]; ok {
			summary["TableArn"] = tableArn
		}
		if cloudWatchLogGroupArn, ok := importData["CloudWatchLogGroupArn"]; ok {
			summary["CloudWatchLogGroupArn"] = cloudWatchLogGroupArn
		}
		if s3BucketSource, ok := importData["S3BucketSource"]; ok {
			summary["S3BucketSource"] = s3BucketSource
		}
		if inputFormat, ok := importData["InputFormat"]; ok {
			summary["InputFormat"] = inputFormat
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
	pageSize := 25 // Default page size for ListImports
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	// For NextToken pagination, we would decode it to get the start index
	// For simplicity, we'll use a basic implementation

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
		// In a real implementation, this would be an encoded token
		response["NextToken"] = "nextPageToken"
	}

	return s.jsonResponse(200, response)
}
