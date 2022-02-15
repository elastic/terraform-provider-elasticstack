package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorGsub(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorGsub,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "field", "field1"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "json", expectedJsonGsub),
				),
			},
		},
	})
}

const expectedJsonGsub = `{
	"gsub": {
		"field": "field1",
		"ignore_failure": false,
		"ignore_missing": false,
		"pattern": "\\.",
		"replacement": "-"
	}
}`

const testAccDataSourceIngestProcessorGsub = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_gsub" "test" {
  field       = "field1"
  pattern     = "\\."
  replacement = "-"
}
`
