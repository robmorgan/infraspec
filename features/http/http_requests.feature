Feature: HTTP Requests
  As a DevOps Engineer
  I want to test HTTP endpoints
  So that I can validate API functionality and infrastructure

  Scenario: Test basic GET request
    When I make a GET request to "https://httpbin.org/get"
    Then the HTTP response status should be 200

  Scenario: Test GET request with JSON response
    When I make a GET request to "https://httpbin.org/json"
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON

  Scenario: Test POST request with status assertion
    Given the POST request to "https://httpbin.org/post" should return status 200

  Scenario: Test response content validation
    When the GET response from "https://httpbin.org/user-agent" should contain "User-Agent"

  Scenario: Test custom headers
    When I make a GET request to "https://httpbin.org/headers" with headers:
      | Name          | Value             |
      | Authorization | Bearer test-token |
      | Content-Type  | application/json  |
    Then the HTTP response status should be 200

  Scenario: Test POST with body and headers
    When I make a POST request to "https://httpbin.org/post" with body "test data" and headers:
      | Name         | Value            |
      | Content-Type | application/json |
    Then the HTTP response status should be 200
    And the HTTP response should contain "test data"

  Scenario: Test response header validation
    When I make a GET request to "https://httpbin.org/response-headers?Content-Type=application/json"
    Then the HTTP response status should be 200
    And the HTTP response header "Content-Type" should be "application/json"

  Scenario: Test file upload
    Given I have a test file "test-file.txt" with content "Hello, World!"
    When I upload file "test-file.txt" to "https://httpbin.org/post" as field "file"
    Then the HTTP response status should be 200

  Scenario: Test file upload with form data
    Given I have a test file "test-file.txt" with content "Hello, World!"
    When I upload file "test-file.txt" to "https://httpbin.org/post" as field "file" with form data:
      | Name | Value                           |
      | uuid | 191152a9-0bd6-4db0-999d-12787295f1ec |
      | type | document                        |
    Then the HTTP response status should be 200

  Scenario: Test API endpoint existence
    Then the http_endpoint "https://httpbin.org/get" should exist

  Scenario: Test multiple response validations
    When I make a GET request to "https://httpbin.org/json"
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "slideshow"