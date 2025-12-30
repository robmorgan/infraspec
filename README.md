<h1>
<p align="center">
  <img src="https://github.com/user-attachments/assets/d744b90a-1e44-4b1e-9f5b-35f948991620" alt="InfraSpec Logo" width="128">
  <br>InfraSpec
</h1>
  <p align="center">
    <strong>✅ Test your AWS infrastructure code in plain English using virtual or real AWS APIs.</strong>
  </p>
</p>

<p align="center">
  <a href="https://github.com/robmorgan/infraspec/actions"><img src="https://github.com/robmorgan/infraspec/workflows/Test/badge.svg" alt="Build Status"></a>
  <a href="https://github.com/robmorgan/infraspec/blob/main/LICENSE.md"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/robmorgan/infraspec"><img src="https://goreportcard.com/badge/github.com/robmorgan/infraspec" alt="Go Report Card"></a>
  <a href="https://github.com/robmorgan/infraspec/releases"><img src="https://img.shields.io/github/v/release/robmorgan/infraspec" alt="Release"></a>
</p>

---

Testing infrastructure code shouldn't require learning complex testing frameworks or writing hundreds of lines of code.
InfraSpec lets you write infrastructure tests in **plain English** using the battle-tested Gherkin syntax.

Traditional infrastructure testing solutions like Terratest require programming knowledge, so writing the tests can take
as long as writing the infrastructure itself. InfraSpec focuses on **intent**. Tests read like documentation, are easy
to review, and remain understandable over time. This makes collaboration possible not just for engineers, but also for
platform, security, and compliance stakeholders.

InfraSpec runs tests against a built-in virtual cloud emulator for fast, deterministic local and CI workflows, or
against real AWS APIs for end-to-end validation. Pre-built testing patterns and natural-language specifications ensure
tests remain readable, maintainable, and executable from day one.

## ⚡ Quick Example

Here's how easy it is to test a Terraform S3 bucket configuration:

```gherkin
Feature: S3 Bucket Creation
  As a DevOps Engineer
  I want to create an S3 bucket with security guardrails
  So that I can store my data securely

  Scenario: Create a secure S3 bucket
    Given I have a Terraform configuration in "./examples/aws/s3/s3-bucket"
    And I set variable "bucket_name" to "my-data-bucket" with a random suffix
    When I run Terraform apply
    Then the S3 bucket from output "bucket_name" should exist
    And the S3 bucket from output "bucket_name" should have versioning enabled
    And the S3 bucket from output "bucket_name" should have a public access block
    And the S3 bucket from output "bucket_name" should have encryption enabled
```

Run it:

```bash
infraspec features/s3_bucket.feature
```

That's it! No code to write, no frameworks to learn. InfraSpec handles the rest.

