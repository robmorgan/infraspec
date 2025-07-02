<h1>
<p align="center">
  <img src="https://github.com/user-attachments/assets/d744b90a-1e44-4b1e-9f5b-35f948991620" alt="InfraSpec Logo" width="128">
  <br>InfraSpec
</h1>
  <p align="center">
    <strong>✅ Test your cloud infrastructure in plain English, no code required.</strong>
  </p>
</p>

## About

Write infrastructure tests for Terraform, Docker, and Kubernetes in plain English, without writing a single line of
code. InfraSpec pairs a vast library of common testing patterns with a domain-specific language for testing
infrastructure.

Tests are written using easy to learn [Gherkin](https://cucumber.io/docs/gherkin/) syntax, which is suitable for both
technical and non-technical team members.

## Quick Start

Here's how easy it is to test a Terraform S3 bucket configuration:

```gherkin
Feature: S3 Bucket Creation
  As a DevOps Engineer
  I want to create an S3 bucket with guardrails
  So that I can store my data securely

  Scenario: Create an S3 bucket with a name
    Given I have a Terraform configuration in "./s3-bucket"
    And I set variable "bucket_name" to "my-bucket"
    When I run Terraform apply
    Then the S3 bucket "my-bucket" should exist
    And the S3 bucket "my-bucket" should have a versioning configuration
    And the S3 bucket "my-bucket" should have a public access block
    And the S3 bucket "my-bucket" should have a server access logging configuration
    And the S3 bucket "my-bucket" should have a encryption configuration
```

InfraSpec automatically translates your natural language specifications into executable infrastructure tests.

> [!WARNING]
> This project is still in heavy development and is likely to change!

## ✨ Features

- 🗣️ **Plain English syntax** - Write tests that read like documentation.
- 👥 **Team-friendly** - Non-technical stakeholders can contribute and review.
- 🚀 **Zero setup** - Works with your existing Terraform, Docker, and Kubernetes configurations.
- 📚 **Rich test library** - Hundreds of pre-built testing patterns for common scenarios.
- ⚡ **Fast feedback** - Catch infrastructure issues before they reach production.

## What can you test?

🏗️ Terraform

- Resource configurations and relationships
- Security policies and compliance
- Cost optimization rules
- Multi-environment consistency

🐳 Docker

- Image security and vulnerabilities
- Container resource limits
- Network configurations
- Build optimization

☸️ Kubernetes

- Deployment strategies and rollouts
- Resource quotas and limits
- Network policies and security contexts
- Service mesh configurations

## Roadmap & Status

At the moment, only a subset of AWS infrastructure is supported, but over time we hope to support other clouds and
tooling.

| **Product**   | **Description**     | **Status**   |
| ------------- | ------------------- | ------------ |
| API Gateway   | Not Implemented     |       ⏳     |
| DynamoDB      | Partially Supported |       ✅     |
| ElastiCache   | Not Implemented     |       ⏳     |
| RDS           | Partially Supported |       ✅     |
| RDS Aurora    | Not Implemented     |       ⏳     |
| S3            | Partially Supported |       ✅     |

## Getting Started

1. Install InfraSpec using Homebrew:

```sh
brew tap robmorgan/infraspec
brew install infraspec
```

Or if you have Go installed, you can install InfraSpec using:

```sh
go install github.com/robmorgan/infraspec@latest
```

2. Initialize a repo containing your infrastructure code:

```sh
infraspec init # creates a ./features directory if it doesn't already exist
```

3. Create your first infrastructure test:

```sh
infraspec new dynamodb.feature
```

4. Run the tests

```sh
infraspec features/dynamodb.feature
```

> [!TIP]
> If your using VS Code, we recommend installing the [Cucumber (Gherkin) Full Support](https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete)
extension for syntax highlighting.

## Contributions

Contributions are welcome! Please open an issue or submit a pull request. Please note, that this project is still in
it's infancy and many internal APIs are likely to change.

**Note:** Our tests use [LocalStack](https://github.com/localstack/localstack), which emulates the AWS APIs, in order to
save both time and money.

## License

[Apache License 2.0](https://github.com/robmorgan/infraspec/blob/main/LICENSE.md)
