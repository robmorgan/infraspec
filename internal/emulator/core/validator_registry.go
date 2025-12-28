package emulator

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// Add more services as needed:
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	// "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// RegisterRDSActions registers all RDS actions with the validator
func RegisterRDSActions(validator *SchemaValidator) {
	actions := map[string]reflect.Type{
		"CreateDBInstance":    reflect.TypeOf(rds.CreateDBInstanceInput{}),
		"DescribeDBInstances": reflect.TypeOf(rds.DescribeDBInstancesInput{}),
		"DeleteDBInstance":    reflect.TypeOf(rds.DeleteDBInstanceInput{}),
		"ModifyDBInstance":    reflect.TypeOf(rds.ModifyDBInstanceInput{}),
		"StartDBInstance":     reflect.TypeOf(rds.StartDBInstanceInput{}),
		"StopDBInstance":      reflect.TypeOf(rds.StopDBInstanceInput{}),
		"RebootDBInstance":    reflect.TypeOf(rds.RebootDBInstanceInput{}),
	}
	validator.RegisterService("rds", actions)
}

// RegisterS3Actions registers all S3 actions with the validator
// Uncomment when github.com/aws/aws-sdk-go-v2/service/s3 is installed
func RegisterS3Actions(validator *SchemaValidator) {
	actions := map[string]reflect.Type{
		"CreateBucket":  reflect.TypeOf(s3.CreateBucketInput{}),
		"ListBuckets":   reflect.TypeOf(s3.ListBucketsInput{}),
		"DeleteBucket":  reflect.TypeOf(s3.DeleteBucketInput{}),
		"PutObject":     reflect.TypeOf(s3.PutObjectInput{}),
		"GetObject":     reflect.TypeOf(s3.GetObjectInput{}),
		"DeleteObject":  reflect.TypeOf(s3.DeleteObjectInput{}),
		"ListObjects":   reflect.TypeOf(s3.ListObjectsInput{}),
		"ListObjectsV2": reflect.TypeOf(s3.ListObjectsV2Input{}),
		"HeadObject":    reflect.TypeOf(s3.HeadObjectInput{}),
		"CopyObject":    reflect.TypeOf(s3.CopyObjectInput{}),
	}
	validator.RegisterService("s3", actions)
}

// RegisterDynamoDBActions registers all DynamoDB actions with the validator
// Uncomment when github.com/aws/aws-sdk-go-v2/service/dynamodb is installed
/*
func RegisterDynamoDBActions(validator *SchemaValidator) {
	actions := map[string]reflect.Type{
		"CreateTable":      reflect.TypeOf(dynamodb.CreateTableInput{}),
		"DescribeTable":    reflect.TypeOf(dynamodb.DescribeTableInput{}),
		"ListTables":       reflect.TypeOf(dynamodb.ListTablesInput{}),
		"DeleteTable":      reflect.TypeOf(dynamodb.DeleteTableInput{}),
		"PutItem":          reflect.TypeOf(dynamodb.PutItemInput{}),
		"GetItem":          reflect.TypeOf(dynamodb.GetItemInput{}),
		"UpdateItem":       reflect.TypeOf(dynamodb.UpdateItemInput{}),
		"DeleteItem":       reflect.TypeOf(dynamodb.DeleteItemInput{}),
		"Query":            reflect.TypeOf(dynamodb.QueryInput{}),
		"Scan":             reflect.TypeOf(dynamodb.ScanInput{}),
		"BatchGetItem":     reflect.TypeOf(dynamodb.BatchGetItemInput{}),
		"BatchWriteItem":   reflect.TypeOf(dynamodb.BatchWriteItemInput{}),
	}
	validator.RegisterService("dynamodb", actions)
}
*/

// RegisterSQSActions registers all SQS actions with the validator
// Uncomment when github.com/aws/aws-sdk-go-v2/service/sqs is installed
/*
func RegisterSQSActions(validator *SchemaValidator) {
	actions := map[string]reflect.Type{
		"CreateQueue":             reflect.TypeOf(sqs.CreateQueueInput{}),
		"DeleteQueue":             reflect.TypeOf(sqs.DeleteQueueInput{}),
		"GetQueueUrl":             reflect.TypeOf(sqs.GetQueueUrlInput{}),
		"GetQueueAttributes":      reflect.TypeOf(sqs.GetQueueAttributesInput{}),
		"ListQueues":              reflect.TypeOf(sqs.ListQueuesInput{}),
		"SendMessage":             reflect.TypeOf(sqs.SendMessageInput{}),
		"ReceiveMessage":          reflect.TypeOf(sqs.ReceiveMessageInput{}),
		"DeleteMessage":           reflect.TypeOf(sqs.DeleteMessageInput{}),
		"DeleteMessageBatch":      reflect.TypeOf(sqs.DeleteMessageBatchInput{}),
		"SendMessageBatch":        reflect.TypeOf(sqs.SendMessageBatchInput{}),
		"ChangeMessageVisibility": reflect.TypeOf(sqs.ChangeMessageVisibilityInput{}),
	}
	validator.RegisterService("sqs", actions)
}
*/

// RegisterAllServices registers all available services with the validator
func RegisterAllServices(validator *SchemaValidator) {
	RegisterRDSActions(validator)
	RegisterS3Actions(validator)
	// Uncomment as you add more services:
	// RegisterDynamoDBActions(validator)
	// RegisterSQSActions(validator)
}
