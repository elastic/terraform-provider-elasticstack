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

package dashboard_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardWaffle(t *testing.T) {
	dashboardTitle := "Test Dashboard with Waffle " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "vis"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.title", "Sample Waffle"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.description", "Test waffle visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.query.language", "kql"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.query.expression", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.size", "m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.values.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.values.0", "absolute"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.sampling", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.data_source_json"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.metrics.0.config", regexp.MustCompile(`"operation"\s*:\s*"count"`)),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.value_display.mode"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.value_display.percent_decimals"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complete"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.title", "Complete Waffle"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.description", "Complete waffle visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.sampling", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.size", "s"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.values.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.values.0", "absolute"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.visible", "visible"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.truncate_after_lines", "8"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.value_display.mode", "percentage"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.value_display.percent_decimals", "1"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.metrics.0.config", regexp.MustCompile(`"operation"\s*:\s*"count"`)),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.filters.#", "1"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.filters.0.filter_json", regexp.MustCompile(`"field":"host.os.keyword"`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("esql"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.title", "ESQL Waffle"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.description", "Waffle visualization using ES|QL"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.data_source_json", regexp.MustCompile(`"type"\s*:\s*"esql"`)),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.query.language"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.query.expression"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.metrics.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.esql_metrics.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.esql_metrics.0.column", "c"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.esql_metrics.0.label"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.esql_metrics.0.format_json", regexp.MustCompile(`"type"\s*:\s*"number"`)),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.esql_metrics.0.color.type", "static"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.esql_metrics.0.color.color", "#006BB4"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.esql_group_by.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.legend.size", "m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.waffle_config.sampling", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"panels.0.waffle_config.title",
					"panels.0.waffle_config.description",
					"panels.0.waffle_config.metrics.0.config",
					"panels.0.waffle_config.group_by.0.config",
					"panels.0.waffle_config.data_source_json",
					"panels.0.waffle_config.ignore_global_filters",
					"panels.0.waffle_config.sampling",
					// Kibana may retain legend/value_display through panel updates; import read can
					// diverge from apply state after exercising multiple waffle_config shapes.
					"panels.0.waffle_config.legend.visible",
					"panels.0.waffle_config.value_display",
				},
			},
		},
	})
}
