output "public_subnet_ids" {
  value       = module.subnets.public_subnet_ids
  description = "Public subnet IDs"
}

output "public_subnet_arns" {
  value       = module.subnets.public_subnet_arns
  description = "Public subnet ARNs"
}

output "public_subnet_cidrs" {
  value       = module.subnets.public_subnet_cidrs
  description = "Public subnet CIDRs"
}

output "public_subnet_ipv6_cidrs" {
  value       = module.subnets.public_subnet_ipv6_cidrs
  description = "Public subnet IPv6 CIDR blocks"
}

output "private_subnet_ids" {
  value       = module.subnets.private_subnet_ids
  description = "Private subnet IDs"
}

output "private_subnet_arns" {
  value       = module.subnets.private_subnet_arns
  description = "Private subnet ARNs"
}

output "private_subnet_cidrs" {
  value       = module.subnets.private_subnet_cidrs
  description = "Private subnet CIDRs"
}

output "private_subnet_ipv6_cidrs" {
  value       = module.subnets.private_subnet_ipv6_cidrs
  description = "Private subnet IPv6 CIDR blocks"
}

output "subnets" {
  value = {
    public : {
      ids : module.subnets.public_subnet_ids
      cidr : module.subnets.public_subnet_cidrs
    }
    private : {
      ids : module.subnets.private_subnet_ids
      cidr : module.subnets.private_subnet_cidrs
    }
  }
  description = "Subnets info map"
}

output "named_subnets" {
  value = {
    for scope, obj in {
      public : module.subnets.named_public_subnets_map
      private : module.subnets.named_private_subnets_map
    } :
    scope => obj
  }

  description = <<-EOT
    Map of subnets IDs, keyed by subnets_per_az_names.
    If subnets_per_az_names is not set, items are grouped by key 'common'
    EOT
}

output "vpc_default_network_acl_id" {
  value       = module.vpc.vpc_default_network_acl_id
  description = "The ID of the network ACL created by default on VPC creation"
}

output "vpc_default_security_group_id" {
  value       = module.vpc.vpc_default_security_group_id
  description = "The ID of the security group created by default on VPC creation"
}

output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "VPC ID"
}

output "vpc_cidr" {
  value       = module.vpc.vpc_cidr_block
  description = "VPC CIDR"
}

output "vpc" {
  value = {
    id : module.vpc.vpc_id
    cidr : module.vpc.vpc_cidr_block
    subnet_type_tag_key : var.subnet_type_tag_key
    # subnet_type_tag_value_format : var.subnet_type_tag_value_format
  }
  description = "VPC info map"
}

output "private_route_table_ids" {
  value       = module.subnets.private_route_table_ids
  description = "Private subnet route table IDs"
}

output "public_route_table_ids" {
  value       = module.subnets.public_route_table_ids
  description = "Public subnet route table IDs"
}

output "public_network_acl_id" {
  value       = module.subnets.public_network_acl_id
  description = "ID of the Network ACL created for public subnets"
}

output "private_network_acl_id" {
  value       = module.subnets.private_network_acl_id
  description = "ID of the Network ACL created for private subnets"
}

output "route_tables" {
  value = {
    public : {
      ids : module.subnets.public_route_table_ids
    }
    private : {
      ids : module.subnets.private_route_table_ids
    }
  }
  description = "Route tables info map"
}

output "named_route_tables" {
  value = {
    for scope, obj in {
      public  = module.subnets.named_public_route_table_ids_map
      private = module.subnets.named_private_route_table_ids_map
    } :
    scope => obj
  }

  description = <<-EOT
    Map of route table IDs, keyed by subnets_per_az_names.
    If subnets_per_az_names is not set, items are grouped by key 'common'
  EOT
}

output "nat_gateway_ids" {
  value       = module.subnets.nat_gateway_ids
  description = "NAT Gateway IDs"
}

output "nat_instance_ids" {
  value       = module.subnets.nat_instance_ids
  description = "NAT Instance IDs"
}

output "nat_instance_ami_id" {
  value       = module.subnets.nat_instance_ami_id
  description = "ID of AMI used by NAT instance"
}

output "nat_ips" {
  value       = module.subnets.nat_ips
  description = "Elastic IP Addresses in use by NAT"
}

