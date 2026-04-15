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

const xyChartDataLayerBreakdownExpected = `{"collapse_by":"avg","color":{"mapping":[{"color":{"type":"color_code","value":"#54B399"},` +
	`"values":["host-a"]}],"mode":"categorical","palette":"default","unassigned":{"type":"color_code","value":"#D3DAE6"}},` +
	`"column":"host.name","format":{"type":"number"}}`

func TestAccResourceDashboardXYChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with XY Chart " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.title", "Sample XY Chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.description", "Test XY chart visualization"),
					// Check axis fields
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.title.value", "Timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.title.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.y.scale", "linear"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.y.title.value", "Count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.y.title.visible", "true"),
					// Check decorations fields
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.fill_opacity", "0.3"),
					// Check fitting
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.type", "none"),
					// Check legend
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.visibility", "visible"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.position", "right"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.inside", "false"),
					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.query.language", "kql"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.query.expression", ""),
					// Check layers
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.type", "line"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.data_source_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.y.#", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("axis"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.ticks", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.grid", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.label_orientation", "angled"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.domain_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.y.ticks", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.y.grid", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.y.label_orientation", "horizontal"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.y.domain_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.secondary_y.scale", "sqrt"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.secondary_y.title.value", "Rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.secondary_y.title.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.secondary_y.ticks", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.secondary_y.grid", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.secondary_y.label_orientation", "vertical"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.secondary_y.domain_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("decorations"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.line_interpolation", "smooth"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.minimum_bar_height", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.fill_opacity", "0.3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.show_value_labels", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.filters.#", "1"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.filters.0.filter_json", regexp.MustCompile(`"field":"log.level"`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("fitting"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.type", "linear"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.dotted", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.end_value", "nearest"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("legend_outside"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.inside", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.position", "right"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.size", "l"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.truncate_after_lines", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.0", "avg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.1", "max"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("legend_inside"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.inside", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.columns", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.alignment", "top_left"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.truncate_after_lines", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.0", "count"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("layers"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.sampling", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.x_json"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_dashboard.test",
						"panels.0.xy_chart_config.layers.0.data_layer.breakdown_by_json",
						xyChartDataLayerBreakdownExpected,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.y.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.y.0.config_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("layers_reference"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.type", "reference_lines"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.data_source_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.sampling", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.thresholds.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.thresholds.0.value_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("layers_reference"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"panels.0.xy_chart_config.title",
					"panels.0.xy_chart_config.axis.x.domain_json",
					"panels.0.xy_chart_config.axis.x.grid",
					"panels.0.xy_chart_config.axis.x.label_orientation",
					"panels.0.xy_chart_config.axis.x.ticks",
					"panels.0.xy_chart_config.axis.y.domain_json",
					"panels.0.xy_chart_config.axis.y.grid",
					"panels.0.xy_chart_config.axis.y.label_orientation",
					"panels.0.xy_chart_config.axis.y.ticks",
					"panels.0.xy_chart_config.decorations.fill_opacity",
					"panels.0.xy_chart_config.decorations.line_interpolation",
					"panels.0.xy_chart_config.decorations.point_visibility",
					"panels.0.xy_chart_config.decorations.show_current_time_marker",
					"panels.0.xy_chart_config.decorations.show_end_zones",
					"panels.0.xy_chart_config.legend.truncate_after_lines",
					"panels.0.xy_chart_config.layers.0.data_layer.data_source_json",
					"panels.0.xy_chart_config.layers.0.data_layer.y.0.config_json",
					"panels.0.xy_chart_config.layers.1.reference_line_layer.data_source_json",
					"panels.0.xy_chart_config.layers.1.reference_line_layer.thresholds.0.value_json",
				},
			},
		},
	})
}
