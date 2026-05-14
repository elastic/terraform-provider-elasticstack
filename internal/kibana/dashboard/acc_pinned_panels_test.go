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

func TestAccResourceKibanaDashboardPinnedPanels(t *testing.T) {
	dashboardTitle := "Test Dashboard with pinned panels " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_both"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.#", "2"),

					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.type", "options_list_control"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.field_name", "status"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.search_technique", "prefix"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.single_select", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.display_settings.placeholder", "Select status..."),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.display_settings.hide_sort", "true"),

					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.type", "range_slider_control"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.field_name", "source.bytes"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.title", "Bytes Range"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.use_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.ignore_validations", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.value.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.value.0", "100"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.value.1", "500"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.step", "10"),

					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.grid"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.grid"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_both"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_both"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.field_name", "host.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.search_technique", "wildcard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.single_select", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.display_settings.placeholder", "Pick a host..."),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.0.options_list_control_config.display_settings.hide_sort", "false"),

					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.title", "Bytes Range Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.use_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.ignore_validations", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.value.0", "50"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.value.1", "400"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "pinned_panels.1.range_slider_control_config.step", "5"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceKibanaDashboardPinnedPanelsInvalidConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("mismatch"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)Pinned panel control does not match type`),
			},
		},
	})
}
