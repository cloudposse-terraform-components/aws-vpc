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

// Test constants - centralized for easier maintenance
const (
	defaultStack      = "default-test"
	defaultRegion     = "us-east-2"
	expectedCIDR      = "172.16.0.0/16"
	vpcFlowLogsBucket = "vpc-flow-logs-bucket"
)

type ComponentSuite struct {
	helper.TestSuite
}

// Helper function to validate common VPC properties
func (s *ComponentSuite) validateVPCProperties(vpc *aws.Vpc, expectedName string) {
	s.T().Helper() // Mark this as a helper function for better error reporting

	assert.Equal(s.T(), expectedName, vpc.Name, "VPC name mismatch")
	assert.Equal(s.T(), expectedCIDR, *vpc.CidrAssociations[0], "CIDR association mismatch")
	assert.Equal(s.T(), expectedCIDR, *vpc.CidrBlock, "CIDR block mismatch")
	assert.Nil(s.T(), vpc.Ipv6CidrAssociations, "IPv6 CIDR should be nil")

	// Validate tags
	assert.Equal(s.T(), "ue2", vpc.Tags["Environment"], "Environment tag mismatch")
	assert.Equal(s.T(), "eg", vpc.Tags["Namespace"], "Namespace tag mismatch")
	assert.Equal(s.T(), "test", vpc.Tags["Stage"], "Stage tag mismatch")
	assert.Equal(s.T(), "default", vpc.Tags["Tenant"], "Tenant tag mismatch")
}

// Helper function to setup S3 bucket cleanup
func (s *ComponentSuite) setupS3Cleanup(stack, region string) {
	s.T().Helper()

	vpcFlowLogsBucketOptions := s.GetAtmosOptions(vpcFlowLogsBucket, stack, nil)
	bucketName := atmos.Output(s.T(), vpcFlowLogsBucketOptions, "vpc_flow_logs_bucket_id")

	// Clean up S3 bucket before component destroy
	s.T().Cleanup(func() {
		aws.EmptyS3Bucket(s.T(), region, bucketName)
	})
}

// TestPrivateVPC tests a VPC with only private subnets and no NAT Gateways
func (s *ComponentSuite) TestPrivateVPC() {
	const component = "vpc/private"

	// Setup S3 cleanup before component destroy
	s.setupS3Cleanup(defaultStack, defaultRegion)

	defer s.DestroyAtmosComponent(s.T(), component, defaultStack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, defaultStack, nil)

	// Validate VPC CIDR
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")
	assert.Equal(s.T(), expectedCIDR, cidrBlock, "VPC CIDR mismatch")

	// Get VPC ID and validate format
	vpcID := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcID, "vpc-"), "VPC ID should have 'vpc-' prefix")

	// Get VPC details from AWS
	vpc := aws.GetVpcById(s.T(), vpcID, defaultRegion)

	// Validate VPC properties using helper function
	expectedName := fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", s.Config.RandomIdentifier)
	s.validateVPCProperties(vpc, expectedName)

	// Validate subnet counts
	subnets := vpc.Subnets
	require.Equal(s.T(), 2, len(subnets), "Should have 2 private subnets")

	// Validate no public subnets exist
	publicSubnetIDs := atmos.OutputList(s.T(), options, "public_subnet_ids")
	assert.Empty(s.T(), publicSubnetIDs, "Should have no public subnets")

	publicSubnetCIDRs := atmos.OutputList(s.T(), options, "public_subnet_cidrs")
	assert.Empty(s.T(), publicSubnetCIDRs, "Should have no public subnet CIDRs")

	// Validate private subnets
	privateSubnetIDs := atmos.OutputList(s.T(), options, "private_subnet_ids")
	require.Equal(s.T(), 2, len(privateSubnetIDs), "Should have 2 private subnets")

	assert.Contains(s.T(), privateSubnetIDs, subnets[0].Id, "First subnet should be in private subnet IDs")
	assert.Contains(s.T(), privateSubnetIDs, subnets[1].Id, "Second subnet should be in private subnet IDs")

	// Validate subnets are private (no route to IGW)
	assert.False(s.T(), aws.IsPublicSubnet(s.T(), subnets[0].Id, defaultRegion), "First subnet should be private")
	assert.False(s.T(), aws.IsPublicSubnet(s.T(), subnets[1].Id, defaultRegion), "Second subnet should be private")

	// Validate no NAT Gateways exist
	nats, err := awshelper.GetNatGatewaysByVpcIdE(s.T(), context.Background(), vpcID, defaultRegion)
	assert.NoError(s.T(), err, "Should be able to query NAT Gateways")
	assert.Equal(s.T(), 0, len(nats), "Should have no NAT Gateways in private VPC")

	// Validate no drift
	s.DriftTest(component, defaultStack, nil)
}

