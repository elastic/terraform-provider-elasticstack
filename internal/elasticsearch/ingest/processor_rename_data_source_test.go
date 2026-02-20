package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorRename(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorRename,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "field", "provider"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_rename.test", "json", expectedJSONRename),
				),
			},
		},
	})
}

const expectedJSONRename = `{
	"rename": {
		"field": "provider",
		"target_field": "cloud.provider",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorRename = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_rename" "test" {
  field        = "provider"
  target_field = "cloud.provider"
}
`
