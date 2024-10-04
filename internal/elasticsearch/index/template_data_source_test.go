package index_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIndexTemplateDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexTemplateDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.ilm-history-7", "name", "ilm-history-7"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.ilm-history-7", "index_patterns.0", "ilm-history-7*"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.ilm-history-7", "priority", "2147483647"),
				),
			},
		},
	})
}

const testAccIndexTemplateDataSourceConfig = `
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}
data "elasticstack_elasticsearch_index_template" "ilm-history-7" {
	name = "ilm-history-7"
}
`