// TestPublicVPC tests a VPC with both public and private subnets with NAT Gateway
func (s *ComponentSuite) TestPublicVPC() {
	const component = "vpc/public"

	// Setup S3 cleanup before component destroy
	s.setupS3Cleanup(defaultStack, defaultRegion)

	defer s.DestroyAtmosComponent(s.T(), component, defaultStack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, defaultStack, nil)

	// Validate VPC CIDR
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")
	assert.Equal(s.T(), expectedCIDR, cidrBlock, "VPC CIDR mismatch")

	// Get VPC ID and validate format
	vpcID := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcID, "vpc-"), "VPC ID should have 'vpc-' prefix")

	// Get VPC details from AWS
	vpc := aws.GetVpcById(s.T(), vpcID, defaultRegion)

	// Validate VPC properties using helper function
	expectedName := fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", s.Config.RandomIdentifier)
	s.validateVPCProperties(vpc, expectedName)

	// Validate total subnet count
	subnets := vpc.Subnets
	require.Equal(s.T(), 4, len(subnets), "Should have 4 total subnets (2 public + 2 private)")

	// Validate public subnets
	publicSubnetIDs := atmos.OutputList(s.T(), options, "public_subnet_ids")
	assert.Equal(s.T(), 2, len(publicSubnetIDs), "Should have 2 public subnets")

	publicSubnetCIDRs := atmos.OutputList(s.T(), options, "public_subnet_cidrs")
	assert.Equal(s.T(), 2, len(publicSubnetCIDRs), "Should have 2 public subnet CIDRs")

	// Validate private subnets
	privateSubnetIDs := atmos.OutputList(s.T(), options, "private_subnet_ids")
	assert.Equal(s.T(), 2, len(privateSubnetIDs), "Should have 2 private subnets")

	privateSubnetCIDRs := atmos.OutputList(s.T(), options, "private_subnet_cidrs")
	assert.Equal(s.T(), 2, len(privateSubnetCIDRs), "Should have 2 private subnet CIDRs")

	// Validate subnet types (public vs private)
	assert.False(s.T(), aws.IsPublicSubnet(s.T(), privateSubnetIDs[0], defaultRegion), "First private subnet should not be public")
	assert.False(s.T(), aws.IsPublicSubnet(s.T(), privateSubnetIDs[1], defaultRegion), "Second private subnet should not be public")

	assert.True(s.T(), aws.IsPublicSubnet(s.T(), publicSubnetIDs[0], defaultRegion), "First public subnet should be public")
	assert.True(s.T(), aws.IsPublicSubnet(s.T(), publicSubnetIDs[1], defaultRegion), "Second public subnet should be public")

	// Validate NAT Gateway count
	nats, err := awshelper.GetNatGatewaysByVpcIdE(s.T(), context.Background(), vpcID, defaultRegion)
	assert.NoError(s.T(), err, "Should be able to query NAT Gateways")
	assert.Equal(s.T(), 1, len(nats), "Should have 1 NAT Gateway (cost-optimized configuration)")

	// Validate no drift
	s.DriftTest(component, defaultStack, nil)
}

