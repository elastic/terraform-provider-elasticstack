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

// TestAccResourceDashboardLensDashboardApp tests creation and import of lens-dashboard-app panels
// in both by-reference and by-value modes, with and without optional fields.
func TestAccResourceDashboardLensDashboardApp(t *testing.T) {
	dashboardTitle := "Test Dashboard Lens Dashboard App " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// 6.1: by-reference mode (required saved_object_id only)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference_minimal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.saved_object_id", "test-saved-object-id-001"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value"),
				),
			},
			// Import round-trip for by-reference
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference_minimal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// 6.2: by-value mode (required attributes_json)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_value_minimal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.attributes_json"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference"),
				),
			},
			// 6.3: by-reference with optional title, description, hide_title, hide_border
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_reference_with_optionals"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.saved_object_id", "test-saved-object-id-002"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.title", "My Shared Visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.description", "A shared Lens visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.hide_border", "false"),
				),
			},
			// 6.4: by-value with optional references_json
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_value_with_references"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.attributes_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.references_json"),
				),
			},
			// 6.5: with optional time_range block
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_time_range"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.saved_object_id", "test-saved-object-id-003"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.time_range.from", "now-7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.time_range.to", "now"),
				),
			},
		},
	})
}

// TestAccResourceDashboardLensDashboardAppInvalidBothSubblocks tests plan-time rejection
// when both by_value and by_reference are set simultaneously (6.6).
func TestAccResourceDashboardLensDashboardAppInvalidBothSubblocks(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("main"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)mutually exclusive|by_value.*by_reference|by_reference.*by_value`),
			},
		},
	})
}

// TestAccResourceDashboardLensDashboardAppInvalidNeitherSubblock tests plan-time rejection
// when neither by_value nor by_reference is set (6.7).
func TestAccResourceDashboardLensDashboardAppInvalidNeitherSubblock(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("main"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)exactly one|by_value.*by_reference`),
			},
		},
	})
}

// TestAccResourceDashboardLensDashboardAppConfigJSONRejected tests that config_json is rejected
// for lens-dashboard-app panels (6.12).
func TestAccResourceDashboardLensDashboardAppConfigJSONRejected(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("main"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)lens-dashboard-app|not allowed|not supported|Invalid Configuration`),
			},
		},
	})
}
