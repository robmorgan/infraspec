variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "authorization_type" {
  description = "Authorization type for the function URL (NONE or AWS_IAM)"
  type        = string
  default     = "NONE"

  validation {
    condition     = contains(["NONE", "AWS_IAM"], var.authorization_type)
    error_message = "authorization_type must be either NONE or AWS_IAM"
  }
}

variable "cors_allow_origins" {
  description = "CORS allowed origins"
  type        = list(string)
  default     = ["*"]
}

variable "cors_allow_methods" {
  description = "CORS allowed methods"
  type        = list(string)
  default     = ["GET", "POST", "PUT", "DELETE"]
}

variable "tags" {
  description = "Tags for the resources"
  type        = map(string)
  default = {
    ManagedBy = "terraform"
  }
}
