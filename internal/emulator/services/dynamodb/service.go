package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

type DynamoDBService struct {
	state     emulator.StateManager
	validator emulator.Validator
}

func NewDynamoDBService(state emulator.StateManager, validator emulator.Validator) *DynamoDBService {
	return &DynamoDBService{
		state:     state,
		validator: validator,
	}
}

func (s *DynamoDBService) ServiceName() string {
	return "dynamodb_20120810"
}

func (s *DynamoDBService) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	action := s.extractAction(req)
	if action == "" {
		return s.errorResponse(400, "InvalidAction", "Missing or invalid action"), nil
	}

	params, err := s.parseParameters(req)
	if err != nil {
		return s.errorResponse(400, "InvalidParameterValue", err.Error()), nil
	}

	if err := s.validator.ValidateAction(action, params); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	switch action {
	case "CreateTable":
		input, err := emulator.ParseJSONRequest[CreateTableInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.createTable(ctx, input)
	case "DescribeTable":
		input, err := emulator.ParseJSONRequest[DescribeTableInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeTable(ctx, input)
	case "DeleteTable":
		input, err := emulator.ParseJSONRequest[DeleteTableInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteTable(ctx, input)
	case "ListTables":
		input, err := emulator.ParseJSONRequest[ListTablesInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listTables(ctx, input)
	case "UpdateTable":
		input, err := emulator.ParseJSONRequest[UpdateTableInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.updateTable(ctx, input)
	case "DescribeContinuousBackups":
		input, err := emulator.ParseJSONRequest[DescribeContinuousBackupsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeContinuousBackups(ctx, input)
	case "UpdateContinuousBackups":
		input, err := emulator.ParseJSONRequest[UpdateContinuousBackupsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.updateContinuousBackups(ctx, input)
	case "DescribeTimeToLive":
		input, err := emulator.ParseJSONRequest[DescribeTimeToLiveInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeTimeToLive(ctx, input)
	case "UpdateTimeToLive":
		input, err := emulator.ParseJSONRequest[UpdateTimeToLiveInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.updateTimeToLive(ctx, input)
	case "ListTagsOfResource":
		input, err := emulator.ParseJSONRequest[ListTagsOfResourceInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listTagsOfResource(ctx, input)
	case "TagResource":
		input, err := emulator.ParseJSONRequest[TagResourceInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.tagResource(ctx, input)
	case "UntagResource":
		input, err := emulator.ParseJSONRequest[UntagResourceInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.untagResource(ctx, input)
	case "PutItem":
		input, err := emulator.ParseJSONRequest[PutItemInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.putItem(ctx, input)
	case "GetItem":
		input, err := emulator.ParseJSONRequest[GetItemInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.getItem(ctx, input)
	case "DeleteItem":
		input, err := emulator.ParseJSONRequest[DeleteItemInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteItem(ctx, input)
	case "Query":
		input, err := emulator.ParseJSONRequest[QueryInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.query(ctx, input)
	case "Scan":
		input, err := emulator.ParseJSONRequest[ScanInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.scan(ctx, input)
	case "CreateBackup":
		input, err := emulator.ParseJSONRequest[CreateBackupInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.createBackup(ctx, input)
	case "CreateGlobalTable":
		input, err := emulator.ParseJSONRequest[CreateGlobalTableInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.createGlobalTable(ctx, input)
	case "DeleteBackup":
		input, err := emulator.ParseJSONRequest[DeleteBackupInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteBackup(ctx, input)
	case "DeleteResourcePolicy":
		input, err := emulator.ParseJSONRequest[DeleteResourcePolicyInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteResourcePolicy(ctx, input)
	case "DescribeBackup":
		input, err := emulator.ParseJSONRequest[DescribeBackupInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeBackup(ctx, input)
	case "DescribeContributorInsights":
		input, err := emulator.ParseJSONRequest[DescribeContributorInsightsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeContributorInsights(ctx, input)
	case "DescribeEndpoints":
		return s.describeEndpoints(ctx)
	case "DescribeExport":
		input, err := emulator.ParseJSONRequest[DescribeExportInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeExport(ctx, input)
	case "DescribeGlobalTable":
		input, err := emulator.ParseJSONRequest[DescribeGlobalTableInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeGlobalTable(ctx, input)
	case "DescribeGlobalTableSettings":
		input, err := emulator.ParseJSONRequest[DescribeGlobalTableSettingsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeGlobalTableSettings(ctx, input)
	case "DescribeImport":
		input, err := emulator.ParseJSONRequest[DescribeImportInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeImport(ctx, input)
	case "DescribeKinesisStreamingDestination":
		input, err := emulator.ParseJSONRequest[DescribeKinesisStreamingDestinationInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeKinesisStreamingDestination(ctx, input)
	case "DescribeLimits":
		return s.describeLimits(ctx)
	case "DescribeTableReplicaAutoScaling":
		input, err := emulator.ParseJSONRequest[DescribeTableReplicaAutoScalingInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.describeTableReplicaAutoScaling(ctx, input)
	case "GetResourcePolicy":
		input, err := emulator.ParseJSONRequest[GetResourcePolicyInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.getResourcePolicy(ctx, input)
	case "ListBackups":
		input, err := emulator.ParseJSONRequest[ListBackupsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listBackups(ctx, input)
	case "ListContributorInsights":
		input, err := emulator.ParseJSONRequest[ListContributorInsightsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listContributorInsights(ctx, input)
	case "ListExports":
		input, err := emulator.ParseJSONRequest[ListExportsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listExports(ctx, input)
	case "ListGlobalTables":
		input, err := emulator.ParseJSONRequest[ListGlobalTablesInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listGlobalTables(ctx, input)
	case "ListImports":
		input, err := emulator.ParseJSONRequest[ListImportsInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listImports(ctx, input)
	case "PutResourcePolicy":
		input, err := emulator.ParseJSONRequest[PutResourcePolicyInput](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.putResourcePolicy(ctx, input)
	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

func (s *DynamoDBService) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	// DynamoDB uses X-Amz-Target header: "DynamoDB_20120810.CreateTable"
	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func (s *DynamoDBService) parseParameters(req *emulator.AWSRequest) (map[string]interface{}, error) {
	if req.Parameters != nil {
		return req.Parameters, nil
	}

	// DynamoDB uses JSON for requests
	var params map[string]interface{}
	if err := json.Unmarshal(req.Body, &params); err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %w", err)
	}
	return params, nil
}

func (s *DynamoDBService) createTable(ctx context.Context, input *CreateTableInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	// Check if table already exists
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var existingTable map[string]interface{}
	if err := s.state.Get(key, &existingTable); err == nil {
		return s.errorResponse(400, "ResourceInUseException", fmt.Sprintf("Table already exists: %s", tableName)), nil
	}

	// Extract table configuration
	billingMode := "PROVISIONED"
	if input.BillingMode != "" {
		billingMode = string(input.BillingMode)
	}

	// Build table description
	now := time.Now().Unix()
	tableDesc := map[string]interface{}{
		"TableName":                 tableName,
		"TableStatus":               "ACTIVE", // In emulator, table is immediately active
		"TableArn":                  fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s", tableName),
		"TableId":                   uuid.New().String(),
		"CreationDateTime":          float64(now),
		"TableSizeBytes":            0,
		"ItemCount":                 0,
		"DeletionProtectionEnabled": false,
		"BillingModeSummary": map[string]interface{}{
			"BillingMode": billingMode,
		},
	}

	// Add key schema
	if len(input.KeySchema) > 0 {
		keySchema := make([]interface{}, len(input.KeySchema))
		for i, ks := range input.KeySchema {
			keySchema[i] = map[string]interface{}{
				"AttributeName": ks.AttributeName,
				"KeyType":       ks.KeyType,
			}
		}
		tableDesc["KeySchema"] = keySchema
	}

	// Add attribute definitions
	if len(input.AttributeDefinitions) > 0 {
		attrDefs := make([]interface{}, len(input.AttributeDefinitions))
		for i, ad := range input.AttributeDefinitions {
			attrDefs[i] = map[string]interface{}{
				"AttributeName": ad.AttributeName,
				"AttributeType": ad.AttributeType,
			}
		}
		tableDesc["AttributeDefinitions"] = attrDefs
	}

	// Add provisioned throughput if specified
	if billingMode == "PROVISIONED" {
		if input.ProvisionedThroughput != nil {
			tableDesc["ProvisionedThroughput"] = map[string]interface{}{
				"ReadCapacityUnits":      input.ProvisionedThroughput.ReadCapacityUnits,
				"WriteCapacityUnits":     input.ProvisionedThroughput.WriteCapacityUnits,
				"NumberOfDecreasesToday": 0,
			}
		} else {
			// Default values
			tableDesc["ProvisionedThroughput"] = map[string]interface{}{
				"ReadCapacityUnits":      5,
				"WriteCapacityUnits":     5,
				"NumberOfDecreasesToday": 0,
			}
		}
	}

	// Add tags if specified
	if len(input.Tags) > 0 {
		tags := make([]interface{}, len(input.Tags))
		for i, tag := range input.Tags {
			tags[i] = map[string]interface{}{
				"Key":   tag.Key,
				"Value": tag.Value,
			}
		}
		tableDesc["Tags"] = tags
	}

	// Add global secondary indexes if specified (always include field)
	if len(input.GlobalSecondaryIndexes) > 0 {
		gsi := make([]interface{}, len(input.GlobalSecondaryIndexes))
		for i, idx := range input.GlobalSecondaryIndexes {
			gsi[i] = map[string]interface{}{
				"IndexName": idx.IndexName,
			}
		}
		tableDesc["GlobalSecondaryIndexes"] = gsi
	} else {
		tableDesc["GlobalSecondaryIndexes"] = []interface{}{}
	}

	// Add local secondary indexes if specified (always include field)
	if len(input.LocalSecondaryIndexes) > 0 {
		lsi := make([]interface{}, len(input.LocalSecondaryIndexes))
		for i, idx := range input.LocalSecondaryIndexes {
			lsi[i] = map[string]interface{}{
				"IndexName": idx.IndexName,
			}
		}
		tableDesc["LocalSecondaryIndexes"] = lsi
	} else {
		tableDesc["LocalSecondaryIndexes"] = []interface{}{}
	}

	// Add stream ARN and label if streaming is enabled
	if input.StreamSpecification != nil && input.StreamSpecification.StreamEnabled != nil && *input.StreamSpecification.StreamEnabled {
		tableDesc["LatestStreamArn"] = fmt.Sprintf("arn:aws:dynamodb:us-east-1:000000000000:table/%s/stream/%s", tableName, uuid.New().String())
		tableDesc["LatestStreamLabel"] = fmt.Sprintf("%d", now)
		tableDesc["StreamSpecification"] = map[string]interface{}{
			"StreamEnabled":  *input.StreamSpecification.StreamEnabled,
			"StreamViewType": input.StreamSpecification.StreamViewType,
		}
	}

	// Add SSE description only if explicitly configured (not included when SSE not specified)
	if input.SSESpecification != nil && input.SSESpecification.Enabled != nil && *input.SSESpecification.Enabled {
		kmsKeyArn := "arn:aws:kms:us-east-1:000000000000:key/" + uuid.New().String()
		if input.SSESpecification.KMSMasterKeyId != nil {
			kmsKeyArn = *input.SSESpecification.KMSMasterKeyId
		}
		tableDesc["SSEDescription"] = map[string]interface{}{
			"Status":          "ENABLED",
			"SSEType":         "KMS",
			"KMSMasterKeyArn": kmsKeyArn,
		}
	}

	// Add table class and table class summary
	tableDesc["TableClass"] = "STANDARD"
	tableDesc["TableClassSummary"] = map[string]interface{}{
		"TableClass": "STANDARD",
	}

	// Note: ArchivalSummary should only be present if table is archived, so we don't include it

	// Add TTL description
	tableDesc["TimeToLiveDescription"] = map[string]interface{}{
		"TimeToLiveStatus": "DISABLED",
	}

	// Add replicas (empty array for non-global tables)
	tableDesc["Replicas"] = []interface{}{}

	// Save table to state
	if err := s.state.Set(key, tableDesc); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to create table"), nil
	}

	// Return response
	response := map[string]interface{}{
		"TableDescription": tableDesc,
	}

	// Debug: log the complete JSON response
	if jsonBytes, err := json.Marshal(response); err == nil {
		fmt.Printf("DEBUG CreateTable Response:\n%s\n", string(jsonBytes))
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) describeTable(ctx context.Context, input *DescribeTableInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Simulate table creation: transition from CREATING to ACTIVE
	if status, ok := tableDesc["TableStatus"].(string); ok && status == "CREATING" {
		tableDesc["TableStatus"] = "ACTIVE"
		// Update in state so subsequent calls return ACTIVE
		if err := s.state.Set(key, tableDesc); err != nil {
			return s.errorResponse(500, "InternalServerError", "Failed to update table status"), nil
		}
	}

	// Add WarmThroughput if table has provisioned throughput
	if pt, ok := tableDesc["ProvisionedThroughput"].(map[string]interface{}); ok {
		if tableDesc["WarmThroughput"] == nil {
			readUnits := 5
			writeUnits := 5
			if rcu, ok := pt["ReadCapacityUnits"].(float64); ok {
				readUnits = int(rcu)
			}
			if wcu, ok := pt["WriteCapacityUnits"].(float64); ok {
				writeUnits = int(wcu)
			}
			tableDesc["WarmThroughput"] = map[string]interface{}{
				"ReadUnitsPerSecond":  readUnits,
				"WriteUnitsPerSecond": writeUnits,
				"Status":              "ACTIVE",
			}
			// Update in state
			if err := s.state.Set(key, tableDesc); err != nil {
				return s.errorResponse(500, "InternalServerError", "Failed to update table"), nil
			}
		}
	}

	response := map[string]interface{}{
		"Table": tableDesc,
	}

	// Debug: log the complete JSON response
	if jsonBytes, err := json.Marshal(response); err == nil {
		fmt.Printf("DEBUG DescribeTable Response:\n%s\n", string(jsonBytes))
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) deleteTable(ctx context.Context, input *DeleteTableInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Update status to DELETING
	tableDesc["TableStatus"] = "DELETING"

	// Delete from state
	if err := s.state.Delete(key); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to delete table"), nil
	}

	response := map[string]interface{}{
		"TableDescription": tableDesc,
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) listTables(ctx context.Context, input *ListTablesInput) (*emulator.AWSResponse, error) {
	// List all tables (input may contain ExclusiveStartTableName and Limit for pagination)
	keys, err := s.state.List("dynamodb:table:")
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to list tables"), nil
	}

	tableNames := []string{}
	for _, key := range keys {
		// Extract table name from key "dynamodb:table:tablename"
		parts := strings.Split(key, ":")
		if len(parts) >= 3 {
			tableNames = append(tableNames, strings.Join(parts[2:], ":"))
		}
	}

	response := map[string]interface{}{
		"TableNames": tableNames,
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) updateTable(ctx context.Context, input *UpdateTableInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Update provisioned throughput if specified
	if input.ProvisionedThroughput != nil {
		tableDesc["ProvisionedThroughput"] = map[string]interface{}{
			"ReadCapacityUnits":  input.ProvisionedThroughput.ReadCapacityUnits,
			"WriteCapacityUnits": input.ProvisionedThroughput.WriteCapacityUnits,
		}
	}

	// Save updated table
	if err := s.state.Set(key, tableDesc); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to update table"), nil
	}

	response := map[string]interface{}{
		"TableDescription": tableDesc,
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) putItem(ctx context.Context, input *PutItemInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	if len(input.Item) == 0 {
		return s.errorResponse(400, "ValidationException", "Item is required"), nil
	}

	// Store item
	itemKey := fmt.Sprintf("dynamodb:item:%s:%s", tableName, uuid.New().String())
	if err := s.state.Set(itemKey, input.Item); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to put item"), nil
	}

	response := map[string]interface{}{}
	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) getItem(ctx context.Context, input *GetItemInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}

	// For simplicity, just return empty for now
	response := map[string]interface{}{}
	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) deleteItem(ctx context.Context, input *DeleteItemInput) (*emulator.AWSResponse, error) {
	response := map[string]interface{}{}
	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) query(ctx context.Context, input *QueryInput) (*emulator.AWSResponse, error) {
	response := map[string]interface{}{
		"Items": []interface{}{},
		"Count": 0,
	}
	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) scan(ctx context.Context, input *ScanInput) (*emulator.AWSResponse, error) {
	response := map[string]interface{}{
		"Items": []interface{}{},
		"Count": 0,
	}
	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) describeContinuousBackups(ctx context.Context, input *DescribeContinuousBackupsInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	// Verify table exists
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Return continuous backups configuration
	// For testing purposes, return a default configuration
	response := map[string]interface{}{
		"ContinuousBackupsDescription": map[string]interface{}{
			"ContinuousBackupsStatus": "ENABLED",
			"PointInTimeRecoveryDescription": map[string]interface{}{
				"PointInTimeRecoveryStatus": "DISABLED",
			},
		},
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) updateContinuousBackups(ctx context.Context, input *UpdateContinuousBackupsInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	// Verify table exists
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Extract point in time recovery specification
	pitrStatus := "DISABLED"
	if input.PointInTimeRecoverySpecification != nil && input.PointInTimeRecoverySpecification.PointInTimeRecoveryEnabled != nil && *input.PointInTimeRecoverySpecification.PointInTimeRecoveryEnabled {
		pitrStatus = "ENABLED"
	}

	// Return updated continuous backups configuration
	response := map[string]interface{}{
		"ContinuousBackupsDescription": map[string]interface{}{
			"ContinuousBackupsStatus": "ENABLED",
			"PointInTimeRecoveryDescription": map[string]interface{}{
				"PointInTimeRecoveryStatus": pitrStatus,
			},
		},
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) describeTimeToLive(ctx context.Context, input *DescribeTimeToLiveInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	// Verify table exists
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Return TTL configuration
	// For testing purposes, return a default disabled TTL configuration
	response := map[string]interface{}{
		"TimeToLiveDescription": map[string]interface{}{
			"TimeToLiveStatus": "DISABLED",
		},
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) updateTimeToLive(ctx context.Context, input *UpdateTimeToLiveInput) (*emulator.AWSResponse, error) {
	if input.TableName == nil || *input.TableName == "" {
		return s.errorResponse(400, "ValidationException", "TableName is required"), nil
	}
	tableName := *input.TableName

	// Verify table exists
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Extract TTL specification
	ttlStatus := "DISABLED"
	attributeName := ""
	if input.TimeToLiveSpecification != nil {
		if input.TimeToLiveSpecification.Enabled != nil && *input.TimeToLiveSpecification.Enabled {
			ttlStatus = "ENABLED"
		}
		if input.TimeToLiveSpecification.AttributeName != nil {
			attributeName = *input.TimeToLiveSpecification.AttributeName
		}
	}

	// Return updated TTL configuration
	response := map[string]interface{}{
		"TimeToLiveSpecification": map[string]interface{}{
			"Enabled":       ttlStatus == "ENABLED",
			"AttributeName": attributeName,
		},
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) listTagsOfResource(ctx context.Context, input *ListTagsOfResourceInput) (*emulator.AWSResponse, error) {
	if input.ResourceArn == nil || *input.ResourceArn == "" {
		return s.errorResponse(400, "ValidationException", "ResourceArn is required"), nil
	}
	resourceArn := *input.ResourceArn

	// Extract table name from ARN
	// ARN format: arn:aws:dynamodb:us-east-1:000000000000:table/tablename
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}
	tableName := parts[len(parts)-1]

	// Verify table exists and get tags
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Get tags from table description
	tags := []interface{}{}
	if existingTags, ok := tableDesc["Tags"].([]interface{}); ok {
		tags = existingTags
	}

	response := map[string]interface{}{
		"Tags": tags,
	}

	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) tagResource(ctx context.Context, input *TagResourceInput) (*emulator.AWSResponse, error) {
	if input.ResourceArn == nil || *input.ResourceArn == "" {
		return s.errorResponse(400, "ValidationException", "ResourceArn is required"), nil
	}
	resourceArn := *input.ResourceArn

	if len(input.Tags) == 0 {
		return s.errorResponse(400, "ValidationException", "Tags is required"), nil
	}

	// Extract table name from ARN
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}
	tableName := parts[len(parts)-1]

	// Verify table exists
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Get existing tags
	existingTags := []interface{}{}
	if existing, ok := tableDesc["Tags"].([]interface{}); ok {
		existingTags = existing
	}

	// Merge new tags with existing tags (new tags override existing ones with same key)
	tagMap := make(map[string]interface{})
	for _, tag := range existingTags {
		if tagObj, ok := tag.(map[string]interface{}); ok {
			if key, ok := tagObj["Key"].(string); ok {
				tagMap[key] = tagObj
			}
		}
	}
	for _, tag := range input.Tags {
		if tag.Key != nil {
			tagMap[*tag.Key] = map[string]interface{}{
				"Key":   *tag.Key,
				"Value": tag.Value,
			}
		}
	}

	// Convert map back to slice
	mergedTags := []interface{}{}
	for _, tag := range tagMap {
		mergedTags = append(mergedTags, tag)
	}

	// Update table with new tags
	tableDesc["Tags"] = mergedTags
	if err := s.state.Set(key, tableDesc); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to update tags"), nil
	}

	response := map[string]interface{}{}
	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) untagResource(ctx context.Context, input *UntagResourceInput) (*emulator.AWSResponse, error) {
	if input.ResourceArn == nil || *input.ResourceArn == "" {
		return s.errorResponse(400, "ValidationException", "ResourceArn is required"), nil
	}
	resourceArn := *input.ResourceArn

	if len(input.TagKeys) == 0 {
		return s.errorResponse(400, "ValidationException", "TagKeys is required"), nil
	}

	// Extract table name from ARN
	parts := strings.Split(resourceArn, "/")
	if len(parts) < 2 {
		return s.errorResponse(400, "ValidationException", "Invalid ResourceArn format"), nil
	}
	tableName := parts[len(parts)-1]

	// Verify table exists
	key := fmt.Sprintf("dynamodb:table:%s", tableName)
	var tableDesc map[string]interface{}
	if err := s.state.Get(key, &tableDesc); err != nil {
		return s.errorResponse(400, "ResourceNotFoundException", fmt.Sprintf("Requested resource not found: Table: %s not found", tableName)), nil
	}

	// Get existing tags
	existingTags := []interface{}{}
	if existing, ok := tableDesc["Tags"].([]interface{}); ok {
		existingTags = existing
	}

	// Remove tags with specified keys
	keysToRemove := make(map[string]bool)
	for _, k := range input.TagKeys {
		keysToRemove[k] = true
	}

	filteredTags := []interface{}{}
	for _, tag := range existingTags {
		if tagObj, ok := tag.(map[string]interface{}); ok {
			if tagKey, ok := tagObj["Key"].(string); ok {
				if !keysToRemove[tagKey] {
					filteredTags = append(filteredTags, tag)
				}
			}
		}
	}

	// Update table with filtered tags
	tableDesc["Tags"] = filteredTags
	if err := s.state.Set(key, tableDesc); err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to update tags"), nil
	}

	response := map[string]interface{}{}
	return s.jsonResponse(200, response)
}

func (s *DynamoDBService) jsonResponse(statusCode int, data interface{}) (*emulator.AWSResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return s.errorResponse(500, "InternalServerError", "Failed to marshal response"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/x-amz-json-1.0",
			"x-amzn-RequestId": uuid.New().String(),
			"x-amz-crc32":      "0",
		},
		Body: body,
	}, nil
}

func (s *DynamoDBService) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	errorData := map[string]interface{}{
		"__type":  code,
		"message": message,
	}

	body, _ := json.Marshal(errorData)

	return &emulator.AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/x-amz-json-1.0",
			"x-amzn-RequestId": uuid.New().String(),
			"x-amzn-ErrorType": code,
		},
		Body: body,
	}
}

func getStringOrDefault(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return defaultValue
}
