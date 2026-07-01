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

package mlanomalyswimlane_test

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

func TestAccResourceDashboardMlAnomalySwimlaneOverall(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Swimlane Overall " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("overall"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "ml_anomaly_swimlane"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.swimlane_type", "overall"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.view_by"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.per_page"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.description"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_border"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("overall"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("overall"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.view_by"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.per_page"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.description"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_border"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("overall_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.swimlane_type", "overall"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.1", "fake-job-beta"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.view_by"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalySwimlaneViewBy(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Swimlane ViewBy " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("view_by"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "ml_anomaly_swimlane"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.swimlane_type", "viewBy"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.view_by", "host.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.0", "fake-job-alpha"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("view_by"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("view_by_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.swimlane_type", "viewBy"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.view_by", "service.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.0", "fake-job-alpha"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.1", "fake-job-beta"),
				),
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalySwimlaneOptionalFields(t *testing.T) {
	dashboardTitle := "Test Dashboard ML Anomaly Swimlane Optionals " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.swimlane_type", "viewBy"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.view_by", "host.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.job_ids.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.per_page", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.title", "Anomaly Swim Lane"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.description", "View-by swim lane panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_border", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range.from", "now-7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range.mode", "relative"),
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
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_optionals_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.per_page", "25"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.title", "Updated Swim Lane"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.description", "Updated description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_title", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.hide_border", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range.from", "now-30d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.ml_anomaly_swimlane_config.time_range.mode", "absolute"),
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
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalySwimlaneInvalidSwimlaneType(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("view_by_missing"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)view_by`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("overall_with_view_by"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)view_by`),
			},
		},
	})
}

func TestAccResourceDashboardMlAnomalySwimlaneInvalidConfig(t *testing.T) {
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
				ExpectError: regexp.MustCompile(`Missing ML anomaly swim lane panel configuration`),
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
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_job_ids"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)job_ids`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("omitted_job_ids"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?s)job_ids`),
			},
		},
	})
}
