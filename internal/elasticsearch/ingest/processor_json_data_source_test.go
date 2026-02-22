package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorJSON(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorJSON,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "field", "string_source"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_json.test", "json", expectedJSONJSON),
				),
			},
		},
	})
}

const expectedJSONJSON = `{
	"json": {
		"field": "string_source",
		"ignore_failure": false,
		"target_field": "json_target"
	}
}`

const testAccDataSourceIngestProcessorJSON = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field        = "string_source"
  target_field = "json_target"
}
`
