package data_view_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))
var minFullDataviewSupport = version.Must(version.NewVersion("8.8.0"))

func TestAccResourceDataView(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				Config:   testAccResourceDataViewPre8_8DV(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				Config:   testAccResourceDataViewBasicDV(indexName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				Config:   testAccResourceDataViewBasicDVUpdated(indexName),
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
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				Config:            testAccResourceDataViewBasicDVUpdated(indexName),
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_data_view.dv",
			},
		},
	})
}

func testAccResourceDataViewPre8_8DV(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
	data_view = {
	  title = "%s*"
	}
}`, indexName, indexName)
}

func testAccResourceDataViewBasicDV(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
	override = true
	data_view = {
		title           = "%s*"
		name            = "%s"
		time_field_name = "@timestamp"
		source_filters  = ["event_time", "machine.ram"]
		allow_no_index  = true
		namespaces      = ["default", "foo", "bar"]
		field_formats = {
			event_time = {
				id = "date_nanos"
			}
			"machine.ram" = {
				id = "number"
				params = {
					pattern = "0,0.[000] b"
				}
			}
		}
		runtime_field_map = {
			runtime_shape_name = {
				type          = "keyword"
				script_source = "emit(doc['shape_name'].value)"
			}
		}
		field_attrs = {
		  ingest_failure = { custom_label = "error.ingest_failure", count = 6 },
		}
	}
}`, indexName, indexName, indexName)
}

func testAccResourceDataViewBasicDVUpdated(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
	override = false
	data_view = {
		title           = "%s*"
		name            = "%s"
		time_field_name = "@timestamp"
		allow_no_index  = true
	}
}`, indexName, indexName, indexName)
}

func TestAccResourceDataViewColorFieldFormat(t *testing.T) {
	indexName := "my-color-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				Config:   testAccResourceDataViewColorFieldFormat(indexName),
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
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				Config:            testAccResourceDataViewColorFieldFormat(indexName),
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_data_view.color_dv",
			},
		},
	})
}

func testAccResourceDataViewColorFieldFormat(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "my_color_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "color_dv" {
	override = true
	data_view = {
		title           = "%s*"
		name            = "%s"
		time_field_name = "@timestamp"
		allow_no_index  = true
		field_formats = {
			status = {
				id = "color"
				params = {
					field_type = "string"
					colors = [
						{
							range      = "-Infinity:Infinity"
							regex      = "Completed"
							text       = "#000000"
							background = "#54B399"
						},
						{
							range      = "-Infinity:Infinity"
							regex      = "Error"
							text       = "#FFFFFF"
							background = "#BD271E"
						}
					]
				}
			}
		}
	}
}`, indexName, indexName, indexName)
}
