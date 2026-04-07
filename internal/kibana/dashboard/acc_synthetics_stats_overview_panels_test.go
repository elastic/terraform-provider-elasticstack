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

// TestAccResourceDashboardSyntheticsStatsOverviewMinimal tests a synthetics_stats_overview panel
// with no config block (shows all monitors).
func TestAccResourceDashboardSyntheticsStatsOverviewMinimal(t *testing.T) {
	dashboardTitle := "Test Dashboard Synthetics Stats Overview Minimal " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "synthetics_stats_overview"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceDashboardSyntheticsStatsOverviewDisplaySettings tests display settings
// (title, description, hide_title, hide_border).
func TestAccResourceDashboardSyntheticsStatsOverviewDisplaySettings(t *testing.T) {
	dashboardTitle := "Test Dashboard Synthetics Stats Overview Display " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_display_settings"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "synthetics_stats_overview"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.title", "Synthetics Overview"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.description", "Shows all monitor statuses"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.hide_border", "false"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_display_settings"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceDashboardSyntheticsStatsOverviewWithFilters tests the filters sub-block.
func TestAccResourceDashboardSyntheticsStatsOverviewWithFilters(t *testing.T) {
	dashboardTitle := "Test Dashboard Synthetics Stats Overview Filters " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "synthetics_stats_overview"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.filters.projects.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.filters.projects.0.label", "My Project"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.filters.projects.0.value", "my-project"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.filters.monitor_types.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.filters.monitor_types.0.value", "http"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceDashboardSyntheticsStatsOverviewWithDrilldowns tests URL drilldowns.
func TestAccResourceDashboardSyntheticsStatsOverviewWithDrilldowns(t *testing.T) {
	dashboardTitle := "Test Dashboard Synthetics Stats Overview Drilldowns " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_drilldowns"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "synthetics_stats_overview"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.drilldowns.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.drilldowns.0.url", "https://example.com/{{context.panel.title}}"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.drilldowns.0.label", "View details"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.drilldowns.0.trigger", "on_open_panel_menu"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.drilldowns.0.type", "url_drilldown"),
					// Optional bools omitted — confirm null-preservation in state.
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.drilldowns.0.encode_url"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_stats_overview_config.drilldowns.0.open_in_new_tab"),
				),
			},
		},
	})
}

// TestAccResourceDashboardSyntheticsStatsOverviewInvalidConfig tests validation errors.
func TestAccResourceDashboardSyntheticsStatsOverviewInvalidConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// config_json rejected for synthetics_stats_overview panel type (schema validator).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_config_json"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
			// synthetics_stats_overview_config rejected for non-synthetics_stats_overview panel type.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_panel_type"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
		},
	})
}
