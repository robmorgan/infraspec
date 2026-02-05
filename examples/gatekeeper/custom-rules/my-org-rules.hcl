# My Organization Rules
# Custom security rules for ACME Corp

# Custom naming convention rule
rule "ACME_001" {
  name          = "S3 bucket must follow naming convention"
  description   = "All S3 buckets must start with 'acme-' prefix"
  severity      = "error"
  resource_type = "aws_s3_bucket"

  condition {
    check {
      attribute = "bucket"
      operator  = "matches"
      value     = "^acme-.*"
    }
  }

  message = "S3 bucket '{{.resource_name}}' does not follow naming convention (must start with 'acme-')"

  remediation = <<-EOT
    Rename your bucket to start with 'acme-':

    resource "aws_s3_bucket" "example" {
      bucket = "acme-my-bucket-name"
    }
  EOT

  tags = ["naming", "standards"]
}

# Custom tagging rule
rule "ACME_002" {
  name          = "All resources must have cost-center tag"
  description   = "Required for cost allocation tracking"
  severity      = "warning"
  resource_type = "aws_s3_bucket"

  condition {
    check {
      attribute = "tags.cost-center"
      operator  = "exists"
    }
  }

  message = "S3 bucket '{{.resource_name}}' is missing required 'cost-center' tag"

  remediation = <<-EOT
    Add a cost-center tag to your resource:

    resource "aws_s3_bucket" "example" {
      bucket = "my-bucket"

      tags = {
        cost-center = "engineering"  # Add this
      }
    }
  EOT

  tags = ["tagging", "cost"]
}

# Custom environment validation
rule "ACME_003" {
  name          = "Environment tag must be valid"
  description   = "Environment must be one of: dev, staging, production"
  severity      = "error"
  resource_type = "aws_s3_bucket"

  condition {
    any {
      check {
        attribute = "tags.Environment"
        operator  = "not_exists"
      }
      check {
        attribute = "tags.Environment"
        operator  = "one_of"
        value     = ["dev", "staging", "production"]
      }
    }
  }

  message = "S3 bucket '{{.resource_name}}' has invalid Environment tag (must be dev, staging, or production)"

  remediation = <<-EOT
    Set a valid environment tag:

    tags = {
      Environment = "production"  # Must be: dev, staging, or production
    }
  EOT

  tags = ["tagging", "standards"]
}

# Instance type restrictions
rule "ACME_004" {
  name          = "EC2 instances must use approved types"
  description   = "Only approved instance types are allowed for cost control"
  severity      = "warning"
  resource_type = "aws_instance"

  condition {
    check {
      attribute = "instance_type"
      operator  = "one_of"
      value     = ["t3.micro", "t3.small", "t3.medium", "t3.large", "m5.large", "m5.xlarge"]
    }
  }

  message = "EC2 instance '{{.resource_name}}' uses non-approved instance type"

  remediation = <<-EOT
    Use an approved instance type:
    - t3.micro, t3.small, t3.medium, t3.large
    - m5.large, m5.xlarge

    For larger instances, file a request with the Cloud Team.
  EOT

  tags = ["cost", "compute"]
}
