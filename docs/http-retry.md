# HTTP Retry Functionality

InfraSpec now supports retrying HTTP requests until the response contains expected content. This is useful for handling eventual consistency, temporary failures, or waiting for services to become available.

## Usage

### Basic Retry with Default Values

```gherkin
Given I have a HTTP endpoint at "http://localhost:8000/status"
When I retry the HTTP request until the response contains "ready"
Then the HTTP response status should be 200
```

This uses default values:
- **Max retries**: 5
- **Timeout**: 30 seconds
- **Sleep between retries**: 1 second

### Custom Retry Configuration

```gherkin
Given I have a HTTP endpoint at "http://localhost:8000/api/data"
When I retry the HTTP request until the response contains "processed" with max 10 retries and a 60 second timeout
Then the HTTP response should contain "data_id"
```

### With Authentication

```gherkin
Given I have a HTTP endpoint at "http://localhost:8000/health"
And I set basic auth credentials with username "admin" and password "secret"
When I retry the HTTP request until the response contains "healthy" with max 5 retries and a 30 second timeout
Then the HTTP response status should be 200
```

### With Bearer Token

```gherkin
Given I have a HTTP endpoint at "http://localhost:8000/api/data"
And I am authenticated with a valid bearer token
When I retry the HTTP request until the response contains "processed"
Then the HTTP response should contain "data_id"
```

### With POST Request and Body

```gherkin
Given I have a HTTP endpoint at "http://localhost:8000/jobs"
And I set the headers to
  | Name         | Value            |
  | Content-Type | application/json |
And I set the request body to "{\"job_type\": \"data_processing\"}"
When I retry the HTTP request until the response contains "completed" with max 20 retries and a 120 second timeout
Then the HTTP response should be valid JSON
```

## How It Works

1. **Setup**: Configure your HTTP endpoint and any authentication/headers as usual
2. **Retry Step**: Use the retry step to specify what content you're waiting for
3. **Retry Logic**: The system will:
   - Make the HTTP request
   - Check if the response contains the expected string
   - If not found, wait 1 second and retry
   - Continue until either:
     - The expected content is found (success)
     - Max retries are exceeded (failure)
     - Timeout is reached (failure)
4. **Assertions**: Use standard HTTP assertions on the final successful response

## Error Handling

The retry step will fail with a descriptive error message if:
- The maximum number of retries is exceeded
- The timeout is reached
- The HTTP endpoint is not configured
- Any HTTP request fails

## Use Cases

- **Service Health Checks**: Wait for a service to become healthy
- **Job Completion**: Wait for background jobs to complete
- **Data Processing**: Wait for data to be processed and available
- **Deployment Verification**: Wait for new deployments to be ready
- **Eventual Consistency**: Handle systems that may take time to propagate changes 