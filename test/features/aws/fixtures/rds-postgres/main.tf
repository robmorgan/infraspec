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
  region = "us-west-2"
}

variable "name" {
  description = "The name of the RDS instance"
  type        = string
}

variable "engine" {
  description = "The database engine to use"
  type        = string
  default     = "postgres"
}

variable "engine_version" {
  description = "The engine version to use"
  type        = string
  default     = "13.7"
}

variable "instance_class" {
  description = "The instance type of the RDS instance"
  type        = string
  default     = "db.t3.micro"
}

variable "allocated_storage" {
  description = "The amount of allocated storage in gigabytes"
  type        = number
  default     = 20
}

variable "multi_az" {
  description = "Specifies if the RDS instance is multi-AZ"
  type        = bool
  default     = false
}

variable "storage_encrypted" {
  description = "Specifies whether the DB instance is encrypted"
  type        = bool
  default     = true
}

variable "tags" {
  description = "A map of tags to apply to the resource"
  type        = map(string)
  default     = {}
}

resource "aws_db_instance" "default" {
  identifier          = var.name
  engine              = var.engine
  engine_version      = var.engine_version
  instance_class      = var.instance_class
  allocated_storage   = var.allocated_storage
  multi_az            = var.multi_az
  storage_encrypted   = var.storage_encrypted
  username            = "dbadmin"
  password            = "Password123!"
  skip_final_snapshot = true
  publicly_accessible = false

  tags = merge(
    var.tags,
    {
      Name = var.name
    }
  )
}

output "db_instance_id" {
  description = "The ID of the RDS instance"
  value       = aws_db_instance.default.id
}

output "db_instance_endpoint" {
  description = "The connection endpoint"
  value       = aws_db_instance.default.endpoint
}

output "db_instance_arn" {
  description = "The ARN of the RDS instance"
  value       = aws_db_instance.default.arn
}
