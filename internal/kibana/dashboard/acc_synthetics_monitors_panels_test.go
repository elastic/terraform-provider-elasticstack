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

// TestAccResourceDashboardSyntheticsMonitors tests the synthetics_monitors panel type.
// The test cases cover: bare panel (no config block), panel with some filters, and panel
// with all five filter dimensions. Plan stability is verified by a PlanOnly step after create.
func TestAccResourceDashboardSyntheticsMonitors(t *testing.T) {
	dashboardTitle := "Test Dashboard Synthetics Monitors " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// 3.1: Bare panel — no synthetics_monitors_config block.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("bare_panel"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "synthetics_monitors"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config"),
				),
			},
			// 3.4: Plan stability — no changes after bare panel create.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("bare_panel"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			// Import bare panel.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("bare_panel"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// 3.2: Panel with some filters (projects and tags).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_some_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "synthetics_monitors"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.projects.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.projects.0.label", "My Project"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.projects.0.value", "my-project"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.tags.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.tags.0.value", "production"),
				),
			},
			// 3.4: Plan stability after filters create.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_some_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			// 3.3: Panel with all five filter dimensions.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_filter_dimensions"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "synthetics_monitors"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.projects.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.tags.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.monitor_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.monitor_ids.0.value", "monitor-a-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.locations.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.locations.0.value", "us-east"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.monitor_types.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.synthetics_monitors_config.filters.monitor_types.0.value", "http"),
				),
			},
			// Import with all filter dimensions.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_filter_dimensions"),
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

// TestAccResourceDashboardSyntheticsMonitorsInvalidConfig covers schema-level validation:
// 3.7: config_json on a synthetics_monitors panel is rejected.
// synthetics_monitors_config on a non-synthetics_monitors panel is rejected.
func TestAccResourceDashboardSyntheticsMonitorsInvalidConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// synthetics_monitors_config on type=lens is rejected.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_panel_type"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
			// config_json on type=synthetics_monitors is rejected.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_config_json"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
		},
	})
}
