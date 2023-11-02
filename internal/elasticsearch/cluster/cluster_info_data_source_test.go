package cluster_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceClusterInfo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityUser,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "tagline", "You Know, for Search"),
				),
			},
		},
	})
}

const testAccDataSourceSecurityUser = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_info" "test" {
}
`
