# Example: Non-compliant Terraform Configuration
# This configuration has security issues that gatekeeper will detect

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# VIOLATION: S3_001 - No encryption configured
# VIOLATION: S3_002 - Versioning not enabled
# VIOLATION: S3_004 - No logging configured
resource "aws_s3_bucket" "insecure_bucket" {
  bucket = "my-insecure-bucket"

  tags = {
    Name = "Insecure Bucket"
  }
}

# VIOLATION: S3_003 - Public access not blocked
resource "aws_s3_bucket_public_access_block" "insecure_bucket" {
  bucket = aws_s3_bucket.insecure_bucket.id

  block_public_acls       = false  # Should be true
  block_public_policy     = false  # Should be true
  ignore_public_acls      = false  # Should be true
  restrict_public_buckets = false  # Should be true
}

# VIOLATION: SG_001 - SSH open to the world
# VIOLATION: SG_003 - Unrestricted ingress
# VIOLATION: SG_004 - Unrestricted egress
resource "aws_security_group" "insecure_sg" {
  name        = "insecure-sg"
  description = "Insecure security group with open ports"
  vpc_id      = aws_vpc.main.id

  # SSH open to the world - BAD!
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH from anywhere"
  }

  # All ports open to the world - VERY BAD!
  ingress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All ports open"
  }

  # Unrestricted egress
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All traffic allowed"
  }

  tags = {
    Name = "insecure-sg"
  }
}

# VIOLATION: SG_002 - RDP open to the world
resource "aws_security_group" "windows_sg" {
  name        = "windows-sg"
  description = "Windows security group"
  vpc_id      = aws_vpc.main.id

  # RDP open to the world - BAD!
  ingress {
    from_port   = 3389
    to_port     = 3389
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "RDP from anywhere"
  }

  egress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS outbound"
  }

  tags = {
    Name = "windows-sg"
  }
}

# VPC without flow logs
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "main-vpc"
  }
}

# VIOLATION: VPC_002 - Default security group has rules
resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.main.id

  # Default SG should not have any rules
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "default-sg"
  }
}

# VIOLATION: IAM_001 - Has inline policy
resource "aws_iam_role" "admin_role" {
  name = "admin-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  # Inline policy - should use managed policy instead
  inline_policy {
    name = "admin-policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect   = "Allow"
          Action   = "*"
          Resource = "*"
        }
      ]
    })
  }
}

# VIOLATION: IAM_002 - Wildcard actions
# VIOLATION: IAM_003 - Wildcard resources
resource "aws_iam_policy" "overly_permissive" {
  name = "overly-permissive-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
      }
    ]
  })
}

output "insecure_bucket_name" {
  value = aws_s3_bucket.insecure_bucket.id
}
