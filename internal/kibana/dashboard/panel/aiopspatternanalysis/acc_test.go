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

package aiopspatternanalysis_test

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

// TestAccResourceDashboardAiopsPatternAnalysis tests the aiops_pattern_analysis panel type.
func TestAccResourceDashboardAiopsPatternAnalysis(t *testing.T) {
	dashboardTitle := "Test Dashboard AIOps Pattern Analysis " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Required-only create.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "aiops_pattern_analysis"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.data_view_id", "logs-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.field_name", "message"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.minimum_time_range"),
				),
			},
			// Plan stability.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			// Import.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("required_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// All optional fields.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_optional"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "aiops_pattern_analysis"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.data_view_id", "logs-*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.field_name", "message"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.minimum_time_range", "1_week"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.random_sampler_mode", "on_manual"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.random_sampler_probability", "0.01"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.title", "Patterns"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.aiops_pattern_analysis_config.time_range.from", "now-7d"),
				),
			},
			// Plan stability after all-optional create.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_optional"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}

// TestAccResourceDashboardAiopsPatternAnalysisInvalidConfig covers plan-time validation:
// probability out of range, invalid enum values, config_json conflict, and wrong panel type.
func TestAccResourceDashboardAiopsPatternAnalysisInvalidConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// random_sampler_probability out of range is rejected by the Between validator.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_probability"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`random_sampler_probability`),
			},
			// Invalid minimum_time_range is rejected by OneOf.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_minimum_time_range"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`minimum_time_range`),
			},
			// Invalid random_sampler_mode is rejected by OneOf.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_random_sampler_mode"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`random_sampler_mode`),
			},
			// config_json on an aiops_pattern_analysis panel is rejected.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_config_json"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Unsupported panel type for config_json`),
			},
			// aiops_pattern_analysis_config on a non-aiops panel is rejected.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_panel_type"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
		},
	})
}
