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

// REQ-004: a repeated refresh after an external reset of the last tracked
// setting adopts nothing: the first refresh nulls the declared typed
// attribute so the drift shows, apply re-establishes the declared value
// without sending null for any other setting, and later refreshes show an
// empty plan.
func TestAccResourceIndexSettings_repeatedRefreshAfterReset(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("typed_refresh_interval"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					checkIndexSettingValue(indexName, "index.refresh_interval", "10s"),
				),
			},
			{
				// The last tracked setting is reset outside Terraform and an
				// unrelated setting is set out of band: the refresh before
				// apply shows drift, and apply restores the declared value
				// without adopting or nulling any other setting.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("typed_refresh_interval"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				PreConfig: func() {
					updateIndexSettingsOutOfBand(t, indexName, map[string]any{
						"index.refresh_interval":           nil,
						"index.mapping.total_fields.limit": 8000,
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "refresh_interval", "10s"),
					checkIndexSettingValue(indexName, "index.refresh_interval", "10s"),
					checkIndexSettingValue(indexName, "index.mapping.total_fields.limit", "8000"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "mapping_total_fields_limit"),
				),
			},
			{
				// First refresh after the declared value is re-established:
				// nothing is adopted and the plan stays empty.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("typed_refresh_interval"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				PlanOnly:                 true,
			},
			{
				// Repeated refresh: still nothing adopted, still an empty plan.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("typed_refresh_interval"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				PlanOnly:                 true,
			},
		},
	})
}
