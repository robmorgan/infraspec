# InfraSpec Development Guide

This document provides guidance for AI coding assistants working on the InfraSpec project.

## Project Overview

**InfraSpec** is a tool for testing your cloud infrastructure written in Go that allows users to write infrastructure
tests in plain English using Gherkin syntax. The project tests infrastructure code for Terraform, Docker, and Kubernetes
without requiring users to write traditional test code using frameworks like Terratest.

### Key Technologies

- **Language**: Go 1.24.4
- **Testing Framework**: Cucumber/Godog for BDD testing
- **Cloud Integration**: AWS SDK v2 (DynamoDB, RDS, S3, EC2, SSM)
- **Website**: Next.js with Nextra documentation theme
- **CLI**: Cobra for command-line interface

### Project Structure

- `cmd/` - CLI commands and main entry point
- `pkg/` - Public packages (assertions, helpers, provisioners)
- `internal/` - Private packages (config, runners, generators)
- `examples/` - Infrastructure as Code examples for testing
- `features/` - Gherkin feature files for testing
- `test/` - Integration tests and test helpers
- `website/` - Next.js documentation website

## Development Setup

### Prerequisites

- Go 1.24.4 or later
- Make (for build automation)
- InfraSpec API (for AWS emulation during testing)

### Getting Started

1. Clone the repository.
2. Fix a bug, improve documentation, or add a new feature.
3. Install dependencies: `make deps`.
4. Run tests: `make go-test-cover`.
5. Format code: `make fmt`.
6. Lint code: `make lint`.

### Build Commands

- `make deps` - Install all dependencies and development tools
- `make tidy` - Run `go mod tidy`
- `make fmt` - Format code using gofumpt, goimports, and gci
- `make lint` - Run golangci-lint
- `make go-test-cover` - Run tests with coverage report

## Coding Standards

### Go Code Style

- **Formatting**: Use `gofumpt` for stricter formatting than `gofmt`
- **Imports**: Organize imports with `gci` in the order: standard, default, project-specific
- **Import Organization**: `goimports` for automatic import management
- **Linting**: All code must pass `golangci-lint` with project configuration

### Code Organization

- Follow Go project layout standards
- Use meaningful package names and avoid generic names like `utils`
- Keep public APIs in `pkg/` and internal logic in `internal/`
- Maintain clear separation between CLI, core logic, and cloud providers

### Testing Patterns

- Use `github.com/stretchr/testify` for assertions
- Write integration tests that work with InfraSpec API
- Follow BDD patterns with Gherkin feature files
- Test both positive and negative scenarios
- Include retry logic for flaky cloud operations

### Error Handling

- Provide clear, actionable error messages
- Log appropriately using `go.uber.org/zap`
- Handle AWS SDK errors gracefully

## Commit Guidelines

This project uses **Conventional Commits** specification. All commit messages must follow this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes

### Examples

```
feat(s3): add bucket encryption validation
fix(dynamodb): handle missing table gracefully
docs: update README with new installation methods
test(rds): add integration tests for MySQL instances
```

### Scopes

Use these scopes when relevant:

- `s3`, `dynamodb`, `rds` - AWS service specific changes
- `terraform` - Terraform-related changes
- `cli` - Command-line interface changes
- `docs` - Documentation changes
- `website` - Website-specific changes

## Architecture Guidelines

## Adding New Assertion Functions

1. Create provider specific assertion functions below `pkg/assertions`. (e.g Create AWS service-specific assertion
   functions in `pkg/assertions/aws`).
2. All assertion function names should begin with `Assert`.

### Adding New AWS Cloud Services

1. Create AWS service-specific assertion functions in `pkg/assertions/aws/`
2. Add corresponding step definitions in `pkg/steps/aws/`
3. Create feature examples in `examples/aws/`
4. Write Gherkin feature files in `features/aws/`
5. Update documentation and roadmap

### Testing Philosophy

- **BDD First**: Write Gherkin scenarios before implementation
- **InfraSpec API Integration**: Use InfraSpec API for AWS service emulation
- **Real-world Examples**: Include practical Terraform configurations
- **Error Scenarios**: Test both success and failure paths

### CLI Design

- Use Cobra for consistent command structure
- Provide helpful error messages and suggestions
- Support both interactive and CI/CD usage
- Include progress indicators for long-running operations

## Dependencies Management

### Go Modules

- Keep dependencies minimal and well-maintained
- Prefer AWS SDK v2 over v1
- Use official libraries when possible
- Regular dependency updates with testing

### Key Dependencies

- `github.com/cucumber/godog` - BDD testing framework
- `github.com/aws/aws-sdk-go-v2` - AWS SDK
- `github.com/spf13/cobra` - CLI framework
- `github.com/stretchr/testify` - Testing utilities
- `go.uber.org/zap` - Structured logging

## Security Considerations

- Never log or expose AWS credentials
- Use AWS IAM roles and policies appropriately
- Sanitize user inputs in CLI commands
- Handle sensitive Terraform state files carefully
- Follow AWS security best practices in examples

