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

package ilm_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue1115 reproduces the race between a destroy and create
// for the same underlying ILM policy when a Terraform resource is renamed
// (foo → bar) without a moved{} block while the name attribute stays the same.
//
// Terraform treats the rename as an independent destroy-foo + create-bar pair.
// With default parallelism both run concurrently. Two failure modes are possible:
//   - Race condition: the DELETE (from destroy-foo) races with the read-after-write
//     GET (from create-bar), causing "Provider produced inconsistent result after apply".
//   - Sequential create-then-destroy: create-bar succeeds, destroy-foo then deletes
//     the same policy, leaving bar's state pointing to a non-existent resource and
//     a non-empty subsequent plan.
//
// Related: https://github.com/elastic/terraform-provider-elasticstack/issues/1115
func TestAccReproduceIssue1115(t *testing.T) {
	policyName := "issue-1115-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceILMDestroy,
		Steps: []resource.TestStep{
			// Step 1: create the ILM policy under resource address .foo
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step1_foo"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.TestCheckResourceAttr(
					"elasticstack_elasticsearch_index_lifecycle.foo", "name", policyName,
				),
			},
			// Step 2: rename the resource to .bar, keeping the same name attribute.
			// Terraform will concurrently destroy .foo and create .bar for the same
			// underlying ILM policy. The plan after this apply should be non-empty
			// because destroy-foo deletes the policy that bar's state references, or
			// the apply itself errors with "inconsistent result after apply" if the
			// race condition fires during create-bar's read-after-write.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("step2_bar"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
