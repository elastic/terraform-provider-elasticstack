package kibana_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKibanaSpaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSpaces,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.myspaces", "spaces[0].name", "my_space"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.myspaces", "spaces[0].space_id", "my_space_id"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.myspaces", "spaces[0].description", "My Space"),
				),
			},
		},
	})
}

const testAccDataSourceSpaces = `
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_kibana_space" "myspace" {
	name = "my_space"
	space_id = "my_space_id"
	description = "My Space"
	disabled_features = ["dev_tools"]
}

data "elasticstack_kibana_spaces" "myspaces" {
	search = elasticstack_kibana_space.myspace.name
}
`
