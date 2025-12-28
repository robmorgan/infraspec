package sqs

import (
	"time"
)

// ============================================================================
// Internal Storage Types
// ============================================================================

// Queue represents an SQS queue stored in state
type Queue struct {
	QueueName                 string            `json:"queueName"`
	QueueUrl                  string            `json:"queueUrl"`
	QueueArn                  string            `json:"queueArn"`
	CreatedTimestamp          int64             `json:"createdTimestamp"`
	LastModifiedTimestamp     int64             `json:"lastModifiedTimestamp"`
	VisibilityTimeout         int32             `json:"visibilityTimeout"`
	MaximumMessageSize        int32             `json:"maximumMessageSize"`
	MessageRetentionPeriod    int32             `json:"messageRetentionPeriod"`
	DelaySeconds              int32             `json:"delaySeconds"`
	ReceiveMessageWaitTime    int32             `json:"receiveMessageWaitTime"`
	ApproximateNumberOfMsgs   int64             `json:"approximateNumberOfMessages"`
	ApproximateNumMsgsNotVis  int64             `json:"approximateNumberOfMessagesNotVisible"`
	ApproximateNumMsgsDelayed int64             `json:"approximateNumberOfMessagesDelayed"`
	FifoQueue                 bool              `json:"fifoQueue"`
	ContentBasedDeduplication bool              `json:"contentBasedDeduplication"`
	DeduplicationScope        string            `json:"deduplicationScope,omitempty"`
	FifoThroughputLimit       string            `json:"fifoThroughputLimit,omitempty"`
	KmsMasterKeyId            string            `json:"kmsMasterKeyId,omitempty"`
	KmsDataKeyReusePeriod     int32             `json:"kmsDataKeyReusePeriodSeconds,omitempty"`
	SqsManagedSseEnabled      bool              `json:"sqsManagedSseEnabled"`
	Policy                    string            `json:"policy,omitempty"`
	RedrivePolicy             string            `json:"redrivePolicy,omitempty"`
	RedriveAllowPolicy        string            `json:"redriveAllowPolicy,omitempty"`
	Tags                      map[string]string `json:"tags,omitempty"`
}

// StoredMessage represents an SQS message in internal storage
// Named StoredMessage to avoid conflict with Smithy-generated Message type
type StoredMessage struct {
	MessageId               string            `json:"messageId"`
	ReceiptHandle           string            `json:"receiptHandle,omitempty"`
	MD5OfBody               string            `json:"md5OfBody"`
	Body                    string            `json:"body"`
	Attributes              map[string]string `json:"attributes,omitempty"`
	MD5OfMessageAttributes  string            `json:"md5OfMessageAttributes,omitempty"`
	MessageAttributes       map[string]string `json:"messageAttributes,omitempty"`
	SentTimestamp           int64             `json:"sentTimestamp"`
	FirstReceiveTimestamp   int64             `json:"firstReceiveTimestamp,omitempty"`
	ApproximateReceiveCount int               `json:"approximateReceiveCount"`
	VisibleAt               time.Time         `json:"visibleAt"`
	DelayUntil              time.Time         `json:"delayUntil,omitempty"`
	SequenceNumber          string            `json:"sequenceNumber,omitempty"`
	MessageGroupId          string            `json:"messageGroupId,omitempty"`
	MessageDeduplicationId  string            `json:"messageDeduplicationId,omitempty"`
}

// QueueMessages stores messages for a queue
type QueueMessages struct {
	Messages []StoredMessage `json:"messages"`
}

// EmptyResult for operations that return no data (Delete, Purge, etc.)
type EmptyResult struct{}

// ============================================================================
// JSON Response Types for AWS SDK v2
// ============================================================================

// JSONCreateQueueResult is the JSON result for CreateQueue
type JSONCreateQueueResult struct {
	QueueUrl string `json:"QueueUrl"`
}

// JSONGetQueueUrlResult is the JSON result for GetQueueUrl
type JSONGetQueueUrlResult struct {
	QueueUrl string `json:"QueueUrl"`
}

// JSONListQueuesResult is the JSON result for ListQueues
type JSONListQueuesResult struct {
	QueueUrls []string `json:"QueueUrls,omitempty"`
	NextToken string   `json:"NextToken,omitempty"`
}

// JSONGetQueueAttributesResult is the JSON result for GetQueueAttributes
type JSONGetQueueAttributesResult struct {
	Attributes map[string]string `json:"Attributes,omitempty"`
}

// JSONSendMessageResult is the JSON result for SendMessage
type JSONSendMessageResult struct {
	MessageId              string `json:"MessageId"`
	MD5OfMessageBody       string `json:"MD5OfMessageBody"`
	MD5OfMessageAttributes string `json:"MD5OfMessageAttributes,omitempty"`
	SequenceNumber         string `json:"SequenceNumber,omitempty"`
}

// JSONReceiveMessageResult is the JSON result for ReceiveMessage
type JSONReceiveMessageResult struct {
	Messages []JSONReceivedMessage `json:"Messages,omitempty"`
}

// JSONReceivedMessage represents a message in JSON format
type JSONReceivedMessage struct {
	MessageId              string                               `json:"MessageId"`
	ReceiptHandle          string                               `json:"ReceiptHandle"`
	MD5OfBody              string                               `json:"MD5OfBody"`
	Body                   string                               `json:"Body"`
	Attributes             map[string]string                    `json:"Attributes,omitempty"`
	MD5OfMessageAttributes string                               `json:"MD5OfMessageAttributes,omitempty"`
	MessageAttributes      map[string]JSONMessageAttributeValue `json:"MessageAttributes,omitempty"`
}

// JSONMessageAttributeValue represents a message attribute value in JSON
type JSONMessageAttributeValue struct {
	DataType    string `json:"DataType"`
	StringValue string `json:"StringValue,omitempty"`
	BinaryValue []byte `json:"BinaryValue,omitempty"`
}

// JSONSendMessageBatchResult is the JSON result for SendMessageBatch
type JSONSendMessageBatchResult struct {
	Successful []JSONSendMessageBatchResultEntry `json:"Successful,omitempty"`
	Failed     []JSONBatchResultErrorEntry       `json:"Failed,omitempty"`
}

// JSONSendMessageBatchResultEntry represents a successful batch send
type JSONSendMessageBatchResultEntry struct {
	Id                     string `json:"Id"`
	MessageId              string `json:"MessageId"`
	MD5OfMessageBody       string `json:"MD5OfMessageBody"`
	MD5OfMessageAttributes string `json:"MD5OfMessageAttributes,omitempty"`
	SequenceNumber         string `json:"SequenceNumber,omitempty"`
}

// JSONDeleteMessageBatchResult is the JSON result for DeleteMessageBatch
type JSONDeleteMessageBatchResult struct {
	Successful []JSONDeleteMessageBatchResultEntry `json:"Successful,omitempty"`
	Failed     []JSONBatchResultErrorEntry         `json:"Failed,omitempty"`
}

// JSONDeleteMessageBatchResultEntry represents a successful batch delete
type JSONDeleteMessageBatchResultEntry struct {
	Id string `json:"Id"`
}

// JSONBatchResultErrorEntry represents a failed batch entry
type JSONBatchResultErrorEntry struct {
	Id          string `json:"Id"`
	SenderFault bool   `json:"SenderFault"`
	Code        string `json:"Code"`
	Message     string `json:"Message,omitempty"`
}

// JSONListQueueTagsResult is the JSON result for ListQueueTags
type JSONListQueueTagsResult struct {
	Tags map[string]string `json:"Tags,omitempty"`
}
