# IAM Security Rules
# Security rules for AWS IAM resources

rule "IAM_001" {
  name          = "IAM role should not have inline policy"
  description   = "IAM roles should use managed policies instead of inline policies for better governance"
  severity      = "warning"
  resource_type = "aws_iam_role"

  condition {
    check {
      attribute = "inline_policy"
      operator  = "not_exists"
    }
  }

  message = "IAM role '{{.resource_name}}' has an inline policy - use managed policies instead"

  remediation = <<-EOT
    Replace inline policies with managed policies:

    resource "aws_iam_role_policy_attachment" "example" {
      role       = aws_iam_role.{{.resource_name}}.name
      policy_arn = aws_iam_policy.example.arn
    }

    Benefits of managed policies:
    - Reusable across multiple roles
    - Version controlled
    - Easier to audit and update
  EOT

  tags = ["security", "iam", "best-practice"]
}

rule "IAM_002" {
  name          = "IAM policy should not use wildcard actions"
  description   = "IAM policies should use specific actions instead of wildcards for least privilege"
  severity      = "error"
  resource_type = "aws_iam_policy"

  condition {
    not {
      check {
        attribute = "policy"
        operator  = "contains"
        value     = "\"Action\": \"*\""
      }
    }
  }

  message = "IAM policy '{{.resource_name}}' uses wildcard (*) actions"

  remediation = <<-EOT
    Replace wildcard actions with specific actions:

    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "s3:GetObject",
            "s3:PutObject",
            "s3:ListBucket"
          ],
          "Resource": "arn:aws:s3:::my-bucket/*"
        }
      ]
    }

    Use the principle of least privilege - only grant actions that are needed.
  EOT

  tags = ["security", "iam", "least-privilege"]
}

rule "IAM_003" {
  name          = "IAM policy should not use wildcard resources"
  description   = "IAM policies should specify resources instead of using wildcards"
  severity      = "warning"
  resource_type = "aws_iam_policy"

  condition {
    not {
      check {
        attribute = "policy"
        operator  = "contains"
        value     = "\"Resource\": \"*\""
      }
    }
  }

  message = "IAM policy '{{.resource_name}}' uses wildcard (*) resources"

  remediation = <<-EOT
    Replace wildcard resources with specific ARNs:

    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": ["s3:GetObject"],
          "Resource": [
            "arn:aws:s3:::my-bucket/*",
            "arn:aws:s3:::my-other-bucket/*"
          ]
        }
      ]
    }

    Note: Some actions require Resource: "*" (e.g., s3:ListAllMyBuckets).
    Use condition keys to further restrict access when possible.
  EOT

  tags = ["security", "iam", "least-privilege"]
}
