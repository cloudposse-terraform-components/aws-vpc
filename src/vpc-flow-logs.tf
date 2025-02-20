# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/flow_log
locals {
  account_name    = lookup(module.this.descriptors, "account_name", module.this.stage)
  log_destination = format("%v/%v/%v/", one(module.vpc_flow_logs_bucket[*].outputs.vpc_flow_logs_bucket_arn), local.account_name, module.this.id)
}

resource "aws_flow_log" "default" {
  count                = local.vpc_flow_logs_enabled ? 1 : 0
  log_destination      = local.log_destination
  log_destination_type = var.vpc_flow_logs_log_destination_type
  traffic_type         = var.vpc_flow_logs_traffic_type
  vpc_id               = module.vpc.vpc_id

  destination_options {
    file_format                = var.vpc_flow_logs_destination_options_file_format
    hive_compatible_partitions = var.vpc_flow_logs_destination_options_hive_compatible_partitions
    per_hour_partition         = var.vpc_flow_logs_destination_options_per_hour_partition
  }

  tags = module.this.tags
}
