package output_ds_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionOutput = version.Must(version.NewVersion("8.6.0"))

func TestAccDataSourceOutput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccDataSourceOutput,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "name", "default"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_output.test", "output_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccDataSourceOutputSpace,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "name", "default"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_output.test", "output_id"),
				),
			},
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:      testAccDataSourceOutputMissing,
				ExpectError: regexp.MustCompile("Output not found"),
			},
		},
	})
}

const testAccDataSourceOutput = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_output" "test" {
  name = "default"
}
`

const testAccDataSourceOutputSpace = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_output" "test" {
  name = "default"
  space_id = "default"
}
`

const testAccDataSourceOutputMissing = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_output" "test" {
  name = "missing"
}
`
