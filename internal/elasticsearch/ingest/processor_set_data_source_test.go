package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorSet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorSet,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "field", "count"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_set.test", "json", expectedJsonSet),
				),
			},
		},
	})
}

const expectedJsonSet = `{
	"set": {
		"field": "count",
		"ignore_empty_value": false,
		"ignore_failure": false,
		"media_type": "application/json",
		"override": true,
		"value": "1"
	}
}`

const testAccDataSourceIngestProcessorSet = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set" "test" {
  field = "count"
  value = 1
}
`
