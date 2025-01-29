package test

import (
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	"github.com/stretchr/testify/assert"
)

type VpcComponentSuite struct {
	helper.TestSuite
}

func (s *VpcComponentSuite) TestPrivateVPC() {
	const component = "vpc/private"
	const stack = "default-test"

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, nil)
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")

	assert.Equal(s.T(), "172.16.0.0/16", cidrBlock)
}

func (s *VpcComponentSuite) TestPublicVPC() {
	const component = "vpc/public"
	const stack = "default-test"

	defer s.DestroyAtmosComponent(s.T(), component, stack, nil)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, nil)
	cidrBlock := atmos.Output(s.T(), options, "vpc_cidr")
	assert.Equal(s.T(), "172.16.0.0/16", cidrBlock)
}

func (s *VpcComponentSuite) TestEnabledFlag() {
	const component = "vpc/disabled"
	const stack = "default-test"
	s.VerifyEnabledFlag(component, stack, nil)
}

func TestRunVPCSuite(t *testing.T) {
	suite := new(VpcComponentSuite)
	helper.Run(t, suite)
}
