package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorJson(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorJson,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "field", "string_source"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_json.test", "json", expectedJsonJson),
				),
			},
		},
	})
}

const expectedJsonJson = `{
	"json": {
		"field": "string_source",
		"ignore_failure": false,
		"target_field": "json_target"
	}
}`

const testAccDataSourceIngestProcessorJson = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field        = "string_source"
  target_field = "json_target"
}
`
