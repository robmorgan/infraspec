Feature: Lambda Versions and Aliases
  As a DevOps Engineer
  I want to create Lambda functions with versions and aliases
  So that I can manage deployments safely

  Scenario: Create a Lambda function with an alias
    Given I have a Terraform configuration in "../../../examples/aws/lambda/function-with-alias"
    And I set the variable "region" to "us-east-1"
    And I set variable "function_name" to "test-lambda-alias" with a random suffix
    And I set the variable "alias_name" to "live"
    When I run Terraform apply
    Then the Lambda function from output "function_name" should exist
    And the Lambda function from output "function_name" alias "live" should exist

  Scenario: Create a Lambda function with a custom alias name
    Given I have a Terraform configuration in "../../../examples/aws/lambda/function-with-alias"
    And I set the variable "region" to "us-east-1"
    And I set variable "function_name" to "test-lambda-prod" with a random suffix
    And I set the variable "alias_name" to "production"
    When I run Terraform apply
    Then the Lambda function from output "function_name" should exist
    And the Lambda function from output "function_name" alias "production" should exist
