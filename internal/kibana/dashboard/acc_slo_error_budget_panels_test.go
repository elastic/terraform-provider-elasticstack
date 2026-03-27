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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardSloErrorBudgetMinimal(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Error Budget Minimal " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "slo_error_budget"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.slo_id", "my-slo-id"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.slo_instance_id"),
				),
			},
			// Import: verify round-trip
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
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

func TestAccResourceDashboardSloErrorBudgetSloInstanceIDNullPreservation(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Error Budget No Instance " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create without slo_instance_id; verify no drift on re-plan
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.slo_instance_id"),
				),
			},
			// Apply again — should produce no diff (no drift)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.slo_instance_id"),
				),
			},
		},
	})
}

func TestAccResourceDashboardSloErrorBudgetWithDrilldowns(t *testing.T) {
	dashboardTitle := "Test Dashboard SLO Error Budget Drilldowns " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "slo_error_budget"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.slo_id", "my-slo-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.drilldowns.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.drilldowns.0.url", "https://example.com/{{context.panel.title}}"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.drilldowns.0.label", "Open Example"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.drilldowns.0.trigger", "on_open_panel_menu"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.drilldowns.0.type", "url_drilldown"),
					// encode_url and open_in_new_tab are omitted in config; should not appear in state
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.drilldowns.0.encode_url"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_error_budget_config.drilldowns.0.open_in_new_tab"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_drilldowns"),
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
