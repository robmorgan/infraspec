Feature: Lambda Function Creation
  As a DevOps Engineer
  I want to create Lambda functions with proper configuration
  So that I can run serverless workloads

  Scenario: Create a basic Lambda function
    Given I have a Terraform configuration in "../../../examples/aws/lambda/basic-function"
    And I set the variable "region" to "us-east-1"
    And I set variable "function_name" to "test-lambda" with a random suffix
    And I set the variable "runtime" to "python3.12"
    And I set the variable "handler" to "index.handler"
    And I set the variable "timeout" to "30"
    And I set the variable "memory_size" to "128"
    When I run Terraform apply
    Then the Lambda function from output "function_name" should exist
    And the Lambda function from output "function_name" runtime should be "python3.12"
    And the Lambda function from output "function_name" handler should be "index.handler"
    And the Lambda function from output "function_name" timeout should be 30 seconds
    And the Lambda function from output "function_name" memory should be 128 MB

  Scenario: Create a Lambda function with environment variables
    Given I have a Terraform configuration in "../../../examples/aws/lambda/basic-function"
    And I set the variable "region" to "us-east-1"
    And I set variable "function_name" to "test-lambda-env" with a random suffix
    And I set the variable "environment_variables" to
      | Key         | Value       |
      | ENVIRONMENT | production  |
      | LOG_LEVEL   | debug       |
    When I run Terraform apply
    Then the Lambda function from output "function_name" should exist
    And the Lambda function from output "function_name" should have environment variable "ENVIRONMENT" with value "production"
