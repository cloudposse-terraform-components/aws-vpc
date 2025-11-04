---
tags:
  - component/vpc
  - layer/network
  - provider/aws
  - nat-gateway
  - subnets
  - vpc-flow-logs
  - vpc-endpoints
  - cost-optimization
---

# Component: `vpc`

This component is responsible for provisioning a VPC and corresponding Subnets with advanced configuration capabilities.

**Key Features:**
- Independent control over public and private subnet counts per Availability Zone
- Flexible NAT Gateway placement (index-based or name-based)
- Named subnets with different naming schemes for public vs private
- Cost optimization through strategic NAT Gateway placement
- VPC Flow Logs support for auditing and compliance
- VPC Endpoints for AWS services (S3, DynamoDB, and interface endpoints)
- AWS Shield Advanced protection for NAT Gateway EIPs (optional)

**What's New in v3.0.1:**
- Uses `terraform-aws-dynamic-subnets` v3.0.1 with enhanced subnet configuration
- Separate public/private subnet counts and names per AZ
- Precise NAT Gateway placement control for cost optimization
- NAT Gateway IDs exposed in subnet stats outputs
- Requires AWS Provider v5.0+
- Fixes critical bug in NAT routing when `max_nats < num_azs`
## Usage

**Stack Level**: Regional

## Basic Configuration

Here's a basic example using legacy configuration (fully backward compatible):

```yaml
# catalog/vpc/defaults
components:
  terraform:
    vpc/defaults:
      metadata:
        type: abstract
        component: vpc
      settings:
        spacelift:
          workspace_enabled: true
      vars:
        enabled: true
        name: vpc
        availability_zones:
          - "a"
          - "b"
          - "c"
        nat_gateway_enabled: true
        nat_instance_enabled: false
        max_subnet_count: 3
        vpc_flow_logs_enabled: true
        vpc_flow_logs_bucket_environment_name: <environment>
        vpc_flow_logs_bucket_stage_name: audit
        vpc_flow_logs_traffic_type: "ALL"
        subnet_type_tag_key: "example.net/subnet/type"
        # Legacy subnet configuration (still supported)
        subnets_per_az_count: 1
        subnets_per_az_names: ["common"]
```

```yaml
# stacks/ue2-dev.yaml
import:
  - catalog/vpc

components:
  terraform:
    vpc:
      metadata:
        component: vpc
        inherits:
          - vpc/defaults
      vars:
        ipv4_primary_cidr_block: "10.111.0.0/18"
```

## Cost-Optimized NAT Configuration

Reduce NAT Gateway costs by placing NAT Gateways in only one public subnet per AZ:

```yaml
components:
  terraform:
    vpc:
      vars:
        # Create 2 public subnets per AZ
        public_subnets_per_az_count: 2
        public_subnets_per_az_names: ["loadbalancer", "web"]

        # Create 3 private subnets per AZ
        private_subnets_per_az_count: 3
        private_subnets_per_az_names: ["app", "database", "cache"]

        # Place NAT Gateway ONLY in the first public subnet (index 0)
        # This saves ~67% on NAT Gateway costs compared to NAT in all public subnets
        nat_gateway_public_subnet_indices: [0]
```

**Cost Savings Example (3 AZs, us-east-1):**
- Without optimization: 6 NAT Gateways (2 per AZ) = ~$270/month
- With optimization: 3 NAT Gateways (1 per AZ) = ~$135/month
- **Monthly Savings: ~$135 (~$1,620/year)**

## Named NAT Gateway Placement

Place NAT Gateways by subnet name instead of index:

```yaml
components:
  terraform:
    vpc:
      vars:
        public_subnets_per_az_names: ["loadbalancer", "web"]
        private_subnets_per_az_names: ["app", "database"]

        # Place NAT Gateway only in "loadbalancer" subnet
        nat_gateway_public_subnet_names: ["loadbalancer"]
```

## High-Availability NAT Configuration

For production environments requiring redundancy:

```yaml
components:
  terraform:
    vpc:
      vars:
        public_subnets_per_az_count: 2
        nat_gateway_public_subnet_indices: [0, 1]  # NAT in both public subnets per AZ
```

## Separate Public/Private Subnet Architecture

Different subnet counts and names for public vs private:

```yaml
components:
  terraform:
    vpc:
      vars:
        # 2 public subnets per AZ for load balancers and public services
        public_subnets_per_az_count: 2
        public_subnets_per_az_names: ["alb", "nat"]

        # 4 private subnets per AZ for different application tiers
        private_subnets_per_az_count: 4
        private_subnets_per_az_names: ["web", "app", "data", "cache"]

        # NAT Gateway in "nat" subnet
        nat_gateway_public_subnet_names: ["nat"]
```

## VPC Endpoints Configuration

