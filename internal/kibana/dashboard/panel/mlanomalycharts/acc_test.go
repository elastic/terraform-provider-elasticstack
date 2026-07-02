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

package mlanomalycharts_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlanomalycharts"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardMlAnomalyChartsNamedSeverities(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Charts Named " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlanomalycharts.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("named_severities"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "ml_anomaly_charts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.0.severity", "critical"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.1.severity", "major"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("named_severities"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalyChartsRawRange(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Charts Raw " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlanomalycharts.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("raw_range"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.0.min", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.0.max", "20"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.0.severity"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("raw_range"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalyChartsRawRangeCanonicalCoincidence(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Charts Raw Canonical " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlanomalycharts.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("raw_range_canonical"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.0.min", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.0.max", "25"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.severity_threshold.0.severity"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("raw_range_canonical"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalyChartsJobIDsUpdate(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Charts Jobs " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlanomalycharts.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_job"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.0", "job-a"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated_jobs"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.0", "job-a"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.1", "job-b"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalyChartsOptionalFields(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Charts Optionals " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, mlanomalycharts.MinKibanaAPISupport, versionutils.FlavorAny)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.max_series_to_plot", "12"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.title", "Anomaly Charts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.description", "ML anomaly charts panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.hide_border", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.time_range.from", "now-7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_charts_config.time_range.mode", "relative"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_optionals"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
