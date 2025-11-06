# HTTP Bearer Token Authentication

InfraSpec now supports Bearer token authentication for HTTP requests. This allows you to test API endpoints that require Bearer token authentication.

## Usage

### Setting Bearer Token

Use the `I am authenticated with a valid bearer token` step to configure Bearer token authentication. This step requires the `INFRASPEC_BEARER_TOKEN` environment variable to be set:

```gherkin
Given I have a HTTP endpoint at "https://api.example.com/protected"
And I am authenticated with a valid bearer token
When I send a GET request
Then the HTTP response status should be 200
```

### Environment Variable Setup

Before running scenarios that use Bearer token authentication, set the `INFRASPEC_BEARER_TOKEN` environment variable:

```bash
export INFRASPEC_BEARER_TOKEN="your-secret-token-here"
```

### Complete Example

```gherkin
Feature: API with Bearer Token Authentication

Scenario: Test protected API endpoint
  Given I have a HTTP endpoint at "https://api.example.com/protected"
  And I am authenticated with a valid bearer token
  And I set the headers to
    | Name         | Value            |
    | Content-Type | application/json |
  When I send a GET request
  Then the HTTP response status should be 200
  And the HTTP response should be valid JSON
  And the HTTP response should contain "authenticated"
```

## How It Works

When you use the `I am authenticated with a valid bearer token` step, InfraSpec will:

1. Check for the `INFRASPEC_BEARER_TOKEN` environment variable
2. If the environment variable is not set, the scenario will fail with a clear error message
3. If the environment variable is set, it will automatically add the `Authorization: Bearer <token>` header to all subsequent HTTP requests for that scenario

The Bearer token is stored in the scenario context and applied to all HTTP requests until the scenario ends or a new Bearer token is set.

## Compatibility

- Bearer token authentication works with all HTTP methods (GET, POST, PUT, DELETE, etc.)
- Can be used alongside other HTTP features like custom headers, form data, file uploads, etc.
- Compatible with basic authentication (though typically you'd use one or the other, not both)

## Error Handling

### Missing Environment Variable

If the `INFRASPEC_BEARER_TOKEN` environment variable is not set when using the `I am authenticated with a valid bearer token` step, the scenario will fail with a clear error message:

```
BEARER_TOKEN environment variable is not set. Please set it before running this scenario
```

### Missing Authentication

If an API endpoint requires Bearer token authentication but none is provided, the server should return a 401 Unauthorized status code, which you can test for:

```gherkin
Given I have a HTTP endpoint at "https://api.example.com/protected"
When I send a GET request
Then the HTTP response status should be 401
```

## Security Best Practices

- Never commit Bearer tokens to version control
- Use environment variables or secure secret management systems
- Consider using different tokens for different environments (dev, staging, prod)
- Rotate tokens regularly 