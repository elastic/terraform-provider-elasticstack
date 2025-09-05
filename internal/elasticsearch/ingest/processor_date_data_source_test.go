package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorDate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorDate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "field", "initial_date"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_date.test", "json", expectedJsonDate),
				),
			},
		},
	})
}

const expectedJsonDate = `{
  "date": {
    "field": "initial_date",
    "formats": [
      "dd/MM/yyyy HH:mm:ss"
    ],
    "ignore_failure": false,
    "locale": "ENGLISH",
    "output_format": "yyyy-MM-dd'T'HH:mm:ss.SSSXXX",
    "target_field": "timestamp",
    "timezone": "Europe/Amsterdam"
  }
}
`

const testAccDataSourceIngestProcessorDate = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date" "test" {
  field        = "initial_date"
  target_field = "timestamp"
  formats      = ["dd/MM/yyyy HH:mm:ss"]
  timezone     = "Europe/Amsterdam"
}
`
