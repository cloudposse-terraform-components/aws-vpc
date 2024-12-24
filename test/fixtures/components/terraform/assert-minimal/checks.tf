check "that" {
  assert {
    condition     = startswith(module.vpc.outputs.vpc_id, "vpc-")
    error_message = "Err: VPC ID does not start with 'vpc-'"
  }

  assert {
    condition     = startswith(module.vpc.outputs.vpc_id, "vpc-")
    error_message = "Err: VPC ID does not start with 'vpc-'"
  }
}
