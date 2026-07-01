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

package aiopschangepointchart_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/aiopschangepointchart"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccResourceDashboardAiopsChangePointChart tests the aiops_change_point_chart panel type.
func TestAccResourceDashboardAiopsChangePointChart(t *testing.T) {
	dashboardTitle := "Test Dashboard AIOps Change Point Chart " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, aiopschangepointchart.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Required-only create.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "aiops_change_point_chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.data_view_id", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.metric_field", "system.cpu.total.pct"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.partitions"),
				),
			},
			// Plan stability.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			// Import.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Kibana returns server-side defaults for these optional enum/float fields even when the
				// practitioner omits them. The create+read path null-preserves them (REQ-009), while import
				// (no prior state) populates them from the API — so they legitimately differ and must be
				// excluded from ImportStateVerify on the required-only import step.
				ImportStateVerifyIgnore: []string{
					"panels.0.aiops_change_point_chart_config.aggregation_function",
					"panels.0.aiops_change_point_chart_config.split_field",
					"panels.0.aiops_change_point_chart_config.partitions",
					"panels.0.aiops_change_point_chart_config.max_series_to_plot",
					"panels.0.aiops_change_point_chart_config.view_type",
				},
			},
			// All optional fields.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_optional"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "aiops_change_point_chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.data_view_id", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.metric_field", "system.cpu.total.pct"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.aggregation_function", "avg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.split_field", "host.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.partitions.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.max_series_to_plot", "6"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config.view_type", "charts"),
				),
			},
			// Plan stability after all-optional create.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_optional"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}

// TestAccResourceDashboardAiopsChangePointChartInvalidConfig covers plan-time validation:
// invalid aggregation_function, invalid view_type, config_json conflict, and wrong panel type.
func TestAccResourceDashboardAiopsChangePointChartInvalidConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Invalid aggregation_function is rejected by OneOf.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_aggregation_function"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`aggregation_function`),
			},
			// Invalid view_type is rejected by OneOf.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_view_type"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`view_type`),
			},
			// config_json on an aiops_change_point_chart panel is rejected.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_config_json"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
			// aiops_change_point_chart_config on a non-aiops panel is rejected.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_panel_type"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
		},
	})
}

// TestAccResourceDashboardAiopsMultiPanel verifies a single dashboard containing all three AIOps
// panel types: each panel's own config block is populated while sibling *_config blocks stay null
// (sibling mutual exclusion), and a subsequent plan shows no changes.
func TestAccResourceDashboardAiopsMultiPanel(t *testing.T) {
	dashboardTitle := "Test Dashboard AIOps Multi Panel " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, aiopschangepointchart.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_three"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "3"),
					// Panel 0: log rate analysis — own block set, siblings null.
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "aiops_log_rate_analysis"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_log_rate_analysis_config.data_view_id", "logs-*"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_change_point_chart_config"),
					// Panel 1: pattern analysis — own block set, siblings null.
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.type", "aiops_pattern_analysis"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.aiops_pattern_analysis_config.data_view_id", "logs-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.aiops_pattern_analysis_config.field_name", "message"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.aiops_log_rate_analysis_config"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.aiops_change_point_chart_config"),
					// Panel 2: change point chart — own block set, siblings null.
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.2.type", "aiops_change_point_chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.2.aiops_change_point_chart_config.data_view_id", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.2.aiops_change_point_chart_config.metric_field", "system.cpu.total.pct"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.2.aiops_log_rate_analysis_config"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.2.aiops_pattern_analysis_config"),
				),
			},
			// Plan stability — no changes after multi-panel create.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_three"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}