output "nat_eip_allocation_ids" {
  value       = module.subnets.nat_eip_allocation_ids
  description = "Elastic IP allocations in use by NAT"
}

output "nat_gateway_public_ips" {
  value       = module.subnets.nat_gateway_public_ips
  description = "NAT Gateway public IPs"
}

output "max_subnet_count" {
  value       = local.max_subnet_count
  description = "Maximum allowed number of subnets before all subnet CIDRs need to be recomputed"
}

output "nat_eip_protections" {
  description = "List of AWS Shield Advanced Protections for NAT Elastic IPs."
  value       = aws_shield_protection.nat_eip_shield_protection
}

output "interface_vpc_endpoints" {
  description = "Map of Interface VPC Endpoints in this VPC."
  value       = try(module.vpc_endpoints.interface_vpc_endpoints_map, {})
}

output "gateway_vpc_endpoints" {
  description = "Map of Gateway VPC Endpoints in this VPC, keyed by service (e.g. \"s3\")."
  value       = try(module.vpc_endpoints.gateway_vpc_endpoints_map, {})
}

output "vpc_endpoint_s3_prefix_list_id" {
  description = "Prefix list ID for S3 gateway endpoint"
  value       = try(module.vpc_endpoints.gateway_vpc_endpoints_map["s3"]["prefix_list_id"], null)
}

output "vpc_endpoint_s3_id" {
  description = "ID of the S3 gateway endpoint"
  value       = try(module.vpc_endpoints.gateway_vpc_endpoints_map["s3"]["id"], null)
}

output "vpc_endpoint_dynamodb_prefix_list_id" {
  description = "Prefix list ID for DynamoDB gateway endpoint"
  value       = try(module.vpc_endpoints.gateway_vpc_endpoints_map["dynamodb"]["prefix_list_id"], null)
}

output "vpc_endpoint_dynamodb_id" {
  description = "ID of the DynamoDB gateway endpoint"
  value       = try(module.vpc_endpoints.gateway_vpc_endpoints_map["dynamodb"]["id"], null)
}

output "vpc_endpoint_interface_security_group_id" {
  description = "Security group ID for interface VPC endpoints"
  value       = try(module.endpoint_security_groups[local.interface_endpoint_security_group_key].id, null)
}

output "availability_zones" {
  description = "List of Availability Zones where subnets were created"
  value       = module.subnets.availability_zones
}

output "availability_zone_ids" {
  description = "List of Availability Zones IDs where subnets were created, when available"
  value       = module.subnets.availability_zone_ids
}

output "az_private_subnets_map" {
  description = "Map of AZ names to list of private subnet IDs in the AZs"
  value       = module.subnets.az_private_subnets_map
}

output "az_public_subnets_map" {
  description = "Map of AZ names to list of public subnet IDs in the AZs"
  value       = module.subnets.az_public_subnets_map
}

output "az_private_route_table_ids_map" {
  description = "Map of AZ names to list of private route table IDs in the AZs"
  value       = module.subnets.az_private_route_table_ids_map
}

output "az_public_route_table_ids_map" {
  description = "Map of AZ names to list of public route table IDs in the AZs"
  value       = module.subnets.az_public_route_table_ids_map
}

output "igw_id" {
  description = "The ID of the Internet Gateway"
  value       = module.vpc.igw_id
}

output "named_private_subnets_stats_map" {
  description = "Map of subnet names (specified in `private_subnets_per_az_names` or `subnets_per_az_names` variable) to lists of objects with each object having four items: AZ, private subnet ID, private route table ID, NAT Gateway ID (the NAT Gateway that this private subnet routes to for egress)"
  value       = module.subnets.named_private_subnets_stats_map
}

output "named_public_subnets_stats_map" {
  description = "Map of subnet names (specified in `public_subnets_per_az_names` or `subnets_per_az_names` variable) to lists of objects with each object having four items: AZ, public subnet ID, public route table ID, NAT Gateway ID (the NAT Gateway in this public subnet, if any)"
  value       = module.subnets.named_public_subnets_stats_map
}

output "flow_log_destination" {
  description = "Destination bucket for VPC flow logs"
  value       = local.log_destination
}

output "flow_log_id" {
  description = "ID of the VPC flow log"
  value       = try(aws_flow_log.default[*].id, "")
}
