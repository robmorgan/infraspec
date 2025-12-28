package sqs

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
)

const (
	defaultAccountID              = "123456789012"
	defaultRegion                 = "us-east-1"
	defaultVisibilityTimeout      = 30
	defaultMaxMessageSize         = 262144 // 256 KB
	defaultMessageRetentionPeriod = 345600 // 4 days
	defaultDelaySeconds           = 0
	defaultReceiveWaitTime        = 0
	defaultKmsReusePeriod         = 300
)

// SQSService implements the AWS SQS service emulator
type SQSService struct {
	state     emulator.StateManager
	validator emulator.Validator
}

// NewSQSService creates a new SQS service instance
func NewSQSService(state emulator.StateManager, validator emulator.Validator) *SQSService {
	return &SQSService{
		state:     state,
		validator: validator,
	}
}

// ServiceName returns the service identifier
func (s *SQSService) ServiceName() string {
	return "sqs"
}

// SupportedActions returns the list of AWS API actions this service handles.
// Used by the router to determine which service handles a given Query Protocol request.
func (s *SQSService) SupportedActions() []string {
	return []string{
		// Queue operations
		"CreateQueue",
		"DeleteQueue",
		"ListQueues",
		"GetQueueUrl",
		"GetQueueAttributes",
		"SetQueueAttributes",
		"PurgeQueue",
		// Message operations
		"SendMessage",
		"ReceiveMessage",
		"DeleteMessage",
		"ChangeMessageVisibility",
		// Batch operations
		"SendMessageBatch",
		"DeleteMessageBatch",
		// Tag operations
		"TagQueue",
		"UntagQueue",
		"ListQueueTags",
	}
}

