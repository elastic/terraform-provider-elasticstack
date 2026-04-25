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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Note: A `by_value` end-to-end acceptance test is not included because Kibana enriches
// the inline panel config on read, so the first apply can fail post-apply consistency checks
// on `by_value.config_json` versus the read-back payload. The by-value path is covered in unit
// tests (e.g. TestLensDashboardAppByValueToAPI_sendsConfigAsAPI, TestLensDashboardAppByValueToAPI_UnknownConfigJSON).

func TestAccResourceDashboardLensDashboardAppByReference(t *testing.T) {
	dashboardTitle := "Acc lens app by-ref " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{{
			ProtoV6ProviderFactories: acctest.Providers,
			SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
			ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
			ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.ref_id", "lensRef"),
				resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.title", "Ref title"),
				resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.description", "By reference desc"),
				resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.hide_title", "true"),
				resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.hide_border", "false"),
				resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.mode", "relative"),
				resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.references_json"),
			),
		}},
	})
}

func TestAccResourceDashboardLensDashboardAppByReferenceAbsoluteTimeMode(t *testing.T) {
	dashboardTitle := "Acc lens app abs " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{{
			ProtoV6ProviderFactories: acctest.Providers,
			SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
			ConfigDirectory:          acctest.NamedTestCaseDirectory("absolute_time_mode"),
			ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.mode", "absolute"),
			),
		}},
	})
}

func TestAccResourceDashboardLensDashboardAppPlan(t *testing.T) {
	dashboardTitle := "Acc lens app plan " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	// Plan-time failures use different summaries (e.g. schema "Invalid Configuration" vs block "Invalid lens_dashboard_app_config").
	expectPlanErr := regexp.MustCompile(`Invalid`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("wrong_type"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("missing_config"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("sibling_config_conflict"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("vis_type_with_lens_block"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("invalid_time_mode"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: regexp.MustCompile(`(?i)invalid|Invalid|expected`)},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("both_value_and_reference"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("neither_value_nor_reference"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("panel_config_json_forbidden"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("panel_config_json_with_lens_block"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: expectPlanErr},
		},
	})
}
