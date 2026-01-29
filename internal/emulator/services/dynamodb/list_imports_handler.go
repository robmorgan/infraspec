package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports returns a list of completed imports within the past 90 days.
func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
	// List all imports from state
	keys, err := s.state.List("dynamodb:import:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list imports"), nil
	}

	importSummaries := []ImportSummary{}

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
		summary := ImportSummary{}

		if importArn, ok := importData["ImportArn"].(string); ok {
			summary.ImportArn = &importArn
		}

		if status, ok := importData["ImportStatus"].(string); ok {
			summary.ImportStatus = ImportStatus(status)
		}

		if inputFormat, ok := importData["InputFormat"].(string); ok {
			summary.InputFormat = InputFormat(inputFormat)
		}

		if tableArn, ok := importData["TableArn"].(string); ok {
			summary.TableArn = &tableArn
		}

		if cloudWatchLogGroupArn, ok := importData["CloudWatchLogGroupArn"].(string); ok {
			summary.CloudWatchLogGroupArn = &cloudWatchLogGroupArn
		}

		// Add S3BucketSource if present
		if s3BucketSource, ok := importData["S3BucketSource"].(map[string]interface{}); ok {
			// Convert to S3BucketSource type if needed
			// For now, just include it in the summary
			// This would need proper type conversion in a real implementation
		}

		importSummaries = append(importSummaries, summary)
	}

	// Apply pagination if specified
	pageSize := 100 // Default page size
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	// Note: NextToken would be used for pagination in a real implementation
	// For simplicity, we'll just apply the page size limit

	// Apply pagination
	endIndex := startIndex + pageSize
	if endIndex > len(importSummaries) {
		endIndex = len(importSummaries)
	}

	paginatedSummaries := []ImportSummary{}
	if startIndex < len(importSummaries) {
		paginatedSummaries = importSummaries[startIndex:endIndex]
	}

	// Build response
	response := ListImportsOutput{
		ImportSummaryList: paginatedSummaries,
	}

	// Add NextToken if there are more results
	if endIndex < len(importSummaries) {
		nextToken := "has-more-results" // Simplified token
		response.NextToken = &nextToken
	}

	return s.jsonResponse(200, response)
}
