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

package fieldstatstable_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/fieldstatstable"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardFieldStatsTableByDataview(t *testing.T) {
	dashboardTitle := "Test Dashboard with Field Stats Table (dataview) " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, fieldstatstable.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_dataview_with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "field_stats_table"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.show_distributions", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.title", "Field statistics — logs view"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.description", "Field stats table panel (dataview)"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.hide_title", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.hide_border", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.time_range.from", "now-24h"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.time_range.to", "now"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_dataview_with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_dataview_empty_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "field_stats_table"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.data_view_id"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.show_distributions"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.description"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.hide_title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.hide_border"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview.time_range"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_dataview_empty_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceDashboardFieldStatsTableByEsql(t *testing.T) {
	dashboardTitle := "Test Dashboard with Field Stats Table (esql) " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, fieldstatstable.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_esql_with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "field_stats_table"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.query", "FROM logs-* | LIMIT 100"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.show_distributions", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.title", "Field statistics — logs by service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.description", "Field stats table panel (esql)"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.hide_border", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.time_range.from", "now-24h"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.time_range.to", "now"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_esql_with_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_esql_empty_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.query", "FROM logs-* | LIMIT 50"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.show_distributions"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.description"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.hide_title"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.hide_border"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_esql.time_range"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.field_stats_table_config.by_dataview"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("by_esql_empty_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceDashboardFieldStatsTableInvalidConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, fieldstatstable.MinKibanaAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("both_branches"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)Exactly one of .by_dataview. or .by_esql.`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)Exactly one of .by_dataview. or .by_esql.`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_config"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable("unused"),
				},
				ExpectError: regexp.MustCompile(`(?i)field_stats_table_config`),
			},
		},
	})
}
