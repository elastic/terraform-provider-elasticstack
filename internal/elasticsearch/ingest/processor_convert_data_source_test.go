package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorConvert(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorConvert,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "field", "id"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_convert.test", "json", expectedJsonConvert),
				),
			},
		},
	})
}

const expectedJsonConvert = `{
	"convert": {
		"description": "converts the content of the id field to an integer",
		"field": "id",
		"type": "integer",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorConvert = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  description = "converts the content of the id field to an integer"
  field       = "id"
  type        = "integer"
}
`
