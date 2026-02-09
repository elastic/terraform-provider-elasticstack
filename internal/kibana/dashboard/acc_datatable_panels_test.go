package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardDatatableChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with Datatable " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.no_esql.title", "Sample Datatable"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.no_esql.description", "Test datatable visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.no_esql.density.mode", "compact"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.no_esql.metrics.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.no_esql.paging", "10"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.title", "Sample ESQL Datatable"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.description", "Test ES|QL datatable visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.density.mode", "expanded"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.metrics.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.rows.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.split_metrics_by.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.paging", "20"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.dataset"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.metrics.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.rows.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.split_metrics_by.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.datatable_config.esql.sort_by"),
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
			},
		},
	})
}
