Feature: HTTP Requests
  As a DevOps Engineer
  I want to test HTTP endpoints
  So that I can validate API functionality and infrastructure

  Scenario: Test basic GET request
    Given I have a HTTP endpoint at "http://localhost:9000/get"
    When I make a GET request
    Then the HTTP response status should be 200

  Scenario: Test GET request with JSON response
    Given I have a HTTP endpoint at "http://localhost:9000/json"
    When I make a GET request
    Then the HTTP response status should be 200
    And the response should be valid JSON

  Scenario: Test POST request with status assertion
    Given I have a HTTP endpoint at "http://localhost:9000/post"
    When I make a POST request
    Then the HTTP response status should be 200

  Scenario: Test response content validation
    Given I have a HTTP endpoint at "http://localhost:9000/user-agent"
    When I make a GET request
    Then the HTTP response should contain "user-agent"

  Scenario: Test custom headers
    Given I have a HTTP endpoint at "http://localhost:9000/headers"
    And I set the headers to
      | Name          | Value             |
      | Authorization | Bearer test-token |
      | Content-Type  | application/json  |
    When I make a GET request
    Then the HTTP response status should be 200

  Scenario: Test POST with body and headers
    Given I have a HTTP endpoint at "http://localhost:9000/post"
    And I set the headers to
      | Name         | Value            |
      | Content-Type | application/json |
    And I set the request body to "test data"
    When I make a POST request
    Then the HTTP response status should be 200
    And the HTTP response should contain "test data"

  Scenario: Test response header validation
    Given I have a HTTP endpoint at "http://localhost:9000/response-headers?Content-Type=application/json"
    When I make a GET request
    Then the HTTP response status should be 200
    And the HTTP response header "Content-Type" should be "application/json"

  Scenario: Test file upload
    Given I have a HTTP endpoint at "http://localhost:9000/post"
    And I have a file "../../examples/http/test-file.txt" as field "file"
    When I make a POST request
    Then the HTTP response status should be 200

  Scenario: Test file upload with form data
    Given I have a HTTP endpoint at "http://localhost:9000/post"
    And I have a file "../../examples/http/test-file.txt" as field "file"
    And I set content type to "multipart/form-data"
    And I set the form data to:
      | Name | Value                                |
      | uuid | 191152a9-0bd6-4db0-999d-12787295f1ec |
      | type | document                             |
    When I make a POST request
    Then the HTTP response status should be 200

  Scenario: Test multiple response validations
    Given I have a HTTP endpoint at "http://localhost:9000/json"
    When I make a GET request
    Then the HTTP response status should be 200
    And the HTTP response should be valid JSON
    And the HTTP response should contain "slideshow"
