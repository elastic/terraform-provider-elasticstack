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

func TestAccResourceDashboardESQLControl(t *testing.T) {
	dashboardTitle := "Test Dashboard with ES|QL Control " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create with full config
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "esql_control"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.variable_name", "my_var"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.variable_type", "values"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.esql_query", "FROM logs-* | STATS count = COUNT(*) BY host.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.control_type", "STATIC_VALUES"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.selected_options.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.selected_options.0", "option_a"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.selected_options.1", "option_b"),
				),
			},
			// Refresh/plan: no drift
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update to empty config block
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "esql_control"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.variable_name"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.variable_type"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.esql_query"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config.control_type"),
				),
			},
			// Update to no config block
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "esql_control"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.esql_control_config"),
				),
			},
		},
	})
}

func TestAccResourceDashboardESQLControlInvalidEnum(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				Config: `
resource "elasticstack_kibana_dashboard" "test" {
  title = "invalid-enum-test"
  panels = [{
    type = "esql_control"
    grid = { x = 0, y = 0 }
    esql_control_config = {
      selected_options = []
      variable_name    = "v"
      variable_type    = "unsupported_type"
      esql_query       = "FROM *"
      control_type     = "STATIC_VALUES"
    }
  }]
}
`,
				ExpectError: regexp.MustCompile(`(?i)unsupported_type`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				Config: `
resource "elasticstack_kibana_dashboard" "test" {
  title = "invalid-control-type-test"
  panels = [{
    type = "esql_control"
    grid = { x = 0, y = 0 }
    esql_control_config = {
      selected_options = []
      variable_name    = "v"
      variable_type    = "values"
      esql_query       = "FROM *"
      control_type     = "UNSUPPORTED"
    }
  }]
}
`,
				ExpectError: regexp.MustCompile(`(?i)UNSUPPORTED`),
			},
		},
	})
}
