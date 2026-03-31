package basic_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestCompliantOrdinaryStep uses ConfigDirectory with acctest.NamedTestCaseDirectory – compliant.
func TestCompliantOrdinaryStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
			},
		},
	})
}

// TestCompliantCompatibilityStep uses ExternalProviders with inline Config – compliant.
func TestCompliantCompatibilityStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: `resource "aws_instance" "example" {}`,
			},
		},
	})
}

// TestCompliantImportOnlyStep has neither Config nor ConfigDirectory – compliant (import-only).
func TestCompliantImportOnlyStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// No Config, no ConfigDirectory – import-only step, not in scope.
			},
		},
	})
}

// TestViolationInlineConfigNoExternalProviders sets Config without ExternalProviders – violation.
func TestViolationInlineConfigNoExternalProviders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `resource "null_resource" "example" {}`, // want `resource.TestStep sets Config without ExternalProviders`
			},
		},
	})
}

// TestViolationConfigDirectoryNotHelper uses ConfigDirectory with a non-helper value – violation.
func TestViolationConfigDirectoryNotHelper(t *testing.T) {
	// Use config.TestNameDirectory() which is not the accepted helper.
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(), // want `resource.TestStep sets ConfigDirectory to a value other than acctest.NamedTestCaseDirectory`
			},
		},
	})
}

// TestViolationExternalProvidersWithConfigDirectory sets both – violation.
func TestViolationExternalProvidersWithConfigDirectory(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{ // want `resource.TestStep sets both ExternalProviders and ConfigDirectory`
					"aws": {Source: "hashicorp/aws"},
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
			},
		},
	})
}

// TestViolationExternalProvidersWithoutConfig sets ExternalProviders but no Config and no ConfigDirectory – not in scope.
// ExternalProviders alone with no Config and no ConfigDir would be non-violating by the "not in scope" rule.
// ExternalProviders + ConfigDirectory = violation (ExternalProvidersWithConfigDirectory takes precedence).
// ExternalProviders + no Config, no ConfigDir = not in scope (neither Config nor ConfigDir).

// TestViolationExternalProvidersNoConfigNoDir has ExternalProviders but no Config or ConfigDirectory – not in scope.
func TestViolationExternalProvidersNoConfigNoDir(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws"},
				},
				// No Config, no ConfigDirectory – not in scope per spec.
			},
		},
	})
}
