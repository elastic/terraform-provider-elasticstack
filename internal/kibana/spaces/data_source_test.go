package spaces_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSpacesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccSpacesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.name", "Default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.description", "This is your default space!"),
				),
			},
		},
	})
}

const testAccSpacesDataSourceConfig = `
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

data "elasticstack_kibana_spaces" "all_spaces" {

}
`
