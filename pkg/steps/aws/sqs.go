package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/assertions"
	"github.com/robmorgan/infraspec/pkg/assertions/aws"
	"github.com/robmorgan/infraspec/pkg/iacprovisioner"
)

// SQS Step Definitions
func registerSQSSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I have the necessary IAM permissions to describe SQS queues$`, newVerifyAWSSQSDescribeQueuesStep)
	sc.Step(`^the SQS queue "([^"]*)" should exist$`, newSQSQueueExistsStep)
	sc.Step(`^the SQS queue "([^"]*)" should have visibility timeout (\d+)$`, newSQSQueueVisibilityTimeoutStep)
	sc.Step(`^the SQS queue "([^"]*)" should have delay seconds (\d+)$`, newSQSQueueDelaySecondsStep)
	sc.Step(`^the SQS queue "([^"]*)" should have max message size (\d+)$`, newSQSQueueMaxMessageSizeStep)
	sc.Step(`^the SQS queue "([^"]*)" should have message retention period (\d+)$`, newSQSQueueMessageRetentionPeriodStep)
	sc.Step(`^the SQS queue "([^"]*)" should have receive message wait time (\d+)$`, newSQSQueueReceiveMessageWaitTimeStep)
	sc.Step(`^the SQS queue "([^"]*)" should be a FIFO queue$`, newSQSQueueIsFifoStep)
	sc.Step(`^the SQS queue "([^"]*)" should have a dead letter queue$`, newSQSQueueHasDeadLetterQueueStep)
	sc.Step(`^the SQS queue "([^"]*)" should have tags$`, newSQSQueueTagsStep)
	sc.Step(`^the SQS queue "([^"]*)" should be encrypted$`, newSQSQueueEncryptedStep)
	sc.Step(`^the SQS queue "([^"]*)" should not be encrypted$`, newSQSQueueNotEncryptedStep)

	// Steps that read queue name from Terraform output
	sc.Step(`^the SQS queue from output "([^"]*)" should exist$`, newSQSQueueFromOutputExistsStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should have visibility timeout (\d+)$`, newSQSQueueFromOutputVisibilityTimeoutStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should have delay seconds (\d+)$`, newSQSQueueFromOutputDelaySecondsStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should have max message size (\d+)$`, newSQSQueueFromOutputMaxMessageSizeStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should have message retention period (\d+)$`, newSQSQueueFromOutputMessageRetentionPeriodStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should have receive message wait time (\d+)$`, newSQSQueueFromOutputReceiveMessageWaitTimeStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should be a FIFO queue$`, newSQSQueueFromOutputIsFifoStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should have a dead letter queue$`, newSQSQueueFromOutputHasDeadLetterQueueStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should have tags$`, newSQSQueueFromOutputTagsStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should be encrypted$`, newSQSQueueFromOutputEncryptedStep)
	sc.Step(`^the SQS queue from output "([^"]*)" should not be encrypted$`, newSQSQueueFromOutputNotEncryptedStep)
}

func newVerifyAWSSQSDescribeQueuesStep(ctx context.Context) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertSQSDescribeQueues()
}

func newSQSQueueExistsStep(ctx context.Context, queueName string) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueExists(queueName)
}

func newSQSQueueVisibilityTimeoutStep(ctx context.Context, queueName string, timeout int) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueVisibilityTimeout(queueName, timeout)
}

func newSQSQueueDelaySecondsStep(ctx context.Context, queueName string, delay int) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueDelaySeconds(queueName, delay)
}

func newSQSQueueMaxMessageSizeStep(ctx context.Context, queueName string, size int) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueMaxMessageSize(queueName, size)
}

func newSQSQueueMessageRetentionPeriodStep(ctx context.Context, queueName string, period int) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueMessageRetentionPeriod(queueName, period)
}

func newSQSQueueReceiveMessageWaitTimeStep(ctx context.Context, queueName string, waitTime int) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueReceiveMessageWaitTime(queueName, waitTime)
}

