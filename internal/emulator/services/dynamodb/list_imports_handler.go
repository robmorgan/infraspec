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

		if importArn, ok := importData["ImportArn"].(string); ok {
			summary["ImportArn"] = importArn
		}
		if importStatus, ok := importData["ImportStatus"].(string); ok {
			summary["ImportStatus"] = importStatus
		}
		if tableArn, ok := importData["TableArn"].(string); ok {
			summary["TableArn"] = tableArn
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
		if startTime, ok := importData["StartTime"]; ok {
			summary["StartTime"] = startTime
		}
		if endTime, ok := importData["EndTime"]; ok {
			summary["EndTime"] = endTime
		}

		importSummaries = append(importSummaries, summary)
	}

	// Apply pagination if specified
	pageSize := 100 // Default page size
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	// Note: NextToken handling would require encoding/decoding the position
	// For simplicity in the emulator, we'll skip detailed pagination logic

	// Apply limit
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