Add VPC Endpoints for AWS services to reduce data transfer costs and improve security:

```yaml
components:
  terraform:
    vpc:
      vars:
        # Gateway endpoints (no hourly charges)
        gateway_vpc_endpoints:
          - "s3"
          - "dynamodb"

        # Interface endpoints (hourly charges apply)
        interface_vpc_endpoints:
          - "ec2"
          - "ecr.api"
          - "ecr.dkr"
          - "logs"
          - "secretsmanager"
```

## Complete Production Example

```yaml
components:
  terraform:
    vpc:
      vars:
        enabled: true
        name: vpc
        ipv4_primary_cidr_block: "10.0.0.0/16"

        availability_zones:
          - "a"
          - "b"
          - "c"

        # Public subnets for ALB and NAT
        public_subnets_per_az_count: 2
        public_subnets_per_az_names: ["loadbalancer", "nat"]

        # Private subnets for different tiers
        private_subnets_per_az_count: 3
        private_subnets_per_az_names: ["app", "database", "cache"]

        # Cost-optimized NAT placement
        nat_gateway_enabled: true
        nat_gateway_public_subnet_names: ["nat"]

        # VPC Flow Logs
        vpc_flow_logs_enabled: true
        vpc_flow_logs_bucket_environment_name: mgmt
        vpc_flow_logs_bucket_stage_name: audit
        vpc_flow_logs_traffic_type: "ALL"

        # VPC Endpoints
        gateway_vpc_endpoints:
          - "s3"
          - "dynamodb"
        interface_vpc_endpoints:
          - "ecr.api"
          - "ecr.dkr"
          - "logs"

        subnet_type_tag_key: "example.net/subnet/type"
```