func newSQSQueueIsFifoStep(ctx context.Context, queueName string) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueIsFifo(queueName)
}

func newSQSQueueHasDeadLetterQueueStep(ctx context.Context, queueName string) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueHasDeadLetterQueue(queueName)
}

func newSQSQueueTagsStep(ctx context.Context, queueName string, table *godog.Table) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}

	// Convert Gherkin table to map[string]string
	tags := make(map[string]string)
	for _, row := range table.Rows[1:] { // Skip header row
		tags[row.Cells[0].Value] = row.Cells[1].Value
	}

	return sqsAssert.AssertQueueTags(queueName, tags)
}

func newSQSQueueEncryptedStep(ctx context.Context, queueName string) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueEncryption(queueName, true)
}

func newSQSQueueNotEncryptedStep(ctx context.Context, queueName string) error {
	sqsAssert, err := getSQSAsserter(ctx)
	if err != nil {
		return err
	}
	return sqsAssert.AssertQueueEncryption(queueName, false)
}

func getSQSAsserter(ctx context.Context) (aws.SQSAsserter, error) {
	asserter, err := contexthelpers.GetAsserter(ctx, assertions.AWS)
	if err != nil {
		return nil, err
	}

	sqsAssert, ok := asserter.(aws.SQSAsserter)
	if !ok {
		return nil, fmt.Errorf("asserter does not implement SQSAsserter")
	}
	return sqsAssert, nil
}

// Step functions that read queue name from Terraform output

func newSQSQueueFromOutputExistsStep(ctx context.Context, outputName string) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueExistsStep(ctx, queueName)
}

func newSQSQueueFromOutputVisibilityTimeoutStep(ctx context.Context, outputName string, timeout int) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueVisibilityTimeoutStep(ctx, queueName, timeout)
}

func newSQSQueueFromOutputDelaySecondsStep(ctx context.Context, outputName string, delay int) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueDelaySecondsStep(ctx, queueName, delay)
}

func newSQSQueueFromOutputMaxMessageSizeStep(ctx context.Context, outputName string, size int) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueMaxMessageSizeStep(ctx, queueName, size)
}

func newSQSQueueFromOutputMessageRetentionPeriodStep(ctx context.Context, outputName string, period int) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueMessageRetentionPeriodStep(ctx, queueName, period)
}

func newSQSQueueFromOutputReceiveMessageWaitTimeStep(ctx context.Context, outputName string, waitTime int) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueReceiveMessageWaitTimeStep(ctx, queueName, waitTime)
}

func newSQSQueueFromOutputIsFifoStep(ctx context.Context, outputName string) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueIsFifoStep(ctx, queueName)
}

func newSQSQueueFromOutputHasDeadLetterQueueStep(ctx context.Context, outputName string) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueHasDeadLetterQueueStep(ctx, queueName)
}

func newSQSQueueFromOutputTagsStep(ctx context.Context, outputName string, table *godog.Table) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueTagsStep(ctx, queueName, table)
}

func newSQSQueueFromOutputEncryptedStep(ctx context.Context, outputName string) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueEncryptedStep(ctx, queueName)
}

func newSQSQueueFromOutputNotEncryptedStep(ctx context.Context, outputName string) error {
	queueName, err := getQueueNameFromOutput(ctx, outputName)
	if err != nil {
		return err
	}
	return newSQSQueueNotEncryptedStep(ctx, queueName)
}

// Helper function to get queue name from Terraform output
func getQueueNameFromOutput(ctx context.Context, outputName string) (string, error) {
	options := contexthelpers.GetIacProvisionerOptions(ctx)
	queueName, err := iacprovisioner.Output(options, outputName)
	if err != nil {
		return "", fmt.Errorf("failed to get queue name from output %s: %w", outputName, err)
	}
	return queueName, nil
}

// Helper for strconv.Atoi error handling (not currently used but useful for future)
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
