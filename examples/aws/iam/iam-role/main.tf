terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.72.1"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# IAM Role for EC2 instances
resource "aws_iam_role" "main" {
  name                 = var.role_name
  path                 = var.path
  max_session_duration = var.max_session_duration
  tags                 = var.tags

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
}

# IAM Policy with S3 read access
resource "aws_iam_policy" "main" {
  name        = "${var.role_name}-policy"
  description = "Policy for ${var.role_name}"
  tags        = var.tags

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::*",
          "arn:aws:s3:::*/*"
        ]
      }
    ]
  })
}

# Attach the policy to the role
resource "aws_iam_role_policy_attachment" "main" {
  role       = aws_iam_role.main.name
  policy_arn = aws_iam_policy.main.arn
}

# Instance profile for EC2
resource "aws_iam_instance_profile" "main" {
  name = "${var.role_name}-profile"
  role = aws_iam_role.main.name
  tags = var.tags
}
