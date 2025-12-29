variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "alias_name" {
  description = "Name of the Lambda alias"
  type        = string
  default     = "live"
}

variable "tags" {
  description = "Tags for the resources"
  type        = map(string)
  default = {
    ManagedBy = "terraform"
  }
}