// TestVPCFlowLogs tests VPC Flow Logs configuration and S3 bucket integration
func (s *ComponentSuite) TestVPCFlowLogs() {
	const component = "vpc/with_flowlogs"

	// Setup S3 cleanup before component destroy
	s.setupS3Cleanup(defaultStack, defaultRegion)

	defer s.DestroyAtmosComponent(s.T(), component, defaultStack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, defaultStack, nil)

	// Validate VPC CIDR
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")
	assert.Equal(s.T(), expectedCIDR, cidrBlock, "VPC CIDR mismatch")

	// Get VPC ID and validate format
	vpcID := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcID, "vpc-"), "VPC ID should have 'vpc-' prefix")

	// Get VPC details from AWS
	vpc := aws.GetVpcById(s.T(), vpcID, defaultRegion)

	// Validate VPC properties using helper function
	expectedName := fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", s.Config.RandomIdentifier)
	s.validateVPCProperties(vpc, expectedName)

	// Validate subnet count
	subnets := vpc.Subnets
	require.Equal(s.T(), 1, len(subnets), "Should have 1 subnet")

	// Validate flow log destination
	flowLogDestinations := atmos.Output(s.T(), options, "flow_log_destination")
	require.NotEmpty(s.T(), flowLogDestinations, "Flow log destination should not be empty")
	require.True(s.T(), strings.HasPrefix(flowLogDestinations, "arn:aws:s3:::eg-default-ue2-test-vpc-flow-logs-bucket"),
		"Flow log destination should be S3 bucket ARN")

	// Validate flow log IDs
	flowLogIDs := atmos.OutputList(s.T(), options, "flow_log_id")
	require.NotEmpty(s.T(), flowLogIDs, "Flow log IDs should not be empty")
	require.True(s.T(), strings.HasPrefix(flowLogIDs[0], "fl-"), "Flow log ID should have 'fl-' prefix")

	// Validate no drift
	s.DriftTest(component, defaultStack, nil)
}

// TestVPCWithEndpoints tests VPC Endpoints (Gateway and Interface) configuration
func (s *ComponentSuite) TestVPCWithEndpoints() {
	const component = "vpc/with_endpoints"

	// Setup S3 cleanup before component destroy
	s.setupS3Cleanup(defaultStack, defaultRegion)

	defer s.DestroyAtmosComponent(s.T(), component, defaultStack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, defaultStack, nil)

	// Validate VPC CIDR
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")
	assert.Equal(s.T(), expectedCIDR, cidrBlock, "VPC CIDR mismatch")

	// Get VPC ID and validate format
	vpcID := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcID, "vpc-"), "VPC ID should have 'vpc-' prefix")

	// Test Gateway VPC Endpoints
	s.T().Run("GatewayEndpoints", func(t *testing.T) {
		gatewayEndpoints := atmos.OutputMap(t, options, "gateway_vpc_endpoints")
		assert.NotEmpty(t, gatewayEndpoints, "Gateway VPC endpoints should not be empty")

		// Validate S3 gateway endpoint
		s3EndpointID := atmos.Output(t, options, "vpc_endpoint_s3_id")
		assert.NotEmpty(t, s3EndpointID, "S3 endpoint ID should not be empty")
		assert.True(t, strings.HasPrefix(s3EndpointID, "vpce-"), "S3 endpoint ID should have 'vpce-' prefix")

		s3PrefixListID := atmos.Output(t, options, "vpc_endpoint_s3_prefix_list_id")
		assert.NotEmpty(t, s3PrefixListID, "S3 prefix list ID should not be empty")
		assert.True(t, strings.HasPrefix(s3PrefixListID, "pl-"), "S3 prefix list ID should have 'pl-' prefix")

		// Validate DynamoDB gateway endpoint
		dynamodbEndpointID := atmos.Output(t, options, "vpc_endpoint_dynamodb_id")
		assert.NotEmpty(t, dynamodbEndpointID, "DynamoDB endpoint ID should not be empty")
		assert.True(t, strings.HasPrefix(dynamodbEndpointID, "vpce-"), "DynamoDB endpoint ID should have 'vpce-' prefix")

		dynamodbPrefixListID := atmos.Output(t, options, "vpc_endpoint_dynamodb_prefix_list_id")
		assert.NotEmpty(t, dynamodbPrefixListID, "DynamoDB prefix list ID should not be empty")
		assert.True(t, strings.HasPrefix(dynamodbPrefixListID, "pl-"), "DynamoDB prefix list ID should have 'pl-' prefix")
	})

	// Test Interface VPC Endpoints
	s.T().Run("InterfaceEndpoints", func(t *testing.T) {
		interfaceEndpoints := atmos.OutputMap(t, options, "interface_vpc_endpoints")
		assert.NotEmpty(t, interfaceEndpoints, "Interface VPC endpoints should not be empty")
		assert.Equal(t, 2, len(interfaceEndpoints), "Should have 2 interface endpoints (ec2 and ssm)")

		// Validate interface endpoint security group
		interfaceSecurityGroupID := atmos.Output(t, options, "vpc_endpoint_interface_security_group_id")
		assert.NotEmpty(t, interfaceSecurityGroupID, "Interface endpoint security group ID should not be empty")
		assert.True(t, strings.HasPrefix(interfaceSecurityGroupID, "sg-"), "Security group ID should have 'sg-' prefix")
	})

	// Validate no drift
	s.DriftTest(component, defaultStack, nil)
}

