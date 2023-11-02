package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorDateIndexName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorDateIndexName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "field", "date1"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "json", expectedJsonDateIndexName),
				),
			},
		},
	})
}

const expectedJsonDateIndexName = `{
  "date_index_name": {
		"date_rounding": "M",
		"description": "monthly date-time index naming",
		"field": "date1",
		"ignore_failure": false,
		"index_name_format": "yyyy-MM-dd",
		"index_name_prefix": "my-index-",
		"locale": "ENGLISH",
		"timezone": "UTC"
	}
}
`

const testAccDataSourceIngestProcessorDateIndexName = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date_index_name" "test" {
  description       = "monthly date-time index naming"
  field             = "date1"
  index_name_prefix = "my-index-"
  date_rounding     = "M"
}
`
