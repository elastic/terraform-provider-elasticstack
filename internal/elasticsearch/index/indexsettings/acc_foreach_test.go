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

package indexsettings_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// REQ-006: two resource instances via for_each over two concrete index
// names, each independently managing its own index's settings.
//
// Steps avoid the legacy Check/CheckDestroy helpers: terraform-plugin-testing's
// legacy terraform.State shim aborts on managed resources with string instance
// keys ("for_each is not supported"), and the deferred end-of-run state
// retrieval only passes on an empty state, so the final step destroys the
// instances. ConfigStateChecks (tfjson state) and ConfigPlanChecks
// (tfjsonplan) support instance-keyed addresses.
func TestAccResourceIndexSettings_forEachIndependently(t *testing.T) {
	indexNameA := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	indexNameB := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	instanceAddress := func(indexName string) string {
		return fmt.Sprintf("%s[%q]", indexSettingsResourceName, indexName)
	}

	variables := config.Variables{
		"index_names": config.SetVariable(
			config.StringVariable(indexNameA),
			config.StringVariable(indexNameB),
		),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("manages_both"),
				ConfigVariables:          variables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						instanceAddress(indexNameA),
						tfjsonpath.New("index"),
						knownvalue.StringExact(indexNameA),
					),
					statecheck.ExpectKnownValue(
						instanceAddress(indexNameA),
						tfjsonpath.New("number_of_replicas"),
						knownvalue.Int64Exact(2),
					),
					statecheck.ExpectKnownValue(
						instanceAddress(indexNameB),
						tfjsonpath.New("index"),
						knownvalue.StringExact(indexNameB),
					),
					statecheck.ExpectKnownValue(
						instanceAddress(indexNameB),
						tfjsonpath.New("number_of_replicas"),
						knownvalue.Int64Exact(2),
					),
				},
			},
			{
				// Only instance A's index changes out of band; the plan
				// shows drift and the apply repairs instance A without
				// affecting instance B.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("manages_both"),
				ConfigVariables:          variables,
				PreConfig: func() {
					updateIndexSettingsOutOfBand(t, indexNameA, map[string]any{"index.number_of_replicas": 3})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						instanceAddress(indexNameA),
						tfjsonpath.New("number_of_replicas"),
						knownvalue.Int64Exact(2),
					),
					statecheck.ExpectKnownValue(
						instanceAddress(indexNameB),
						tfjsonpath.New("number_of_replicas"),
						knownvalue.Int64Exact(2),
					),
				},
			},
			{
				// Destroy empties the state so the legacy end-of-run state
				// retrieval passes.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("manages_both"),
				ConfigVariables:          variables,
				Destroy:                  true,
			},
		},
	})
}