## Documentation

- Update README.md for user-facing changes
- Maintain examples in `examples/` directory
- Keep feature files as living documentation
- Update website documentation for major changes
- Include inline code comments for complex logic

## Performance

- Use retries with exponential backoff for AWS operations
- Implement timeouts for all external calls
- Consider pagination for large result sets
- Profile code for performance bottlenecks
- Use appropriate AWS service limits

## Virtual Cloud Integration

InfraSpec uses Virtual Cloud (the embedded AWS emulator) by default for fast, cost-free testing.

**Note:** Emulation is the default behavior. Use `--live` flag to test against real AWS.

### Running Tests

```bash
# Default: Uses embedded emulator (no flag needed)
./infraspec features/aws/s3/s3_bucket.feature

# To test against real AWS, use --live
./infraspec features/aws/s3/s3_bucket.feature --live
```

### How It Works

1. CLI starts the embedded AWS emulator by default
2. Configures Terraform with custom AWS endpoints via environment variables
3. All AWS API calls route to the emulator instead of real AWS
4. Emulator returns AWS-compatible responses

### Service Endpoint Configuration

The endpoint mapping is in `pkg/steps/terraform/terraform.go`:

```go
serviceMap := map[string]string{
    "DYNAMODB":                 "dynamodb",
    "STS":                      "sts",
    "RDS":                      "rds",
    "S3":                       "s3",
    "S3_CONTROL":               "s3",  // Note: underscore required
    "EC2":                      "ec2",
    "SSM":                      "ssm",
    "APPLICATION_AUTO_SCALING": "autoscaling",
}
```

**Critical:** Service keys must match AWS SDK expectations exactly. For example:

- `S3_CONTROL` (correct) - generates `AWS_ENDPOINT_URL_S3_CONTROL`
- `S3CONTROL` (wrong) - generates incorrect env var, requests go to real AWS

### Adding New Service Endpoints

When adding support for a new AWS service:

1. Add entry to `serviceMap` in `pkg/steps/terraform/terraform.go`
2. Use the correct AWS SDK service identifier (check AWS docs)
3. Ensure the builtin AWS emulator implements the service.

## Troubleshooting

### Common Issues

- **InfraSpec API connectivity**: Ensure InfraSpec API is running and accessible on port 3687
- **AWS credentials**: Check AWS configuration and permissions
- **Go module issues**: Run `make tidy` to resolve dependencies
- **Test failures**: Verify InfraSpec API services are running
- **Virtual Cloud 403 errors**: Check service endpoint mapping uses correct AWS SDK identifier
- **Virtual Cloud 404 errors**: Service operation may not be implemented in the AWS emulator

### Development Tools

- Use `make help` to see all available commands
- Check `cover.html` for test coverage reports
- Use Go's built-in profiling tools for performance analysis

## Embedded Virtual Cloud (Emulator)

The Virtual Cloud AWS emulator is embedded directly in this repository under `internal/emulator/`.

### Emulator Architecture

```
internal/emulator/
├── auth/           # SigV4 authentication middleware
├── core/           # Router, state management, types, validator
├── graph/          # Resource relationship graph
├── helpers/        # Utility functions
├── metadata/       # Instance metadata service
├── server/         # HTTP server setup and middleware
├── services/       # AWS service implementations
│   ├── applicationautoscaling/
│   ├── dynamodb/
│   ├── ec2/
│   ├── iam/
│   ├── lambda/
│   ├── rds/
│   ├── s3/
│   ├── sqs/
│   └── sts/
└── testing/        # Test helpers
```

### ⛔ MANDATORY: Response Building Rules

**STOP! Read this section BEFORE implementing ANY handler.**

#### Banned Patterns - NEVER use these:

```go
// BANNED - manual XML construction
responseXML := fmt.Sprintf(`<?xml version="1.0"...`, ...)
xml.MarshalIndent(data, "    ", "  ")  // Only allowed in response_builder.go
errorXML := fmt.Sprintf(`<ErrorResponse>...`)
```

#### Required Pattern - Always use service helpers:

```go
// For Query Protocol services (IAM, RDS, EC2, STS)
type CreateRoleResult struct {
    XMLName xml.Name `xml:"CreateRoleResult"`
    Role    Role     `xml:"Role"`
}

result := CreateRoleResult{Role: role}
return s.successResponse("CreateRole", result)
```

#### Protocol Requirements

| Protocol  | Services                | Content-Type                 | Response Builder          |
| --------- | ----------------------- | ---------------------------- | ------------------------- |
| Query     | RDS, EC2, IAM, STS, SQS | `text/xml`                   | `BuildQueryResponse()`    |
| JSON      | DynamoDB, CloudWatch    | `application/x-amz-json-1.0` | `BuildJSONResponse()`     |
| REST-XML  | S3                      | `application/xml`            | `BuildRESTXMLResponse()`  |
| REST-JSON | Lambda, API Gateway     | `application/json`           | `BuildRESTJSONResponse()` |

#### Mandatory Verification

