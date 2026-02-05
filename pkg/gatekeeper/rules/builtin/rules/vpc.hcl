# VPC Security Rules
# Security rules for AWS VPC resources

rule "VPC_001" {
  name          = "VPC should have flow logs enabled"
  description   = "VPC flow logs should be enabled to capture network traffic information"
  severity      = "warning"
  resource_type = "aws_flow_log"

  condition {
    check {
      attribute = "vpc_id"
      operator  = "exists"
    }
  }

  message = "VPC flow log '{{.resource_name}}' configuration found - ensure it covers all VPCs"

  remediation = <<-EOT
    Enable VPC flow logs for your VPC:

    resource "aws_flow_log" "example" {
      iam_role_arn    = aws_iam_role.flow_log.arn
      log_destination = aws_cloudwatch_log_group.flow_log.arn
      traffic_type    = "ALL"
      vpc_id          = aws_vpc.main.id
    }

    Note: Flow logs help with network troubleshooting and security analysis.
  EOT

  tags = ["security", "vpc", "logging", "network"]
}

rule "VPC_002" {
  name          = "Default security group should restrict all traffic"
  description   = "The default security group should not have any rules allowing traffic"
  severity      = "warning"
  resource_type = "aws_default_security_group"

  condition {
    all {
      check {
        attribute = "ingress"
        operator  = "not_exists"
      }
      check {
        attribute = "egress"
        operator  = "not_exists"
      }
    }
  }

  message = "Default security group '{{.resource_name}}' should not have ingress or egress rules"

  remediation = <<-EOT
    Remove all rules from the default security group:

    resource "aws_default_security_group" "default" {
      vpc_id = aws_vpc.main.id

      # No ingress or egress rules - use explicit security groups instead
    }

    Create dedicated security groups for your resources instead of using the default.
  EOT

  tags = ["security", "vpc", "network"]
}
