Feature: VPC Creation
  As a DevOps engineer
  I want to create a VPC with associated resources
  So that I can ensure my network infrastructure meets requirements

  Scenario: Create a VPC with subnet, internet gateway, and security group
    Given I have a Terraform configuration in "../../../examples/aws/ec2/vpc"
    And I set the variable "region" to "us-east-1"
    And I set the variable "name" to "test-vpc" with a random suffix
    And I set the variable "cidr_block" to "10.0.0.0/16"
    And I set the variable "subnet_cidr_block" to "10.0.1.0/24"
    And I set the variable "availability_zone" to "us-east-1a"
    And I set the variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the VPC from output "vpc_id" should exist
    And the VPC from output "vpc_id" state should be "available"
    And the VPC from output "vpc_id" CIDR block should be "10.0.0.0/16"
    And the VPC from output "vpc_id" should have the tags
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    And the subnet from output "subnet_id" should exist
    And the subnet from output "subnet_id" state should be "available"
    And the subnet from output "subnet_id" CIDR block should be "10.0.1.0/24"
    And the subnet from output "subnet_id" availability zone should be "us-east-1a"
    And the internet gateway from output "internet_gateway_id" should exist
    And the security group from output "security_group_id" should exist

  Scenario: Create a VPC with custom CIDR block
    Given I have a Terraform configuration in "../../../examples/aws/ec2/vpc"
    And I set the variable "region" to "us-east-1"
    And I set the variable "name" to "custom-vpc" with a random suffix
    And I set the variable "cidr_block" to "172.16.0.0/16"
    And I set the variable "subnet_cidr_block" to "172.16.1.0/24"
    And I set the variable "availability_zone" to "us-east-1b"
    When I run Terraform apply
    Then the VPC from output "vpc_id" should exist
    And the VPC from output "vpc_id" CIDR block should be "172.16.0.0/16"
    And the subnet from output "subnet_id" CIDR block should be "172.16.1.0/24"
    And the subnet from output "subnet_id" availability zone should be "us-east-1b"
