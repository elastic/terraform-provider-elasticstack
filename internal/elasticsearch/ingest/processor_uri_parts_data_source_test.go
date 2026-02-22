package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorURIParts(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorURIParts,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "field", "input_field"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "json", expectedJSONURIParts),
				),
			},
		},
	})
}

const expectedJSONURIParts = `{
	"uri_parts": {
		"field": "input_field",
		"ignore_failure": false,
		"keep_original": true,
		"remove_if_successful": false,
		"target_field": "url"
	}
}`

const testAccDataSourceIngestProcessorURIParts = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uri_parts" "test" {
  field                = "input_field"
  target_field         = "url"
  keep_original        = true
  remove_if_successful = false
}
`
