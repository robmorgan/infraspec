# InfraSpec Gatekeeper Examples

This directory contains examples demonstrating how to use InfraSpec Gatekeeper
for pre-apply security checks on Terraform configurations.

## Quick Start

### Check Compliant Terraform

The `terraform-good/` directory contains a Terraform configuration that follows
security best practices:

```bash
# Should pass all checks
infraspec check ./terraform-good
```

### Check Non-Compliant Terraform

The `terraform-bad/` directory contains a Terraform configuration with
intentional security issues:

```bash
# Will report violations
infraspec check ./terraform-bad
```

Expected output:
```
InfraSpec Gatekeeper
Checking 1 file(s)...

=== Violations ===

[ERROR] S3_001: S3 bucket must have encryption
  Resource: aws_s3_bucket.insecure_bucket
  File: main.tf:19
  Message: S3 bucket 'insecure_bucket' does not have server-side encryption configured

[ERROR] SG_001: No SSH from 0.0.0.0/0
  Resource: aws_security_group.insecure_sg
  File: main.tf:40
  Message: Security group 'insecure_sg' allows SSH access from 0.0.0.0/0

... more violations ...

=== Summary ===
Result: FAIL
Files: 1 | Resources: 8 | Rules: 13
Violations: 6 error(s), 4 warning(s)
```

### Use Custom Rules

The `custom-rules/` directory contains example custom rules:

```bash
# Check with custom rules
infraspec check ./terraform-bad --rules ./custom-rules/my-org-rules.yaml
```

### JSON Output for CI

```bash
# Output JSON for parsing in CI pipelines
infraspec check ./terraform-bad --format json > results.json
```

### List Available Rules

```bash
# List all built-in rules
infraspec check --list-rules

# List with descriptions
infraspec check --list-rules --verbose
```

### Exclude/Include Rules

```bash
# Exclude specific rules
infraspec check ./terraform --exclude S3_004,VPC_001

# Only run specific rules
infraspec check ./terraform --include S3_001,S3_002,SG_001
```

### Filter by Severity

```bash
# Only show errors (ignore warnings and info)
infraspec check ./terraform --severity error

# Show warnings and above (default)
infraspec check ./terraform --severity warning
```

## Directory Structure

```
examples/gatekeeper/
├── README.md                 # This file
├── terraform-good/           # Compliant Terraform
│   └── main.tf
├── terraform-bad/            # Non-compliant Terraform
│   └── main.tf
├── custom-rules/             # Example custom rules
│   └── my-org-rules.yaml
└── .github/
    └── workflows/
        └── infraspec-check.yml  # GitHub Actions example
```

## Writing Custom Rules

See `custom-rules/my-org-rules.yaml` for examples. Rules are defined in YAML:

```yaml
version: "1"
metadata:
  name: "My Rules"

rules:
  - id: MY_001
    name: "Rule name"
    description: "Longer description"
    severity: error  # error, warning, or info
    resource_type: aws_s3_bucket
    condition:
      attribute: encryption
      operator: exists
    message: "Bucket '{{.resource_name}}' needs encryption"
    remediation: "Add encryption block..."
    tags: [security]
```

### Available Operators

- `exists` / `not_exists` - Check attribute presence
- `equals` / `not_equals` - Exact match
- `contains` / `not_contains` - Array/string contains
- `matches` - Regex match
- `greater_than` / `less_than` - Numeric comparison
- `one_of` - Value is in list
- `all` / `any` / `not` - Logical combinators

### Attribute Paths

- `attribute` - Simple attribute
- `nested.path` - Nested object
- `array[0].field` - Array index
- `array[*].field` - All array elements

## CI/CD Integration

See `.github/workflows/infraspec-check.yml` for a GitHub Actions example.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks passed |
| 1 | One or more violations found |
| 2 | Configuration or parse error |
