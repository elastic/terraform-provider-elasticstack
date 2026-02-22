package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardTagcloud(t *testing.T) {
	dashboardTitle := "Test Dashboard with Tagcloud " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					// Check tagcloud config
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.title", "Sample Tagcloud"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.description", "Test tagcloud visualization"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.orientation", "horizontal"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.font_size.min", "18"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.font_size.max", "72"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.ignore_global_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.sampling", "1"),
					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.query.query", ""),
					// Check JSON fields are set
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.dataset_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.metric_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.tag_by_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Filters"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" with Filters"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					// Check tagcloud config with filters
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.title", "Filtered Tagcloud"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.description", "Tagcloud with filters and custom settings"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.orientation", "vertical"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.font_size.min", "12"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.font_size.max", "100"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.sampling", "0.5"),
					// Check query with filter
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.query.query", "service.name:*"),
					// Check filters
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.tagcloud_config.filters.0.query", "log.level:error OR log.level:warning"),
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
				// Ignore JSON fields with API defaults
				ImportStateVerifyIgnore: []string{
					"panels.0.tagcloud_config.metric_json",
					"panels.0.tagcloud_config.tag_by_json",
				},
			},
		},
	})
}
