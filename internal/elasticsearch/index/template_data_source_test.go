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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.logs", "name", "logs"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.logs", "index_patterns.0", "logs-*-*"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.logs", "priority", "100"),
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
data "elasticstack_elasticsearch_index_template" "logs" {
	name = "logs"
}
`
