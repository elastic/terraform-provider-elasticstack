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

// TestAccReproduceIssue3402 covers the regression where bar_stacked layers
// caused Kibana to return an empty string for fitting.type on read-back even
// though the practitioner wrote "none". The provider previously surfaced
// "Provider produced inconsistent result after apply" because the empty string
// was stored as a known value, skipping the plan-preservation alignment step.
//
// The fix uses typeutils.NonEmptyStringishValue so the empty string maps to
// types.StringNull(), letting alignXYFittingStateFromPlan restore the
// practitioner's "none" intent.
//
// Related to: https://github.com/elastic/terraform-provider-elasticstack/issues/3402
func TestAccReproduceIssue3402(t *testing.T) {
	dashboardTitle := "Repro Issue 3402 bar_stacked " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("bar_stacked"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.repro_3402", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.repro_3402", "panels.0.vis_config.by_value.xy_chart_config.layers.0.type", "bar_stacked"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.repro_3402", "panels.0.vis_config.by_value.xy_chart_config.fitting.type", "none"),
				),
			},
			{
				// Plan-only follow-up must show no drift after a clean apply.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("bar_stacked"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly: true,
			},
		},
	})
}
