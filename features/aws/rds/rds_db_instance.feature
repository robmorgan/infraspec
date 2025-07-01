Feature: RDS Instance Creation
  As a DevOps engineer
  I want to create an RDS instance with specific settings
  So that I can ensure it meets our application requirements

  Background:
    Given I have the necessary IAM permissions to describe RDS instances

  Scenario: Create a PostgreSQL RDS instance with basic configuration
    Given I have a Terraform configuration in "../../../examples/aws/rds/postgres"
    And I set the variable "region" to a random stable AWS region
    And I set the variable "name" to "test-postgres-db"
    And I set the variable "engine" to "postgres"
    And I set the variable "engine_version" to "17.5"
    And I set the variable "instance_class" to "db.t4g.micro"
    And I set the variable "allocated_storage" to "20"
    And I set the variable "multi_az" to "false"
    And I set the variable "storage_encrypted" to "true"
    And I set the variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the output "db_instance_arn" should contain "test-postgres-db"
    And the RDS instance "test-postgres-db" should exist
    And the RDS instance "test-postgres-db" instance class should be "db.t4g.micro"
    And the RDS instance "test-postgres-db" engine should be "postgres"
    And the RDS instance "test-postgres-db" allocated storage should be 20
    And the RDS instance "test-postgres-db" MultiAZ should be "false"
    And the RDS instance "test-postgres-db" encryption should be "true"
    And the RDS instance "test-postgres-db" status should be "available"
    And the RDS instance "test-postgres-db" should not be publicly accessible
    And the RDS instance "test-postgres-db" should have the tags
      | Key         | Value            |
      | Name        | test-postgres-db |
      | Environment | test             |
      | Project     | infratest        |

  Scenario: Create a MySQL RDS instance with high availability
    Given I have a Terraform configuration in "../../../examples/aws/rds/postgres"
    And I set the variable "region" to a random stable AWS region
    And I set the variable "name" to "test-mysql-db"
    And I set the variable "engine" to "mysql"
    And I set the variable "engine_version" to "8.4.5"
    And I set the variable "instance_class" to "db.t4g.micro"
    And I set the variable "allocated_storage" to "30"
    And I set the variable "multi_az" to "true"
    And I set the variable "storage_encrypted" to "true"
    And I set the variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the output "db_instance_arn" should contain "test-mysql-db"
    And the RDS instance "test-mysql-db" should exist
    And the RDS instance "test-mysql-db" instance class should be "db.t4g.micro"
    And the RDS instance "test-mysql-db" engine should be "mysql"
    And the RDS instance "test-mysql-db" allocated storage should be 30
    And the RDS instance "test-mysql-db" MultiAZ should be "true"
    And the RDS instance "test-mysql-db" encryption should be "true"
    And the RDS instance "test-mysql-db" status should be "available"
    And the RDS instance "test-mysql-db" should not be publicly accessible
    And the RDS instance "test-mysql-db" should have the tags
      | Key         | Value         |
      | Name        | test-mysql-db |
      | Environment | test          |
      | Project     | infratest     |
