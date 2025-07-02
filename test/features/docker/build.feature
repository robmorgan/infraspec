Feature: Docker image build
    As a developer
    I want to build Docker images
    So that I can containerize my applications

    Scenario: Building a basic Docker image
        Given I have a Dockerfile in in "./fixtures/node-yarn-alpine"
        When I run the docker build command
        Then the build should complete successfully
        And a new Docker image should be created

    Scenario: Building a Docker image with tags
        Given I have a Dockerfile in in "./fixtures/basic-node-app"
        When I run the docker build command with tag "myapp:latest"
        Then the build should complete successfully
        And the image should have the tag "myapp:latest"

    Scenario: Building a Docker image with build arguments
        Given I have a Dockerfile with build arguments
        And I set build argument "VERSION" to "1.0.0"
        When I run the docker build command with the build arguments
        Then the build should complete successfully
        And the image should be built with the specified arguments

    Scenario: Building a Docker image with custom context
        Given I have a Dockerfile in "./docker" directory
        When I run the docker build command with context "./docker"
        Then the build should complete successfully
        And a new Docker image should be created

    Scenario: Failed Docker build with invalid Dockerfile
        Given I have an invalid Dockerfile
        When I run the docker build command
        Then the build should fail
        And an error message should be displayed
