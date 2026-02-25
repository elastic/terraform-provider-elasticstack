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

func TestAccResourceDashboardHeatmap(t *testing.T) {
	dashboardTitle := "Test Dashboard with Heatmap " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.title", "Sample Heatmap"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.description", "Test heatmap visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.labels.orientation", "horizontal"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.labels.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.y.labels.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.cells.labels.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.size", "medium"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.position", "right"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.query.query", ""),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.dataset_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.metric_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.x_axis_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.y_axis_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complete"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.title", "Complete Heatmap"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.description", "Complete heatmap visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.sampling", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.size", "small"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.position", "left"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.truncate_after_lines", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.labels.orientation", "vertical"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.labels.visible", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.title.value", "Custom X Axis"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.title.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.y.labels.visible", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.y.title.value", "Custom Y Axis"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.y.title.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.cells.labels.visible", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.filters.0.query", "host.os.keyword: \"linux\""),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.filters.0.language", "kuery"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"panels.0.heatmap_config.metric_json",
				},
			},
		},
	})
}