// HandleRequest routes incoming requests to the appropriate handler
func (s *SQSService) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	action := s.extractAction(req)
	if action == "" {
		return s.errorResponse(400, "InvalidAction", "Missing or invalid action"), nil
	}

	switch action {
	// Queue operations
	case "CreateQueue":
		input, err := emulator.ParseJSONRequest[CreateQueueRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.createQueue(ctx, input)
	case "DeleteQueue":
		input, err := emulator.ParseJSONRequest[DeleteQueueRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteQueue(ctx, input)
	case "ListQueues":
		input, err := emulator.ParseJSONRequest[ListQueuesRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listQueues(ctx, input)
	case "GetQueueUrl":
		input, err := emulator.ParseJSONRequest[GetQueueUrlRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.getQueueUrl(ctx, input)
	case "GetQueueAttributes":
		input, err := emulator.ParseJSONRequest[GetQueueAttributesRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.getQueueAttributes(ctx, input)
	case "SetQueueAttributes":
		input, err := emulator.ParseJSONRequest[SetQueueAttributesRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.setQueueAttributes(ctx, input)
	case "PurgeQueue":
		input, err := emulator.ParseJSONRequest[PurgeQueueRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.purgeQueue(ctx, input)

	// Message operations
	case "SendMessage":
		input, err := emulator.ParseJSONRequest[SendMessageRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.sendMessage(ctx, input)
	case "ReceiveMessage":
		input, err := emulator.ParseJSONRequest[ReceiveMessageRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.receiveMessage(ctx, input)
	case "DeleteMessage":
		input, err := emulator.ParseJSONRequest[DeleteMessageRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteMessage(ctx, input)
	case "ChangeMessageVisibility":
		input, err := emulator.ParseJSONRequest[ChangeMessageVisibilityRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.changeMessageVisibility(ctx, input)

	// Batch operations
	case "SendMessageBatch":
		input, err := emulator.ParseJSONRequest[SendMessageBatchRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.sendMessageBatch(ctx, input)
	case "DeleteMessageBatch":
		input, err := emulator.ParseJSONRequest[DeleteMessageBatchRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.deleteMessageBatch(ctx, input)

	// Tag operations
	case "TagQueue":
		input, err := emulator.ParseJSONRequest[TagQueueRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.tagQueue(ctx, input)
	case "UntagQueue":
		input, err := emulator.ParseJSONRequest[UntagQueueRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.untagQueue(ctx, input)
	case "ListQueueTags":
		input, err := emulator.ParseJSONRequest[ListQueueTagsRequest](req.Body)
		if err != nil {
			return s.errorResponse(400, "SerializationException", err.Error()), nil
		}
		return s.listQueueTags(ctx, input)

	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

func (s *SQSService) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

// ============================================================================
// Queue Operations
// ============================================================================

func (s *SQSService) createQueue(ctx context.Context, input *CreateQueueRequest) (*emulator.AWSResponse, error) {
	if input.QueueName == nil || *input.QueueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueName is required"), nil
	}
	queueName := *input.QueueName

	// Validate queue name
	if err := validateQueueName(queueName); err != nil {
		return s.errorResponse(400, "InvalidParameterValue", err.Error()), nil
	}

	// Check if FIFO queue (name must end with .fifo)
	isFifo := strings.HasSuffix(queueName, ".fifo")

	// Check if queue already exists
	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	if s.state.Exists(stateKey) {
		// Return existing queue URL (idempotent)
		var existingQueue Queue
		if err := s.state.Get(stateKey, &existingQueue); err == nil {
			result := JSONCreateQueueResult{QueueUrl: existingQueue.QueueUrl}
			return s.successResponse("CreateQueue", result)
		}
	}

	now := time.Now().Unix()
	queueUrl := fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/%s", defaultRegion, defaultAccountID, queueName)
	queueArn := fmt.Sprintf("arn:aws:sqs:%s:%s:%s", defaultRegion, defaultAccountID, queueName)

	queue := Queue{
		QueueName:              queueName,
		QueueUrl:               queueUrl,
		QueueArn:               queueArn,
		CreatedTimestamp:       now,
		LastModifiedTimestamp:  now,
		VisibilityTimeout:      defaultVisibilityTimeout,
		MaximumMessageSize:     defaultMaxMessageSize,
		MessageRetentionPeriod: defaultMessageRetentionPeriod,
		DelaySeconds:           defaultDelaySeconds,
		ReceiveMessageWaitTime: defaultReceiveWaitTime,
		FifoQueue:              isFifo,
		SqsManagedSseEnabled:   true, // Default to SSE enabled
		Tags:                   make(map[string]string),
	}

	// Apply attributes from input
	s.applyQueueAttributesFromMap(&queue, input.Attributes)

	// Apply tags from input
	if len(input.Tags) > 0 {
		queue.Tags = input.Tags
	}

	if err := s.state.Set(stateKey, &queue); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store queue"), nil
	}

	// Initialize empty message store
	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	if err := s.state.Set(msgKey, &QueueMessages{Messages: []StoredMessage{}}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to initialize message store"), nil
	}

	result := JSONCreateQueueResult{QueueUrl: queueUrl}
	return s.successResponse("CreateQueue", result)
}

func (s *SQSService) deleteQueue(ctx context.Context, input *DeleteQueueRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// Delete queue
	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete queue"), nil
	}

	// Delete messages
	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	s.state.Delete(msgKey)

	return s.successResponse("DeleteQueue", EmptyResult{})
}

func (s *SQSService) listQueues(ctx context.Context, input *ListQueuesRequest) (*emulator.AWSResponse, error) {
	var queueNamePrefix string
	if input.QueueNamePrefix != nil {
		queueNamePrefix = *input.QueueNamePrefix
	}

	keys, err := s.state.List("sqs:queue:")
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to list queues"), nil
	}

	var queueUrls []string
	for _, key := range keys {
		var queue Queue
		if err := s.state.Get(key, &queue); err == nil {
			if queueNamePrefix == "" || strings.HasPrefix(queue.QueueName, queueNamePrefix) {
				queueUrls = append(queueUrls, queue.QueueUrl)
			}
		}
	}

	result := JSONListQueuesResult{QueueUrls: queueUrls}
	return s.successResponse("ListQueues", result)
}

func (s *SQSService) getQueueUrl(ctx context.Context, input *GetQueueUrlRequest) (*emulator.AWSResponse, error) {
	if input.QueueName == nil || *input.QueueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueName is required"), nil
	}
	queueName := *input.QueueName

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	result := JSONGetQueueUrlResult{QueueUrl: queue.QueueUrl}
	return s.successResponse("GetQueueUrl", result)
}

func (s *SQSService) getQueueAttributes(ctx context.Context, input *GetQueueAttributesRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// Convert requested attributes to string slice
	var requestedAttrs []string
	for _, attr := range input.AttributeNames {
		requestedAttrs = append(requestedAttrs, string(attr))
	}

	// Build attributes map (JSON format uses map, not array)
	attrs := s.buildQueueAttributesMap(&queue, requestedAttrs)

	result := JSONGetQueueAttributesResult{Attributes: attrs}
	return s.successResponse("GetQueueAttributes", result)
}

func (s *SQSService) setQueueAttributes(ctx context.Context, input *SetQueueAttributesRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// Apply new attributes
	s.applyQueueAttributesFromMap(&queue, input.Attributes)
	queue.LastModifiedTimestamp = time.Now().Unix()

	if err := s.state.Set(stateKey, &queue); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update queue"), nil
	}

	return s.successResponse("SetQueueAttributes", EmptyResult{})
}

func (s *SQSService) purgeQueue(ctx context.Context, input *PurgeQueueRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// Clear all messages
	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	if err := s.state.Set(msgKey, &QueueMessages{Messages: []StoredMessage{}}); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to purge queue"), nil
	}

	return s.successResponse("PurgeQueue", EmptyResult{})
}

// ============================================================================
// Message Operations
// ============================================================================

func (s *SQSService) sendMessage(ctx context.Context, input *SendMessageRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	if input.MessageBody == nil || *input.MessageBody == "" {
		return s.errorResponse(400, "InvalidParameterValue", "MessageBody is required"), nil
	}
	messageBody := *input.MessageBody

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// Validate message size
	if len(messageBody) > int(queue.MaximumMessageSize) {
		return s.errorResponse(400, "InvalidParameterValue", "Message body exceeds maximum message size"), nil
	}

	// Calculate MD5
	md5Hash := md5.Sum([]byte(messageBody))
	md5Str := hex.EncodeToString(md5Hash[:])

	// Create message
	messageId := uuid.New().String()
	now := time.Now()

	delaySeconds := queue.DelaySeconds
	if input.DelaySeconds != nil {
		delaySeconds = *input.DelaySeconds
	}

	msg := StoredMessage{
		MessageId:     messageId,
		MD5OfBody:     md5Str,
		Body:          messageBody,
		SentTimestamp: now.Unix() * 1000, // milliseconds
		VisibleAt:     now.Add(time.Duration(delaySeconds) * time.Second),
	}

	// Handle FIFO queue specifics
	if queue.FifoQueue {
		if input.MessageGroupId != nil {
			msg.MessageGroupId = *input.MessageGroupId
		}
		if msg.MessageGroupId == "" {
			return s.errorResponse(400, "MissingParameter", "MessageGroupId is required for FIFO queues"), nil
		}

		if input.MessageDeduplicationId != nil {
			msg.MessageDeduplicationId = *input.MessageDeduplicationId
		}
		if msg.MessageDeduplicationId == "" && !queue.ContentBasedDeduplication {
			return s.errorResponse(400, "MissingParameter", "MessageDeduplicationId is required when ContentBasedDeduplication is disabled"), nil
		}
		if msg.MessageDeduplicationId == "" {
			// Use content-based deduplication
			msg.MessageDeduplicationId = md5Str
		}

		msg.SequenceNumber = generateSequenceNumber()
	}

	// Store message
	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	var queueMsgs QueueMessages
	if err := s.state.Get(msgKey, &queueMsgs); err != nil {
		queueMsgs = QueueMessages{Messages: []StoredMessage{}}
	}

	queueMsgs.Messages = append(queueMsgs.Messages, msg)
	if err := s.state.Set(msgKey, &queueMsgs); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store message"), nil
	}

	result := JSONSendMessageResult{
		MessageId:        messageId,
		MD5OfMessageBody: md5Str,
		SequenceNumber:   msg.SequenceNumber,
	}
	return s.successResponse("SendMessage", result)
}

func (s *SQSService) receiveMessage(ctx context.Context, input *ReceiveMessageRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	maxMessages := int32(1)
	if input.MaxNumberOfMessages != nil {
		maxMessages = *input.MaxNumberOfMessages
	}
	if maxMessages < 1 || maxMessages > 10 {
		maxMessages = 1
	}

	visibilityTimeout := queue.VisibilityTimeout
	if input.VisibilityTimeout != nil {
		visibilityTimeout = *input.VisibilityTimeout
	}

	// Get messages
	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	var queueMsgs QueueMessages
	if err := s.state.Get(msgKey, &queueMsgs); err != nil {
		queueMsgs = QueueMessages{Messages: []StoredMessage{}}
	}

	now := time.Now()
	var receivedMsgs []JSONReceivedMessage
	var updatedMsgs []StoredMessage

	for i := range queueMsgs.Messages {
		msg := &queueMsgs.Messages[i]

		// Skip messages that are not yet visible
		if msg.VisibleAt.After(now) {
			updatedMsgs = append(updatedMsgs, *msg)
			continue
		}

		// Skip messages still in delay
		if !msg.DelayUntil.IsZero() && msg.DelayUntil.After(now) {
			updatedMsgs = append(updatedMsgs, *msg)
			continue
		}

		if len(receivedMsgs) < int(maxMessages) {
			// Generate receipt handle
			receiptHandle := generateReceiptHandle()
			msg.ReceiptHandle = receiptHandle
			msg.ApproximateReceiveCount++
			if msg.FirstReceiveTimestamp == 0 {
				msg.FirstReceiveTimestamp = now.Unix() * 1000
			}
			msg.VisibleAt = now.Add(time.Duration(visibilityTimeout) * time.Second)

			// Build message attributes for JSON response (map instead of array)
			attrs := map[string]string{
				"SenderId":                         defaultAccountID,
				"SentTimestamp":                    strconv.FormatInt(msg.SentTimestamp, 10),
				"ApproximateReceiveCount":          strconv.Itoa(msg.ApproximateReceiveCount),
				"ApproximateFirstReceiveTimestamp": strconv.FormatInt(msg.FirstReceiveTimestamp, 10),
			}

			jsonMsg := JSONReceivedMessage{
				MessageId:     msg.MessageId,
				ReceiptHandle: receiptHandle,
				MD5OfBody:     msg.MD5OfBody,
				Body:          msg.Body,
				Attributes:    attrs,
			}

			receivedMsgs = append(receivedMsgs, jsonMsg)
		}

		updatedMsgs = append(updatedMsgs, *msg)
	}

	// Update message store
	queueMsgs.Messages = updatedMsgs
	if err := s.state.Set(msgKey, &queueMsgs); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update messages"), nil
	}

	result := JSONReceiveMessageResult{Messages: receivedMsgs}
	return s.successResponse("ReceiveMessage", result)
}

func (s *SQSService) deleteMessage(ctx context.Context, input *DeleteMessageRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	if input.ReceiptHandle == nil || *input.ReceiptHandle == "" {
		return s.errorResponse(400, "InvalidParameterValue", "ReceiptHandle is required"), nil
	}
	receiptHandle := *input.ReceiptHandle

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// Find and delete message
	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	var queueMsgs QueueMessages
	if err := s.state.Get(msgKey, &queueMsgs); err != nil {
		return s.errorResponse(400, "ReceiptHandleIsInvalid", "The receipt handle provided is not valid"), nil
	}

	found := false
	newMsgs := make([]StoredMessage, 0, len(queueMsgs.Messages))
	for _, msg := range queueMsgs.Messages {
		if msg.ReceiptHandle == receiptHandle {
			found = true
		} else {
			newMsgs = append(newMsgs, msg)
		}
	}

	if !found {
		return s.errorResponse(400, "ReceiptHandleIsInvalid", "The receipt handle provided is not valid"), nil
	}

	queueMsgs.Messages = newMsgs
	if err := s.state.Set(msgKey, &queueMsgs); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to delete message"), nil
	}

	return s.successResponse("DeleteMessage", EmptyResult{})
}

func (s *SQSService) changeMessageVisibility(ctx context.Context, input *ChangeMessageVisibilityRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	if input.ReceiptHandle == nil || *input.ReceiptHandle == "" {
		return s.errorResponse(400, "InvalidParameterValue", "ReceiptHandle is required"), nil
	}
	receiptHandle := *input.ReceiptHandle

	visibilityTimeout := int32(0)
	if input.VisibilityTimeout != nil {
		visibilityTimeout = *input.VisibilityTimeout
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	var queueMsgs QueueMessages
	if err := s.state.Get(msgKey, &queueMsgs); err != nil {
		return s.errorResponse(400, "ReceiptHandleIsInvalid", "The receipt handle provided is not valid"), nil
	}

	found := false
	for i := range queueMsgs.Messages {
		if queueMsgs.Messages[i].ReceiptHandle == receiptHandle {
			queueMsgs.Messages[i].VisibleAt = time.Now().Add(time.Duration(visibilityTimeout) * time.Second)
			found = true
			break
		}
	}

	if !found {
		return s.errorResponse(400, "ReceiptHandleIsInvalid", "The receipt handle provided is not valid"), nil
	}

	if err := s.state.Set(msgKey, &queueMsgs); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update message visibility"), nil
	}

	return s.successResponse("ChangeMessageVisibility", EmptyResult{})
}

// ============================================================================
// Batch Operations
// ============================================================================

func (s *SQSService) sendMessageBatch(ctx context.Context, input *SendMessageBatchRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	var successful []JSONSendMessageBatchResultEntry
	var failed []JSONBatchResultErrorEntry

	// Process batch entries from typed input
	for _, entry := range input.Entries {
		if entry.Id == nil || entry.MessageBody == nil {
			continue
		}

		id := *entry.Id
		body := *entry.MessageBody

		// Create message
		md5Hash := md5.Sum([]byte(body))
		md5Str := hex.EncodeToString(md5Hash[:])
		messageId := uuid.New().String()

		msg := StoredMessage{
			MessageId:     messageId,
			MD5OfBody:     md5Str,
			Body:          body,
			SentTimestamp: time.Now().Unix() * 1000,
			VisibleAt:     time.Now(),
		}

		// Store message
		msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
		var queueMsgs QueueMessages
		if err := s.state.Get(msgKey, &queueMsgs); err != nil {
			queueMsgs = QueueMessages{Messages: []StoredMessage{}}
		}
		queueMsgs.Messages = append(queueMsgs.Messages, msg)
		if err := s.state.Set(msgKey, &queueMsgs); err != nil {
			failed = append(failed, JSONBatchResultErrorEntry{
				Id:          id,
				SenderFault: false,
				Code:        "InternalFailure",
				Message:     "Failed to store message",
			})
			continue
		}

		successful = append(successful, JSONSendMessageBatchResultEntry{
			Id:               id,
			MessageId:        messageId,
			MD5OfMessageBody: md5Str,
		})
	}

	result := JSONSendMessageBatchResult{
		Successful: successful,
		Failed:     failed,
	}
	return s.successResponse("SendMessageBatch", result)
}

func (s *SQSService) deleteMessageBatch(ctx context.Context, input *DeleteMessageBatchRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	if !s.state.Exists(stateKey) {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	msgKey := fmt.Sprintf("sqs:messages:%s", queueName)
	var queueMsgs QueueMessages
	if err := s.state.Get(msgKey, &queueMsgs); err != nil {
		queueMsgs = QueueMessages{Messages: []StoredMessage{}}
	}

	var successful []JSONDeleteMessageBatchResultEntry
	var failed []JSONBatchResultErrorEntry

	// Process batch entries from typed input
	for _, entry := range input.Entries {
		if entry.Id == nil || entry.ReceiptHandle == nil {
			continue
		}

		id := *entry.Id
		handle := *entry.ReceiptHandle

		// Find and remove message
		found := false
		newMsgs := make([]StoredMessage, 0, len(queueMsgs.Messages))
		for _, msg := range queueMsgs.Messages {
			if msg.ReceiptHandle == handle {
				found = true
			} else {
				newMsgs = append(newMsgs, msg)
			}
		}

		if found {
			queueMsgs.Messages = newMsgs
			successful = append(successful, JSONDeleteMessageBatchResultEntry{Id: id})
		} else {
			failed = append(failed, JSONBatchResultErrorEntry{
				Id:          id,
				SenderFault: true,
				Code:        "ReceiptHandleIsInvalid",
				Message:     "The receipt handle provided is not valid",
			})
		}
	}

	if err := s.state.Set(msgKey, &queueMsgs); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update message store"), nil
	}

	result := JSONDeleteMessageBatchResult{
		Successful: successful,
		Failed:     failed,
	}
	return s.successResponse("DeleteMessageBatch", result)
}

// ============================================================================
// Tag Operations
// ============================================================================

func (s *SQSService) tagQueue(ctx context.Context, input *TagQueueRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	if queue.Tags == nil {
		queue.Tags = make(map[string]string)
	}

	// Apply tags from typed input
	for k, v := range input.Tags {
		queue.Tags[k] = v
	}

	if err := s.state.Set(stateKey, &queue); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update queue tags"), nil
	}

	return s.successResponse("TagQueue", EmptyResult{})
}

func (s *SQSService) untagQueue(ctx context.Context, input *UntagQueueRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// Remove tag keys from typed input
	for _, key := range input.TagKeys {
		delete(queue.Tags, key)
	}

	if err := s.state.Set(stateKey, &queue); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update queue tags"), nil
	}

	return s.successResponse("UntagQueue", EmptyResult{})
}

func (s *SQSService) listQueueTags(ctx context.Context, input *ListQueueTagsRequest) (*emulator.AWSResponse, error) {
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return s.errorResponse(400, "InvalidParameterValue", "QueueUrl is required"), nil
	}

	queueName := extractQueueNameFromUrl(*input.QueueUrl)
	if queueName == "" {
		return s.errorResponse(400, "InvalidParameterValue", "Invalid QueueUrl"), nil
	}

	stateKey := fmt.Sprintf("sqs:queue:%s", queueName)
	var queue Queue
	if err := s.state.Get(stateKey, &queue); err != nil {
		return s.errorResponse(400, "AWS.SimpleQueueService.NonExistentQueue", "The specified queue does not exist"), nil
	}

	// JSON format uses map[string]string for tags
	result := JSONListQueueTagsResult{Tags: queue.Tags}
	return s.successResponse("ListQueueTags", result)
}

// ============================================================================
// Helper Functions
// ============================================================================

func (s *SQSService) successResponse(action string, data interface{}) (*emulator.AWSResponse, error) {
	// SQS uses JSON protocol in AWS SDK v2
	return emulator.BuildJSONResponse(200, data)
}

func (s *SQSService) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	// SQS uses JSON protocol in AWS SDK v2
	return emulator.BuildJSONErrorResponse(statusCode, code, message)
}

func (s *SQSService) applyQueueAttributesFromMap(queue *Queue, attrs map[string]string) {
	for name, value := range attrs {
		s.setQueueAttribute(queue, name, value)
	}
}

func (s *SQSService) setQueueAttribute(queue *Queue, name, value string) {
	switch name {
	case "VisibilityTimeout":
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			queue.VisibilityTimeout = int32(v)
		}
	case "MaximumMessageSize":
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			queue.MaximumMessageSize = int32(v)
		}
	case "MessageRetentionPeriod":
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			queue.MessageRetentionPeriod = int32(v)
		}
	case "DelaySeconds":
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			queue.DelaySeconds = int32(v)
		}
	case "ReceiveMessageWaitTimeSeconds":
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			queue.ReceiveMessageWaitTime = int32(v)
		}
	case "Policy":
		queue.Policy = value
	case "RedrivePolicy":
		queue.RedrivePolicy = value
	case "RedriveAllowPolicy":
		queue.RedriveAllowPolicy = value
	case "KmsMasterKeyId":
		queue.KmsMasterKeyId = value
	case "KmsDataKeyReusePeriodSeconds":
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			queue.KmsDataKeyReusePeriod = int32(v)
		}
	case "SqsManagedSseEnabled":
		queue.SqsManagedSseEnabled = value == "true"
	case "ContentBasedDeduplication":
		queue.ContentBasedDeduplication = value == "true"
	case "DeduplicationScope":
		queue.DeduplicationScope = value
	case "FifoThroughputLimit":
		queue.FifoThroughputLimit = value
	}
}

