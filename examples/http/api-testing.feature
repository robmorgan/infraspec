Feature: API Testing Example
  As a DevOps Engineer
  I want to test my API endpoints
  So that I can ensure they are working correctly after deployment

  Background:
    Given I have deployed my API to a test environment

  Scenario: Test health endpoint
    When the GET request to "https://api.example.com/health" should return status 200
    And the GET response from "https://api.example.com/health" should be valid JSON
    And the GET response from "https://api.example.com/health" should contain "healthy"

  Scenario: Test authentication endpoint
    When I make a POST request to "https://api.example.com/auth/login" with body "{"username":"test","password":"test"}" and headers:
      | Name         | Value            |
      | Content-Type | application/json |
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "token"

  Scenario: Test protected endpoint with authentication
    When I make a GET request to "https://api.example.com/protected" with headers:
      | Name          | Value             |
      | Authorization | Bearer test-token |
    Then the HTTP response status should be 200

  Scenario: Test file upload endpoint
    Given I have a test document "document.pdf"
    When I upload file "document.pdf" to "https://api.example.com/upload" as field "document" with form data:
      | Name     | Value      |
      | category | document   |
      | userId   | test-user  |
    Then the HTTP response status should be 201
    And the HTTP response should be valid JSON
    And the HTTP response should contain "uploaded"

  Scenario: Test API error handling
    When the GET request to "https://api.example.com/nonexistent" should return status 404
    And the GET response from "https://api.example.com/nonexistent" should contain "not found"

  Scenario: Test API response headers
    When I make a GET request to "https://api.example.com/cors-test"
    Then the HTTP response status should be 200
    And the HTTP response header "Access-Control-Allow-Origin" should be "*"
    And the HTTP response header "Content-Type" should be "application/json"

  Scenario: Test API performance (status check)
    # This could be extended with timing assertions in future versions
    When the GET request to "https://api.example.com/fast-endpoint" should return status 200