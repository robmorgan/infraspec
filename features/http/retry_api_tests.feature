Feature: HTTP Retry Testing
  As a DevOps Engineer
  I want to retry HTTP requests until they contain expected content
  So that I can handle eventual consistency and temporary failures

  Scenario: Basic HTTP request test
    Given I have a HTTP endpoint at "http://localhost:8000/json"
    And I want to retry the HTTP request until the response contains "completed" with max 5 retries and 10 second timeout
    When I retry the HTTP request until the response contains "completed" with max 5 retries and 10 second timeout
    Then the HTTP response status should be 200
    And the HTTP response should contain "Hello, World!"
    And the HTTP response should be valid JSON

  Scenario: HTTP request with headers test
    Given I have a HTTP endpoint at "http://localhost:8000/headers"
    And I set the headers to
      | Name         | Value            |
      | Content-Type | application/json |
    When I retry the HTTP request until the response contains "headers"
    Then the HTTP response status should be 200

  Scenario: HTTP request with basic auth test
    Given I have a HTTP endpoint at "http://localhost:8000/json"
    And I set basic auth credentials with username "testuser" and password "testpass"
    When I send a GET request
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON

  Scenario: HTTP request with bearer token test
    Given I have a HTTP endpoint at "http://localhost:8000/bearer"
    And I am authenticated with a valid bearer token
    When I send a GET request
    Then the HTTP response status should be 200
    And the HTTP response should contain "authenticated"
