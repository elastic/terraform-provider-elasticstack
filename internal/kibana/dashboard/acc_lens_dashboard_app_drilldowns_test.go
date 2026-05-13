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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const lensAppByRefDashboard = "elasticstack_kibana_dashboard.test"

func TestAccResourceDashboardLensDashboardAppByReference_dashboardDrilldown(t *testing.T) {
	dashboardTitle := "Acc lens dd dash " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.lens_dashboard_app_config.by_reference.drilldowns.0.dashboard"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lensAppByRefDashboard, "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, "panels.0.lens_dashboard_app_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".dashboard_id", "11111111-1111-1111-1111-111111111111"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".label", "Open detail dashboard"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".use_filters", "false"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".use_time_range", "true"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".open_in_new_tab", "true"),
					resource.TestCheckNoResourceAttr(lensAppByRefDashboard, "panels.0.lens_dashboard_app_config.by_reference.drilldowns.0.url"),
					resource.TestCheckNoResourceAttr(lensAppByRefDashboard, "panels.0.lens_dashboard_app_config.by_reference.drilldowns.0.discover"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             lensAppByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.lens_dashboard_app_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardLensDashboardAppByReference_discoverDrilldown(t *testing.T) {
	dashboardTitle := "Acc lens dd disc " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.lens_dashboard_app_config.by_reference.drilldowns.0.discover"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lensAppByRefDashboard, "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, "panels.0.lens_dashboard_app_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".label", "Open in Discover"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".open_in_new_tab", "false"),
					resource.TestCheckNoResourceAttr(lensAppByRefDashboard, "panels.0.lens_dashboard_app_config.by_reference.drilldowns.0.dashboard"),
					resource.TestCheckNoResourceAttr(lensAppByRefDashboard, "panels.0.lens_dashboard_app_config.by_reference.drilldowns.0.url"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             lensAppByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.lens_dashboard_app_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardLensDashboardAppByReference_urlDrilldownExplicitTrigger(t *testing.T) {
	dashboardTitle := "Acc lens dd urle " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.lens_dashboard_app_config.by_reference.drilldowns.0.url"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lensAppByRefDashboard, "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, "panels.0.lens_dashboard_app_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".url", "https://example.com/{{event.field}}"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".label", "External"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".trigger", "on_click_value"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".encode_url", "false"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, p+".open_in_new_tab", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             lensAppByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.lens_dashboard_app_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardLensDashboardAppByReference_urlDrilldown_triggerRequired_planRejected(t *testing.T) {
	dashboardTitle := "Acc lens dd trig req " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	triggerRequired := regexp.MustCompile(`(?i)trigger`)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly:                 true,
				ExpectError:              triggerRequired,
			},
		},
	})
}

func TestAccResourceDashboardLensDashboardAppByReference_mixedDrilldowns(t *testing.T) {
	dashboardTitle := "Acc lens dd mix " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	dd := "panels.0.lens_dashboard_app_config.by_reference.drilldowns"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lensAppByRefDashboard, "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".#", "3"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".0.dashboard.dashboard_id", "22222222-2222-2222-2222-222222222222"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".0.dashboard.label", "Dashboard drill"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".1.url.url", "https://mixed.example/{{event.field}}"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".1.url.label", "URL drill"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".1.url.trigger", "on_open_panel_menu"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".1.url.encode_url", "true"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".1.url.open_in_new_tab", "true"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".2.discover.label", "Discover drill"),
					resource.TestCheckResourceAttr(lensAppByRefDashboard, dd+".2.discover.open_in_new_tab", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             lensAppByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.lens_dashboard_app_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}
