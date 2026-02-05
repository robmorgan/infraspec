# Spec file for terraform-good module
# Rules specific to this Terraform configuration

rule "LOCAL_001" {
  name          = "Production buckets require lifecycle rules"
  description   = "Production S3 buckets should have lifecycle rules for cost management"
  severity      = "info"
  resource_type = "aws_s3_bucket"

  condition {
    any {
      check {
        attribute = "lifecycle_rule"
        operator  = "exists"
      }
      # Allow buckets that aren't production
      check {
        attribute = "tags.Environment"
        operator  = "not_equals"
        value     = "production"
      }
    }
  }

  message = "Production S3 bucket '{{.resource_name}}' should have lifecycle rules configured"

  remediation = <<-EOT
    Add lifecycle rules to manage object lifecycle:

    resource "aws_s3_bucket_lifecycle_configuration" "example" {
      bucket = aws_s3_bucket.example.id

      rule {
        id     = "expire-old-versions"
        status = "Enabled"

        noncurrent_version_expiration {
          noncurrent_days = 90
        }
      }
    }
  EOT

  tags = ["cost", "lifecycle", "s3"]
}
