package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	awshelper "github.com/cloudposse/test-helpers/pkg/aws"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestPrivateVPC() {
	const component = "vpc/private"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, nil)
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")

	// Ensure S3 bucket is empty before destroy
	vpcFlowLogsBucketOptions := s.GetAtmosOptions("vpc-flow-logs-bucket", stack, nil)
	bucketName := atmos.Output(s.T(), vpcFlowLogsBucketOptions, "vpc_flow_logs_bucket_id")
	defer aws.EmptyS3Bucket(s.T(), awsRegion, bucketName)

	assert.Equal(s.T(), "172.16.0.0/16", cidrBlock)

	vpcId := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcId, "vpc-"))

	vpc := aws.GetVpcById(s.T(), vpcId, awsRegion)

	assert.Equal(s.T(), vpc.Name, fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", s.Config.RandomIdentifier))
	assert.Equal(s.T(), *vpc.CidrAssociations[0], "172.16.0.0/16")
	assert.Equal(s.T(), *vpc.CidrBlock, "172.16.0.0/16")
	assert.Nil(s.T(), vpc.Ipv6CidrAssociations)
	assert.Equal(s.T(), vpc.Tags["Environment"], "ue2")
	assert.Equal(s.T(), vpc.Tags["Namespace"], "eg")
	assert.Equal(s.T(), vpc.Tags["Stage"], "test")
	assert.Equal(s.T(), vpc.Tags["Tenant"], "default")

	subnets := vpc.Subnets
	require.Equal(s.T(), 2, len(subnets))

	public_subnet_ids := atmos.OutputList(s.T(), options, "public_subnet_ids")
	assert.Empty(s.T(), public_subnet_ids)

	public_subnet_cidrs := atmos.OutputList(s.T(), options, "public_subnet_cidrs")
	assert.Empty(s.T(), public_subnet_cidrs)

	private_subnet_ids := atmos.OutputList(s.T(), options, "private_subnet_ids")
	assert.Equal(s.T(), 2, len(private_subnet_ids))

	assert.Contains(s.T(), private_subnet_ids, subnets[0].Id)
	assert.Contains(s.T(), private_subnet_ids, subnets[1].Id)

	assert.False(s.T(), aws.IsPublicSubnet(s.T(), subnets[0].Id, awsRegion))
	assert.False(s.T(), aws.IsPublicSubnet(s.T(), subnets[1].Id, awsRegion))

	nats, err := awshelper.GetNatGatewaysByVpcIdE(s.T(), context.Background(), vpcId, awsRegion)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, len(nats))

	s.DriftTest(component, stack, nil)
}

func (s *ComponentSuite) TestPublicVPC() {
	const component = "vpc/public"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, nil)
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")

	// Ensure S3 bucket is empty before destroy
	vpcFlowLogsBucketOptions := s.GetAtmosOptions("vpc-flow-logs-bucket", stack, nil)
	bucketName := atmos.Output(s.T(), vpcFlowLogsBucketOptions, "vpc_flow_logs_bucket_id")
	defer aws.EmptyS3Bucket(s.T(), awsRegion, bucketName)

	assert.Equal(s.T(), "172.16.0.0/16", cidrBlock)

	vpcId := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcId, "vpc-"))

	vpc := aws.GetVpcById(s.T(), vpcId, awsRegion)

	assert.Equal(s.T(), vpc.Name, fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", s.Config.RandomIdentifier))
	assert.Equal(s.T(), *vpc.CidrAssociations[0], "172.16.0.0/16")
	assert.Equal(s.T(), *vpc.CidrBlock, "172.16.0.0/16")
	assert.Nil(s.T(), vpc.Ipv6CidrAssociations)
	assert.Equal(s.T(), vpc.Tags["Environment"], "ue2")
	assert.Equal(s.T(), vpc.Tags["Namespace"], "eg")
	assert.Equal(s.T(), vpc.Tags["Stage"], "test")
	assert.Equal(s.T(), vpc.Tags["Tenant"], "default")

	subnets := vpc.Subnets
	require.Equal(s.T(), 4, len(subnets))

	public_subnet_ids := atmos.OutputList(s.T(), options, "public_subnet_ids")
	assert.Equal(s.T(), 2, len(public_subnet_ids))

	public_subnet_cidrs := atmos.OutputList(s.T(), options, "public_subnet_cidrs")
	assert.Equal(s.T(), 2, len(public_subnet_cidrs))

	private_subnet_ids := atmos.OutputList(s.T(), options, "private_subnet_ids")
	assert.Equal(s.T(), 2, len(private_subnet_ids))

	private_subnet_cidrs := atmos.OutputList(s.T(), options, "private_subnet_cidrs")
	assert.Equal(s.T(), 2, len(private_subnet_cidrs))

	assert.False(s.T(), aws.IsPublicSubnet(s.T(), private_subnet_ids[0], awsRegion))
	assert.False(s.T(), aws.IsPublicSubnet(s.T(), private_subnet_ids[1], awsRegion))

	assert.True(s.T(), aws.IsPublicSubnet(s.T(), public_subnet_ids[0], awsRegion))
	assert.True(s.T(), aws.IsPublicSubnet(s.T(), public_subnet_ids[1], awsRegion))

	nats, err := awshelper.GetNatGatewaysByVpcIdE(s.T(), context.Background(), vpcId, awsRegion)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(nats))

	s.DriftTest(component, stack, nil)
}

