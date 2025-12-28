package emulator

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// Add more services as needed:
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	// "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// ResponseRegistry holds the mapping of actions to their output types for each service
type ResponseRegistry struct {
	responses map[string]reflect.Type // key: "servicename:ActionName"
}

// NewResponseRegistry creates a new response registry
func NewResponseRegistry() *ResponseRegistry {
	return &ResponseRegistry{
		responses: make(map[string]reflect.Type),
	}
}

// RegisterServiceResponse registers all response types for a service
func (r *ResponseRegistry) RegisterServiceResponse(serviceName string, actionOutputTypes map[string]reflect.Type) {
	for action, outputType := range actionOutputTypes {
		key := formatServiceActionKey(serviceName, action)
		r.responses[key] = outputType
	}
}

// RegisterActionResponse registers a single action response type
func (r *ResponseRegistry) RegisterActionResponse(serviceName, action string, outputType reflect.Type) {
	key := formatServiceActionKey(serviceName, action)
	r.responses[key] = outputType
}

// GetResponseType returns the output type for a service and action
func (r *ResponseRegistry) GetResponseType(serviceName, action string) reflect.Type {
	key := formatServiceActionKey(serviceName, action)
	return r.responses[key]
}

// GetResponseTypeForAction searches for action across all services
func (r *ResponseRegistry) GetResponseTypeForAction(action string) reflect.Type {
	// Search through all registered responses
	for key, outputType := range r.responses {
		// key format: "servicename:ActionName"
		parts := splitServiceActionKey(key)
		if len(parts) == 2 && parts[1] == action {
			return outputType
		}
	}
	return nil
}

// GetRegisteredResponses returns all registered response types for debugging
func (r *ResponseRegistry) GetRegisteredResponses() []string {
	responses := make([]string, 0, len(r.responses))
	for key := range r.responses {
		responses = append(responses, key)
	}
	return responses
}

// formatServiceActionKey formats a service:action key
func formatServiceActionKey(serviceName, action string) string {
	return serviceName + ":" + action
}

// splitServiceActionKey splits a service:action key
func splitServiceActionKey(key string) []string {
	// Simple split on ":"
	parts := make([]string, 0, 2)
	idx := 0
	for i, c := range key {
		if c == ':' {
			if idx < i {
				parts = append(parts, key[idx:i])
			}
			idx = i + 1
		}
	}
	if idx < len(key) {
		parts = append(parts, key[idx:])
	}
	return parts
}

// RegisterRDSResponses registers all RDS response types
func RegisterRDSResponses(registry *ResponseRegistry) {
	responses := map[string]reflect.Type{
		"CreateDBInstance":    reflect.TypeOf(rds.CreateDBInstanceOutput{}),
		"DescribeDBInstances": reflect.TypeOf(rds.DescribeDBInstancesOutput{}),
		"DeleteDBInstance":    reflect.TypeOf(rds.DeleteDBInstanceOutput{}),
		"ModifyDBInstance":    reflect.TypeOf(rds.ModifyDBInstanceOutput{}),
		"StartDBInstance":     reflect.TypeOf(rds.StartDBInstanceOutput{}),
		"StopDBInstance":      reflect.TypeOf(rds.StopDBInstanceOutput{}),
		"RebootDBInstance":    reflect.TypeOf(rds.RebootDBInstanceOutput{}),
	}
	registry.RegisterServiceResponse("rds", responses)
}

