package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorDrop(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorDrop,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_drop.test", "json", expectedJsonDrop),
				),
			},
		},
	})
}

const expectedJsonDrop = `{
  "drop": {
		"ignore_failure": false,
		"if" : "ctx.network_name == 'Guest'"
	}
}
`

const testAccDataSourceIngestProcessorDrop = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_drop" "test" {
  if = "ctx.network_name == 'Guest'"
}
`
