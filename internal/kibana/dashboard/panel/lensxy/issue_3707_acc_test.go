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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3707 covers the regression where using operation="percentile"
// (with a numeric percentile value) in a bar_horizontal layer's y config_json failed:
// the provider unconditionally injected empty_as_null=false, which the Kibana percentile
// metric schema rejects with an HTTP 400. The fix gates the empty_as_null default on the
// metric operation, so a percentile layer that omits empty_as_null now applies cleanly
// and round-trips without drift.
//
// Related to: https://github.com/elastic/terraform-provider-elasticstack/issues/3707
func TestAccReproduceIssue3707(t *testing.T) {
	dashboardTitle := "Repro Issue 3707 percentile bar " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("percentile_bar_horizontal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.repro_3707", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.repro_3707", "panels.0.vis_config.by_value.xy_chart_config.layers.0.type", "bar_horizontal"),
				),
			},
			{
				// Plan-only follow-up must show no drift after a clean apply.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("percentile_bar_horizontal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}

// TestAccDashboardXYMetricEmptyAsNullGating verifies the empty_as_null gating across two
// representative XY metric operations:
//   - average maps to the stats-metric schema, which (like percentile) does not accept
//     empty_as_null; a clean apply confirms the gate prevents the HTTP 400.
//   - count maps to the count-metric schema, which does accept empty_as_null; the
//     provider still injects empty_as_null=false and it round-trips without drift.
func TestAccDashboardXYMetricEmptyAsNullGating(t *testing.T) {
	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	for _, tc := range []struct {
		name string
		dir  string
	}{
		{name: "average", dir: "average_bar_horizontal"},
		{name: "count", dir: "count_bar_horizontal"},
	} {
		dashboardTitle := "XY empty_as_null " + tc.name + " " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
		configVars := config.Variables{
			"dashboard_title": config.StringVariable(dashboardTitle),
		}

		resource.Test(t, resource.TestCase{
			PreCheck: func() { acctest.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory(tc.dir),
					ConfigVariables:          configVars,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
						resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.vis_config.by_value.xy_chart_config.layers.0.type", "bar_horizontal"),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory(tc.dir),
					ConfigVariables:          configVars,
					PlanOnly:                 true,
				},
			},
		})
	}
}
