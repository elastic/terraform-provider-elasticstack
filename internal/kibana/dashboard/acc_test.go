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
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_access_control"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Access Control"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle+" with Access Control"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "access_control.access_mode", "write_restricted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "access_control.owner", "elastic"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_access_control"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle + " with Access Control"),
				},
				ResourceName:      "elasticstack_kibana_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Access control is not returned by the GET Dashboard API.
				ImportStateVerifyIgnore: []string{"access_control"},
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

func TestAccResourceDashboardPanels(t *testing.T) {
	dashboardTitle := "Test Dashboard with Panel " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.embeddable_config.content", "First markdown panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.embeddable_config.title", "My Markdown Panel"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multiple_panels"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.h", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.embeddable_config.title", "My Markdown Panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.grid.h", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.embeddable_config.content", "Second markdown panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.embeddable_config.title", "My Markdown Panel"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_sections"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.title", "My Section"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "sections.0.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.type", "DASHBOARD_MARKDOWN"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.grid.h", "10"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.grid.w", "24"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.grid.x", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.grid.y", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.embeddable_config.content", "First markdown panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.embeddable_config.title", "My First Markdown Panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.embeddable_config.hide_panel_titles", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi_sections_single_panel_each"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.title", "Section One"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "sections.0.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.embeddable_config.content", "Section one - panel one"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.title", "Section Two"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "sections.1.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.panels.0.embeddable_config.content", "Section two - panel one"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("multi_sections_multi_panels_each"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.title", "Section One"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "sections.0.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.embeddable_config.content", "Section one - panel one"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.1.embeddable_config.content", "Section one - panel two"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.title", "Section Two"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "sections.1.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.panels.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.panels.0.embeddable_config.content", "Section two - panel one"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.panels.1.embeddable_config.content", "Section two - panel two"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("panels_and_sections"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "title", dashboardTitle),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.embeddable_config.content", "Top-level panel one"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.1.embeddable_config.content", "Top-level panel two"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.title", "Section One"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "sections.0.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.0.panels.0.embeddable_config.content", "Section one - panel one"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.title", "Section Two"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_dashboard.test", "sections.1.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.panels.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "sections.1.panels.0.embeddable_config.content", "Section two - panel one"),
				),
			},
		},
	})
}
