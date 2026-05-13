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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func skipDashboardOrKqlSLOUnsupported() func() (bool, error) {
	return func() (bool, error) {
		if skip, err := versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport)(); skip || err != nil {
			return skip, err
		}
		return versionutils.CheckIfVersionMeetsConstraints(slo.SLOKqlAccTestConstraints)()
	}
}

// TestAccResourceDashboardSloAlerts_panel_round_trip creates an SLO via elasticstack_kibana_slo,
// references it from an slo_alerts panel without slo_instance_id, sets envelope + explicit drilldown flags.
func TestAccResourceDashboardSloAlerts_panel_round_trip(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	dashboardTitle := "Acc slo alerts " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	// Serial (not ParallelTest): this setup combines elasticsearch_index + kibana_slo + dashboard;
	// parallel runs alongside other acc tests have intermittently failed teardown with
	// "elasticsearch client is not configured" during terraform destroy.
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipDashboardOrKqlSLOUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"suffix":          config.StringVariable(suffix),
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "slo_alerts"),
					resource.TestCheckResourceAttrPair(
						"elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.slos.0.slo_id",
						"elasticstack_kibana_slo.slo", "slo_id",
					),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.slos.0.slo_instance_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.title", "Open violations"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.description", "SLO alerts fixture panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.hide_border", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.drilldowns.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.drilldowns.0.url", "https://example.com/alerts?slo={{context.panel.title}}"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.drilldowns.0.label", "Investigate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.drilldowns.0.encode_url", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.slo_alerts_config.drilldowns.0.open_in_new_tab", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipDashboardOrKqlSLOUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables: config.Variables{
					"suffix":          config.StringVariable(suffix),
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
