Feature: RDS Database Instance Verification
    As a DevOps engineer
    I want to verify that an RDS database instance exists and is properly configured
    So that I can ensure our infrastructure is deployed correctly

    Background:
        Given I have access to AWS RDS service
        And I have the necessary IAM permissions to describe RDS instances

    Scenario: Verify existing RDS instance exists and is available
        Given I have an RDS instance with identifier "test-postgres-db"
        When I describe the RDS instance
        Then the RDS instance "test-postgres-db" should exist
        And the RDS instance "test-postgres-db" status should be "available"

# Scenario: Verify RDS instance has correct engine configuration
#     Given an RDS instance with identifier "test-postgres-db" exists
#     When I retrieve the instance details
#     Then the database engine should be "postgres"
#     And the engine version should be "8.0.35"
#     And the instance class should be "db.t3.micro"

# Scenario: Verify RDS instance network configuration
#     Given an RDS instance with identifier "test-postgres-db" exists
#     When I check the network configuration
#     Then the instance should have a valid endpoint address
#     And the endpoint port should be 3306
#     And the instance should not be publicly accessible

# Scenario: Verify RDS instance storage configuration
#     Given an RDS instance with identifier "test-postgres-db" exists
#     When I check the storage configuration
#     Then the allocated storage should be at least 20 GB
#     And the storage type should be "gp2"
#     And backup retention should be configured

# Scenario: Handle non-existent RDS instance
#     Given an RDS instance with identifier "non-existent-instance" should not exist
#     When I attempt to describe the RDS instance
#     Then I should receive a "DBInstanceNotFoundFault" error
#     And the error message should indicate the instance was not found

# Scenario Outline: Verify multiple RDS instances exist
#     Given an RDS instance with identifier "<instance_id>" should exist
#     When I describe the RDS instance
#     Then the instance should be found
#     And the instance status should be "<expected_status>"

#     Examples:
#         | instance_id      | expected_status |
#         | test-postgres-db | available       |
#         | dev-1            | available       |

# Scenario: Verify RDS instance tags are properly set
#     Given an RDS instance with identifier "test-postgres-db" exists
#     When I retrieve the instance tags
#     Then the instance should have a tag with key "Environment" and value "test"
#     And the instance should have a tag with key "Project" and value "go-rds-example"
#     And all required tags should be present
