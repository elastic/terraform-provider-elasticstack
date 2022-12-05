package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorSplit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorSplit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_split.test", "field", "my_field"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_split.test", "json", expectedJsonSplit),
				),
			},
		},
	})
}

const expectedJsonSplit = `{
	"split": {
		"field": "my_field",
		"separator": "\\s+",
		"preserve_trailing": false,
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorSplit = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_split" "test" {
  field     = "my_field"
  separator = "\\s+"
}
`
