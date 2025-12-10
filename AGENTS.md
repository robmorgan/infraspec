# AI Agent Guidelines for InfraSpec Development

This document provides specific guidance for AI coding assistants working on the InfraSpec project.

## General Principles

### Terraform/OpenTofu Configuration

**Always prefer Terraform environment variables over generating providers.tf files.**

When implementing features that require Terraform/OpenTofu provider configuration:

- ✅ **DO**: Use environment variables that Terraform recognizes (e.g., `AWS_S3_USE_PATH_STYLE`, `AWS_ENDPOINT_URL_*`)
- ✅ **DO**: Set environment variables in the `options.EnvVars` map within the provisioner options
- ✅ **DO**: Document which environment variables are being set and why

- ❌ **DON'T**: Auto-generate `providers.tf` or `provider.tf` files
- ❌ **DON'T**: Create temporary configuration files that need to be cleaned up
- ❌ **DON'T**: Modify user's existing Terraform configuration files

**Rationale**: Environment variables are:
- Non-invasive and don't modify the user's codebase
- Easily overridable by users if needed
- Standard practice in Terraform/OpenTofu workflows
- Simpler to maintain and debug

**Example**:
```go
// Good: Using environment variables
options.EnvVars["AWS_ENDPOINT_URL_S3"] = endpoint

// Bad: Generating provider files
// DO NOT DO THIS
providerContent := `provider "aws" { ... }`
os.WriteFile("providers.tf", []byte(providerContent), 0644)
```

### Available Terraform Environment Variables

When working with the AWS provider, prefer these environment variables:

- `AWS_ENDPOINT_URL` - General AWS endpoint override
- `AWS_ENDPOINT_URL_<SERVICE>` - Service-specific endpoint (e.g., `AWS_ENDPOINT_URL_S3`, `AWS_ENDPOINT_URL_DYNAMODB`)
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `AWS_REGION` - Default AWS region

**Note**: InfraSpec uses S3 virtual hosted-style URLs by default when `--virtual-cloud` is enabled (e.g., `bucket-name.s3.infraspec.sh`). Path-style URLs are no longer supported.

See the [Terraform AWS Provider documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) for the complete list of supported environment variables.

### Terraform Module Usage

**Prefer community modules from `terraform-aws-modules` over plain Terraform resources.**

When writing Terraform configurations for AWS infrastructure:

- ✅ **DO**: Use modules from the [`terraform-aws-modules`](https://github.com/terraform-aws-modules) GitHub organization when available
- ✅ **DO**: Check the registry at `registry.terraform.io/modules/terraform-aws-modules/` for available modules
- ✅ **DO**: Pin module versions explicitly for reproducibility

- ❌ **DON'T**: Write raw `aws_*` resources when a well-maintained module exists
- ❌ **DON'T**: Reinvent common patterns that modules already handle

**Rationale**: The `terraform-aws-modules` community modules:
- Follow AWS and Terraform best practices
- Handle edge cases and common configurations
- Are actively maintained and widely used
- Reduce boilerplate and potential misconfigurations
- Include sensible defaults while remaining configurable

**Example**:
```hcl
# Good: Using terraform-aws-modules
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.1.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]
}

# Avoid: Manual resource definitions for common infrastructure
# resource "aws_vpc" "main" { ... }
# resource "aws_subnet" "private" { ... }
# resource "aws_internet_gateway" "main" { ... }
# ... many more resources
```

**Common modules to use**:
- `terraform-aws-modules/vpc/aws` - VPC, subnets, NAT gateways
- `terraform-aws-modules/eks/aws` - EKS clusters
- `terraform-aws-modules/rds/aws` - RDS instances and clusters
- `terraform-aws-modules/s3-bucket/aws` - S3 buckets
- `terraform-aws-modules/lambda/aws` - Lambda functions
- `terraform-aws-modules/security-group/aws` - Security groups

## Project-Specific Guidelines

For other project-specific coding standards and guidelines, see [CLAUDE.md](./CLAUDE.md).