Run this before completing any service work:

```bash
grep -rn 'fmt\.Sprintf.*<?xml\|fmt\.Sprintf.*<.*Response>\|xml\.MarshalIndent' internal/emulator/services/<service-name>/
# Expected: NO OUTPUT. Any matches must be fixed.
```

### Adding a New AWS Service

1. **Analyze with CloudMirror:**

   ```bash
   cd tools/cloudmirror && go build -o ../../bin/cloudmirror ./cmd/cloudmirror && cd ../..
   ./bin/cloudmirror analyze --service=<name> --output=markdown
   ```

2. **Generate scaffold:**

   ```bash
   ./bin/cloudmirror scaffold --service=<name>
   ```

3. **Generate response types with correct XML tags:**

   ```bash
   ./bin/cloudmirror gentypes --service=<name>
   ```

4. **Implement handlers** following IAM service patterns in `internal/emulator/services/iam/`

### Emulator Code Patterns

#### Handler Pattern

```go
func (s *MyService) handleAction(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
    // 1. Extract and validate parameters
    name := emulator.GetStringParam(params, "Name", "")
    if name == "" {
        return s.errorResponse(400, "ValidationError", "Name is required"), nil
    }

    // 2. Perform business logic and update state
    resource := &types.Resource{Name: name}
    s.state.Set(fmt.Sprintf("myservice:resources:%s", name), resource)

    // 3. Return response using helper
    return s.successResponse("CreateResource", CreateResourceResult{Resource: resource})
}
```

#### State Key Pattern

Use consistent keys: `<service>:<resource-type>:<identifier>`

- `rds:instances:my-database`
- `s3:buckets:my-bucket`
- `iam:roles:my-role`

#### File Naming Convention

Use **snake_case** for handler file names:

- `DeleteScheduledAction` → `delete_scheduled_action_handler.go`
- `CreateDBInstance` → `create_db_instance_handler.go`

### Service File Decomposition

When a service file exceeds ~500 lines, decompose using this structure:

```
internal/emulator/services/<service>/
├── service.go              (~300-500 lines - routing only)
├── types.go                (shared types)
├── response_types.go       (response structs)
├── helpers.go              (parsing utilities, validation)
├── graph_helpers.go        (resource graph methods)
├── responses.go            (response builder methods)
├── create_<resource>_handler.go
├── describe_<resource>_handler.go
└── delete_<resource>_handler.go
```

See `internal/emulator/services/ec2/` for a fully decomposed reference.

### Unit Testing Requirements

Every new AWS operation MUST have unit tests.

```go
func TestMyOperation_Success(t *testing.T) {
    state := emulator.NewMemoryStateManager()
    validator := emulator.NewSchemaValidator()
    service := NewMyService(state, validator)

    req := &emulator.AWSRequest{
        Method:  "POST",
        Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
        Body:    []byte("Action=MyOperation&Param1=value1"),
        Action:  "MyOperation",
    }

    resp, err := service.HandleRequest(context.Background(), req)
    require.NoError(t, err)
    testhelpers.AssertResponseStatus(t, resp, 200)
}
```

### Resource Relationship Graph

The graph (`internal/emulator/graph/`) models AWS resource dependencies for:

- **Dependency validation** - Block deletion of resources with dependents
- **Relationship tracking** - Model containment, references, attachments
- **Cross-service dependencies** - EC2 instances → IAM instance profiles

#### Trust the Graph - Avoid Manual Dependency Checks

```go
// BAD - manual checks
if len(profile.Roles) > 0 {
    return s.errorResponse(409, "DeleteConflict", "...")
}

// GOOD - let graph handle it
if err := s.unregisterResource("role", roleName); err != nil {
    return s.errorResponse(409, "DeleteConflict", fmt.Sprintf("Cannot delete: %v", err)), nil
}
s.state.Delete(stateKey)  // Only after graph validation succeeds
```

#### Pre-Defined AWS Relationships (`internal/emulator/graph/aws_schema.go`)

```go
"ec2:subnet -> ec2:vpc":              {Type: RelContains, Required: true}
"ec2:security-group -> ec2:vpc":      {Type: RelContains, Required: true}
"ec2:instance -> ec2:subnet":         {Type: RelReferences}
"iam:policy -> iam:role":             {Type: RelAssociatedWith}
"iam:instance-profile -> iam:role":   {Type: RelContains}
```

### CloudMirror Commands

```bash
# Build CloudMirror
cd tools/cloudmirror && go build -o ../../bin/cloudmirror ./cmd/cloudmirror && cd ../..

# Common commands
./bin/cloudmirror list implemented                    # List implemented services
./bin/cloudmirror list missing                        # List unimplemented services
./bin/cloudmirror analyze --service=<name>            # Analyze service coverage
./bin/cloudmirror scaffold --service=<name>           # Generate service scaffold
./bin/cloudmirror gentypes --service=<name>           # Generate Go types from Smithy models
./bin/cloudmirror check --target=internal/emulator/services/rds/  # Code pattern analysis
```
