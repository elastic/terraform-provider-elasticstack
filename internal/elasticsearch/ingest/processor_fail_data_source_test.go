package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorFail(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorFail,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_fail.test", "json", expectedJsonFail),
				),
			},
		},
	})
}

const expectedJsonFail = `{
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
