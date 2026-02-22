package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorFail(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorFail,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fail.test", "json", expectedJSONFail),
				),
			},
		},
	})
}

const expectedJSONFail = `{
  "fail": {
		"message": "The production tag is not present, found tags: {{{tags}}}",
		"ignore_failure": false,
		"if" : "ctx.tags.contains('production') != true"
	}
}
`

const testAccDataSourceIngestProcessorFail = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fail" "test" {
  if      = "ctx.tags.contains('production') != true"
  message = "The production tag is not present, found tags: {{{tags}}}"
}
`
