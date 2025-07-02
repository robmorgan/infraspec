Feature: S3 Bucket Creation
  As a DevOps Engineer
  I want to create an S3 bucket with guardrails
  So that I can store my data securely

 Background:
    Given I have the necessary IAM permissions to describe S3 buckets

  Scenario: Create an S3 bucket with a name
    Given I have a Terraform configuration in "../../../examples/aws/s3/s3-bucket"
    And I set the variable "region" to a random stable AWS region
    And I set variable "bucket_name" to "my-bucket"
    And I set the variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the S3 bucket "my-bucket" should exist
    And the S3 bucket "my-bucket" should have a versioning configuration
    And the S3 bucket "my-bucket" should have a public access block
    And the S3 bucket "my-bucket" should have a server access logging configuration
    And the S3 bucket "my-bucket" should have an encryption configuration