Feature: DynamoDB Table Creation
    As a DevOps engineer
    I want to create a DynamoDB table with specific settings
    So that I can ensure it meets our application requirements

    Scenario: Create DynamoDB table with basic configuration
        Given I have a Terraform configuration in "./fixtures/terraform/dynamodb-with-autoscaling"
        And I set variable "name" to "test-xyzg23"
        And I set variable "hash_key" to "id"
        And I set variable "billing_mode" to "PROVISIONED"
        And I set variable "tags" to
            | Key         | Value     |
            | Environment | test      |
            | Project     | infratest |
        When I run Terraform apply
        Then the output "table_arn" should contain "test-xyzg23"
        And the AWS resource "aws_dynamodb_table.main" should exist
        And the DynamoDB table "test-xyzg23" should have billing mode "PAY_PER_REQUEST"
        And the DynamoDB table "test-xyzg23" should have read capacity 5
        And the DynamoDB table "test-xyzg23" should have write capacity 5
        And the DynamoDB table "test-xyzg23" should have tags
            | Key         | Value       |
            | Name        | test-xyzg23 |
            | Environment | test        |
            | Project     | infratest   |
