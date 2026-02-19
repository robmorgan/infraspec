package dynamodb

import (
	"context"

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
		var importDesc map[string]interface{}
		if err := s.state.Get(key, &importDesc); err != nil {
			continue
		}

		// Filter by table ARN if specified
		if input.TableArn != nil && *input.TableArn != "" {
			tableArn, ok := importDesc["TableArn"].(string)
			if !ok || tableArn != *input.TableArn {
				continue
			}
		}

		// Build import summary with key fields
		summary := map[string]interface{}{}

		if importArn, ok := importDesc["ImportArn"]; ok {
			summary["ImportArn"] = importArn
		}
		if importStatus, ok := importDesc["ImportStatus"]; ok {
			summary["ImportStatus"] = importStatus
		}
		if tableArn, ok := importDesc["TableArn"]; ok {
			summary["TableArn"] = tableArn
		}
		if s3BucketSource, ok := importDesc["S3BucketSource"]; ok {
			summary["S3BucketSource"] = s3BucketSource
		}
		if cloudWatchLogGroupArn, ok := importDesc["CloudWatchLogGroupArn"]; ok {
			summary["CloudWatchLogGroupArn"] = cloudWatchLogGroupArn
		}
		if inputFormat, ok := importDesc["InputFormat"]; ok {
			summary["InputFormat"] = inputFormat
		}

		importSummaries = append(importSummaries, summary)
	}

	response := map[string]interface{}{
		"ImportSummaryList": importSummaries,
	}

	return s.jsonResponse(200, response)
}
