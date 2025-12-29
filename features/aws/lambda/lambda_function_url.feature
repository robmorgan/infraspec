Feature: Lambda Function URLs
  As a DevOps Engineer
  I want to create Lambda functions with public URLs
  So that I can expose functions as HTTP endpoints

  Scenario: Create a Lambda function with a public URL
    Given I have a Terraform configuration in "../../../examples/aws/lambda/function-url"
    And I set the variable "region" to "us-east-1"
    And I set variable "function_name" to "test-lambda-url" with a random suffix
    And I set the variable "authorization_type" to "NONE"
    When I run Terraform apply
    Then the Lambda function from output "function_name" should exist
    And the Lambda function from output "function_name" should have a function URL
    And the Lambda function from output "function_name" function URL auth type should be "NONE"

  Scenario: Create a Lambda function with IAM-authenticated URL
    Given I have a Terraform configuration in "../../../examples/aws/lambda/function-url"
    And I set the variable "region" to "us-east-1"
    And I set variable "function_name" to "test-lambda-url-iam" with a random suffix
    And I set the variable "authorization_type" to "AWS_IAM"
    When I run Terraform apply
    Then the Lambda function from output "function_name" should exist
    And the Lambda function from output "function_name" should have a function URL
    And the Lambda function from output "function_name" function URL auth type should be "AWS_IAM"
