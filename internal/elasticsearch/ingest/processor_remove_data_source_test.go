package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorRemove(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorRemove,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_remove.test", "json", expectedJSONRemove),
				),
			},
		},
	})
}

const expectedJSONRemove = `{
	"remove": {
		"field": ["user_agent"],
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorRemove = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_remove" "test" {
  field = ["user_agent"]
}
`
