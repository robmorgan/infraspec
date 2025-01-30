Feature: DynamoDB Table Creation
    As a DevOps engineer
    I want to create a DynamoDB table with specific settings
    So that I can ensure it meets our application requirements

    Scenario: Create DynamoDB table with basic configuration
        Given I have a Terraform configuration in "./fixtures/terraform/dynamodb-with-autoscaling"
        And I set variable "table_name" to "test-xyzg23"
        And I set variable "billing_mode" to "PAY_PER_REQUEST"
        When I run Terraform apply
        Then the output "table_arn" should contain "test-xyzg23"
        And the AWS resource "aws_dynamodb_table.main" should exist
        And the DynamoDB table "test-xyzg23" should have billing mode "PAY_PER_REQUEST"
        And the DynamoDB table "test-xyzg23" should have read capacity 5
        And the DynamoDB table "test-xyzg23" should have write capacity 5
        And the DynamoDB table "test-xyzg23" should have tags
            | Key         | Value     |
            | Environment | test      |
            | Project     | infratest |

# Scenario: Create DynamoDB table with autoscaling
#     Given I have a Terraform configuration in "./terraform/dynamodb"
#     And I generate a random resource name with prefix "orders-"
#     And I set variable "table_name" to "${resource_name}"
#     And I set variable "billing_mode" to "PROVISIONED"
#     And I set variable "read_capacity" to "5"
#     And I set variable "write_capacity" to "5"
#     When I run Terraform apply
#     Then the DynamoDB table "${resource_name}" should have billing mode "PROVISIONED"
#     And the DynamoDB table "${resource_name}" should have read capacity 5
#     And the DynamoDB table "${resource_name}" should have write capacity 5
#     And the DynamoDB table "${resource_name}" should have point in time recovery enabled
