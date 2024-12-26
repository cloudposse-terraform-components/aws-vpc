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

		public_subnet_ids := asArray(t, suites.Output(t, component, "public_subnet_ids"))
		assert.Empty(t, public_subnet_ids)

		public_subnet_cidrs := asArray(t, suites.Output(t, component, "public_subnet_cidrs"))
		assert.Empty(t, public_subnet_cidrs)

		//private_subnet_ids := suites.Output(t, component, "private_subnet_ids")
		//assert.Empty(t, private_subnet_ids)
	})
}
