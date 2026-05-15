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

package lensdatatable_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardDatatableChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with Datatable " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.title", "Sample Datatable"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.description", "Test datatable visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.styling.density.mode", "compact"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.metrics.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.styling.paging", "10"),
				),
			},
			// Skipping this case until the metric format is correctly described in the API spec
			// and returned by the API.
			//
			// {
			// 	ProtoV6ProviderFactories: acctest.Providers,
			// 	SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(dashboardacctest.MinDashboardAPISupport),
			// 	ConfigDirectory:          acctest.NamedTestCaseDirectory("esql"),
			// 	ConfigVariables: config.Variables{
			// 		"dashboard_title": config.StringVariable(dashboardTitle + " ESQL"),
			// 	},
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" ESQL"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "vis"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.title", "count"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.metrics.#", "2"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.styling.density.mode", "default"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.ignore_global_filters", "false"),
			// 		resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.sampling", "1"),
			// 		resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.data_source_json"),
			// 		resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.metrics.0.config_json"),
			// 		resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.esql.metrics.1.config_json"),
			// 	),
			// },
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"panels.0.vis_config.by_value.datatable_config.no_esql.title",
					"panels.0.vis_config.by_value.datatable_config.no_esql.description",
					"panels.0.vis_config.by_value.datatable_config.no_esql.data_source_json",
					"panels.0.vis_config.by_value.datatable_config.no_esql.metrics.0.config_json",
				},
			},
		},
	})
}

func TestAccResourceDashboardDatatableChart_lensPresentationCrossCutting(t *testing.T) {
	dashboardTitle := "Test Dashboard Datatable presentation " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("cross_cutting"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.time_range.from", "now-30d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.time_range.to", "now-1d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_dashboard.test",
						"panels.0.vis_config.by_value.datatable_config.no_esql.drilldowns.0.url_drilldown.url",
						"https://example.com/{{event.field}}",
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.drilldowns.0.url_drilldown.label", "Open URL"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.datatable_config.no_esql.drilldowns.0.url_drilldown.trigger", "on_click_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("cross_cutting"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}
