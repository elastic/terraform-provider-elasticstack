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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// TestAccResourceDashboardImage_url_drilldowns_round_trip covers URL src, mixed drilldown order
// (dashboard_drilldown then url_drilldown), optional envelope fields, and omits object_fit so
// REQ-009 keeps it null against the API default "contain". Includes create + empty second plan.
func TestAccResourceDashboardImage_url_drilldowns_round_trip(t *testing.T) {
	dashboardTitle := "Acc image url dd " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("url_round_trip"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "image"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.src.url.url", "https://example.com/logo.png"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.alt_text", "Logo"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.background_color", "#111111"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.title", "Image panel title"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.description", "Image panel description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.hide_title", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.hide_border", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.object_fit"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.drilldowns.#", "2"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.image_config.drilldowns.0.dashboard_drilldown.dashboard_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.drilldowns.0.dashboard_drilldown.label", "Open dashboard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.drilldowns.0.dashboard_drilldown.trigger", "on_click_image"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.drilldowns.1.url_drilldown.url", "https://example.com/details/{{context.panel.title}}"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.drilldowns.1.url_drilldown.label", "External link"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.drilldowns.1.url_drilldown.trigger", "on_click_image"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("url_round_trip"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccResourceDashboardImage_object_fit_explicit sets object_fit explicitly (coverage vs unset).
func TestAccResourceDashboardImage_object_fit_explicit(t *testing.T) {
	dashboardTitle := "Acc image fit " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("object_fit"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.object_fit", "cover"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("object_fit"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccResourceDashboardImage_file_src_round_trip uses src.file with a placeholder file_id.
// Kibana may reject unknown file IDs at apply time; if so, skip until a file-upload resource exists.
func TestAccResourceDashboardImage_file_src_round_trip(t *testing.T) {
	dashboardTitle := "Acc image file " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("file_src"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.image_config.src.file.file_id", "acc-test-placeholder-file-id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("file_src"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
