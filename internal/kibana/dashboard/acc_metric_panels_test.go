package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardMetricChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with Metric Chart " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					// Check metric chart config
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.title", "Sample Metric Chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.description", "Test metric chart visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.sampling", "1"),
					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.query.query", ""),
					// Check JSON fields are set
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.dataset"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.metrics.0.config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_breakdown"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Filters"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" with Filters"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					// Check metric chart config with filters
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.title", "Sample Metric Chart with Filters"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.description", "Test metric chart with filters visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.sampling", "1"),
					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.query.query", "status:active"),
					// Check JSON fields are set
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.dataset"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.metrics.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.metrics.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.breakdown_by"),
					// Check filters
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.filters.0.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.filters.0.query", "event.category:web"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_secondary_metric"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Secondary Metric"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" with Secondary Metric"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					// Check metric chart config with secondary metric
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.title", "Sample Metric Chart with Secondary Metric"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.description", "Test metric chart with secondary metric"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.dataset"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.metrics.#", "2"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.metrics.0.config"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.metric_chart_config.metrics.1.config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_breakdown"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Filters"),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore JSON fields that may have API/defaults differences
				ImportStateVerifyIgnore: []string{
					"panels.0.metric_chart_config.dataset",
					"panels.0.metric_chart_config.metrics.0.config",
					"panels.0.metric_chart_config.breakdown_by",
				},
			},
		},
	})
}
