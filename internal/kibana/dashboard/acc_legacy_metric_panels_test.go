package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardLegacyMetricChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with Legacy Metric " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.title", "Legacy Metric"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.description", "Legacy metric chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.sampling", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.query.query", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.filters.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.dataset"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.legacy_metric_config.metric"),
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
