package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/robmorgan/infraspec/pkg/awshelpers"
)

// Ensure the `AWSAsserter` struct implements the `SQSAsserter` interface.
var _ SQSAsserter = (*AWSAsserter)(nil)

// SQSAsserter defines SQS-specific assertions
type SQSAsserter interface {
	AssertSQSDescribeQueues() error
	AssertQueueExists(queueName string) error
	AssertQueueVisibilityTimeout(queueName string, timeout int) error
	AssertQueueDelaySeconds(queueName string, delay int) error
	AssertQueueMaxMessageSize(queueName string, size int) error
	AssertQueueMessageRetentionPeriod(queueName string, period int) error
	AssertQueueReceiveMessageWaitTime(queueName string, waitTime int) error
	AssertQueueIsFifo(queueName string) error
	AssertQueueHasDeadLetterQueue(queueName string) error
	AssertQueueTags(queueName string, expectedTags map[string]string) error
	AssertQueueEncryption(queueName string, expectEncrypted bool) error
}

// AssertSQSDescribeQueues checks if the AWS account has permission to list SQS queues
func (a *AWSAsserter) AssertSQSDescribeQueues() error {
	client, err := a.createSQSClient()
	if err != nil {
		return err
	}

	// List queues to verify access
	_, err = client.ListQueues(context.TODO(), &sqs.ListQueuesInput{})
	if err != nil {
		return fmt.Errorf("error listing SQS queues: %w", err)
	}

	return nil
}

// AssertQueueExists checks if an SQS queue exists
func (a *AWSAsserter) AssertQueueExists(queueName string) error {
	client, err := a.createSQSClient()
	if err != nil {
		return err
	}

	// Get queue URL to verify it exists
	_, err = client.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return fmt.Errorf("queue %s does not exist or is not accessible: %w", queueName, err)
	}

	return nil
}

// AssertQueueVisibilityTimeout checks if a queue has the expected visibility timeout
func (a *AWSAsserter) AssertQueueVisibilityTimeout(queueName string, timeout int) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{types.QueueAttributeNameVisibilityTimeout})
	if err != nil {
		return err
	}

	actualTimeout, ok := attrs[string(types.QueueAttributeNameVisibilityTimeout)]
	if !ok {
		return fmt.Errorf("queue %s does not have VisibilityTimeout attribute", queueName)
	}

	actualTimeoutInt, err := strconv.Atoi(actualTimeout)
	if err != nil {
		return fmt.Errorf("invalid VisibilityTimeout value: %s", actualTimeout)
	}

	if actualTimeoutInt != timeout {
		return fmt.Errorf("queue %s has VisibilityTimeout %d, expected %d", queueName, actualTimeoutInt, timeout)
	}

	return nil
}

// AssertQueueDelaySeconds checks if a queue has the expected delay seconds
func (a *AWSAsserter) AssertQueueDelaySeconds(queueName string, delay int) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{types.QueueAttributeNameDelaySeconds})
	if err != nil {
		return err
	}

	actualDelay, ok := attrs[string(types.QueueAttributeNameDelaySeconds)]
	if !ok {
		return fmt.Errorf("queue %s does not have DelaySeconds attribute", queueName)
	}

	actualDelayInt, err := strconv.Atoi(actualDelay)
	if err != nil {
		return fmt.Errorf("invalid DelaySeconds value: %s", actualDelay)
	}

	if actualDelayInt != delay {
		return fmt.Errorf("queue %s has DelaySeconds %d, expected %d", queueName, actualDelayInt, delay)
	}

	return nil
}

// AssertQueueMaxMessageSize checks if a queue has the expected max message size
func (a *AWSAsserter) AssertQueueMaxMessageSize(queueName string, size int) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{types.QueueAttributeNameMaximumMessageSize})
	if err != nil {
		return err
	}

	actualSize, ok := attrs[string(types.QueueAttributeNameMaximumMessageSize)]
	if !ok {
		return fmt.Errorf("queue %s does not have MaximumMessageSize attribute", queueName)
	}

	actualSizeInt, err := strconv.Atoi(actualSize)
	if err != nil {
		return fmt.Errorf("invalid MaximumMessageSize value: %s", actualSize)
	}

	if actualSizeInt != size {
		return fmt.Errorf("queue %s has MaximumMessageSize %d, expected %d", queueName, actualSizeInt, size)
	}

	return nil
}

// AssertQueueMessageRetentionPeriod checks if a queue has the expected message retention period
func (a *AWSAsserter) AssertQueueMessageRetentionPeriod(queueName string, period int) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{types.QueueAttributeNameMessageRetentionPeriod})
	if err != nil {
		return err
	}

	actualPeriod, ok := attrs[string(types.QueueAttributeNameMessageRetentionPeriod)]
	if !ok {
		return fmt.Errorf("queue %s does not have MessageRetentionPeriod attribute", queueName)
	}

	actualPeriodInt, err := strconv.Atoi(actualPeriod)
	if err != nil {
		return fmt.Errorf("invalid MessageRetentionPeriod value: %s", actualPeriod)
	}

	if actualPeriodInt != period {
		return fmt.Errorf("queue %s has MessageRetentionPeriod %d, expected %d", queueName, actualPeriodInt, period)
	}

	return nil
}

