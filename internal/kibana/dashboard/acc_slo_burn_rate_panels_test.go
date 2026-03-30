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

func TestAccResourceDashboardSloBurnRate(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Burn Rate " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create with required fields only
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "slo_burn_rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "6"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.slo_id", "test-slo-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.duration", "72h"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.slo_instance_id"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
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

func TestAccResourceDashboardSloBurnRateWithInstanceID(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Burn Rate Instance " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_instance_id"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "slo_burn_rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.slo_id", "test-slo-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.duration", "6d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.slo_instance_id", "host-a"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.title", "Burn Rate: host-a"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.description", "Monitors the 6-day burn rate for host-a"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_instance_id"),
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

func TestAccResourceDashboardSloBurnRateWithDrilldowns(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Burn Rate Drilldowns " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "slo_burn_rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.slo_instance_id", "host-a"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.drilldowns.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.drilldowns.0.url", "https://example.com/{{context.panel.title}}"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.drilldowns.0.label", "View details"),
					// Optional drilldown bools omitted in config — confirm null-preservation in state.
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.drilldowns.0.encode_url"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.drilldowns.0.open_in_new_tab"),
				),
			},
		},
	})
}

// TestAccResourceDashboardSloBurnRateDisplayOptions covers hide_title, hide_border, title,
// description, and update behavior (duration change, bool flip).
func TestAccResourceDashboardSloBurnRateDisplayOptions(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Burn Rate Display " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create with display options: hide_title=true, hide_border=false, duration="5m" (minutes unit)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_display_options"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "slo_burn_rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.slo_id", "test-slo-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.duration", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.title", "My Burn Rate Panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.description", "Monitors the 5-minute burn rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.hide_border", "false"),
				),
			},
			// Update: change duration, flip hide_title/hide_border, remove title/description
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("display_options_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.duration", "24h"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.hide_title", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.hide_border", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.description"),
				),
			},
		},
	})
}

// Invalid duration is rejected at plan time.
func TestAccResourceDashboardSloBurnRateInvalidDuration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_duration"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)duration.*\\d\+\[mhd\]`),
			},
		},
	})
}

func TestAccResourceDashboardSloBurnRateInvalidConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_panel_type"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
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

// slo_instance_id null-preservation: when not configured, stays null after Kibana read-back.
func TestAccResourceDashboardSloBurnRateSloInstanceIDNullPreservation(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Burn Rate Null Preservation " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create without slo_instance_id — should be null after read-back.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_burn_rate_config.slo_instance_id"),
				),
			},
			// Re-apply — no plan changes expected (null-preservation).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}
