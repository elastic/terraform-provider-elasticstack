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

package mlsinglemetricviewer_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlsinglemetricviewer"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardMlSingleMetricViewerSelectedEntities(t *testing.T) {
	dashboardTitle := "Test Dashboard ML SMV SelectedEntities " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlsinglemetricviewer.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("selected_entities"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "ml_single_metric_viewer"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.airline.string_value", "AAL"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.airline.numeric_value"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.region_code.numeric_value", "4"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.region_code.string_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("selected_entities"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Kibana returns selected_detector_index=0 when unset; import has no prior
				// Terraform state to null-preserve against, so the default surfaces as 0.
				ImportStateVerifyIgnore: []string{"panels.0.ml_single_metric_viewer_config.selected_detector_index"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("selected_entities"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.airline.string_value", "AAL"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.airline.numeric_value"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.region_code.numeric_value", "4"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.region_code.string_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("selected_entities_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.job_ids.0", "fake-job-beta"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.host.string_value", "web-01"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.host.numeric_value"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.shard.numeric_value", "7"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.shard.string_value"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerInvalidSelectedEntities(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("both_values"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)Invalid selected_entities entry.*not both`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("neither_value"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)Invalid selected_entities entry.*Exactly one of ` + "`string_value` or `numeric_value` must be set"),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerInvalidJobIds(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("too_many_job_ids"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)at most 1 elements, got: 2`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_job_ids"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)at least 1 elements, got: 0`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("omitted_job_ids"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)attribute "job_ids" is required`),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerForecastAndFunction(t *testing.T) {
	dashboardTitle := "Test Dashboard ML SMV Forecast " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlsinglemetricviewer.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_forecast_function"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.forecast_id", "fake-forecast-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.function_description", "mean"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_forecast_function"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Kibana returns selected_detector_index=0 when unset; import has no prior
				// Terraform state to null-preserve against, so the default surfaces as 0.
				ImportStateVerifyIgnore: []string{"panels.0.ml_single_metric_viewer_config.selected_detector_index"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_forecast_function_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.forecast_id", "another-fake-forecast-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.function_description", "max"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_forecast_function_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.forecast_id", "another-fake-forecast-id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.function_description", "max"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerInvalidFunctionDescription(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_function_description"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)must be one of: \["min" "max" "mean"\], got: "median"`),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerNullPreservation(t *testing.T) {
	dashboardTitle := "Test Dashboard ML SMV NullPreservation " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlsinglemetricviewer.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_selected_entities"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.%"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_selected_entities"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_entities.%"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_detector_index"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.forecast_id"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.function_description"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.description"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_border"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerSelectedDetectorIndex(t *testing.T) {
	dashboardTitle := "Test Dashboard ML SMV DetectorIndex " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlsinglemetricviewer.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_detector_index"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_detector_index", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_detector_index"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_detector_index_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_detector_index", "5"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_detector_index_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.selected_detector_index", "5"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerOptionalFields(t *testing.T) {
	dashboardTitle := "Test Dashboard ML SMV Optionals " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlsinglemetricviewer.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_optionals"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.title", "Single Metric Viewer"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.description", "SMV panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_border", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.from", "now-7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.mode", "relative"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_optionals"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Kibana returns selected_detector_index=0 when unset; import has no prior
				// Terraform state to null-preserve against, so the default surfaces as 0.
				ImportStateVerifyIgnore: []string{"panels.0.ml_single_metric_viewer_config.selected_detector_index"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_optionals_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.title", "Updated SMV"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.description", "Updated description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_title", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_border", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.from", "now-30d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.mode", "absolute"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_optionals_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.title", "Updated SMV"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.description", "Updated description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_title", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.hide_border", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.from", "now-30d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_single_metric_viewer_config.time_range.mode", "absolute"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlSingleMetricViewerInvalidConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("missing_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Missing ML single metric viewer panel configuration`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_config_json"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_panel_type"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("sibling_block_conflict"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
		},
	})
}
