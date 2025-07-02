# InfraSpec Development Guide

This document provides guidance for AI coding assistants working on the InfraSpec project.

## Project Overview

**InfraSpec** is a tool for testing your cloud infrastructure written in Go that allows users to write infrastructure
tests in plain English using Gherkin syntax. The project tests infrastructure code for Terraform, Docker, and
Kubernetes without requiring users to write traditional test code using frameworks like Terratest.

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
- LocalStack (for AWS emulation during testing)

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
- Write integration tests that work with LocalStack
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

### Adding New AWS Cloud Services
1. Create AWS service-specific assertion functions in `pkg/assertions/aws/`
2. Add corresponding step definitions in `pkg/steps/aws/`
3. Create feature examples in `examples/aws/`
4. Write Gherkin feature files in `features/aws/`
5. Update documentation and roadmap

### Testing Philosophy
- **BDD First**: Write Gherkin scenarios before implementation
- **LocalStack Integration**: Use LocalStack for AWS service emulation
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

## Troubleshooting

### Common Issues
- **LocalStack connectivity**: Ensure LocalStack is running and accessible
- **AWS credentials**: Check AWS configuration and permissions
- **Go module issues**: Run `make tidy` to resolve dependencies
- **Test failures**: Verify LocalStack services are running

### Development Tools
- Use `make help` to see all available commands
- Check `cover.html` for test coverage reports
- Use Go's built-in profiling tools for performance analysis
