package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
)

// ExternalProvider holds configuration for an external provider.
type ExternalProvider struct {
	VersionConstraint string
	Source            string
}

// TestCase is a single acceptance test case.
type TestCase struct {
	Steps []TestStep
}

// TestStep is a single step within a TestCase.
type TestStep struct {
	Config            string
	ConfigDirectory   config.TestStepConfigFunc
	ExternalProviders map[string]ExternalProvider
}

// Test runs an acceptance test.
func Test(t *testing.T, c TestCase) {
	_ = t
	_ = c
}

// ParallelTest runs an acceptance test in parallel.
func ParallelTest(t *testing.T, c TestCase) {
	_ = t
	_ = c
}
