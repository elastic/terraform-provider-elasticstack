package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardMosaicChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with Mosaic " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),

					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.title", "Sample Mosaic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.description", "Test mosaic visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.sampling", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.legend.size", "auto"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.legend.nested", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.legend.truncate_after_lines", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.legend.visible", "show"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.value_display.mode", "percentage"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.value_display.percent_decimals", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.filters.0.language", "kuery"),

					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.standard.dataset"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.standard.group_by.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.standard.group_breakdown_by.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.standard.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.standard.group_by.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.standard.group_breakdown_by.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.standard.metrics.0.config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("esql"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " ESQL"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" ESQL"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "lens"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "15"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),

					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.title", "Sample Mosaic ESQL"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.description", "Test mosaic visualization (ES|QL)"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.sampling", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.legend.size", "auto"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.value_display.mode", "absolute"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.filters.#", "1"),

					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.esql.dataset"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.esql.group_by.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.esql.group_breakdown_by.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.esql.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.esql.group_by.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.esql.group_breakdown_by.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.mosaic_config.esql.metrics.0.config"),
				),
			},
			{
				PreConfig: func() {
					t.Log("Hello")
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("esql"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore JSON fields with API defaults/normalization.
				ImportStateVerifyIgnore: []string{
					"panels.0.mosaic_config.standard.group_by",
					"panels.0.mosaic_config.standard.group_breakdown_by",
					"panels.0.mosaic_config.standard.metrics",
					"panels.0.mosaic_config.esql.group_by",
					"panels.0.mosaic_config.esql.group_breakdown_by",
					"panels.0.mosaic_config.esql.metrics",
				},
			},
		},
	})
}
