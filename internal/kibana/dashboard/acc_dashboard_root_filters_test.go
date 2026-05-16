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
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Canonical filter_json strings match read-time normalization (lexicographic object keys at every depth).
const (
	accDashboardRootFilterHostCanonical    = `{"condition":{"field":"host.name","operator":"is","value":"web-01"},"type":"condition"}`
	accDashboardRootFilterServiceCanonical = `{"condition":{"field":"service.name","operator":"is","value":"api"},"type":"condition"}`
	accDashboardRootFilterPodCanonical     = `{"condition":{"field":"kubernetes.pod.name","operator":"is","value":"pod-1"},"type":"condition"}`
)

func TestAccResourceDashboardRootSavedFilters(t *testing.T) {
	dashboardTitle := "Test Dashboard root filters " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "filters.#", "2"),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "filters.0.filter_json", accDashboardRootFilterHostCanonical),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "filters.1.filter_json", accDashboardRootFilterServiceCanonical),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:            "elasticstack_kibana_dashboard.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"time_range.mode"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reordered"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "filters.#", "2"),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "filters.0.filter_json", accDashboardRootFilterServiceCanonical),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "filters.1.filter_json", accDashboardRootFilterHostCanonical),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reordered"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reordered"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:            "elasticstack_kibana_dashboard.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"time_range.mode"},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("three_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "filters.#", "3"),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "filters.0.filter_json", accDashboardRootFilterServiceCanonical),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "filters.1.filter_json", accDashboardRootFilterHostCanonical),
					checks.TestCheckResourceAttrJSONSubset("elasticstack_kibana_dashboard.test", "filters.2.filter_json", accDashboardRootFilterPodCanonical),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("three_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceDashboardRootSavedFilters_unset(t *testing.T) {
	dashboardTitle := "Test Dashboard root filters unset " + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, minDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "filters"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:            "elasticstack_kibana_dashboard.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"time_range.mode"},
			},
		},
	})
}