// buildQueueAttributesMap returns attributes as a map for JSON responses
func (s *SQSService) buildQueueAttributesMap(queue *Queue, requestedAttrs []string) map[string]string {
	allAttrs := map[string]string{
		"QueueArn":                              queue.QueueArn,
		"ApproximateNumberOfMessages":           strconv.FormatInt(queue.ApproximateNumberOfMsgs, 10),
		"ApproximateNumberOfMessagesNotVisible": strconv.FormatInt(queue.ApproximateNumMsgsNotVis, 10),
		"ApproximateNumberOfMessagesDelayed":    strconv.FormatInt(queue.ApproximateNumMsgsDelayed, 10),
		"CreatedTimestamp":                      strconv.FormatInt(queue.CreatedTimestamp, 10),
		"LastModifiedTimestamp":                 strconv.FormatInt(queue.LastModifiedTimestamp, 10),
		"VisibilityTimeout":                     strconv.FormatInt(int64(queue.VisibilityTimeout), 10),
		"MaximumMessageSize":                    strconv.FormatInt(int64(queue.MaximumMessageSize), 10),
		"MessageRetentionPeriod":                strconv.FormatInt(int64(queue.MessageRetentionPeriod), 10),
		"DelaySeconds":                          strconv.FormatInt(int64(queue.DelaySeconds), 10),
		"ReceiveMessageWaitTimeSeconds":         strconv.FormatInt(int64(queue.ReceiveMessageWaitTime), 10),
		"SqsManagedSseEnabled":                  strconv.FormatBool(queue.SqsManagedSseEnabled),
	}

	if queue.FifoQueue {
		allAttrs["FifoQueue"] = "true"
		allAttrs["ContentBasedDeduplication"] = strconv.FormatBool(queue.ContentBasedDeduplication)
		if queue.DeduplicationScope != "" {
			allAttrs["DeduplicationScope"] = queue.DeduplicationScope
		}
		if queue.FifoThroughputLimit != "" {
			allAttrs["FifoThroughputLimit"] = queue.FifoThroughputLimit
		}
	}

	if queue.Policy != "" {
		allAttrs["Policy"] = queue.Policy
	}
	if queue.RedrivePolicy != "" {
		allAttrs["RedrivePolicy"] = queue.RedrivePolicy
	}
	if queue.RedriveAllowPolicy != "" {
		allAttrs["RedriveAllowPolicy"] = queue.RedriveAllowPolicy
	}
	if queue.KmsMasterKeyId != "" {
		allAttrs["KmsMasterKeyId"] = queue.KmsMasterKeyId
		allAttrs["KmsDataKeyReusePeriodSeconds"] = strconv.FormatInt(int64(queue.KmsDataKeyReusePeriod), 10)
	}

	// If "All" is requested or no specific attrs requested, return all
	returnAll := len(requestedAttrs) == 0
	for _, a := range requestedAttrs {
		if a == "All" {
			returnAll = true
			break
		}
	}

	if returnAll {
		return allAttrs
	}

	result := make(map[string]string)
	for _, name := range requestedAttrs {
		if value, ok := allAttrs[name]; ok {
			result[name] = value
		}
	}
	return result
}

func validateQueueName(name string) error {
	if len(name) < 1 || len(name) > 80 {
		return fmt.Errorf("queue name must be between 1 and 80 characters")
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return fmt.Errorf("queue name can only contain alphanumeric characters, hyphens, underscores, and periods")
		}
	}

	return nil
}

func extractQueueNameFromUrl(queueUrl string) string {
	// Extract queue name from URL like https://sqs.us-east-1.amazonaws.com/123456789012/my-queue
	parts := strings.Split(queueUrl, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func generateReceiptHandle() string {
	b := make([]byte, 64)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func generateSequenceNumber() string {
	// Generate a sequence number similar to AWS FIFO queues
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
