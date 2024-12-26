package test

import (
	"encoding/json"
	"github.com/cloudposse/test-helpers/pkg/atmos"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/jsonc"
	"strings"
	"testing"

	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
)

func asArray(t *testing.T, value string) []string {
	result := make([]string, 0)
	err := json.Unmarshal(jsonc.ToJSON([]byte(value)), &result)
	assert.NoError(t, err)
	return result
}

func TestComponent(t *testing.T) {
	awsRegion := "us-east-2"
	suites := helper.NewTestSuites(t, "../", awsRegion, "test/fixtures")

	defer suites.TearDown(t)
	suites.SetUp(t, &atmos.Options{})

	t.Parallel()
	t.Run("two-private-subnets", func(t *testing.T) {
		component := suites.CreateAndDeployComponent(t, "vpc", "default-test", &atmos.Options{})
		defer suites.DestroyComponent(t, component, &atmos.Options{})

		vpcId := suites.Output(t, component, "vpc_id")
		require.True(t, strings.HasPrefix(vpcId, "vpc-"))

		subnets := aws.GetSubnetsForVpc(t, vpcId, awsRegion)

		require.Equal(t, 2, len(subnets))
		//// Verify if the network that is supposed to be public is really public
		//assert.True(t, aws.IsPublicSubnet(t, publicSubnetId, awsRegion))
		//// Verify if the network that is supposed to be private is really private
		//assert.False(t, aws.IsPublicSubnet(t, privateSubnetId, awsRegion))

		public_subnet_ids := asArray(t, suites.Output(t, component, "public_subnet_ids"))
		assert.Empty(t, public_subnet_ids)

		public_subnet_cidrs := asArray(t, suites.Output(t, component, "public_subnet_cidrs"))
		assert.Empty(t, public_subnet_cidrs)

		//private_subnet_ids := suites.Output(t, component, "private_subnet_ids")
		//assert.Empty(t, private_subnet_ids)

		//private_subnet_cidrs := asArray(t, suites.Output(t, component, "private_subnet_cidrs"))
		//assert.Empty(t, private_subnet_cidrs)
		//
		//vpc_default_network_acl_id := suites.Output(t, component, "vpc_default_network_acl_id")
		//assert.Empty(t, vpc_default_network_acl_id)
		//
		//vpc_default_security_group_id := suites.Output(t, component, "vpc_default_security_group_id")
		//assert.Empty(t, vpc_default_security_group_id)
		//
		//vpc_cidr := suites.Output(t, component, "vpc_cidr")
		//assert.Empty(t, vpc_cidr)
		//
		//private_route_table_ids := asArray(t, suites.Output(t, component, "private_route_table_ids"))
		//assert.Empty(t, private_route_table_ids)
		//
		//public_route_table_ids := asArray(t, suites.Output(t, component, "public_route_table_ids"))
		//assert.Empty(t, public_route_table_ids)
		//
		//nat_gateway_ids := asArray(t, suites.Output(t, component, "nat_gateway_ids"))
		//assert.Empty(t, nat_gateway_ids)
		//
		//nat_instance_ids := asArray(t, suites.Output(t, component, "nat_instance_ids"))
		//assert.Empty(t, nat_instance_ids)
		//
		//nat_gateway_public_ips := asArray(t, suites.Output(t, component, "nat_gateway_public_ips"))
		//assert.Empty(t, nat_gateway_public_ips)
		//
		//max_subnet_count := suites.Output(t, component, "max_subnet_count")
		//assert.Empty(t, max_subnet_count)
		//
		//nat_eip_protections := suites.Output(t, component, "nat_eip_protections")
		//assert.Empty(t, nat_eip_protections)
		//
		//interface_vpc_endpoints := asArray(t, suites.Output(t, component, "interface_vpc_endpoints"))
		//assert.Empty(t, interface_vpc_endpoints)
		//
		//availability_zones := asArray(t, suites.Output(t, component, "availability_zones"))
		//assert.Empty(t, availability_zones)

		//az_private_subnets_map := suites.Output(t, component, "az_private_subnets_map")
		//assert.Empty(t, az_private_subnets_map)
		//
		//az_public_subnets_map := suites.Output(t, component, "az_public_subnets_map")
		//assert.Empty(t, az_public_subnets_map)
	})
}
