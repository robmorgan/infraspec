variable "name" {
  description = "Name of the DynamoDB table"
  type        = string
}

variable "hash_key" {
  description = "The attribute to use as the hash (partition) key. Must also be defined as an attribute"
  type        = string
  default     = "id"
}

variable "billing_mode" {
  description = "Controls how you are billed for read/write throughput and how you manage capacity. The valid values are PROVISIONED or PAY_PER_REQUEST"
  type        = string
  default     = "PROVISIONED"
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}