// AssertQueueReceiveMessageWaitTime checks if a queue has the expected receive message wait time
func (a *AWSAsserter) AssertQueueReceiveMessageWaitTime(queueName string, waitTime int) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{types.QueueAttributeNameReceiveMessageWaitTimeSeconds})
	if err != nil {
		return err
	}

	actualWaitTime, ok := attrs[string(types.QueueAttributeNameReceiveMessageWaitTimeSeconds)]
	if !ok {
		return fmt.Errorf("queue %s does not have ReceiveMessageWaitTimeSeconds attribute", queueName)
	}

	actualWaitTimeInt, err := strconv.Atoi(actualWaitTime)
	if err != nil {
		return fmt.Errorf("invalid ReceiveMessageWaitTimeSeconds value: %s", actualWaitTime)
	}

	if actualWaitTimeInt != waitTime {
		return fmt.Errorf("queue %s has ReceiveMessageWaitTimeSeconds %d, expected %d", queueName, actualWaitTimeInt, waitTime)
	}

	return nil
}

// AssertQueueIsFifo checks if a queue is a FIFO queue
func (a *AWSAsserter) AssertQueueIsFifo(queueName string) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{types.QueueAttributeNameFifoQueue})
	if err != nil {
		return err
	}

	isFifo, ok := attrs[string(types.QueueAttributeNameFifoQueue)]
	if !ok || isFifo != "true" {
		return fmt.Errorf("queue %s is not a FIFO queue", queueName)
	}

	return nil
}

// AssertQueueHasDeadLetterQueue checks if a queue has a dead letter queue configured
func (a *AWSAsserter) AssertQueueHasDeadLetterQueue(queueName string) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{types.QueueAttributeNameRedrivePolicy})
	if err != nil {
		return err
	}

	redrivePolicy, ok := attrs[string(types.QueueAttributeNameRedrivePolicy)]
	if !ok || redrivePolicy == "" {
		return fmt.Errorf("queue %s does not have a dead letter queue configured", queueName)
	}

	return nil
}

// AssertQueueTags checks if a queue has the expected tags
func (a *AWSAsserter) AssertQueueTags(queueName string, expectedTags map[string]string) error {
	client, err := a.createSQSClient()
	if err != nil {
		return err
	}

	queueUrl, err := a.getQueueUrl(queueName)
	if err != nil {
		return err
	}

	result, err := client.ListQueueTags(context.TODO(), &sqs.ListQueueTagsInput{
		QueueUrl: aws.String(queueUrl),
	})
	if err != nil {
		return fmt.Errorf("error getting tags for queue %s: %w", queueName, err)
	}

	for key, expectedValue := range expectedTags {
		actualValue, ok := result.Tags[key]
		if !ok {
			return fmt.Errorf("queue %s is missing tag %s", queueName, key)
		}
		if actualValue != expectedValue {
			return fmt.Errorf("queue %s tag %s has value %s, expected %s", queueName, key, actualValue, expectedValue)
		}
	}

	return nil
}

// AssertQueueEncryption checks if a queue has encryption enabled or disabled
func (a *AWSAsserter) AssertQueueEncryption(queueName string, expectEncrypted bool) error {
	attrs, err := a.getQueueAttributes(queueName, []types.QueueAttributeName{
		types.QueueAttributeNameKmsMasterKeyId,
		types.QueueAttributeNameSqsManagedSseEnabled,
	})
	if err != nil {
		return err
	}

	kmsKeyId := attrs[string(types.QueueAttributeNameKmsMasterKeyId)]
	sqsManagedSse := attrs[string(types.QueueAttributeNameSqsManagedSseEnabled)]

	isEncrypted := kmsKeyId != "" || sqsManagedSse == "true"

	if expectEncrypted && !isEncrypted {
		return fmt.Errorf("queue %s is not encrypted", queueName)
	}
	if !expectEncrypted && isEncrypted {
		return fmt.Errorf("queue %s is encrypted but expected unencrypted", queueName)
	}

	return nil
}

// Helper method to create an SQS client
func (a *AWSAsserter) createSQSClient() (*sqs.Client, error) {
	cfg, err := awshelpers.NewAuthenticatedSessionWithDefaultRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	opts := make([]func(*sqs.Options), 0)

	if endpoint, ok := awshelpers.GetVirtualCloudEndpoint("sqs"); ok {
		opts = append(opts, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return sqs.NewFromConfig(*cfg, opts...), nil
}

// Helper method to get queue URL
func (a *AWSAsserter) getQueueUrl(queueName string) (string, error) {
	client, err := a.createSQSClient()
	if err != nil {
		return "", err
	}

	result, err := client.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("queue %s does not exist: %w", queueName, err)
	}

	return *result.QueueUrl, nil
}

// Helper method to get queue attributes
func (a *AWSAsserter) getQueueAttributes(queueName string, attributeNames []types.QueueAttributeName) (map[string]string, error) {
	client, err := a.createSQSClient()
	if err != nil {
		return nil, err
	}

	queueUrl, err := a.getQueueUrl(queueName)
	if err != nil {
		return nil, err
	}

	result, err := client.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueUrl),
		AttributeNames: attributeNames,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting attributes for queue %s: %w", queueName, err)
	}

	return result.Attributes, nil
}
