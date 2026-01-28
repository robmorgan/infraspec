package dynamodb

import (
	"context"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// listImports returns a list of DynamoDB table import jobs.
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
			if tableArnInImport, ok := importData["TableArn"].(string); ok {
				if tableArnInImport != *input.TableArn {
					continue
				}
			} else {
				continue
			}
		}

		// Build summary
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
			summary.S3BucketSource = &S3BucketSource{}
			if bucket, ok := s3BucketSource["S3Bucket"].(string); ok {
				summary.S3BucketSource.S3Bucket = &bucket
			}
			if keyPrefix, ok := s3BucketSource["S3KeyPrefix"].(string); ok {
				summary.S3BucketSource.S3KeyPrefix = &keyPrefix
			}
			if bucketOwner, ok := s3BucketSource["S3BucketOwner"].(string); ok {
				summary.S3BucketSource.S3BucketOwner = &bucketOwner
			}
		}

		importSummaries = append(importSummaries, summary)
	}

	// Apply pagination if specified
	pageSize := 100 // Default page size
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
	}

	startIndex := 0
	if input.NextToken != nil && *input.NextToken != "" {
		// In a real implementation, decode the token to find the start index
		// For simplicity, we'll skip pagination token parsing in the emulator
	}

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
		nextToken := "next-token-placeholder"
		response.NextToken = &nextToken
	}

	return s.jsonResponse(200, response)
}
