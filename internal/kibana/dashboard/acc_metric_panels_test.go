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

// TestAccResourceDashboardMetricChartMinimalConfig is a regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/2355.
// It applies a metric_chart_config that deliberately omits all optional attributes
// that have Kibana-side defaults (ignore_global_filters, sampling, query.language,
// and per-metric config_json fields: empty_as_null, color, format.decimals, format.compact).
// If the issue is not fixed the first apply step fails with
// "Provider produced inconsistent result after apply".
// The second plan-only step verifies no drift remains after a clean apply.
func TestAccResourceDashboardMetricChartMinimalConfig(t *testing.T) {
	dashboardTitle := "Test Dashboard Metric Chart Minimal " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "vis"),
				),
			},
			{
				// Same config, plan only — must show no changes after a clean apply.
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceDashboardMetricChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with Metric Chart " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					// Check metric chart config
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.title", "Sample Metric Chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.description", "Test metric chart visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.sampling", "1"),
					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.query.language", "kql"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.query.expression", ""),
					// Check JSON fields are set
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.data_source_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.metrics.0.config_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_breakdown"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Filters"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" with Filters"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					// Check metric chart config with filters
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.title", "Sample Metric Chart with Filters"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.description", "Test metric chart with filters visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.sampling", "1"),
					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.query.language", "kql"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.query.expression", "status:active"),
					// Check JSON fields are set
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.data_source_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.metrics.0.config_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.breakdown_by_json"),
					// Check filters
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.filters.#", "1"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.filters.0.filter_json", regexp.MustCompile(`"field":"event.category"`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secondary_metric"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Secondary Metric"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" with Secondary Metric"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					// Check metric chart config with secondary metric
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.title", "Sample Metric Chart with Secondary Metric"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.description", "Test metric chart with secondary metric"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.data_source_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.metrics.#", "2"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.metrics.0.config_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.metrics.1.config_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_breakdown"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Filters"),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore JSON fields that may have API/defaults differences
				ImportStateVerifyIgnore: []string{
					"panels.0.viz_config.by_value.metric_chart_config.title",
					"panels.0.viz_config.by_value.metric_chart_config.description",
					"panels.0.viz_config.by_value.metric_chart_config.data_source_json",
					"panels.0.viz_config.by_value.metric_chart_config.metrics.0.config_json",
					"panels.0.viz_config.by_value.metric_chart_config.metrics.1.config_json",
					"panels.0.viz_config.by_value.metric_chart_config.breakdown_by_json",
				},
			},
		},
	})
}

func TestAccResourceDashboardMetricChart_lensPresentationCrossCutting(t *testing.T) {
	dashboardTitle := "Test Dashboard Metric presentation " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("cross_cutting"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.time_range.from", "now-30d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.time_range.to", "now-1d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.drilldowns.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.drilldowns.0.discover_drilldown.label", "Open Discover"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.viz_config.by_value.metric_chart_config.drilldowns.0.discover_drilldown.trigger", "on_apply_filter"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("cross_cutting"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}
