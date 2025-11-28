Feature: EC2 Instance Creation
  As a DevOps engineer
  I want to create EC2 instances with specific configurations
  So that I can ensure my compute infrastructure meets requirements

  Scenario: Create an EC2 instance with basic configuration
    Given I have a Terraform configuration in "../../../examples/aws/ec2/instance"
    And I set the variable "region" to "us-east-1"
    And I set the variable "name" to "test-instance" with a random suffix
    And I set the variable "instance_type" to "t3.micro"
    And I set the variable "ami_id" to "ami-12345678"
    And I set the variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the EC2 instance from output "instance_id" should exist
    And the EC2 instance from output "instance_id" state should be "running"
    And the EC2 instance from output "instance_id" instance type should be "t3.micro"
    And the EC2 instance from output "instance_id" AMI should be "ami-12345678"
    And the EC2 instance from output "instance_id" should have the tags
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |

  Scenario: Verify VPC and network resources are created with instance
    Given I have a Terraform configuration in "../../../examples/aws/ec2/instance"
    And I set the variable "region" to "us-east-1"
    And I set the variable "name" to "test-network" with a random suffix
    And I set the variable "instance_type" to "t3.small"
    And I set the variable "ami_id" to "ami-87654321"
    When I run Terraform apply
    Then the VPC from output "vpc_id" should exist
    And the subnet from output "subnet_id" should exist
    And the security group from output "security_group_id" should exist
    And the EC2 instance from output "instance_id" should exist
