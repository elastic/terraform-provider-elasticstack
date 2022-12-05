package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorForeach(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorForeach,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "json", expectedJsonForeach),
				),
			},
		},
	})
}

const expectedJsonForeach = `{
  "foreach": {
		"field": "values",
		"ignore_failure": false,
		"ignore_missing": false,
		"processor": {
			"convert": {
				"field": "_ingest._value",
				"ignore_failure": false,
				"ignore_missing": false,
				"type": "integer"
			}
		}
	}
}
`

const testAccDataSourceIngestProcessorForeach = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  field = "_ingest._value"
  type  = "integer"
}

data "elasticstack_elasticsearch_ingest_processor_foreach" "test" {
  field     = "values"
  processor = data.elasticstack_elasticsearch_ingest_processor_convert.test.json
}
`
