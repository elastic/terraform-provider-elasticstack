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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// REQ-002: an index-only create verifies the target index exists and records
// the resource in state with the computed id, without issuing any settings
// PUT; the next plan is empty.
func TestAccResourceIndexSettings_indexOnlyCreateIssuesNoSettingsCall(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(indexSettingsResourceName, "id", indexSettingsIDRegexpForIndex(indexName)),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "index", indexName),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "mapping_total_fields_limit"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "settings_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// REQ-003: removing the last declared setting sends `null` for the previously
// tracked key only, and the resource remains in state with an empty declared
// subset of settings.
func TestAccResourceIndexSettings_removingLastSettingEmptiesSubset(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("typed"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					checkIndexSettingValue(indexName, "index.mapping.total_fields.limit", "5000"),
				),
			},
			{
				// The only declared setting is removed from the
				// configuration; the update sends the previously tracked
				// key as null, and nothing else is tracked afterwards.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(indexSettingsResourceName, "id", indexSettingsIDRegexpForIndex(indexName)),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "mapping_total_fields_limit"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "settings_json"),
				),
			},
			{
				// Elasticsearch drops the nulled setting from persisted
				// settings; the empty declared subset stays empty and the
				// repeated plan adopts nothing.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}