→ [**Check out our example features**](https://github.com/robmorgan/infraspec/tree/main/features)

## ✨ Features

- 🗣️ **Plain English syntax** - Write tests that read like documentation using Gherkin
- ⚡️ **Fast feedback** - 10-100x faster than creating real AWS infrastructure
- 👥 **Team-friendly** - Non-technical stakeholders can read, review, and contribute
- 🚀 **Zero boilerplate** - Works with your existing Terraform configurations out of the box
- 📚 **Rich assertion library** - Hundreds of pre-built assertions for AWS resources
- 🔄 **CI/CD ready** - Integrates seamlessly with your existing pipelines
- 💰 **Cost effective** - Built-in emulator eliminates AWS testing costs
- 🧪 **Flexible testing** - Test against real AWS or the Virtual Cloud emulator

## 🚀 Installation

### Homebrew (macOS/Linux)

```bash
brew tap robmorgan/infraspec
brew install infraspec
```

### Go Install

```bash
go install github.com/robmorgan/infraspec@latest
```

### Binary Download

Download the latest release for your platform from the [releases page](https://github.com/robmorgan/infraspec/releases).

### Verify Installation

```bash
infraspec --version
```

## 📖 Getting Started

### 1. Initialize Your Project

Navigate to your Terraform project directory and initialize InfraSpec:

```bash
cd my-terraform-project
infraspec init
```

This creates a `features/` directory where your tests will live.

### 2. Create Your First Test

Generate a test template for the service you want to test:

```bash
infraspec new s3_bucket.feature
```

Or create a test manually in `features/s3_bucket.feature`:

```gherkin
Feature: S3 Bucket Security
  Scenario: Bucket has encryption enabled
    Given I have a Terraform configuration in "./terraform/s3"
    And I set variable "bucket_name" to "test-bucket" with a random suffix
    When I run Terraform apply
    Then the S3 bucket from output "bucket_name" should exist
    And the S3 bucket from output "bucket_name" should have encryption enabled
```

### 3. Run Your Tests

```bash
infraspec features/s3_bucket.feature
```

Or run all tests:

```bash
infraspec features/
```

Optionally use the `--live` flag to run against real AWS APIs (be sure to cleanup any dangling resources):

```bash
infraspec --live features/
```

→ [**Learn more about live testing on AWS**](https://infraspec.sh/docs/guides/live-testing)

### 4. Integrate with CI/CD

Add to your GitHub Actions workflow:

```yaml
- name: Run InfraSpec Tests
  run: |
    infraspec features/
```

## 🔍 What Can You Test?

### 🏗️ Terraform

- ✅ Resource configurations and outputs
- ✅ Security policies and compliance rules
- ✅ Cost optimization validations
- ✅ Multi-environment consistency
- ✅ Variable validation

### ☁️ AWS Resources

→ [**See the AWS Compatibility page for the current status**](https://infraspec.sh/docs/compatibility)

### 🌐 HTTP/APIs

- ✅ HTTP(S) endpoints and status codes
- ✅ Response headers and bodies
- ✅ Form data and file uploads
- ✅ JSON/XML response validation

## 📚 Real-World Examples

### DynamoDB Table with GSI

```gherkin
Scenario: DynamoDB table with Global Secondary Index
  Given I have a Terraform configuration in "./terraform/dynamodb"
  And I set variable "table_name" to "users-table" with a random suffix
  When I run Terraform apply
  Then the DynamoDB table from output "table_name" should exist
  And the DynamoDB table from output "table_name" should have encryption enabled
  And the DynamoDB table from output "table_name" should have "PAY_PER_REQUEST" billing mode
  And the DynamoDB table from output "table_name" should have 1 global secondary index
```

### RDS Instance Security

```gherkin
Scenario: RDS instance meets security requirements
  Given I have a Terraform configuration in "./terraform/rds"
  And I set variable "db_identifier" to "production-db" with a random suffix
  When I run Terraform apply
  Then the RDS instance from output "db_instance_id" should exist
  And the RDS instance from output "db_instance_id" should not be publicly accessible
  And the RDS instance from output "db_instance_id" should have encryption enabled
  And the RDS instance from output "db_instance_id" should have automated backups enabled
```

### Multi-Environment Validation

```gherkin
Scenario Outline: S3 bucket configuration across environments
  Given I have a Terraform configuration in "./terraform/s3"
  And I set variable "environment" to "<environment>"
  When I run Terraform apply
  Then the S3 bucket from output "bucket_name" should exist
  And the S3 bucket from output "bucket_name" should have the tag "Environment" with value "<environment>"

  Examples:
    | environment |
    | dev         |
    | staging     |
    | production  |
```

## 🆚 InfraSpec vs. Alternatives

| Feature                    | InfraSpec               | Terratest | Terraform Testing | Conftest     |
| -------------------------- | ----------------------- | --------- | ----------------- | ------------ |
| **Language**               | Plain English (Gherkin) | Go        | HCL               | Rego         |
| **Learning Curve**         | Low                     | High      | Medium            | Medium       |
| **AWS Integration**        | Native                  | Manual    | Limited           | Policy-based |
| **Non-technical Friendly** | ✅ Yes                  | ❌ No     | ⚠️ Partial        | ❌ No        |
| **Live Resource Testing**  | ✅ Yes                  | ✅ Yes    | ❌ No             | ❌ No        |
| **Pre-built Assertions**   | ✅ Hundreds             | ❌ None   | ⚠️ Some           | ❌ None      |

## 🎯 Roadmap

We're actively expanding InfraSpec's capabilities. For more information see the
[compatibility page](https://infraspec.sh/docs/compatibility).

## 💡 Editor Support

### VS Code

Install the
[Cucumber (Gherkin) Full Support](https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete)
extension for:

- Syntax highlighting
- Auto-completion
- Step definition navigation

### IntelliJ IDEA / PyCharm

Enable the built-in Gherkin plugin for full IDE support.

## 🤝 Contributing

We welcome contributions! Whether you're fixing bugs, adding features, or improving documentation, your help makes
InfraSpec better.

### Ways to Contribute

- 🐛 [Report bugs](https://github.com/robmorgan/infraspec/issues/new?template=bug_report.md)
- 💡 [Request features](https://github.com/robmorgan/infraspec/issues/new?template=feature_request.md)
- 📝 Improve documentation
- 🔧 Submit pull requests
- ⭐ Star the project to show support

### Development Setup

```bash
# Clone the repository
git clone https://github.com/robmorgan/infraspec.git
cd infraspec

# Install dependencies
make deps

# Run tests
make test

# Build locally
make build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📞 Community & Support

- 💬 [GitHub Discussions](https://github.com/robmorgan/infraspec/discussions) - Ask questions and share ideas
- 🐛 [Issue Tracker](https://github.com/robmorgan/infraspec/issues) - Report bugs and request features
- 📖 [Documentation](https://infraspec.sh) - Full documentation and guides
- 🐦 [Twitter/X](https://twitter.com/_rjm_) - Follow @\_rjm\_ for updates

## 📄 License

InfraSpec is open source software licensed under the
[Apache License 2.0](https://github.com/robmorgan/infraspec/blob/main/LICENSE.md).

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/robmorgan">Rob Morgan</a> and <a href="https://github.com/robmorgan/infraspec/graphs/contributors">contributors</a>
  <br>
  ⭐ Star us on GitHub to support the project!
</p>
