package dashboard_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Dashboard API is in technical preview and available from 9.3.x onwards
var minDashboardAPISupport = version.Must(version.NewVersion("9.3.0-SNAPSHOT"))

func TestAccResourceEmptyDashboard(t *testing.T) {
	dashboardTitle := "Test Dashboard " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "dashboard_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "description", "Test dashboard description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "time_from", "now-15m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "time_to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "refresh_interval_pause", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "refresh_interval_value", "90000"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "query_language", "kuery"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "query_text", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " Updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "description", "Updated dashboard description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "time_from", "now-30m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "time_to", "now"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "refresh_interval_pause", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "refresh_interval_value", "30000"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_options"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Options"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" with Options"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "time_from", "2024-01-01T00:00:00.000Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "time_to", "2024-01-01T01:00:00.000Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "time_range_mode", "absolute"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "options.hide_panel_titles", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "options.use_margins", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "options.sync_colors", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "options.sync_tooltips", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_options"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Options"),
				},
				ResourceName:            "elasticstack_kibana_dashboard.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"time_range_mode"},
			},
		},
	})
}

func TestAccResourceDashboardWithControlGroupInput(t *testing.T) {
	dashboardTitle := "Test Dashboard CGI " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_control_group_input"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "dashboard_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "description", "Test dashboard with control group input"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.auto_apply_selections", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.chaining_system", "HIERARCHICAL"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.label_position", "oneLine"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_filters", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_query", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_timerange", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_validations", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.type", "optionsListControl"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.order", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.width", "medium"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.grow", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.control_config"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.1.type", "rangeSliderControl"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.1.order", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.1.width", "large"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.1.grow", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "control_group_input.controls.1.control_config"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("control_group_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " Updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.auto_apply_selections", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.chaining_system", "NONE"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.label_position", "twoLine"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_filters", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_query", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_timerange", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.ignore_parent_settings.ignore_validations", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.type", "optionsListControl"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.order", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.width", "small"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "control_group_input.controls.0.grow", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("control_group_updated"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " Updated"),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDashboardInSpace(t *testing.T) {
	spaceName := "test-space-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	dashboardTitle := "Test Dashboard in Space " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("in_space"),
				ConfigVariables: config.Variables{
					"space_name":      config.StringVariable(spaceName),
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test_space", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test_space", "space_id", spaceName),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test_space", "title", dashboardTitle),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("in_space"),
				ConfigVariables: config.Variables{
					"space_name":      config.StringVariable(spaceName),
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test_space",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
