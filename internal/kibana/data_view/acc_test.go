package data_view_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))
var minFullDataviewSupport = version.Must(version.NewVersion("8.8.0"))

func TestAccResourceDataView(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("pre_8_8"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "override", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.event_time.id", "date_nanos"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.machine.ram.params.pattern", "0,0.[000] b"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map.runtime_shape_name.script_source", "emit(doc['shape_name'].value)"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_attrs.ingest_failure.custom_label", "error.ingest_failure"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_updated"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "override", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.name", indexName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_updated"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_data_view.dv",
			},
		},
	})
}

func TestAccResourceDataViewColorFieldFormat(t *testing.T) {
	indexName := "my-color-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.color_dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.id", "color"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.field_type", "string"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.range", "-Infinity:Infinity"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.regex", "Completed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.text", "#000000"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.background", "#54B399"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.1.regex", "Error"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.1.text", "#FFFFFF"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.1.background", "#BD271E"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState: true,
				ImportStateVerifyIgnore: []string{
					"override",
				},
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_data_view.color_dv",
			},
		},
	})
}
