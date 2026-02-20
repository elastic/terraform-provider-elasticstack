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

func TestAccResourceDashboardTreemap(t *testing.T) {
	dashboardTitle := "Test Dashboard with Treemap " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.title", "Sample Treemap"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.description", "Test treemap visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.query.query", ""),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.dataset"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.group_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.metrics"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.size", "medium"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.nested", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.visible", "auto"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.truncate_after_lines", "5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.value_display.mode", "percentage"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.value_display.percent_decimals", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complete"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.title", "Complete Treemap"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.description", "Complete treemap visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.sampling", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.label_position", "hidden"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.filters.0.query", "host.os.keyword: \"linux\""),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.size", "small"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.nested", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.visible", "show"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.legend.truncate_after_lines", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.value_display.mode", "absolute"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.dataset"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.group_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.metrics"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("esql"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.title", "ESQL Treemap"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.description", "Treemap visualization using ES|QL"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.sampling", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.label_position", "hidden"),
					resource.TestMatchResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.dataset", regexp.MustCompile(`"type"\s*:\s*"esql"`)),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.group_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.metrics"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.query.language"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.treemap_config.query.query"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"panels.0.treemap_config.metrics",
					"panels.0.treemap_config.group_by",
					"panels.0.treemap_config.ignore_global_filters",
					"panels.0.treemap_config.sampling",
					"panels.0.treemap_config.label_position",
				},
			},
		},
	})
}
