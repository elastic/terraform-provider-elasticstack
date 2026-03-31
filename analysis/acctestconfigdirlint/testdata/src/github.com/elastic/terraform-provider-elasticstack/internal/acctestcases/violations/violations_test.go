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
				Config: `resource "null_resource" "example" {}`, // want `resource.TestStep sets Config without ExternalProviders`
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
				ConfigDirectory: config.TestNameDirectory(), // want `resource.TestStep sets ConfigDirectory to a value other than acctest.NamedTestCaseDirectory`
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

// TestViolation4_InlineConfigNoExternalProviders_ParallelTest exercises violation 1
// inside resource.ParallelTest to ensure the analyzer checks both entry points.
func TestViolation4_InlineConfigNoExternalProviders_ParallelTest(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `resource "null_resource" "other" {}`, // want `resource.TestStep sets Config without ExternalProviders`
			},
		},
	})
}
