# Security Group Rules
# Security rules for AWS Security Groups

rule "SG_001" {
  name          = "No SSH from 0.0.0.0/0"
  description   = "Security groups should not allow SSH (port 22) access from the internet"
  severity      = "error"
  resource_type = "aws_security_group"

  condition {
    not {
      all {
        check {
          attribute = "ingress[*].from_port"
          operator  = "equals"
          value     = 22
        }
        check {
          attribute = "ingress[*].cidr_blocks"
          operator  = "contains"
          value     = "0.0.0.0/0"
        }
      }
    }
  }

  message = "Security group '{{.resource_name}}' allows SSH access from 0.0.0.0/0"

  remediation = <<-EOT
    Restrict SSH access to specific IP ranges:

    ingress {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = ["10.0.0.0/8"]  # Your internal network
    }

    Or use a bastion host or VPN for SSH access.
  EOT

  tags = ["security", "network", "ssh"]
}

rule "SG_002" {
  name          = "No RDP from 0.0.0.0/0"
  description   = "Security groups should not allow RDP (port 3389) access from the internet"
  severity      = "error"
  resource_type = "aws_security_group"

  condition {
    not {
      all {
        check {
          attribute = "ingress[*].from_port"
          operator  = "equals"
          value     = 3389
        }
        check {
          attribute = "ingress[*].cidr_blocks"
          operator  = "contains"
          value     = "0.0.0.0/0"
        }
      }
    }
  }

  message = "Security group '{{.resource_name}}' allows RDP access from 0.0.0.0/0"

  remediation = <<-EOT
    Restrict RDP access to specific IP ranges:

    ingress {
      from_port   = 3389
      to_port     = 3389
      protocol    = "tcp"
      cidr_blocks = ["10.0.0.0/8"]  # Your internal network
    }

    Or use a bastion host or VPN for RDP access.
  EOT

  tags = ["security", "network", "rdp"]
}

rule "SG_003" {
  name          = "No unrestricted ingress"
  description   = "Security groups should not allow all traffic from the internet (0.0.0.0/0 on all ports)"
  severity      = "error"
  resource_type = "aws_security_group"

  condition {
    not {
      all {
        check {
          attribute = "ingress[*].from_port"
          operator  = "equals"
          value     = 0
        }
        check {
          attribute = "ingress[*].to_port"
          operator  = "equals"
          value     = 65535
        }
        check {
          attribute = "ingress[*].cidr_blocks"
          operator  = "contains"
          value     = "0.0.0.0/0"
        }
      }
    }
  }

  message = "Security group '{{.resource_name}}' allows unrestricted ingress from 0.0.0.0/0"

  remediation = <<-EOT
    Restrict ingress to only the ports and protocols you need:

    ingress {
      from_port   = 443
      to_port     = 443
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]  # Only HTTPS
    }

    Never allow all ports (0-65535) from the internet.
  EOT

  tags = ["security", "network", "ingress"]
}

rule "SG_004" {
  name          = "Egress should be restricted"
  description   = "Security groups should have restricted egress rather than allowing all outbound traffic"
  severity      = "warning"
  resource_type = "aws_security_group"

  condition {
    not {
      all {
        check {
          attribute = "egress[*].from_port"
          operator  = "equals"
          value     = 0
        }
        check {
          attribute = "egress[*].to_port"
          operator  = "equals"
          value     = 0
        }
        check {
          attribute = "egress[*].protocol"
          operator  = "equals"
          value     = "-1"
        }
        check {
          attribute = "egress[*].cidr_blocks"
          operator  = "contains"
          value     = "0.0.0.0/0"
        }
      }
    }
  }

  message = "Security group '{{.resource_name}}' allows unrestricted egress to 0.0.0.0/0"

  remediation = <<-EOT
    Restrict egress to only the destinations you need:

    egress {
      from_port   = 443
      to_port     = 443
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]  # HTTPS only
    }

    egress {
      from_port   = 53
      to_port     = 53
      protocol    = "udp"
      cidr_blocks = ["0.0.0.0/0"]  # DNS
    }
  EOT

  tags = ["security", "network", "egress"]
}