func (s *ComponentSuite) TestVPCFlowLogs() {
	const component = "vpc/with_flowlogs"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)

	// Ensure S3 bucket is empty before destroy
	vpcFlowLogsBucketOptions := s.GetAtmosOptions("vpc-flow-logs-bucket", stack, nil)
	bucketName := atmos.Output(s.T(), vpcFlowLogsBucketOptions, "vpc_flow_logs_bucket_id")
	defer aws.EmptyS3Bucket(s.T(), awsRegion, bucketName)

	options, _ := s.DeployAtmosComponent(s.T(), component, stack, nil)

	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")
	assert.Equal(s.T(), "172.16.0.0/16", cidrBlock)

	vpcId := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcId, "vpc-"))

	vpc := aws.GetVpcById(s.T(), vpcId, awsRegion)

	assert.Equal(s.T(), vpc.Name, fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", s.Config.RandomIdentifier))
	assert.Equal(s.T(), *vpc.CidrAssociations[0], "172.16.0.0/16")
	assert.Equal(s.T(), *vpc.CidrBlock, "172.16.0.0/16")
	assert.Nil(s.T(), vpc.Ipv6CidrAssociations)
	assert.Equal(s.T(), vpc.Tags["Environment"], "ue2")
	assert.Equal(s.T(), vpc.Tags["Namespace"], "eg")
	assert.Equal(s.T(), vpc.Tags["Stage"], "test")
	assert.Equal(s.T(), vpc.Tags["Tenant"], "default")

	subnets := vpc.Subnets
	require.Equal(s.T(), 1, len(subnets))

	nats, err := awshelper.GetNatGatewaysByVpcIdE(s.T(), context.Background(), vpcId, awsRegion)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(nats))

	flow_log_destinations := atmos.Output(s.T(), options, "flow_log_destination")
	require.NotEmpty(s.T(), flow_log_destinations, "Expected at least one flow log destination but found none")
	require.True(s.T(), strings.HasPrefix(flow_log_destinations, "arn:aws:s3:::eg-default-ue2-test-vpc-flow-logs-bucket"))

	flow_log_ids := atmos.OutputList(s.T(), options, "flow_log_id")
	require.NotEmpty(s.T(), flow_log_ids, "Expected at least one flow log ID but found none")
	require.True(s.T(), strings.HasPrefix(flow_log_ids[0], "fl-"), "Flow log ID does not match expected format")

	s.DriftTest(component, stack, nil)
}

func (s *ComponentSuite) TestEnabledFlag() {
	const component = "vpc/disabled"
	const stack = "default-test"
	s.VerifyEnabledFlag(component, stack, nil)
}

func TestRunVPCSuite(t *testing.T) {
	suite := new(ComponentSuite)

	suite.AddDependency(t, "vpc-flow-logs-bucket", "default-test", nil)
	helper.Run(t, suite)
}
