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

package links_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/links"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const dashboardResource = "elasticstack_kibana_dashboard.test"

func TestAccKibanaDashboard_LinksPanel_ByValue(t *testing.T) {
	dashboardTitle := "Acc links by-value " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	bv := "panels.0.links_config.by_value"

	versionutils.SkipIfUnsupported(t, links.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_value"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dashboardResource, "id"),
					resource.TestCheckResourceAttr(dashboardResource, "panels.0.type", "links"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".layout", "vertical"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".title", "Links panel title"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".description", "Links panel description"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".hide_title", "false"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".hide_border", "true"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.#", "2"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.0.type", "dashboard"),
					resource.TestCheckResourceAttrSet(dashboardResource, bv+".links.0.destination"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.0.label", "Dashboard link"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.0.open_in_new_tab", "false"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.0.use_filters", "true"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.0.use_time_range", "true"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.1.type", "external"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.1.destination", "https://example.com"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.1.label", "External link"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.1.open_in_new_tab", "true"),
					resource.TestCheckResourceAttr(dashboardResource, bv+".links.1.encode_url", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_value"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_value"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             dashboardResource,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccKibanaDashboard_LinksPanel_ByReference(t *testing.T) {
	dashboardTitle := "Acc links by-reference " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	br := "panels.0.links_config.by_reference"

	versionutils.SkipIfUnsupported(t, links.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dashboardResource, "id"),
					resource.TestCheckResourceAttr(dashboardResource, "panels.0.type", "links"),
					resource.TestCheckResourceAttr(dashboardResource, br+".ref_id", "links-ref-1"),
					resource.TestCheckResourceAttr(dashboardResource, br+".title", "Linked links panel"),
					resource.TestCheckResourceAttr(dashboardResource, br+".description", "Linked links panel description"),
					resource.TestCheckResourceAttr(dashboardResource, br+".hide_title", "true"),
					resource.TestCheckResourceAttr(dashboardResource, br+".hide_border", "false"),
					resource.TestCheckNoResourceAttr(dashboardResource, "panels.0.links_config.by_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             dashboardResource,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.id",
				},
			},
		},
	})
}
