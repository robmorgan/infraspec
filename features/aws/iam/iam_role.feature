Feature: IAM Role Creation
  As a DevOps Engineer
  I want to create an IAM role with proper configuration
  So that I can grant permissions to AWS services securely

  Background:
    Given I have the necessary IAM permissions to describe IAM roles

  Scenario: Create an IAM role for EC2 with an attached policy
    Given I have a Terraform configuration in "../../../examples/aws/iam/iam-role"
    And I set variable "role_name" to "test-ec2-role" with a random suffix
    And I set the variable "max_session_duration" to "7200"
    And I set the variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the IAM role from output "role_name" should exist
    And the IAM role from output "role_name" path should be "/"
    And the IAM role from output "role_name" max session duration should be 7200
    And the IAM role from output "role_name" should have the tags
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    And the IAM policy from output "policy_arn" should exist
    And the IAM policy from output "policy_arn" should be attached to role from output "role_name"
    And the IAM instance profile from output "instance_profile_name" should exist
    And the IAM instance profile from output "instance_profile_name" should have role from output "role_name"
