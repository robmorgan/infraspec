Feature: RDS Instance Creation
  As a DevOps engineer
  I want to create an RDS instance with specific settings
  So that I can ensure it meets our application requirements

  Scenario: Create a PostgreSQL RDS instance with basic configuration
    Given I have a Terraform configuration in "./fixtures/rds-postgres"
    And I set variable "region" to a random stable AWS region
    And I set variable "name" to "test-postgres-db"
    And I set variable "engine" to "postgres"
    And I set variable "engine_version" to "17.5"
    And I set variable "instance_class" to "db.t4g.micro"
    And I set variable "allocated_storage" to "20"
    And I set variable "multi_az" to "false"
    And I set variable "storage_encrypted" to "true"
    And I set variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the output "db_instance_arn" should contain "test-postgres-db"
    And the RDS instance "test-postgres-db" should exist
    And the RDS instance "test-postgres-db" should have instance class "db.t4g.micro"
    And the RDS instance "test-postgres-db" should have engine "postgres"
    And the RDS instance "test-postgres-db" should have allocated storage 20
    And the RDS instance "test-postgres-db" should have MultiAZ "false"
    And the RDS instance "test-postgres-db" should have encryption "true"
    And the RDS instance "test-postgres-db" should have tags
      | Key         | Value            |
      | Name        | test-postgres-db |
      | Environment | test             |
      | Project     | infratest        |

  Scenario: Create a MySQL RDS instance with high availability
    Given I have a Terraform configuration in "./fixtures/rds-mysql"
    And I set variable "region" to a random stable AWS region
    And I set variable "name" to "test-mysql-db"
    And I set variable "engine" to "mysql"
    And I set variable "engine_version" to "8.4.5"
    And I set variable "instance_class" to "db.t4g.micro"
    And I set variable "allocated_storage" to "30"
    And I set variable "multi_az" to "true"
    And I set variable "storage_encrypted" to "true"
    And I set variable "tags" to
      | Key         | Value     |
      | Environment | test      |
      | Project     | infratest |
    When I run Terraform apply
    Then the output "db_instance_arn" should contain "test-mysql-db"
    And the RDS instance "test-mysql-db" should exist
    And the RDS instance "test-mysql-db" should have instance class "db.t4g.micro"
    And the RDS instance "test-mysql-db" should have engine "mysql"
    And the RDS instance "test-mysql-db" should have allocated storage 30
    And the RDS instance "test-mysql-db" should have MultiAZ "true"
    And the RDS instance "test-mysql-db" should have encryption "true"
    And the RDS instance "test-mysql-db" should have tags
      | Key         | Value         |
      | Name        | test-mysql-db |
      | Environment | test          |
      | Project     | infratest     |
