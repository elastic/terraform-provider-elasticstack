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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// **By-value:** `TestAccResourceDashboardLensDashboardAppByValue` applies the same config twice;
// the second step uses `plancheck.ExpectEmptyPlan()` to assert no post-refresh drift. The
// provider keeps practitioner `by_value.config_json` when the API read is a safe value-superset
// (REQ-035). Residual drift is still possible if Kibana rewrites a user-set value. See
// `preservePriorLensByValueConfigJSON` in `models_lens_dashboard_app_converters.go` and `tasks.md` 6.2.

func TestAccResourceDashboardLensDashboardAppByValue_basic(t *testing.T) {
	dashboardTitle := "Acc lens app by-val " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.config_json", regexp.MustCompile(`"type"\s*:\s*"metric"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.config_json", regexp.MustCompile(`"title"\s*:\s*"Acc by-value"`)),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.config_json", regexp.MustCompile(`"index_pattern"\s*:\s*"metrics-\*"`)),
				),
			},
			// Same config again: require an empty pre-apply plan (no post-apply drift on refresh).
			// PreApply cannot be combined with PlanOnly (framework limitation); a no-op apply is fine.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccResourceDashboardLensDashboardAppByValueTypedMetric applies a typed by-value
// `metric_chart_config` (not raw by_value.config_json) twice with an empty second plan, and
// ensures after apply+read, panel-level `config_json` and by_value `config_json` are not set
// (4.3/4.4). Import has no prior typed selection, so read-back may populate by_value
// `config_json`; the import step only asserts panel `config_json` is still absent.
func TestAccResourceDashboardLensDashboardAppByValueTypedMetric_basic(t *testing.T) {
	dashboardTitle := "Acc lens app typed m " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	dataSourceRe := regexp.MustCompile(`"index_pattern"\s*:\s*"metrics-\*"`)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens-dashboard-app"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.metric_chart_config.data_source_json", dataSourceRe),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.config_json"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_value.config_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             "elasticstack_kibana_dashboard.test",
				ImportState:              true,
				ImportStateVerify:        false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.config_json"),
				),
			},
		},
	})
}

func TestAccResourceDashboardLensDashboardAppByReference(t *testing.T) {
	dashboardTitle := "Acc lens app by-ref " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	refWireID := regexp.MustCompile(`aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee`)
	typeLens := regexp.MustCompile(`"type"\s*:\s*"lens"`)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.from", "now-7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.mode", "relative"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.references_json", refWireID),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.references_json", typeLens),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.title", "Ref title updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.description", "By reference desc updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.hide_title", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.hide_border", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.from", "now-30d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.mode", "relative"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             "elasticstack_kibana_dashboard.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.lens_dashboard_app_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardLensDashboardAppByReferenceAbsoluteTimeMode(t *testing.T) {
	dashboardTitle := "Acc lens app abs " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("absolute_time_mode"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.from", "2024-06-01T00:00:00.000Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.to", "2024-06-01T12:00:00.000Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.lens_dashboard_app_config.by_reference.time_range.mode", "absolute"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("absolute_time_mode"),
				ConfigVariables:          config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				ResourceName:             "elasticstack_kibana_dashboard.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"panels.0.lens_dashboard_app_config.by_reference.references_json",
					"panels.0.id",
				},
			},
		},
	})
}

func TestAccResourceDashboardLensDashboardAppPlan(t *testing.T) {
	dashboardTitle := "Acc lens app plan " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	// Observed: validators emit "Invalid Configuration" or, for the lens block, "Invalid lens_dashboard_app_config".
	// A `vis` panel with only `lens_dashboard_app` may hit `Missing vis panel configuration` first.
	allowedIfMarkdown := regexp.MustCompile(`can only be set when`)
	visWithLensOnly := regexp.MustCompile(`Missing vis panel configuration|can only be set when`)
	requiredIfMissing := regexp.MustCompile(`must be set when`)
	conflictOrInvalid := regexp.MustCompile(`Invalid Configuration`)
	visConfigJSONPlusLens := regexp.MustCompile(`Invalid Configuration`)
	invalidTimeMode := regexp.MustCompile(`weekly|one of|absolute|relative|mode`)
	bothSubblocks := regexp.MustCompile(`not both`)
	neitherSubblock := regexp.MustCompile(`Exactly one of`)
	byValueSourceInvalid := regexp.MustCompile(`Invalid lens_dashboard_app_config\.by_value`)
	panelConfigNotAllowed := regexp.MustCompile(`config_json|can only be set when|markdown|vis`)
	configJSONWithLens := regexp.MustCompile(`Invalid Configuration|conflict`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("wrong_type"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: allowedIfMarkdown},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("wrong_type_vis"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: visWithLensOnly},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("missing_config"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: requiredIfMissing},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("sibling_config_conflict"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: conflictOrInvalid},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("vis_type_with_lens_block"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: visConfigJSONPlusLens},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("invalid_time_mode"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: invalidTimeMode},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("both_value_and_reference"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: bothSubblocks},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("neither_value_nor_reference"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: neitherSubblock},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("panel_config_json_forbidden"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: panelConfigNotAllowed},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("panel_config_json_with_lens_block"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: configJSONWithLens},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("by_value_no_source"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: byValueSourceInvalid},
			{ProtoV6ProviderFactories: acctest.Providers, SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("by_value_config_json_and_metric_typed"), ConfigVariables: config.Variables{"dashboard_title": config.StringVariable(dashboardTitle)},
				PlanOnly: true, ExpectError: byValueSourceInvalid},
		},
	})
}
