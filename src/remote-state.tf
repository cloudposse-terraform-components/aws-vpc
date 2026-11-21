module "vpc_flow_logs_bucket" {
  # Not needed when the bucket ARN is passed in with `vpc_flow_logs_bucket_arn`
  count = local.vpc_flow_logs_enabled && var.vpc_flow_logs_bucket_arn == null ? 1 : 0

  source  = "cloudposse/stack-config/yaml//modules/remote-state"
  version = "2.0.1"

  component   = var.vpc_flow_logs_bucket_component_name
  environment = var.vpc_flow_logs_bucket_environment_name
  stage       = var.vpc_flow_logs_bucket_stage_name
  tenant      = try(coalesce(var.vpc_flow_logs_bucket_tenant_name, module.this.tenant), null)

  context = module.this.context
}
