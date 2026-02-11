package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDashboardXYChart(t *testing.T) {
	dashboardTitle := "Test Dashboard with XY Chart " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.title", "Sample XY Chart"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.description", "Test XY chart visualization"),
					// Check axis fields
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.title.value", "Timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.title.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.left.scale", "linear"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.left.title.value", "Count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.left.title.visible", "true"),
					// Check decorations fields
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.fill_opacity", "0.3"),
					// Check fitting
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.type", "none"),
					// Check legend
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.visibility", "visible"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.position", "right"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.inside", "false"),
					// Check query
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.query.language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.query.query", ""),
					// Check layers
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.type", "line"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.dataset"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.y.#", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("axis"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.ticks", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.grid", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.label_orientation", "angled"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.x.extent"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.left.ticks", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.left.grid", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.left.label_orientation", "horizontal"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.left.extent"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.right.scale", "sqrt"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.right.title.value", "Rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.right.title.visible", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.right.ticks", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.right.grid", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.right.label_orientation", "vertical"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.axis.right.extent"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("decorations"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.line_interpolation", "smooth"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.minimum_bar_height", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.fill_opacity", "0.3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.decorations.show_value_labels", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("filters"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.filters.0.query", "log.level:error"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("fitting"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.type", "linear"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.dotted", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.fitting.end_value", "nearest"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("legend_outside"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.inside", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.position", "bottom"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.size", "large"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.truncate_after_lines", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.0", "avg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.1", "max"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("legend_inside"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.inside", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.columns", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.alignment", "top_left"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.truncate_after_lines", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.legend.statistics.0", "count"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("layers"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.sampling", "0.5"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.x"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.breakdown_by"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.y.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.0.data_layer.y.0.config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("layers_reference"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.type", "referenceLines"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.dataset"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.ignore_global_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.sampling", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.thresholds.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "panels.0.xy_chart_config.layers.1.reference_line_layer.thresholds.0.value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("layers_reference"),
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
