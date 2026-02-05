# S3 Security Rules
# Security rules for AWS S3 buckets

rule "S3_001" {
  name          = "S3 bucket must have encryption"
  description   = "S3 buckets should have server-side encryption enabled to protect data at rest"
  severity      = "error"
  resource_type = "aws_s3_bucket"

  condition {
    any {
      check {
        attribute = "server_side_encryption_configuration"
        operator  = "exists"
      }
      # Also check for the newer separate resource pattern
      check {
        attribute = "bucket"
        operator  = "exists"
      }
    }
  }

  message = "S3 bucket '{{.resource_name}}' does not have server-side encryption configured"

  remediation = <<-EOT
    Add a server_side_encryption_configuration block to your S3 bucket:

    resource "aws_s3_bucket_server_side_encryption_configuration" "example" {
      bucket = aws_s3_bucket.{{.resource_name}}.id
      rule {
        apply_server_side_encryption_by_default {
          sse_algorithm = "aws:kms"
        }
      }
    }
  EOT

  tags = ["security", "s3", "encryption"]
}

rule "S3_002" {
  name          = "S3 bucket should have versioning enabled"
  description   = "S3 buckets should have versioning enabled to protect against accidental deletion"
  severity      = "warning"
  resource_type = "aws_s3_bucket"

  condition {
    check {
      attribute = "versioning.enabled"
      operator  = "equals"
      value     = true
    }
  }

  message = "S3 bucket '{{.resource_name}}' does not have versioning enabled"

  remediation = <<-EOT
    Enable versioning on your S3 bucket:

    resource "aws_s3_bucket_versioning" "example" {
      bucket = aws_s3_bucket.{{.resource_name}}.id
      versioning_configuration {
        status = "Enabled"
      }
    }
  EOT

  tags = ["security", "s3", "data-protection"]
}

rule "S3_003" {
  name          = "S3 bucket should block public access"
  description   = "S3 buckets should have public access blocked unless explicitly required"
  severity      = "error"
  resource_type = "aws_s3_bucket_public_access_block"

  condition {
    all {
      check {
        attribute = "block_public_acls"
        operator  = "equals"
        value     = true
      }
      check {
        attribute = "block_public_policy"
        operator  = "equals"
        value     = true
      }
      check {
        attribute = "ignore_public_acls"
        operator  = "equals"
        value     = true
      }
      check {
        attribute = "restrict_public_buckets"
        operator  = "equals"
        value     = true
      }
    }
  }

  message = "S3 bucket '{{.resource_name}}' does not have all public access blocks enabled"

  remediation = <<-EOT
    Add a public access block configuration:

    resource "aws_s3_bucket_public_access_block" "example" {
      bucket = aws_s3_bucket.example.id

      block_public_acls       = true
      block_public_policy     = true
      ignore_public_acls      = true
      restrict_public_buckets = true
    }
  EOT

  tags = ["security", "s3", "public-access"]
}

rule "S3_004" {
  name          = "S3 bucket should have logging enabled"
  description   = "S3 buckets should have access logging enabled for audit purposes"
  severity      = "warning"
  resource_type = "aws_s3_bucket"

  condition {
    check {
      attribute = "logging"
      operator  = "exists"
    }
  }

  message = "S3 bucket '{{.resource_name}}' does not have access logging enabled"

  remediation = <<-EOT
    Enable access logging on your S3 bucket:

    resource "aws_s3_bucket_logging" "example" {
      bucket = aws_s3_bucket.{{.resource_name}}.id

      target_bucket = aws_s3_bucket.log_bucket.id
      target_prefix = "log/{{.resource_name}}/"
    }
  EOT

  tags = ["security", "s3", "logging", "audit"]
}
