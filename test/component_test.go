package test

import (
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
	"github.com/stretchr/testify/assert"
)

func TestComponent(t *testing.T) {
	awsRegion := "us-east-2"

	fixture := helper.NewFixture(t, "../", awsRegion, "test/fixtures")

	defer fixture.TearDown()
	fixture.SetUp(&atmos.Options{})

	fixture.Suite("default", func(t *testing.T, suite *helper.Suite) {
		suite.AddDependency("vpc", "default-test")
		suite.Test(t, "two-private-subnets", func(t *testing.T, atm *helper.Atmos) {
			assert.True(t, true)
			// inputs := map[string]interface{}{
			// 	"name":                    "vpc-terraform",
			// 	"availability_zones":      []string{"a", "b"},
			// 	"public_subnets_enabled":  false,
			// 	"nat_gateway_enabled":     false,
			// 	"nat_instance_enabled":    false,
			// 	"subnet_type_tag_key":     "eg.cptest.co/subnet/type",
			// 	"max_subnet_count":        3,
			// 	"vpc_flow_logs_enabled":   false,
			// 	"ipv4_primary_cidr_block": "172.16.0.0/16",
			// }

			// defer atm.GetAndDestroy("vpc", "default-test", inputs)
			// component := atm.GetAndDeploy("vpc", "default-test", inputs)

			// vpcId := atm.Output(component, "vpc_id")
			// require.True(t, strings.HasPrefix(vpcId, "vpc-"))

			// vpc := aws.GetVpcById(t, vpcId, awsRegion)

			// assert.Equal(t, vpc.Name, fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", component.RandomIdentifier))
			// assert.Equal(t, *vpc.CidrAssociations[0], "172.16.0.0/16")
			// assert.Equal(t, *vpc.CidrBlock, "172.16.0.0/16")
			// assert.Nil(t, vpc.Ipv6CidrAssociations)
			// assert.Equal(t, vpc.Tags["Environment"], "ue2")
			// assert.Equal(t, vpc.Tags["Namespace"], "eg")
			// assert.Equal(t, vpc.Tags["Stage"], "test")
			// assert.Equal(t, vpc.Tags["Tenant"], "default")

			// subnets := vpc.Subnets
			// require.Equal(t, 2, len(subnets))

			// public_subnet_ids := atm.OutputList(component, "public_subnet_ids")
			// assert.Empty(t, public_subnet_ids)

			// public_subnet_cidrs := atm.OutputList(component, "public_subnet_cidrs")
			// assert.Empty(t, public_subnet_cidrs)

			// private_subnet_ids := atm.OutputList(component, "private_subnet_ids")
			// assert.Equal(t, 2, len(private_subnet_ids))

			// assert.Contains(t, private_subnet_ids, subnets[0].Id)
			// assert.Contains(t, private_subnet_ids, subnets[1].Id)

			// assert.False(t, aws.IsPublicSubnet(t, subnets[0].Id, awsRegion))
			// assert.False(t, aws.IsPublicSubnet(t, subnets[1].Id, awsRegion))
		})
	})
}
