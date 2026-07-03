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

package rangeslider_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardRangeSliderControl(t *testing.T) {
	dashboardTitle := "Test Dashboard with Range Slider Control " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create with full config
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "range_slider_control"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "4"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.data_view_id", "test-data-view-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.field_name", "bytes"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.title", "Bytes Range"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.use_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.ignore_validations", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.value.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.value.0", "100"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.value.1", "500"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.step", "10"),
				),
			},
			// Refresh/plan: ensure no perpetual drift
			{
				ProtoV6ProviderFactories: acctest.Providers,
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update to empty config block (required fields only, optionals omitted)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "range_slider_control"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.data_view_id", "test-data-view-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.field_name", "bytes"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field.value"),
				),
			},
		},
	})
}

func TestAccResourceDashboardRangeSliderControlByEsql(t *testing.T) {
	dashboardTitle := "Test Dashboard with Range Slider Control ByEsql " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	// The by_esql branch only exists on Kibana servers that register range_slider_control's config
	// as a values_source-discriminated union. See dashboardacctest.MinControlByFieldEsqlUnionSupport.
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinControlByFieldEsqlUnionSupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create with by_esql config
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "range_slider_control"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_esql.esql_query", "FROM logs-* | STATS min = MIN(bytes), max = MAX(bytes)"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_esql.values_source", "esql_query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_esql.title", "Bytes Range"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_esql.use_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_esql.ignore_validations", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_esql.step", "10"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.range_slider_control_config.by_field"),
				),
			},
			// Refresh/plan: ensure no perpetual drift
			{
				ProtoV6ProviderFactories: acctest.Providers,
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDashboardRangeSliderControlInvalidConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// config_json is not supported for range_slider_control panels (REQ-010).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_config_json"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
			// value must contain exactly 2 elements (REQ-006 / REQ-028).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_value_length"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)value\s+list\s+must\s+contain\s+at\s+least\s+2\s+elements`),
			},
			// range_slider_control_config is required when type = "range_slider_control"; omitting it must be rejected.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)range_slider_control_config`),
			},
		},
	})
}