// TestNATPlacementByIndex tests NAT Gateway placement using subnet indices
func (s *ComponentSuite) TestNATPlacementByIndex() {
	const component = "vpc/nat-by-index"

	// Setup S3 cleanup before component destroy
	s.setupS3Cleanup(defaultStack, defaultRegion)

	defer s.DestroyAtmosComponent(s.T(), component, defaultStack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, defaultStack, nil)

	// Get VPC ID
	vpcID := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcID, "vpc-"), "VPC ID should have 'vpc-' prefix")

	// Validate we have 2 public subnets per AZ (2 AZs = 4 total)
	publicSubnetIDs := atmos.OutputList(s.T(), options, "public_subnet_ids")
	assert.Equal(s.T(), 4, len(publicSubnetIDs), "Should have 2 public subnets per AZ across 2 AZs = 4 total")

	// Validate NAT Gateway count - should only have NAT in first public subnet of each AZ
	nats, err := awshelper.GetNatGatewaysByVpcIdE(s.T(), context.Background(), vpcID, defaultRegion)
	assert.NoError(s.T(), err, "Should be able to query NAT Gateways")
	assert.Equal(s.T(), 2, len(nats), "Should have 1 NAT per AZ (placed at index 0) = 2 NATs total")

	// Validate NAT Gateways are in "available" state
	for _, nat := range nats {
		assert.Equal(s.T(), "available", nat.State, "NAT Gateway should be in available state")
	}

	// Validate no drift
	s.DriftTest(component, defaultStack, nil)
}

// TestNATPlacementByName tests NAT Gateway placement using subnet names
func (s *ComponentSuite) TestNATPlacementByName() {
	const component = "vpc/nat-by-name"

	// Setup S3 cleanup before component destroy
	s.setupS3Cleanup(defaultStack, defaultRegion)

	defer s.DestroyAtmosComponent(s.T(), component, defaultStack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, defaultStack, nil)

	// Get VPC ID
	vpcID := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcID, "vpc-"), "VPC ID should have 'vpc-' prefix")

	// Validate we have named subnets
	publicSubnetIDs := atmos.OutputList(s.T(), options, "public_subnet_ids")
	assert.Equal(s.T(), 4, len(publicSubnetIDs), "Should have 2 public subnets per AZ across 2 AZs = 4 total")

	// Validate NAT Gateway count - should only have NAT in "nat" named subnet per AZ
	nats, err := awshelper.GetNatGatewaysByVpcIdE(s.T(), context.Background(), vpcID, defaultRegion)
	assert.NoError(s.T(), err, "Should be able to query NAT Gateways")
	assert.Equal(s.T(), 2, len(nats), "Should have NAT only in 'nat' named subnets = 2 NATs total")

	// Validate named subnets output contains expected names
	namedSubnets := atmos.OutputMap(s.T(), options, "named_subnets")
	assert.Contains(s.T(), namedSubnets, "public", "Should have public named subnets map")
	assert.Contains(s.T(), namedSubnets, "private", "Should have private named subnets map")

	// Validate no drift
	s.DriftTest(component, defaultStack, nil)
}

