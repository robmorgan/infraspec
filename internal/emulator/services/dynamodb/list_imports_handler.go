package dynamodb

import (
	"context"
	"strings"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports lists completed imports within the past 90 days.
func (s *DynamoDBService) listImports(ctx context.Context, input *ListImportsInput) (*emulator.AWSResponse, error) {
	// List all import keys from state
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
			tableArn, _ := importData["TableArn"].(string)
			if tableArn != *input.TableArn {
				continue
			}
		}

		// Build import summary with available fields
		summary := map[string]interface{}{}

		if importArn, ok := importData["ImportArn"].(string); ok {
			summary["ImportArn"] = importArn
		} else {
			// Derive ARN from the state key
			// Key format: dynamodb:import:<importArn>
			importArn := strings.TrimPrefix(key, "dynamodb:import:")
			if importArn != "" {
				summary["ImportArn"] = importArn
			}
		}

		if importStatus, ok := importData["ImportStatus"].(string); ok {
			summary["ImportStatus"] = importStatus
		}

		if inputFormat, ok := importData["InputFormat"].(string); ok {
			summary["InputFormat"] = inputFormat
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

		importSummaries = append(importSummaries, summary)
	}

	// Apply PageSize pagination if specified
	pageSize := len(importSummaries)
	if input.PageSize != nil && *input.PageSize > 0 && int(*input.PageSize) < pageSize {
		pageSize = int(*input.PageSize)
	}

	paginatedSummaries := importSummaries
	var nextToken *string
	if pageSize < len(importSummaries) {
		paginatedSummaries = importSummaries[:pageSize]
		// Indicate more results available via NextToken
		if lastSummary, ok := paginatedSummaries[len(paginatedSummaries)-1].(map[string]interface{}); ok {
			if arn, ok := lastSummary["ImportArn"].(string); ok {
				nextToken = &arn
			}
		}
	}

	response := map[string]interface{}{
		"ImportSummaryList": paginatedSummaries,
	}

	if nextToken != nil {
		response["NextToken"] = *nextToken
	}

	return s.jsonResponse(200, response)
}
