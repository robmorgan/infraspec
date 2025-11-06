Feature: Bearer Token Authentication Example
    As a DevOps Engineer
    I want to test API endpoints that require Bearer token authentication
    So that I can ensure my authentication is working correctly

    Scenario: Test API with Bearer token authentication
        Given I have a HTTP endpoint at "https://api.example.com/protected"
        And I am authenticated with a valid bearer token
        When I send a GET request
        Then the HTTP response status should be 200
        And the HTTP response should be valid JSON
        And the HTTP response should contain "authenticated"

    Scenario: Test API with Bearer token and additional headers
        Given I have a HTTP endpoint at "https://api.example.com/protected"
        And I am authenticated with a valid bearer token
        And I set the headers to
            | Name          | Value            |
            | Content-Type  | application/json |
            | X-API-Version | v1               |
        When I send a POST request
        Then the HTTP response status should be 200
        And the HTTP response should be valid JSON

    Scenario: Test API without Bearer token (should fail)
        Given I have a HTTP endpoint at "https://api.example.com/protected"
        When I send a GET request
        Then the HTTP response status should be 401