<!-- markdownlint-disable -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 5.0.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | >= 5.0.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_endpoint_security_groups"></a> [endpoint\_security\_groups](#module\_endpoint\_security\_groups) | cloudposse/security-group/aws | 2.2.0 |
| <a name="module_iam_roles"></a> [iam\_roles](#module\_iam\_roles) | ../account-map/modules/iam-roles | n/a |
| <a name="module_subnets"></a> [subnets](#module\_subnets) | cloudposse/dynamic-subnets/aws | 3.0.1 |
| <a name="module_this"></a> [this](#module\_this) | cloudposse/label/null | 0.25.0 |
| <a name="module_utils"></a> [utils](#module\_utils) | cloudposse/utils/aws | 1.4.0 |
| <a name="module_vpc"></a> [vpc](#module\_vpc) | cloudposse/vpc/aws | 3.0.0 |
| <a name="module_vpc_endpoints"></a> [vpc\_endpoints](#module\_vpc\_endpoints) | cloudposse/vpc/aws//modules/vpc-endpoints | 3.0.0 |
| <a name="module_vpc_flow_logs_bucket"></a> [vpc\_flow\_logs\_bucket](#module\_vpc\_flow\_logs\_bucket) | cloudposse/stack-config/yaml//modules/remote-state | 1.8.0 |

## Resources

| Name | Type |
|------|------|
| [aws_flow_log.default](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/flow_log) | resource |
| [aws_shield_protection.nat_eip_shield_protection](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/shield_protection) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_eip.eip](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/eip) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_additional_tag_map"></a> [additional\_tag\_map](#input\_additional\_tag\_map) | Additional key-value pairs to add to each map in `tags_as_list_of_maps`. Not added to `tags` or `id`.<br/>This is for some rare cases where resources want additional configuration of tags<br/>and therefore take a list of maps with tag key, value, and additional configuration. | `map(string)` | `{}` | no |
| <a name="input_assign_generated_ipv6_cidr_block"></a> [assign\_generated\_ipv6\_cidr\_block](#input\_assign\_generated\_ipv6\_cidr\_block) | When `true`, assign AWS generated IPv6 CIDR block to the VPC.  Conflicts with `ipv6_ipam_pool_id`. | `bool` | `false` | no |
| <a name="input_attributes"></a> [attributes](#input\_attributes) | ID element. Additional attributes (e.g. `workers` or `cluster`) to add to `id`,<br/>in the order they appear in the list. New attributes are appended to the<br/>end of the list. The elements of the list are joined by the `delimiter`<br/>and treated as a single ID element. | `list(string)` | `[]` | no |
| <a name="input_availability_zone_ids"></a> [availability\_zone\_ids](#input\_availability\_zone\_ids) | List of Availability Zones IDs where subnets will be created. Overrides `availability_zones`.<br/>Can be the full name, e.g. `use1-az1`, or just the part after the AZ ID region code, e.g. `-az1`,<br/>to allow reusable values across regions. Consider contention for resources and spot pricing in each AZ when selecting.<br/>Useful in some regions when using only some AZs and you want to use the same ones across multiple accounts. | `list(string)` | `[]` | no |
| <a name="input_availability_zones"></a> [availability\_zones](#input\_availability\_zones) | List of Availability Zones (AZs) where subnets will be created. Ignored when `availability_zone_ids` is set.<br/>Can be the full name, e.g. `us-east-1a`, or just the part after the region, e.g. `a` to allow reusable values across regions.<br/>The order of zones in the list ***must be stable*** or else Terraform will continually make changes.<br/>If no AZs are specified, then `max_subnet_count` AZs will be selected in alphabetical order.<br/>If `max_subnet_count > 0` and `length(var.availability_zones) > max_subnet_count`, the list<br/>will be truncated. We recommend setting `availability_zones` and `max_subnet_count` explicitly as constant<br/>(not computed) values for predictability, consistency, and stability. | `list(string)` | `[]` | no |
| <a name="input_context"></a> [context](#input\_context) | Single object for setting entire context at once.<br/>See description of individual variables for details.<br/>Leave string and numeric variables as `null` to use default value.<br/>Individual variable settings (non-null) override settings in context object,<br/>except for attributes, tags, and additional\_tag\_map, which are merged. | `any` | <pre>{<br/>  "additional_tag_map": {},<br/>  "attributes": [],<br/>  "delimiter": null,<br/>  "descriptor_formats": {},<br/>  "enabled": true,<br/>  "environment": null,<br/>  "id_length_limit": null,<br/>  "label_key_case": null,<br/>  "label_order": [],<br/>  "label_value_case": null,<br/>  "labels_as_tags": [<br/>    "unset"<br/>  ],<br/>  "name": null,<br/>  "namespace": null,<br/>  "regex_replace_chars": null,<br/>  "stage": null,<br/>  "tags": {},<br/>  "tenant": null<br/>}</pre> | no |
| <a name="input_delimiter"></a> [delimiter](#input\_delimiter) | Delimiter to be used between ID elements.<br/>Defaults to `-` (hyphen). Set to `""` to use no delimiter at all. | `string` | `null` | no |
| <a name="input_descriptor_formats"></a> [descriptor\_formats](#input\_descriptor\_formats) | Describe additional descriptors to be output in the `descriptors` output map.<br/>Map of maps. Keys are names of descriptors. Values are maps of the form<br/>`{<br/>  format = string<br/>  labels = list(string)<br/>}`<br/>(Type is `any` so the map values can later be enhanced to provide additional options.)<br/>`format` is a Terraform format string to be passed to the `format()` function.<br/>`labels` is a list of labels, in order, to pass to `format()` function.<br/>Label values will be normalized before being passed to `format()` so they will be<br/>identical to how they appear in `id`.<br/>Default is `{}` (`descriptors` output will be empty). | `any` | `{}` | no |
| <a name="input_enabled"></a> [enabled](#input\_enabled) | Set to false to prevent the module from creating any resources | `bool` | `null` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | ID element. Usually used for region e.g. 'uw2', 'us-west-2', OR role 'prod', 'staging', 'dev', 'UAT' | `string` | `null` | no |
| <a name="input_gateway_vpc_endpoints"></a> [gateway\_vpc\_endpoints](#input\_gateway\_vpc\_endpoints) | A list of Gateway VPC Endpoints to provision into the VPC. Only valid values are "dynamodb" and "s3". | `set(string)` | `[]` | no |
| <a name="input_id_length_limit"></a> [id\_length\_limit](#input\_id\_length\_limit) | Limit `id` to this many characters (minimum 6).<br/>Set to `0` for unlimited length.<br/>Set to `null` for keep the existing setting, which defaults to `0`.<br/>Does not affect `id_full`. | `number` | `null` | no |
| <a name="input_interface_vpc_endpoints"></a> [interface\_vpc\_endpoints](#input\_interface\_vpc\_endpoints) | A list of Interface VPC Endpoints to provision into the VPC. | `set(string)` | `[]` | no |
| <a name="input_ipv4_additional_cidr_block_associations"></a> [ipv4\_additional\_cidr\_block\_associations](#input\_ipv4\_additional\_cidr\_block\_associations) | IPv4 CIDR blocks to assign to the VPC.<br/>`ipv4_cidr_block` can be set explicitly, or set to `null` with the CIDR block derived from `ipv4_ipam_pool_id` using `ipv4_netmask_length`.<br/>Map keys must be known at `plan` time, and are only used to track changes. | <pre>map(object({<br/>    ipv4_cidr_block     = string<br/>    ipv4_ipam_pool_id   = string<br/>    ipv4_netmask_length = number<br/>  }))</pre> | `{}` | no |
| <a name="input_ipv4_cidr_block_association_timeouts"></a> [ipv4\_cidr\_block\_association\_timeouts](#input\_ipv4\_cidr\_block\_association\_timeouts) | Timeouts (in `go` duration format) for creating and destroying IPv4 CIDR block associations | <pre>object({<br/>    create = string<br/>    delete = string<br/>  })</pre> | `null` | no |
| <a name="input_ipv4_cidrs"></a> [ipv4\_cidrs](#input\_ipv4\_cidrs) | Lists of CIDRs to assign to subnets. Order of CIDRs in the lists must not change over time.<br/>Lists may contain more CIDRs than needed. | <pre>list(object({<br/>    private = list(string)<br/>    public  = list(string)<br/>  }))</pre> | `[]` | no |
| <a name="input_ipv4_primary_cidr_block"></a> [ipv4\_primary\_cidr\_block](#input\_ipv4\_primary\_cidr\_block) | The primary IPv4 CIDR block for the VPC.<br/>Either `ipv4_primary_cidr_block` or `ipv4_primary_cidr_block_association` must be set, but not both. | `string` | `null` | no |
| <a name="input_ipv4_primary_cidr_block_association"></a> [ipv4\_primary\_cidr\_block\_association](#input\_ipv4\_primary\_cidr\_block\_association) | Configuration of the VPC's primary IPv4 CIDR block via IPAM. Conflicts with `ipv4_primary_cidr_block`.<br/>One of `ipv4_primary_cidr_block` or `ipv4_primary_cidr_block_association` must be set.<br/>Additional CIDR blocks can be set via `ipv4_additional_cidr_block_associations`. | <pre>object({<br/>    ipv4_ipam_pool_id   = string<br/>    ipv4_netmask_length = number<br/>  })</pre> | `null` | no |
| <a name="input_label_key_case"></a> [label\_key\_case](#input\_label\_key\_case) | Controls the letter case of the `tags` keys (label names) for tags generated by this module.<br/>Does not affect keys of tags passed in via the `tags` input.<br/>Possible values: `lower`, `title`, `upper`.<br/>Default value: `title`. | `string` | `null` | no |
| <a name="input_label_order"></a> [label\_order](#input\_label\_order) | The order in which the labels (ID elements) appear in the `id`.<br/>Defaults to ["namespace", "environment", "stage", "name", "attributes"].<br/>You can omit any of the 6 labels ("tenant" is the 6th), but at least one must be present. | `list(string)` | `null` | no |
| <a name="input_label_value_case"></a> [label\_value\_case](#input\_label\_value\_case) | Controls the letter case of ID elements (labels) as included in `id`,<br/>set as tag values, and output by this module individually.<br/>Does not affect values of tags passed in via the `tags` input.<br/>Possible values: `lower`, `title`, `upper` and `none` (no transformation).<br/>Set this to `title` and set `delimiter` to `""` to yield Pascal Case IDs.<br/>Default value: `lower`. | `string` | `null` | no |
| <a name="input_labels_as_tags"></a> [labels\_as\_tags](#input\_labels\_as\_tags) | Set of labels (ID elements) to include as tags in the `tags` output.<br/>Default is to include all labels.<br/>Tags with empty values will not be included in the `tags` output.<br/>Set to `[]` to suppress all generated tags.<br/>**Notes:**<br/>  The value of the `name` tag, if included, will be the `id`, not the `name`.<br/>  Unlike other `null-label` inputs, the initial setting of `labels_as_tags` cannot be<br/>  changed in later chained modules. Attempts to change it will be silently ignored. | `set(string)` | <pre>[<br/>  "default"<br/>]</pre> | no |
| <a name="input_map_public_ip_on_launch"></a> [map\_public\_ip\_on\_launch](#input\_map\_public\_ip\_on\_launch) | Instances launched into a public subnet should be assigned a public IP address | `bool` | `true` | no |
| <a name="input_max_nats"></a> [max\_nats](#input\_max\_nats) | Upper limit on number of NAT Gateways/Instances to create.<br/>Set to 1 or 2 for cost savings at the expense of availability.<br/>Default creates a NAT Gateway in each public subnet. | `number` | `null` | no |
| <a name="input_max_subnet_count"></a> [max\_subnet\_count](#input\_max\_subnet\_count) | Sets the maximum amount of subnets to deploy. 0 will deploy a subnet for every provided availability zone (in `region_availability_zones` variable) within the region | `number` | `0` | no |
| <a name="input_name"></a> [name](#input\_name) | ID element. Usually the component or solution name, e.g. 'app' or 'jenkins'.<br/>This is the only ID element not also included as a `tag`.<br/>The "name" tag is set to the full `id` string. There is no tag with the value of the `name` input. | `string` | `null` | no |
| <a name="input_namespace"></a> [namespace](#input\_namespace) | ID element. Usually an abbreviation of your organization name, e.g. 'eg' or 'cp', to help ensure generated IDs are globally unique | `string` | `null` | no |
| <a name="input_nat_eip_aws_shield_protection_enabled"></a> [nat\_eip\_aws\_shield\_protection\_enabled](#input\_nat\_eip\_aws\_shield\_protection\_enabled) | Enable or disable AWS Shield Advanced protection for NAT EIPs. If set to 'true', a subscription to AWS Shield Advanced must exist in this account. | `bool` | `false` | no |
| <a name="input_nat_gateway_enabled"></a> [nat\_gateway\_enabled](#input\_nat\_gateway\_enabled) | Flag to enable/disable NAT gateways | `bool` | `true` | no |
| <a name="input_nat_gateway_public_subnet_indices"></a> [nat\_gateway\_public\_subnet\_indices](#input\_nat\_gateway\_public\_subnet\_indices) | Indices (0-based) of public subnets where NAT Gateways should be placed.<br/>Use this for index-based NAT Gateway placement (e.g., [0, 1] to place NATs in first 2 public subnets per AZ).<br/>Conflicts with `nat_gateway_public_subnet_names`.<br/>If both are null, NAT Gateways are placed in all public subnets by default. | `list(number)` | `null` | no |
| <a name="input_nat_gateway_public_subnet_names"></a> [nat\_gateway\_public\_subnet\_names](#input\_nat\_gateway\_public\_subnet\_names) | Names of public subnets where NAT Gateways should be placed.<br/>Use this for name-based NAT Gateway placement (e.g., ["loadbalancer"] to place NATs only in "loadbalancer" subnets).<br/>Conflicts with `nat_gateway_public_subnet_indices`.<br/>If both are null, NAT Gateways are placed in all public subnets by default. | `list(string)` | `null` | no |
| <a name="input_nat_instance_ami_id"></a> [nat\_instance\_ami\_id](#input\_nat\_instance\_ami\_id) | A list optionally containing the ID of the AMI to use for the NAT instance.<br/>If the list is empty (the default), the latest official AWS NAT instance AMI<br/>will be used. NOTE: The Official NAT instance AMI is being phased out and<br/>does not support NAT64. Use of a NAT gateway is recommended instead. | `list(string)` | `[]` | no |
| <a name="input_nat_instance_enabled"></a> [nat\_instance\_enabled](#input\_nat\_instance\_enabled) | Flag to enable/disable NAT instances | `bool` | `false` | no |
| <a name="input_nat_instance_type"></a> [nat\_instance\_type](#input\_nat\_instance\_type) | NAT Instance type | `string` | `"t3.micro"` | no |
| <a name="input_private_subnets_per_az_count"></a> [private\_subnets\_per\_az\_count](#input\_private\_subnets\_per\_az\_count) | The number of private subnets to provision per Availability Zone.<br/>If null, defaults to the value of `subnets_per_az_count` for backward compatibility.<br/>Use this to create different numbers of private and public subnets per AZ. | `number` | `null` | no |
| <a name="input_private_subnets_per_az_names"></a> [private\_subnets\_per\_az\_names](#input\_private\_subnets\_per\_az\_names) | The names of private subnets to provision per Availability Zone.<br/>If null, defaults to the value of `subnets_per_az_names` for backward compatibility.<br/>Use this to create different named private subnets than public subnets. | `list(string)` | `null` | no |
| <a name="input_public_subnets_enabled"></a> [public\_subnets\_enabled](#input\_public\_subnets\_enabled) | If false, do not create public subnets.<br/>Since NAT gateways and instances must be created in public subnets, these will also not be created when `false`. | `bool` | `true` | no |
| <a name="input_public_subnets_per_az_count"></a> [public\_subnets\_per\_az\_count](#input\_public\_subnets\_per\_az\_count) | The number of public subnets to provision per Availability Zone.<br/>If null, defaults to the value of `subnets_per_az_count` for backward compatibility.<br/>Use this to create different numbers of public and private subnets per AZ. | `number` | `null` | no |
| <a name="input_public_subnets_per_az_names"></a> [public\_subnets\_per\_az\_names](#input\_public\_subnets\_per\_az\_names) | The names of public subnets to provision per Availability Zone.<br/>If null, defaults to the value of `subnets_per_az_names` for backward compatibility.<br/>Use this to create different named public subnets than private subnets. | `list(string)` | `null` | no |
| <a name="input_regex_replace_chars"></a> [regex\_replace\_chars](#input\_regex\_replace\_chars) | Terraform regular expression (regex) string.<br/>Characters matching the regex will be removed from the ID elements.<br/>If not set, `"/[^a-zA-Z0-9-]/"` is used to remove all characters other than hyphens, letters and digits. | `string` | `null` | no |
| <a name="input_region"></a> [region](#input\_region) | AWS Region | `string` | n/a | yes |
| <a name="input_stage"></a> [stage](#input\_stage) | ID element. Usually used to indicate role, e.g. 'prod', 'staging', 'source', 'build', 'test', 'deploy', 'release' | `string` | `null` | no |
| <a name="input_subnet_type_tag_key"></a> [subnet\_type\_tag\_key](#input\_subnet\_type\_tag\_key) | Key for subnet type tag to provide information about the type of subnets, e.g. `cpco/subnet/type=private` or `cpcp/subnet/type=public` | `string` | n/a | yes |
| <a name="input_subnets_per_az_count"></a> [subnets\_per\_az\_count](#input\_subnets\_per\_az\_count) | The number of subnet of each type (public or private) to provision per Availability Zone. | `number` | `1` | no |
| <a name="input_subnets_per_az_names"></a> [subnets\_per\_az\_names](#input\_subnets\_per\_az\_names) | The subnet names of each type (public or private) to provision per Availability Zone.<br/>This variable is optional.<br/>If a list of names is provided, the list items will be used as keys in the outputs `named_private_subnets_map`, `named_public_subnets_map`,<br/>`named_private_route_table_ids_map` and `named_public_route_table_ids_map` | `list(string)` | <pre>[<br/>  "common"<br/>]</pre> | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags (e.g. `{'BusinessUnit': 'XYZ'}`).<br/>Neither the tag keys nor the tag values will be modified by this module. | `map(string)` | `{}` | no |
| <a name="input_tenant"></a> [tenant](#input\_tenant) | ID element \_(Rarely used, not included by default)\_. A customer identifier, indicating who this instance of a resource is for | `string` | `null` | no |
| <a name="input_vpc_flow_logs_bucket_component_name"></a> [vpc\_flow\_logs\_bucket\_component\_name](#input\_vpc\_flow\_logs\_bucket\_component\_name) | The name of the VPC flow logs bucket component | `string` | `"vpc-flow-logs-bucket"` | no |
| <a name="input_vpc_flow_logs_bucket_environment_name"></a> [vpc\_flow\_logs\_bucket\_environment\_name](#input\_vpc\_flow\_logs\_bucket\_environment\_name) | The name of the environment where the VPC Flow Logs bucket is provisioned | `string` | `""` | no |
| <a name="input_vpc_flow_logs_bucket_stage_name"></a> [vpc\_flow\_logs\_bucket\_stage\_name](#input\_vpc\_flow\_logs\_bucket\_stage\_name) | The stage (account) name where the VPC Flow Logs bucket is provisioned | `string` | `""` | no |
| <a name="input_vpc_flow_logs_bucket_tenant_name"></a> [vpc\_flow\_logs\_bucket\_tenant\_name](#input\_vpc\_flow\_logs\_bucket\_tenant\_name) | The name of the tenant where the VPC Flow Logs bucket is provisioned.<br/><br/>If the `tenant` label is not used, leave this as `null`. | `string` | `null` | no |
| <a name="input_vpc_flow_logs_destination_options_file_format"></a> [vpc\_flow\_logs\_destination\_options\_file\_format](#input\_vpc\_flow\_logs\_destination\_options\_file\_format) | VPC Flow Logs file format | `string` | `"parquet"` | no |
| <a name="input_vpc_flow_logs_destination_options_hive_compatible_partitions"></a> [vpc\_flow\_logs\_destination\_options\_hive\_compatible\_partitions](#input\_vpc\_flow\_logs\_destination\_options\_hive\_compatible\_partitions) | Flag to enable/disable VPC Flow Logs hive compatible partitions | `bool` | `false` | no |
| <a name="input_vpc_flow_logs_destination_options_per_hour_partition"></a> [vpc\_flow\_logs\_destination\_options\_per\_hour\_partition](#input\_vpc\_flow\_logs\_destination\_options\_per\_hour\_partition) | Flag to enable/disable VPC Flow Logs per hour partition | `bool` | `false` | no |
| <a name="input_vpc_flow_logs_enabled"></a> [vpc\_flow\_logs\_enabled](#input\_vpc\_flow\_logs\_enabled) | Enable or disable the VPC Flow Logs | `bool` | `true` | no |
| <a name="input_vpc_flow_logs_log_destination_type"></a> [vpc\_flow\_logs\_log\_destination\_type](#input\_vpc\_flow\_logs\_log\_destination\_type) | The type of the logging destination. Valid values: `cloud-watch-logs`, `s3` | `string` | `"s3"` | no |
| <a name="input_vpc_flow_logs_traffic_type"></a> [vpc\_flow\_logs\_traffic\_type](#input\_vpc\_flow\_logs\_traffic\_type) | The type of traffic to capture. Valid values: `ACCEPT`, `REJECT`, `ALL` | `string` | `"ALL"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_availability_zone_ids"></a> [availability\_zone\_ids](#output\_availability\_zone\_ids) | List of Availability Zones IDs where subnets were created, when available |
| <a name="output_availability_zones"></a> [availability\_zones](#output\_availability\_zones) | List of Availability Zones where subnets were created |
| <a name="output_az_private_route_table_ids_map"></a> [az\_private\_route\_table\_ids\_map](#output\_az\_private\_route\_table\_ids\_map) | Map of AZ names to list of private route table IDs in the AZs |
| <a name="output_az_private_subnets_map"></a> [az\_private\_subnets\_map](#output\_az\_private\_subnets\_map) | Map of AZ names to list of private subnet IDs in the AZs |
| <a name="output_az_public_route_table_ids_map"></a> [az\_public\_route\_table\_ids\_map](#output\_az\_public\_route\_table\_ids\_map) | Map of AZ names to list of public route table IDs in the AZs |
| <a name="output_az_public_subnets_map"></a> [az\_public\_subnets\_map](#output\_az\_public\_subnets\_map) | Map of AZ names to list of public subnet IDs in the AZs |
| <a name="output_flow_log_destination"></a> [flow\_log\_destination](#output\_flow\_log\_destination) | Destination bucket for VPC flow logs |
| <a name="output_flow_log_id"></a> [flow\_log\_id](#output\_flow\_log\_id) | ID of the VPC flow log |
| <a name="output_gateway_vpc_endpoints"></a> [gateway\_vpc\_endpoints](#output\_gateway\_vpc\_endpoints) | Map of Gateway VPC Endpoints in this VPC, keyed by service (e.g. "s3"). |
| <a name="output_igw_id"></a> [igw\_id](#output\_igw\_id) | The ID of the Internet Gateway |
| <a name="output_interface_vpc_endpoints"></a> [interface\_vpc\_endpoints](#output\_interface\_vpc\_endpoints) | Map of Interface VPC Endpoints in this VPC. |
| <a name="output_max_subnet_count"></a> [max\_subnet\_count](#output\_max\_subnet\_count) | Maximum allowed number of subnets before all subnet CIDRs need to be recomputed |
| <a name="output_named_private_subnets_stats_map"></a> [named\_private\_subnets\_stats\_map](#output\_named\_private\_subnets\_stats\_map) | Map of subnet names (specified in `private_subnets_per_az_names` or `subnets_per_az_names` variable) to lists of objects with each object having four items: AZ, private subnet ID, private route table ID, NAT Gateway ID (the NAT Gateway that this private subnet routes to for egress) |
| <a name="output_named_public_subnets_stats_map"></a> [named\_public\_subnets\_stats\_map](#output\_named\_public\_subnets\_stats\_map) | Map of subnet names (specified in `public_subnets_per_az_names` or `subnets_per_az_names` variable) to lists of objects with each object having four items: AZ, public subnet ID, public route table ID, NAT Gateway ID (the NAT Gateway in this public subnet, if any) |
| <a name="output_named_route_tables"></a> [named\_route\_tables](#output\_named\_route\_tables) | Map of route table IDs, keyed by subnets\_per\_az\_names.<br/>If subnets\_per\_az\_names is not set, items are grouped by key 'common' |
| <a name="output_named_subnets"></a> [named\_subnets](#output\_named\_subnets) | Map of subnets IDs, keyed by subnets\_per\_az\_names.<br/>If subnets\_per\_az\_names is not set, items are grouped by key 'common' |
| <a name="output_nat_eip_allocation_ids"></a> [nat\_eip\_allocation\_ids](#output\_nat\_eip\_allocation\_ids) | Elastic IP allocations in use by NAT |
| <a name="output_nat_eip_protections"></a> [nat\_eip\_protections](#output\_nat\_eip\_protections) | List of AWS Shield Advanced Protections for NAT Elastic IPs. |
| <a name="output_nat_gateway_ids"></a> [nat\_gateway\_ids](#output\_nat\_gateway\_ids) | NAT Gateway IDs |
| <a name="output_nat_gateway_public_ips"></a> [nat\_gateway\_public\_ips](#output\_nat\_gateway\_public\_ips) | NAT Gateway public IPs |
| <a name="output_nat_instance_ami_id"></a> [nat\_instance\_ami\_id](#output\_nat\_instance\_ami\_id) | ID of AMI used by NAT instance |
| <a name="output_nat_instance_ids"></a> [nat\_instance\_ids](#output\_nat\_instance\_ids) | NAT Instance IDs |
| <a name="output_nat_ips"></a> [nat\_ips](#output\_nat\_ips) | Elastic IP Addresses in use by NAT |
| <a name="output_private_network_acl_id"></a> [private\_network\_acl\_id](#output\_private\_network\_acl\_id) | ID of the Network ACL created for private subnets |
| <a name="output_private_route_table_ids"></a> [private\_route\_table\_ids](#output\_private\_route\_table\_ids) | Private subnet route table IDs |
| <a name="output_private_subnet_arns"></a> [private\_subnet\_arns](#output\_private\_subnet\_arns) | Private subnet ARNs |
| <a name="output_private_subnet_cidrs"></a> [private\_subnet\_cidrs](#output\_private\_subnet\_cidrs) | Private subnet CIDRs |
| <a name="output_private_subnet_ids"></a> [private\_subnet\_ids](#output\_private\_subnet\_ids) | Private subnet IDs |
| <a name="output_private_subnet_ipv6_cidrs"></a> [private\_subnet\_ipv6\_cidrs](#output\_private\_subnet\_ipv6\_cidrs) | Private subnet IPv6 CIDR blocks |
| <a name="output_public_network_acl_id"></a> [public\_network\_acl\_id](#output\_public\_network\_acl\_id) | ID of the Network ACL created for public subnets |
| <a name="output_public_route_table_ids"></a> [public\_route\_table\_ids](#output\_public\_route\_table\_ids) | Public subnet route table IDs |
| <a name="output_public_subnet_arns"></a> [public\_subnet\_arns](#output\_public\_subnet\_arns) | Public subnet ARNs |
| <a name="output_public_subnet_cidrs"></a> [public\_subnet\_cidrs](#output\_public\_subnet\_cidrs) | Public subnet CIDRs |
| <a name="output_public_subnet_ids"></a> [public\_subnet\_ids](#output\_public\_subnet\_ids) | Public subnet IDs |
| <a name="output_public_subnet_ipv6_cidrs"></a> [public\_subnet\_ipv6\_cidrs](#output\_public\_subnet\_ipv6\_cidrs) | Public subnet IPv6 CIDR blocks |
| <a name="output_route_tables"></a> [route\_tables](#output\_route\_tables) | Route tables info map |
| <a name="output_subnets"></a> [subnets](#output\_subnets) | Subnets info map |
| <a name="output_vpc"></a> [vpc](#output\_vpc) | VPC info map |
| <a name="output_vpc_cidr"></a> [vpc\_cidr](#output\_vpc\_cidr) | VPC CIDR |
| <a name="output_vpc_default_network_acl_id"></a> [vpc\_default\_network\_acl\_id](#output\_vpc\_default\_network\_acl\_id) | The ID of the network ACL created by default on VPC creation |
| <a name="output_vpc_default_security_group_id"></a> [vpc\_default\_security\_group\_id](#output\_vpc\_default\_security\_group\_id) | The ID of the security group created by default on VPC creation |
| <a name="output_vpc_endpoint_dynamodb_id"></a> [vpc\_endpoint\_dynamodb\_id](#output\_vpc\_endpoint\_dynamodb\_id) | ID of the DynamoDB gateway endpoint |
| <a name="output_vpc_endpoint_dynamodb_prefix_list_id"></a> [vpc\_endpoint\_dynamodb\_prefix\_list\_id](#output\_vpc\_endpoint\_dynamodb\_prefix\_list\_id) | Prefix list ID for DynamoDB gateway endpoint |
| <a name="output_vpc_endpoint_interface_security_group_id"></a> [vpc\_endpoint\_interface\_security\_group\_id](#output\_vpc\_endpoint\_interface\_security\_group\_id) | Security group ID for interface VPC endpoints |
| <a name="output_vpc_endpoint_s3_id"></a> [vpc\_endpoint\_s3\_id](#output\_vpc\_endpoint\_s3\_id) | ID of the S3 gateway endpoint |
| <a name="output_vpc_endpoint_s3_prefix_list_id"></a> [vpc\_endpoint\_s3\_prefix\_list\_id](#output\_vpc\_endpoint\_s3\_prefix\_list\_id) | Prefix list ID for S3 gateway endpoint |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | VPC ID |
<!-- markdownlint-restore -->



## References


- [cloudposse-terraform-components](https://github.com/orgs/cloudposse-terraform-components/repositories) - Cloud Posse's upstream component

- [terraform-aws-vpc](https://github.com/cloudposse/terraform-aws-vpc) - CloudPosse VPC Module v3.0.0

- [terraform-aws-dynamic-subnets](https://github.com/cloudposse/terraform-aws-dynamic-subnets) - CloudPosse Dynamic Subnets Module v3.0.1 - Enhanced subnet configuration with separate public/private control

- [terraform-aws-dynamic-subnets v3.0.1 Release](https://github.com/cloudposse/terraform-aws-dynamic-subnets/releases/tag/v3.0.1) - Patch release fixing NAT routing bug when max_nats < num_azs




[<img src="https://cloudposse.com/logo-300x69.svg" height="32" align="right"/>](https://cpco.io/homepage?utm_source=github&utm_medium=readme&utm_campaign=cloudposse-terraform-components/aws-vpc&utm_content=)

