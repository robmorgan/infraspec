<h1>
<p align="center">
  <img src="https://github.com/user-attachments/assets/d744b90a-1e44-4b1e-9f5b-35f948991620" alt="InfraSpec Logo" width="128">
  <br>InfraSpec
</h1>
  <p align="center">
    <strong>✅ Test your Terraform AWS infrastructure in plain English, no code required.</strong>
  </p>
</p>

<p align="center">
  <a href="https://github.com/robmorgan/infraspec/actions"><img src="https://github.com/robmorgan/infraspec/workflows/Test/badge.svg" alt="Build Status"></a>
  <a href="https://github.com/robmorgan/infraspec/blob/main/LICENSE.md"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/robmorgan/infraspec"><img src="https://goreportcard.com/badge/github.com/robmorgan/infraspec" alt="Go Report Card"></a>
  <a href="https://github.com/robmorgan/infraspec/releases"><img src="https://img.shields.io/github/v/release/robmorgan/infraspec" alt="Release"></a>
</p>

<p align="center">
  <a href="https://infraspec.sh/virtual-cloud">
    <img src="https://img.shields.io/badge/🚀%20Try-InfraSpec%20Virtual%20Cloud-5B47ED?style=for-the-badge" alt="Try Virtual Cloud">
  </a>
  <br>
  <sub>100x faster tests • 90% cost savings • Zero cleanup</sub>
</p>

---

Testing infrastructure shouldn't require learning complex testing frameworks or writing hundreds of lines of code.
InfraSpec lets you write infrastructure tests in **plain English** using the battle-tested Gherkin syntax.

Traditional infrastructure testing solutions like Terratest require programming knowledge, so writing the tests can take
as long as writing the infrastructure itself. They also limit collaboration so non-technical stakeholders can’t review
or contribute, and tests often become difficult to maintain or understand over time.

InfraSpec combines a rich library of pre-built testing patterns with natural language specifications. Write tests that
read like documentation and are executable from day one.

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

## 💰 InfraSpec Virtual Cloud - Test Faster, Pay Less

Stop paying AWS for every test run. **InfraSpec Virtual Cloud** is a high-fidelity AWS emulator designed specifically
for infrastructure testing. Make your tests faster and cheaper:

### Why Virtual Cloud?

**Testing against real AWS is expensive and slow:**

- 💸 **High costs** - $0.10+ per test run adds up fast
- ⏱️ **Slow execution** - Network calls take 2-10 seconds each
- 🔄 **CI/CD bills** - Costs accumulate with every pipeline run
- 🐌 **Poor feedback** - Slow tests reduce developer productivity
- 🗑️ **Cleanup headaches** - Failed tests leave orphaned resources that cost money

**InfraSpec Virtual Cloud solves this:**

- ⚡ **100x faster** - Tests run in milliseconds, not seconds
- 💰 **90% cheaper** - No AWS charges for testing infrastructure
- 🔒 **No credentials needed** - Test safely without AWS access
- 🧹 **No cleanup required** - Failed tests don't leave dangling resources in AWS
- 🎯 **High fidelity** - Accurate AWS behavior without complexity
- 🚀 **CI/CD optimized** - Unlimited parallel tests without cost concerns

### Getting Started with Virtual Cloud

Add the `--virtual-cloud` flag to your tests:

```bash
# Use Virtual Cloud for fast, cost-free testing
infraspec --virtual-cloud features/
```

Or set it as default in your CI/CD:

```yaml
- name: Run InfraSpec Tests
  run: infraspec --virtual-cloud features/
  env:
    INFRASPEC_CLOUD_TOKEN: ${{ secrets.INFRASPEC_CLOUD_TOKEN }}
```

→ [**Get started**](https://infraspec.sh/virtual-cloud) → [View pricing](https://infraspec.sh/pricing) →
[Compare with alternatives](https://infraspec.sh/vs-localstack)

## ✨ Features

- 🗣️ **Plain English syntax** - Write tests that read like documentation using Gherkin
- 👥 **Team-friendly** - Non-technical stakeholders can read, review, and contribute
- 🚀 **Zero boilerplate** - Works with your existing Terraform configurations out of the box
- 📚 **Rich assertion library** - Hundreds of pre-built assertions for AWS resources
- ⚡ **Fast feedback** - Catch infrastructure issues before they reach production
- 🔄 **CI/CD ready** - Integrates seamlessly with your existing pipelines
- 💰 **Cost effective** - Use Virtual Cloud to eliminate AWS testing costs
- 🧪 **Flexible testing** - Test against real AWS or Virtual Cloud emulator

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

| Service         | Status       | Virtual Cloud | Example Assertions                              |
| --------------- | ------------ | ------------- | ----------------------------------------------- |
| **S3**          | ✅ Supported | ✅ Supported  | Versioning, encryption, public access, logging  |
| **DynamoDB**    | ✅ Supported | ✅ Supported  | Tables, indexes, capacity modes, encryption     |
| **RDS**         | ✅ Supported | ✅ Supported  | Instances, security groups, backups, encryption |
| **EC2**         | 🚧 Partial   | ✅ Supported  | Basic instance validation                       |
| **SSM**         | 🚧 Partial   | ✅ Supported  | Parameter store                                 |
| **API Gateway** | ⏳ Planned   | ⏳ Planned    | -                                               |
| **Lambda**      | ⏳ Planned   | ⏳ Planned    | -                                               |

💡 **Tip:** All services marked with ✅ in Virtual Cloud work with the `--virtual-cloud` flag for instant, cost-free
testing.

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

We're actively expanding InfraSpec's capabilities. Here's what's on the horizon:

### Current Status

| AWS Service | Status       | Coverage                                                |
| ----------- | ------------ | ------------------------------------------------------- |
| S3          | ✅ Supported | Buckets, versioning, encryption, logging, public access |
| DynamoDB    | ✅ Supported | Tables, GSI/LSI, billing modes, streams, encryption     |
| RDS         | ✅ Supported | Instances, snapshots, security groups, backups          |
| EC2         | 🚧 Partial   | Basic instance validation                               |
| SSM         | 🚧 Partial   | Parameter store                                         |

### Coming Soon

- 🔜 **Lambda** - Function testing and event validation
- 🔜 **API Gateway** - REST and HTTP API testing
- 🔜 **VPC** - Network configuration validation
- 🔜 **ECS/EKS** - Container orchestration testing
- 🔜 **CloudFront** - CDN configuration validation

[View full roadmap →](https://github.com/users/robmorgan/projects/1)

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

**Note:** Our tests use [InfraSpec Virtual Cloud](https://infraspec.sh/virtual-cloud), a high-fidelity AWS emulator, to
save time and costs during development. The Virtual Cloud API is available as a paid service for production use.

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
