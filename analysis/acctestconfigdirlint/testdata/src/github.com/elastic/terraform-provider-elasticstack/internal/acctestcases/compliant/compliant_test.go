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

package compliant_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestOrdinaryStep verifies that a step using ConfigDirectory: acctest.NamedTestCaseDirectory(...)
// inside resource.Test is compliant and produces no diagnostic.
func TestOrdinaryStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
			},
		},
	})
}

// TestOrdinaryStepParallel verifies that the same compliant ordinary step pattern works
// inside resource.ParallelTest.
func TestOrdinaryStepParallel(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
			},
		},
	})
}

// TestCompatibilityStep verifies that a step using ExternalProviders + Config: "..."
// inside resource.Test is compliant and produces no diagnostic.
func TestCompatibilityStep(t *testing.T) {
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

// TestImportOnlyStep verifies that a step with neither Config nor ConfigDirectory
// (e.g. import-only or refresh-only) produces no diagnostic.
func TestImportOnlyStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// No Config, no ConfigDirectory – import-only step is out of scope.
			},
		},
	})
}

// TestMixedCompliantSteps verifies that multiple compliant steps in one test case
// all pass without diagnostics.
func TestMixedCompliantSteps(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
			},
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
			},
			{
				// import-only step
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws"},
				},
				Config: `resource "aws_instance" "example" {}`,
			},
		},
	})
}
