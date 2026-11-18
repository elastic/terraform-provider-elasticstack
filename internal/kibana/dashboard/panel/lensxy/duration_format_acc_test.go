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

package lensxy_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const durationFormatDecimalsAndCompactJSON = `{"format":{"type":"duration","from":"s","to":"ms","decimals":0,"compact":true}}`

func TestAccResourceDashboardXYChart_durationFormatDecimalsAndCompactOnSupportedStack(t *testing.T) {
	dashboardTitle := "Test Dashboard XY duration format " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	updatedTitle := dashboardTitle + " Updated"

	versionutils.SkipIfUnsupported(t, panelkit.MinKibanaAPISupportDurationFormatParams, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duration_format"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.xy_chart_config.layers.0.type", "line"),
					checks.TestCheckResourceAttrJSONSubset(
						"elasticstack_kibana_dashboard.test",
						"panels.0.vis_config.by_value.xy_chart_config.layers.0.data_layer.y.0.config_json",
						durationFormatDecimalsAndCompactJSON,
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duration_format"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duration_format"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(updatedTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", updatedTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.xy_chart_config.layers.0.type", "line"),
					checks.TestCheckResourceAttrJSONSubset(
						"elasticstack_kibana_dashboard.test",
						"panels.0.vis_config.by_value.xy_chart_config.layers.0.data_layer.y.0.config_json",
						durationFormatDecimalsAndCompactJSON,
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duration_format"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(updatedTitle),
				},
				PlanOnly: true,
			},
		},
	})
}
