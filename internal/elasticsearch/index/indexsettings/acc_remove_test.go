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
)

// REQ-003: removing a declared setting from the configuration sends `null`
// so Elasticsearch resets the setting to its default; the attribute stops
// being tracked and the narrowed configuration converges with no diff.
func TestAccResourceIndexSettings_removedSettingResets(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("typed_5000"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					checkIndexSettingValue(indexName, "index.mapping.total_fields.limit", "5000"),
				),
			},
			{
				// The declared setting is removed from the configuration;
				// only settings_json remains, so the update payload resets
				// index.mapping.total_fields.limit to null.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "mapping_total_fields_limit"),
					checkIndexSettingValue(indexName, "index.refresh_interval", "10s"),
				),
			},
			{
				// Elasticsearch drops the nulled setting from persisted
				// settings; the next plan shows no diff for the narrowed
				// configuration.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				PlanOnly:                 true,
			},
		},
	})
}
