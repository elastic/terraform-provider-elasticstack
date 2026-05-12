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

const vizByRefDashboard = "elasticstack_kibana_dashboard.test"

func TestAccResourceDashboardVizConfigByReference_minimal(t *testing.T) {
	dashboardTitle := "Acc viz by-ref min " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	br := "panels.0.viz_config.by_reference"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vizByRefDashboard, "panels.0.type", "vis"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".ref_id", "lensRef"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.from", "now-7d"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.to", "now"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.mode", "relative"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, br+".references_json"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, br+".title"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, br+".description"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, br+".hide_title"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, br+".hide_border"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".drilldowns.#", "0"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccResourceDashboardVizConfigByReference_full(t *testing.T) {
	dashboardTitle := "Acc viz by-ref full " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	refWire := regexp.MustCompile(`aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee`)
	typeLens := regexp.MustCompile(`"type"\s*:\s*"lens"`)
	br := "panels.0.viz_config.by_reference"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vizByRefDashboard, "panels.0.type", "vis"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".ref_id", "lensRef"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".title", "Ref title"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".description", "By reference desc"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".hide_title", "true"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".hide_border", "false"),
					resource.TestMatchResourceAttr(vizByRefDashboard, br+".references_json", refWire),
					resource.TestMatchResourceAttr(vizByRefDashboard, br+".references_json", typeLens),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.from", "now-7d"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.to", "now"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.mode", "relative"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".drilldowns.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".title", "Ref title updated"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".description", "By reference desc updated"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".hide_title", "false"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".hide_border", "true"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.from", "now-30d"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.to", "now"),
					resource.TestCheckResourceAttr(vizByRefDashboard, br+".time_range.mode", "relative"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             vizByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					br + ".references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardVizConfigByReference_dashboardDrilldown(t *testing.T) {
	dashboardTitle := "Acc viz dd dash " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.viz_config.by_reference.drilldowns.0.dashboard"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".dashboard_id", "11111111-1111-1111-1111-111111111111"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".label", "Open detail dashboard"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".use_filters", "false"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".use_time_range", "true"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".open_in_new_tab", "true"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_reference.drilldowns.0.url"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_reference.drilldowns.0.discover"),
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
		},
	})
}

func TestAccResourceDashboardVizConfigByReference_discoverDrilldown(t *testing.T) {
	dashboardTitle := "Acc viz dd disc " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.viz_config.by_reference.drilldowns.0.discover"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".label", "Open in Discover"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".open_in_new_tab", "false"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_reference.drilldowns.0.dashboard"),
					resource.TestCheckNoResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_reference.drilldowns.0.url"),
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
		},
	})
}

func TestAccResourceDashboardVizConfigByReference_urlDrilldownExplicitTrigger(t *testing.T) {
	dashboardTitle := "Acc viz dd urle " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.viz_config.by_reference.drilldowns.0.url"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vizByRefDashboard, "panels.0.viz_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".url", "https://example.com/{{event.field}}"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".label", "External"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".trigger", "on_click_value"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".encode_url", "false"),
					resource.TestCheckResourceAttr(vizByRefDashboard, p+".open_in_new_tab", "true"),
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
		},
	})
}

func TestAccResourceDashboardVizConfigByReference_urlDrilldown_triggerRequired_planRejected(t *testing.T) {
	dashboardTitle := "Acc viz dd trig req " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
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

func TestAccResourceDashboardVizConfigByReference_mixedDrilldowns(t *testing.T) {
	dashboardTitle := "Acc viz dd mix " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	dd := "panels.0.viz_config.by_reference.drilldowns"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".#", "3"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".0.dashboard.dashboard_id", "22222222-2222-2222-2222-222222222222"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".0.dashboard.label", "Dashboard drill"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".1.url.url", "https://mixed.example/{{event.field}}"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".1.url.label", "URL drill"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".1.url.trigger", "on_open_panel_menu"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".1.url.encode_url", "true"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".1.url.open_in_new_tab", "true"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".2.discover.label", "Discover drill"),
					resource.TestCheckResourceAttr(vizByRefDashboard, dd+".2.discover.open_in_new_tab", "true"),
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
		},
	})
}
