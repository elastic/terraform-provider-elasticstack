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

func TestAccResourceDashboardPieChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with Pie Chart " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens"),

					// Check pie chart config
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.title", "Sample Pie Chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.description", "Test pie chart visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.donut_hole", "small"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.label_position", "inside"),

					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.query.language", "kuery"),

					// Check JSON fields are set
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.dataset"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.legend"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.metrics.0.config"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.group_by.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.group_by.0.config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				// Check full config
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),

					// Check pie chart config
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.title", "Full Pie Chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.description", "Full pie chart visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.donut_hole", "large"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.label_position", "outside"),

					// Check new attributes (using values compatible with current API behavior/defaults)
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.sampling", "1"),

					// Check JSON fields
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.dataset"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.legend"),

					// Check metrics and group_by
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.metrics.0.config"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.group_by.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.pie_chart_config.group_by.0.config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ResourceName:             "elasticstack_kibana_dashboard.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ImportStateVerifyIgnore: []string{
					"panels.0.pie_chart_config.group_by.0.config",
					"panels.0.pie_chart_config.metrics.0.config",
				},
			},
		},
	})
}
