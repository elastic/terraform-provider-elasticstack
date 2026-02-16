package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardHeatmap(t *testing.T) {
	dashboardTitle := "Test Dashboard with Heatmap " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.title", "Sample Heatmap"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.description", "Test heatmap visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.labels.orientation", "horizontal"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.x.labels.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.axes.y.labels.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.cells.labels.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.size", "medium"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.position", "right"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.legend.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.query.query", ""),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.dataset"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.metric"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.x_axis"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.heatmap_config.y_axis"),
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
					"panels.0.heatmap_config.metric",
				},
			},
		},
	})
}
