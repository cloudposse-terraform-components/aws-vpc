package test

import (
	"strings"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
)

func TestComponent(t *testing.T) {
	awsRegion := "us-east-2"
	suites := helper.NewTestSuites(t, "../", awsRegion, "test/fixtures")

	defer suites.TearDown(t)
	suites.SetUp(t, &atmos.Options{})

	t.Parallel()
	suites.Test(t, "two-private-subnets", func(t *testing.T) {
		component := suites.CreateAndDeployComponent(t, "vpc", "default-test", &atmos.Options{})
		defer suites.DestroyComponent(t, component, &atmos.Options{})

		options := suites.GetOptions(t, component)
		vpcId := atmos.Output(t, options, "vpc_id")
		require.True(t, strings.HasPrefix(vpcId, "vpc-"))

		vpc := aws.GetVpcById(t, vpcId, awsRegion)

		assert.Equal(t, vpc.Name, "eg-default-ue2-test-vpc-terraform")
		assert.Equal(t, *vpc.CidrAssociations[0], "172.16.0.0/16")
		assert.Equal(t, *vpc.CidrBlock, "172.16.0.0/16")
		assert.Nil(t, vpc.Ipv6CidrAssociations)
		assert.Equal(t, vpc.Tags["Environment"], "ue2")
		assert.Equal(t, vpc.Tags["Namespace"], "eg")
		assert.Equal(t, vpc.Tags["Stage"], "test")
		assert.Equal(t, vpc.Tags["Tenant"], "default")

		subnets := vpc.Subnets
		require.Equal(t, 2, len(subnets))

		public_subnet_ids := atmos.OutputList(t, options, "public_subnet_ids")
		assert.Empty(t, public_subnet_ids)

		public_subnet_cidrs := atmos.OutputList(t, options, "public_subnet_cidrs")
		assert.Empty(t, public_subnet_cidrs)

		private_subnet_ids := atmos.OutputList(t, options, "private_subnet_ids")
		assert.Equal(t, 2, len(private_subnet_ids))

		assert.Contains(t, private_subnet_ids, subnets[0].Id)
		assert.Contains(t, private_subnet_ids, subnets[1].Id)

		assert.False(t, aws.IsPublicSubnet(t, subnets[0].Id, awsRegion))
		assert.False(t, aws.IsPublicSubnet(t, subnets[1].Id, awsRegion))
	})
}
