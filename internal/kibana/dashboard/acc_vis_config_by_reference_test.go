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
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const (
	visByRefDashboard = "elasticstack_kibana_dashboard.test"
	visByRefPath      = "panels.0.vis_config.by_reference"
)

func TestAccResourceDashboardVisConfigByReference_minimal(t *testing.T) {
	dashboardTitle := "Acc vis by-ref min " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	br := visByRefPath
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.type", "vis"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".ref_id", "lensRef"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.from", "now-7d"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.to", "now"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.mode", "relative"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, br+".references_json"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, br+".title"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, br+".description"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, br+".hide_title"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, br+".hide_border"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".drilldowns.#", "0"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, "panels.0.vis_config.by_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccResourceDashboardVisConfigByReference_full(t *testing.T) {
	dashboardTitle := "Acc vis by-ref full " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	br := visByRefPath
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.type", "vis"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".ref_id", "lensRef"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".title", "Ref title"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".description", "By reference desc"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".hide_title", "true"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".hide_border", "false"),
					checks.TestCheckResourceAttrJSONSubset(visByRefDashboard, br+".references_json", `[{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","type":"lens"}]`),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.from", "now-7d"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.to", "now"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.mode", "relative"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".drilldowns.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, br+".title", "Ref title updated"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".description", "By reference desc updated"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".hide_title", "false"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".hide_border", "true"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.from", "now-30d"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.to", "now"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.mode", "relative"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             visByRefDashboard,
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

func TestAccResourceDashboardVisConfigByReference_dashboardDrilldown(t *testing.T) {
	dashboardTitle := "Acc vis dd dash " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.vis_config.by_reference.drilldowns.0.dashboard"
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.vis_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".dashboard_id", "11111111-1111-1111-1111-111111111111"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".label", "Open detail dashboard"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".use_filters", "false"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".use_time_range", "true"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".open_in_new_tab", "true"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, "panels.0.vis_config.by_reference.drilldowns.0.url"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, "panels.0.vis_config.by_reference.drilldowns.0.discover"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             visByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.vis_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardVisConfigByReference_discoverDrilldown(t *testing.T) {
	dashboardTitle := "Acc vis dd disc " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.vis_config.by_reference.drilldowns.0.discover"
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.vis_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".label", "Open in Discover"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".open_in_new_tab", "false"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, "panels.0.vis_config.by_reference.drilldowns.0.dashboard"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, "panels.0.vis_config.by_reference.drilldowns.0.url"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             visByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.vis_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardVisConfigByReference_urlDrilldownExplicitTrigger(t *testing.T) {
	dashboardTitle := "Acc vis dd urle " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	p := "panels.0.vis_config.by_reference.drilldowns.0.url"
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.vis_config.by_reference.drilldowns.#", "1"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".url", "https://example.com/{{event.field}}"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".label", "External"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".trigger", "on_click_value"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".encode_url", "false"),
					resource.TestCheckResourceAttr(visByRefDashboard, p+".open_in_new_tab", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             visByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.vis_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardVisConfigByReference_valueToReference_update(t *testing.T) {
	dashboardTitle := "Acc vis val2ref " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	br := visByRefPath
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_value"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.type", "vis"),
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.vis_config.by_value.metric_chart_config.title", "Metric Chart"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, br),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, "panels.0.type", "vis"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".ref_id", "lensRef"),
					resource.TestCheckResourceAttr(visByRefDashboard, br+".time_range.from", "now-7d"),
					resource.TestCheckNoResourceAttr(visByRefDashboard, "panels.0.vis_config.by_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             visByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardVisConfigByReference_urlDrilldown_triggerRequired_planRejected(t *testing.T) {
	dashboardTitle := "Acc vis dd trig req " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	triggerRequired := regexp.MustCompile(`(?i)trigger`)
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly:                 true,
				ExpectError:              triggerRequired,
			},
		},
	})
}

func TestAccResourceDashboardVisConfigByReference_mixedDrilldowns(t *testing.T) {
	dashboardTitle := "Acc vis dd mix " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	dd := "panels.0.vis_config.by_reference.drilldowns"
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".#", "3"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".0.dashboard.dashboard_id", "22222222-2222-2222-2222-222222222222"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".0.dashboard.label", "Dashboard drill"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".1.url.url", "https://mixed.example/{{event.field}}"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".1.url.label", "URL drill"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".1.url.trigger", "on_open_panel_menu"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".1.url.encode_url", "true"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".1.url.open_in_new_tab", "true"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".2.discover.label", "Discover drill"),
					resource.TestCheckResourceAttr(visByRefDashboard, dd+".2.discover.open_in_new_tab", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             visByRefDashboard,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.vis_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}
