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
	_ "embed"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//go:embed testdata/sdk_compat/main.tf
var sdkCompatEmbeddedTF string

//go:embed testdata/non_main_compat/compat.tf
var nonMainCompatEmbeddedTF string

var (
	//go:embed testdata/grouped_embed_compat/main.tf
	groupedCompatEmbeddedTF string
)

//go:embed testdata/outer_paren_embed_compat/main.tf
var (
	outerParenCompatEmbeddedTF string
)

var (
	//go:embed testdata/comment_sep_compat/main.tf
	// optional line comment between embed directive and declaration (allowed by go:embed)
	commentSepCompatEmbeddedTF string
)

// TestOrdinaryStep verifies that a step using ConfigDirectory: acctest.NamedTestCaseDirectory(...)
// inside resource.Test is compliant and produces no diagnostic.
func TestOrdinaryStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
			},
		},
	})
}

// TestCompatibilityStep verifies that a step using ExternalProviders + Config referencing
// a package-level //go:embed testdata/.../*.tf variable inside resource.Test is compliant.
func TestCompatibilityStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: sdkCompatEmbeddedTF,
			},
		},
	})
}

// TestCompatibilityStepNonMainFixture verifies non-main .tf fixture names are accepted too.
func TestCompatibilityStepNonMainFixture(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: nonMainCompatEmbeddedTF,
			},
		},
	})
}

// TestCompatibilityStepGroupedVarEmbed verifies //go:embed immediately above a ValueSpec inside var ( ... ).
func TestCompatibilityStepGroupedVarEmbed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: groupedCompatEmbeddedTF,
			},
		},
	})
}

// TestCompatibilityStepOuterEmbedBeforeParen verifies //go:embed above a parenthesized var block.
func TestCompatibilityStepOuterEmbedBeforeParen(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: outerParenCompatEmbeddedTF,
			},
		},
	})
}

// TestCompatibilityStepParenConfig verifies parenthesized Config still resolves to the embedded variable.
func TestCompatibilityStepParenConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: (sdkCompatEmbeddedTF),
			},
		},
	})
}

// TestCompatibilityStepParallel verifies ExternalProviders compatibility wiring under resource.ParallelTest.
func TestCompatibilityStepParallel(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: sdkCompatEmbeddedTF,
			},
		},
	})
}

// TestCompatibilityStepEmbedWithLineCommentBetween verifies //go:embed separated from the var by a // comment.
func TestCompatibilityStepEmbedWithLineCommentBetween(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws", VersionConstraint: "~> 4.0"},
				},
				Config: commentSepCompatEmbeddedTF,
			},
		},
	})
}

// TestImportOnlyStep verifies that a step with neither Config nor ConfigDirectory
// (e.g. import-only or refresh-only) still declares provider wiring via ProtoV6ProviderFactories.
func TestImportOnlyStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				// No Config, no ConfigDirectory – import-only step still needs step-local provider wiring.
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
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				// import-only step
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {Source: "hashicorp/aws"},
				},
				Config: sdkCompatEmbeddedTF,
			},
		},
	})
}
