Feature: API Testing Example
  As a DevOps Engineer
  I want to test my API endpoints
  So that I can ensure they are working correctly after deployment

  Scenario: Test health endpoint
    Given I have a HTTP endpoint at "https://api.example.com/health"
    When I send a GET request
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "healthy"

  Scenario: Test authentication endpoint
    Given I have a HTTP endpoint at "https://api.example.com/auth/login"
    And I set the headers to
      | Name         | Value            |
      | Content-Type | application/json |
    When I send a POST request
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "token"

  Scenario: Test protected endpoint with authentication
    Given I have a HTTP endpoint at "https://api.example.com/protected"
    And I set the headers to
      | Name          | Value             |
      | Authorization | Bearer test-token |
    When I send a GET request
    Then the HTTP response status should be 200

  Scenario: Test file upload endpoint
    Given I have a HTTP endpoint at "https://api.example.com/auth/upload"
    And I have a file "../../../examples/http/test-file.txt" as field "file"
    And I set content type to "multipart/form-data"
    And I set the form data to:
      | Name     | Value     |
      | category | document  |
      | userId   | test-user |
    When I send a POST request
    Then the HTTP response status should be 201
    And the HTTP response should be valid JSON
    And the HTTP response should contain "uploaded"

  Scenario: Test API error handling
    Given I have a HTTP endpoint at "https://api.example.com/nonexistent"
    When I send a GET request
    Then the HTTP response status should be 404
    And the HTTP response should contain "not found"

  Scenario: Test API response headers
    Given I have a HTTP endpoint at "https://api.example.com/cors-test"
    When I send a GET request
    Then the HTTP response status should be 200
    And the HTTP response header "Access-Control-Allow-Origin" should be "*"
    And the HTTP response header "Content-Type" should be "application/json"