// TestSeparateSubnetCounts tests separate public and private subnet counts per AZ
func (s *ComponentSuite) TestSeparateSubnetCounts() {
	const component = "vpc/separate-counts"

	// Setup S3 cleanup before component destroy
	s.setupS3Cleanup(defaultStack, defaultRegion)

	defer s.DestroyAtmosComponent(s.T(), component, defaultStack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, defaultStack, nil)

	// Get VPC ID
	vpcID := atmos.Output(s.T(), options, "vpc_id")
	require.True(s.T(), strings.HasPrefix(vpcID, "vpc-"), "VPC ID should have 'vpc-' prefix")

	// Validate public subnet count: 2 public subnets per AZ * 2 AZs = 4 total
	publicSubnetIDs := atmos.OutputList(s.T(), options, "public_subnet_ids")
	assert.Equal(s.T(), 4, len(publicSubnetIDs), "Should have 2 public subnets per AZ across 2 AZs = 4 total")

	// Validate private subnet count: 3 private subnets per AZ * 2 AZs = 6 total
	privateSubnetIDs := atmos.OutputList(s.T(), options, "private_subnet_ids")
	assert.Equal(s.T(), 6, len(privateSubnetIDs), "Should have 3 private subnets per AZ across 2 AZs = 6 total")

	// Get VPC details from AWS
	vpc := aws.GetVpcById(s.T(), vpcID, defaultRegion)

	// Validate total subnet count in VPC
	subnets := vpc.Subnets
	assert.Equal(s.T(), 10, len(subnets), "Should have 10 total subnets (4 public + 6 private)")

	// Validate all public subnets are actually public (have route to IGW)
	for _, subnetID := range publicSubnetIDs {
		assert.True(s.T(), aws.IsPublicSubnet(s.T(), subnetID, defaultRegion),
			"Public subnet %s should have route to Internet Gateway", subnetID)
	}

	// Validate all private subnets are actually private (no route to IGW)
	for _, subnetID := range privateSubnetIDs {
		assert.False(s.T(), aws.IsPublicSubnet(s.T(), subnetID, defaultRegion),
			"Private subnet %s should not have route to Internet Gateway", subnetID)
	}

	// Validate no drift
	s.DriftTest(component, defaultStack, nil)
}

// TestValidationMutualExclusivity tests that the validation check fails when both NAT placement methods are set
func (s *ComponentSuite) TestValidationMutualExclusivity() {
	const component = "vpc/validation-conflict"

	s.T().Log("Testing that terraform plan fails when both NAT placement methods are specified")

	// Get Atmos options for the component
	options := s.GetAtmosOptions(component, defaultStack, nil)

	// Run terraform init (required before plan)
	atmos.Init(s.T(), options)

	// Run terraform plan - this should FAIL due to validation check
	_, err := atmos.PlanE(s.T(), options)

	// Verify that plan failed
	require.Error(s.T(), err, "Terraform plan should fail when both NAT placement methods are specified")

	// Verify error message contains expected validation message
	errorMessage := err.Error()
	s.T().Logf("Validation error (as expected): %s", errorMessage)

	// The error should mention the mutual exclusivity issue
	assert.Contains(s.T(), errorMessage, "Cannot specify both",
		"Error message should mention mutual exclusivity")

	s.T().Log("Validation test passed - mutual exclusivity check is working correctly at plan time")
}

// TestEnabledFlag tests the enabled flag functionality
func (s *ComponentSuite) TestEnabledFlag() {
	const component = "vpc/disabled"
	s.VerifyEnabledFlag(component, defaultStack, nil)
}

// TestRunVPCSuite is the main test suite runner
func TestRunVPCSuite(t *testing.T) {
	suite := new(ComponentSuite)

	// Add dependency on VPC Flow Logs bucket
	suite.AddDependency(t, vpcFlowLogsBucket, defaultStack, nil)

	// Run the test suite
	helper.Run(t, suite)
}
