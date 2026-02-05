variables {
  environment = "test"
  cidr_block  = "10.0.0.0/16"
}

run "vpc_dns_enabled" {
  module = "./modules/vpc"

  assert {
    condition     = output.vpc_id != ""
    error_message = "VPC ID must not be empty"
  }

  assert {
    condition     = resource.aws_vpc.main.enable_dns_support == true
    error_message = "VPC must have DNS support enabled"
  }
}

run "idempotent" {
  module = "./modules/vpc"
  state  = "./fixtures/applied.tfstate.json"

  assert {
    condition     = length(changes) == 0
    error_message = "Re-apply must produce no changes"
  }
}
