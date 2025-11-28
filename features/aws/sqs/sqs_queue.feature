Feature: SQS Queue Creation
  As a DevOps engineer
  I want to create an SQS queue with specific settings
  So that I can ensure it meets our application requirements

  Background:
    Given I have the necessary IAM permissions to describe SQS queues

  Scenario: Create SQS queue with basic configuration
    Given I have a Terraform configuration in "../../../examples/aws/sqs/sqs-queue"
    And I set the variable "region" to a random stable AWS region
    And I set variable "queue_name" to "test-queue" with a random suffix
    And I set the variable "visibility_timeout_seconds" to "60"
    And I set the variable "delay_seconds" to "5"
    And I set the variable "message_retention_seconds" to "86400"
    And I set the variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the SQS queue from output "queue_name" should exist
    And the SQS queue from output "queue_name" should have visibility timeout 60
    And the SQS queue from output "queue_name" should have delay seconds 5
    And the SQS queue from output "queue_name" should have message retention period 86400
    And the SQS queue from output "queue_name" should be encrypted
    And the SQS queue from output "queue_name" should have tags
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
