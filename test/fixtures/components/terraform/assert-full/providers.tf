provider "aws" {
  region = var.region

  # Profile is deprecated in favor of terraform_role_arn. When profiles are not in use, terraform_profile_name is null.
  profile = module.iam_roles.terraform_profile_name
  default_tags {
    tags = var.default_tags
  }
  dynamic "assume_role" {
    # module.iam_roles.terraform_role_arn may be null, in which case do not assume a role.
    for_each = compact([module.iam_roles.terraform_role_arn])
    content {
      role_arn = assume_role.value
    }
  }
}

module "iam_roles" {
  source  = "cloudposse/components/aws//modules/account-map/modules/iam-roles"
  version = "1.534.0"
  context = module.this.context
}

variable "default_tags" {
  type        = map(string)
  description = "A map of tags to add to every resource"
  default     = {}
}
