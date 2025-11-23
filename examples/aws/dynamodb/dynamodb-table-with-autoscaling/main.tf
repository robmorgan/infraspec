provider "aws" {
  region = "us-east-1"
}


terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.72.1"
    }
  }
}

module "dynamodb" {
  source  = "terraform-aws-modules/dynamodb-table/aws"
  version = "4.2.0"

  name           = var.name
  hash_key       = var.hash_key
  billing_mode   = var.billing_mode
  read_capacity  = 5
  write_capacity = 5

  attributes = [
    {
      name = "id"
      type = "S"
    }
  ]

  autoscaling_enabled = true
  autoscaling_read = {
    scale_in_cooldown  = 50
    scale_out_cooldown = 40
    target_value       = 45
    max_capacity       = 10
  }
  autoscaling_write = {
    scale_in_cooldown  = 50
    scale_out_cooldown = 40
    target_value       = 45
    max_capacity       = 10
  }

  tags = var.tags
}
