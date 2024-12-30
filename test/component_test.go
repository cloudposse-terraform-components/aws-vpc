package test

import (
	"fmt"
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	awsRegion := "us-east-2"

	fixture := helper.NewFixture(t, "../", awsRegion, "test/fixtures")

	defer fixture.TearDown()
	fixture.SetUp(&atmos.Options{})

	fixture.Suite("default", func(t *testing.T, suite *helper.Suite) {
		t.Parallel()
		suite.Test(t, "two-private-subnets", func(t *testing.T, atm *helper.Atmos) {
			t.Parallel()
			inputs := map[string]interface{}{
				"name":                    "vpc-terraform",
				"availability_zones":      []string{"a", "b"},
				"public_subnets_enabled":  false,
				"nat_gateway_enabled":     false,
				"nat_instance_enabled":    false,
				"subnet_type_tag_key":     "eg.cptest.co/subnet/type",
				"max_subnet_count":        3,
				"vpc_flow_logs_enabled":   false,
				"ipv4_primary_cidr_block": "172.16.0.0/16",
			}

			defer atm.GetAndDestroy("vpc/private", "default-test", inputs)
			component := atm.GetAndDeploy("vpc/private", "default-test", inputs)

			vpcId := atm.Output(component, "vpc_id")
			require.True(t, strings.HasPrefix(vpcId, "vpc-"))

			vpc := aws.GetVpcById(t, vpcId, awsRegion)

			assert.Equal(t, vpc.Name, fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", component.RandomIdentifier))
			assert.Equal(t, *vpc.CidrAssociations[0], "172.16.0.0/16")
			assert.Equal(t, *vpc.CidrBlock, "172.16.0.0/16")
			assert.Nil(t, vpc.Ipv6CidrAssociations)
			assert.Equal(t, vpc.Tags["Environment"], "ue2")
			assert.Equal(t, vpc.Tags["Namespace"], "eg")
			assert.Equal(t, vpc.Tags["Stage"], "test")
			assert.Equal(t, vpc.Tags["Tenant"], "default")

			subnets := vpc.Subnets
			require.Equal(t, 2, len(subnets))

			public_subnet_ids := atm.OutputList(component, "public_subnet_ids")
			assert.Empty(t, public_subnet_ids)

			public_subnet_cidrs := atm.OutputList(component, "public_subnet_cidrs")
			assert.Empty(t, public_subnet_cidrs)

			private_subnet_ids := atm.OutputList(component, "private_subnet_ids")
			assert.Equal(t, 2, len(private_subnet_ids))

			assert.Contains(t, private_subnet_ids, subnets[0].Id)
			assert.Contains(t, private_subnet_ids, subnets[1].Id)

			assert.False(t, aws.IsPublicSubnet(t, subnets[0].Id, awsRegion))
			assert.False(t, aws.IsPublicSubnet(t, subnets[1].Id, awsRegion))

			nats, err := GetNatsByVpcIdE(t, vpcId, awsRegion)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(nats))
		})

		suite.Test(t, "public-private-subnets", func(t *testing.T, atm *helper.Atmos) {
			t.Parallel()
			inputs := map[string]interface{}{
				"name":                    "vpc-terraform",
				"availability_zones":      []string{"b", "c"},
				"public_subnets_enabled":  true,
				"nat_gateway_enabled":     true,
				"nat_instance_enabled":    false,
				"subnet_type_tag_key":     "eg.cptest.co/subnet/type",
				"max_nats":                1,
				"max_subnet_count":        3,
				"vpc_flow_logs_enabled":   false,
				"ipv4_primary_cidr_block": "172.16.0.0/16",
			}

			defer atm.GetAndDestroy("vpc/public", "default-test", inputs)
			component := atm.GetAndDeploy("vpc/public", "default-test", inputs)

			vpcId := atm.Output(component, "vpc_id")
			require.True(t, strings.HasPrefix(vpcId, "vpc-"))

			vpc := aws.GetVpcById(t, vpcId, awsRegion)

			assert.Equal(t, vpc.Name, fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", component.RandomIdentifier))
			assert.Equal(t, *vpc.CidrAssociations[0], "172.16.0.0/16")
			assert.Equal(t, *vpc.CidrBlock, "172.16.0.0/16")
			assert.Nil(t, vpc.Ipv6CidrAssociations)
			assert.Equal(t, vpc.Tags["Environment"], "ue2")
			assert.Equal(t, vpc.Tags["Namespace"], "eg")
			assert.Equal(t, vpc.Tags["Stage"], "test")
			assert.Equal(t, vpc.Tags["Tenant"], "default")

			subnets := vpc.Subnets
			require.Equal(t, 4, len(subnets))

			public_subnet_ids := atm.OutputList(component, "public_subnet_ids")
			assert.Equal(t, 2, len(public_subnet_ids))

			public_subnet_cidrs := atm.OutputList(component, "public_subnet_cidrs")
			assert.Equal(t, 2, len(public_subnet_cidrs))

			private_subnet_ids := atm.OutputList(component, "private_subnet_ids")
			assert.Equal(t, 2, len(private_subnet_ids))

			private_subnet_cidrs := atm.OutputList(component, "private_subnet_cidrs")
			assert.Equal(t, 2, len(private_subnet_cidrs))

			assert.False(t, aws.IsPublicSubnet(t, private_subnet_ids[0], awsRegion))
			assert.False(t, aws.IsPublicSubnet(t, private_subnet_ids[1], awsRegion))

			assert.True(t, aws.IsPublicSubnet(t, public_subnet_ids[0], awsRegion))
			assert.True(t, aws.IsPublicSubnet(t, public_subnet_ids[1], awsRegion))

			nats, err := GetNatsByVpcIdE(t, vpcId, awsRegion)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(nats))
		})

	})
}

func GetNatsByVpcIdE(t *testing.T, vpcId string, awsRegion string) ([]*ec2.NatGateway, error) {
	client, err := aws.NewEc2ClientE(t, awsRegion)
	if err != nil {
		return nil, err
	}

	filter := ec2.Filter{Name: awssdk.String("vpc-id"), Values: []*string{&vpcId}}
	response, err := client.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{Filter: []*ec2.Filter{&filter}})
	if err != nil {
		return nil, err
	}
	return response.NatGateways, nil
}
