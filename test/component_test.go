package test

import (
	"testing"

	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
)

func TestComponent(t *testing.T) {
	suites := helper.NewTestSuites(t, "../", "us-east-2", "test/fixtures")
	suites.Run(t, nil)
}
