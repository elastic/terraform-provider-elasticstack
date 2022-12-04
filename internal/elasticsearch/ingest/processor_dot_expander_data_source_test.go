package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorDotExpander(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorDotExpander,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "field", "foo.bar"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "json", expectedJsonDotExpander),
				),
			},
		},
	})
}

const expectedJsonDotExpander = `{
  "dot_expander": {
		"field": "foo.bar",
		"ignore_failure": false,
		"override": false
	}
}
`

const testAccDataSourceIngestProcessorDotExpander = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dot_expander" "test" {
  field = "foo.bar"
}
`
