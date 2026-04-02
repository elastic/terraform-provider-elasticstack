// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package violations_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestViolation1_InlineConfigWithoutExternalProviders exercises violation 1:
// a step sets Config without ExternalProviders.
func TestViolation1_InlineConfigWithoutExternalProviders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   `resource "null_resource" "example" {}`, // want `resource.TestStep sets Config without ExternalProviders`
			},
		},
	})
}

// TestViolation_InlineConfigDeduplicatesMissingProviderWiring ensures a step with only inline
// Config (no ProtoV6ProviderFactories, no ExternalProviders) reports the inline-config
// diagnostic only, not an additional missing-provider-wiring diagnostic.
func TestViolation_InlineConfigDeduplicatesMissingProviderWiring(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `resource "null_resource" "example" {}`, // want `resource.TestStep sets Config without ExternalProviders`
			},
		},
	})
}

// TestViolation_InlineConfigWithConfigDirectoryWithoutExternal reports inline Config even when
// ConfigDirectory is also set (Config branch must not be skipped in favor of ConfigDirectory).
func TestViolation_InlineConfigWithConfigDirectoryWithoutExternal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   `resource "null_resource" "example" {}`, // want `resource.TestStep sets Config without ExternalProviders`
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
			},
		},
	})
}

// TestViolation2_ConfigDirectoryNotNamedHelper exercises violation 2:
// a step sets ConfigDirectory to something other than acctest.NamedTestCaseDirectory(...).
func TestViolation2_ConfigDirectoryNotNamedHelper(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          config.TestNameDirectory(), // want `resource.TestStep sets ConfigDirectory to a value other than acctest.NamedTestCaseDirectory`
			},
		},
	})
}

// TestViolation3_ExternalProvidersWithConfigDirectory exercises violation 4 (the reachable
// ExternalProviders violation): a step sets both ExternalProviders and ConfigDirectory.
func TestViolation3_ExternalProvidersWithConfigDirectory(t *testing.T) {
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

// TestViolation_ExternalProvidersWithoutConfig exercises violation 3 (scenario 8):
// a step sets ExternalProviders but has neither Config nor ConfigDirectory.
func TestViolation_ExternalProvidersWithoutConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{ // want `sets ExternalProviders without inline Config`
					"aws": {Source: "hashicorp/aws"},
				},
				// No Config, no ConfigDirectory
			},
		},
	})
}

// TestViolation4_InlineConfigNoExternalProviders_ParallelTest exercises violation 1
// inside resource.ParallelTest to ensure the analyzer checks both entry points.
func TestViolation4_InlineConfigNoExternalProviders_ParallelTest(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   `resource "null_resource" "other" {}`, // want `resource.TestStep sets Config without ExternalProviders`
			},
		},
	})
}

// TestViolation_TestCaseProtoV6ProviderFactories reports test-case-level ProtoV6 wiring.
func TestViolation_TestCaseProtoV6ProviderFactories(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers, // want `resource.TestCase sets ProtoV6ProviderFactories`
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
			},
		},
	})
}

// TestViolation_MissingStepProviderWiring reports a step with ConfigDirectory but no provider mode.
func TestViolation_MissingStepProviderWiring(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{ // want `resource.TestStep sets neither ProtoV6ProviderFactories nor ExternalProviders`
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
			},
		},
	})
}

// TestViolation_MixedProtoV6AndExternalProviders reports both wiring modes on one step.
func TestViolation_MixedProtoV6AndExternalProviders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers, // want `resource.TestStep sets both ProtoV6ProviderFactories and ExternalProviders`
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws"},
				},
				Config: `resource "aws_instance" "example" {}`,
			},
		},
	})
}
