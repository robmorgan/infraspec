<div align="center">
<h1 align="center">
<img src="_docs/infraspec_logo.jpg" width="200" />
</h1>
<h3>Write infrastructure tests in plain English, without writing a single line of code.</h3>
</div>

:warning: This project is still under heavy development and probably won't work!

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
    And the S3 bucket "my-bucket" should have a tags
            | Key         | Value     |
            | Environment | test      |
            | Project     | infratest |
```

Under the hood, InfraSpec uses [Gherkin](https://cucumber.io/docs/gherkin/) to parse
Go Dog and testing modules from Terratest.

## Why?

Additionally, LLMs are great at generating scenarios using the Gherkin syntax, so you can write tests in plain English
and InfraSpec will translate them into code.

## Installation

```sh
go install github.com/robmorgan/infraspec@latest
```

## Writing Tests

If your using VS Code, we recommend installing the [Cucumber (Gherkin) Full Support](https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete)
extension for syntax highlighting.
