# InfraSpec Configuration
# This file configures InfraSpec Gatekeeper for this repository

config {
  min_severity = "warning"  # Report warnings and errors
  format       = "text"     # Output format (text, json)
  strict       = false      # Don't treat unknowns as violations
  # no_builtin = false      # Use built-in rules
}

# You can also define rules directly in this file
# These rules apply to all Terraform files in the repository

rule "REPO_001" {
  name          = "All S3 buckets must have tags"
  description   = "Repository policy: all S3 buckets must have at least one tag"
  severity      = "warning"
  resource_type = "aws_s3_bucket"

  condition {
    check {
      attribute = "tags"
      operator  = "exists"
    }
  }

  message = "S3 bucket '{{.resource_name}}' has no tags defined"

  remediation = <<-EOT
    Add tags to your S3 bucket:

    resource "aws_s3_bucket" "example" {
      bucket = "my-bucket"

      tags = {
        Name        = "my-bucket"
        Environment = "production"
      }
    }
  EOT

  tags = ["tagging", "repository-policy"]
}
