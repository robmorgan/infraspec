Feature: API Testing Example
  As a DevOps Engineer
  I want to test my API endpoints
  So that I can ensure they are working correctly after deployment

  Scenario: Test authentication endpoint
    Given I have a HTTP endpoint at "http://localhost:8000/basic-auth/testuser/testpass"
    And I set the headers to
      | Name         | Value            |
      | Content-Type | application/json |
    And I set basic auth credentials with username "testuser" and password "testpass"
    When I send a GET request
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "authenticated"

  Scenario: Test protected endpoint with authentication
    Given I have a HTTP endpoint at "http://localhost:8000/bearer"
    And I set the headers to
      | Name          | Value             |
      | Authorization | Bearer test-token |
    When I send a GET request
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "token"

  Scenario: Test file upload endpoint
    Given I have a HTTP endpoint at "http://localhost:8000/post"
    And I have a file "../../examples/http/test-file.txt" as field "file"
    And I set content type to "multipart/form-data"
    And I set the form data to:
      | Name     | Value     |
      | category | document  |
      | userId   | test-user |
    When I send a POST request
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "191152a9-0bd6-4db0-999d-12787295f1ec"

  Scenario: Test API error handling
    Given I have a HTTP endpoint at "http://localhost:8000/status/404"
    When I send a GET request
    Then the HTTP response status should be 404