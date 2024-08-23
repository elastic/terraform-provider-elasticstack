package indices_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIndicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndicesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_indices", "indices.0.name", ".internal.alerts-default.alerts-default-000001"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_indices", "indices.0.number_of_shards", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_indices", "indices.0.alias.0.name", ".alerts-default.alerts-default"),
				),
			},
		},
	})
}

const testAccIndicesDataSourceConfig = `
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

data "elasticstack_elasticsearch_indices" "all_indices" {
	search = ".internal.alerts-default.alerts-default-0*"
}
`
