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

package apmservicemap_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/apmservicemap"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDashboardPanelApmServiceMap_invalidConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_config_json"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`Invalid Configuration`),
			},
		},
	})
}

func TestAccDashboardPanelApmServiceMap_environmentOnly(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map Environment " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("environment_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "apm_service_map"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.environment", "production"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("environment_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("environment_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccDashboardPanelApmServiceMap_environmentAllExplicit covers the read-path guarantee that an
// explicitly configured `environment = "ENVIRONMENT_ALL"` is preserved rather than suppressed to
// null (the suppression only fires when the prior state has no known `environment` value). This
// test intentionally has no import step: on import there is no prior state to distinguish an
// explicit `"ENVIRONMENT_ALL"` from an omitted `environment`, so the import path always suppresses
// the server default to null (see the "Import has no prior plan..." risk in
// openspec/changes/apm-service-map-environment-default/design.md). That is an accepted, documented
// trade-off, not something `ImportStateVerify` should assert against here.
func TestAccDashboardPanelApmServiceMap_environmentAllExplicit(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map Environment All Explicit " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("environment_all_explicit"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "apm_service_map"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.environment", "ENVIRONMENT_ALL"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("environment_all_explicit"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDashboardPanelApmServiceMap_serviceNameOnly(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map Service Name " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("service_name_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "apm_service_map"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.service_name", "checkout"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("service_name_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("service_name_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDashboardPanelApmServiceMap_serviceGroupIdOnly(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map Service Group " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("service_group_id_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "apm_service_map"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.service_group_id", "group-abc"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("service_group_id_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("service_group_id_only"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDashboardPanelApmServiceMap_combinedSelectors(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map Combined " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("combined_selectors"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.environment", "staging"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.service_name", "checkout"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.service_group_id", "group-abc"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("combined_selectors"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("combined_selectors"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDashboardPanelApmServiceMap_allFilters(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map Filters " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.alert_status_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.alert_status_filter.*", "active"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.alert_status_filter.*", "delayed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.anomaly_severity_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.anomaly_severity_filter.*", "major"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.anomaly_severity_filter.*", "critical"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.connection_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.connection_filter.*", "connected"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.connection_filter.*", "orphaned"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.slo_status_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.slo_status_filter.*", "healthy"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.slo_status_filter.*", "noData"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDashboardPanelApmServiceMap_full(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map Full " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.title", "APM Service Map"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.description", "Dependencies overview"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.hide_border", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.service_name", "checkout"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.service_group_id", "group-abc"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.kuery", "service.name : checkout"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.map_orientation", "vertical"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.sync_with_dashboard_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.time_range.from", "now-7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.time_range.mode", "relative"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.alert_status_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.alert_status_filter.*", "active"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.alert_status_filter.*", "recovered"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.anomaly_severity_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.anomaly_severity_filter.*", "warning"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.anomaly_severity_filter.*", "minor"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.connection_filter.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.connection_filter.*", "connected"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.slo_status_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.slo_status_filter.*", "degrading"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config.slo_status_filter.*", "violated"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDashboardPanelApmServiceMap_noConfig(t *testing.T) {
	dashboardTitle := "Test Dashboard APM Service Map No Config " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, apmservicemap.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "apm_service_map"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.apm_service_map_config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_config"),
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