// RegisterS3Responses registers all S3 response types
func RegisterS3Responses(registry *ResponseRegistry) {
	responses := map[string]reflect.Type{
		"CreateBucket":  reflect.TypeOf(s3.CreateBucketOutput{}),
		"ListBuckets":   reflect.TypeOf(s3.ListBucketsOutput{}),
		"DeleteBucket":  reflect.TypeOf(s3.DeleteBucketOutput{}),
		"PutObject":     reflect.TypeOf(s3.PutObjectOutput{}),
		"GetObject":     reflect.TypeOf(s3.GetObjectOutput{}),
		"DeleteObject":  reflect.TypeOf(s3.DeleteObjectOutput{}),
		"ListObjects":   reflect.TypeOf(s3.ListObjectsOutput{}),
		"ListObjectsV2": reflect.TypeOf(s3.ListObjectsV2Output{}),
		"HeadObject":    reflect.TypeOf(s3.HeadObjectOutput{}),
		"CopyObject":    reflect.TypeOf(s3.CopyObjectOutput{}),
	}
	registry.RegisterServiceResponse("s3", responses)
}

// RegisterDynamoDBResponses registers all DynamoDB response types
// Uncomment when github.com/aws/aws-sdk-go-v2/service/dynamodb is installed
/*
func RegisterDynamoDBResponses(registry *ResponseRegistry) {
	responses := map[string]reflect.Type{
		"CreateTable":      reflect.TypeOf(dynamodb.CreateTableOutput{}),
		"DescribeTable":    reflect.TypeOf(dynamodb.DescribeTableOutput{}),
		"ListTables":       reflect.TypeOf(dynamodb.ListTablesOutput{}),
		"DeleteTable":      reflect.TypeOf(dynamodb.DeleteTableOutput{}),
		"PutItem":          reflect.TypeOf(dynamodb.PutItemOutput{}),
		"GetItem":          reflect.TypeOf(dynamodb.GetItemOutput{}),
		"UpdateItem":       reflect.TypeOf(dynamodb.UpdateItemOutput{}),
		"DeleteItem":       reflect.TypeOf(dynamodb.DeleteItemOutput{}),
		"Query":            reflect.TypeOf(dynamodb.QueryOutput{}),
		"Scan":             reflect.TypeOf(dynamodb.ScanOutput{}),
		"BatchGetItem":     reflect.TypeOf(dynamodb.BatchGetItemOutput{}),
		"BatchWriteItem":   reflect.TypeOf(dynamodb.BatchWriteItemOutput{}),
	}
	registry.RegisterServiceResponse("dynamodb", responses)
}
*/

// RegisterSQSResponses registers all SQS response types
// Uncomment when github.com/aws/aws-sdk-go-v2/service/sqs is installed
/*
func RegisterSQSResponses(registry *ResponseRegistry) {
	responses := map[string]reflect.Type{
		"CreateQueue":             reflect.TypeOf(sqs.CreateQueueOutput{}),
		"DeleteQueue":             reflect.TypeOf(sqs.DeleteQueueOutput{}),
		"GetQueueUrl":             reflect.TypeOf(sqs.GetQueueUrlOutput{}),
		"GetQueueAttributes":      reflect.TypeOf(sqs.GetQueueAttributesOutput{}),
		"ListQueues":              reflect.TypeOf(sqs.ListQueuesOutput{}),
		"SendMessage":             reflect.TypeOf(sqs.SendMessageOutput{}),
		"ReceiveMessage":          reflect.TypeOf(sqs.ReceiveMessageOutput{}),
		"DeleteMessage":           reflect.TypeOf(sqs.DeleteMessageOutput{}),
		"DeleteMessageBatch":      reflect.TypeOf(sqs.DeleteMessageBatchOutput{}),
		"SendMessageBatch":        reflect.TypeOf(sqs.SendMessageBatchOutput{}),
		"ChangeMessageVisibility": reflect.TypeOf(sqs.ChangeMessageVisibilityOutput{}),
	}
	registry.RegisterServiceResponse("sqs", responses)
}
*/

// RegisterAllResponseTypes registers all available service response types
func RegisterAllResponseTypes(registry *ResponseRegistry) {
	RegisterRDSResponses(registry)
	RegisterS3Responses(registry)
	// Uncomment as you add more services:
	// RegisterDynamoDBResponses(registry)
	// RegisterSQSResponses(registry)
}

